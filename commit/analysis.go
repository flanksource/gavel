package commit

import (
	"context"
	"fmt"
	"os"
	"strings"

	captainapi "github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/commons/logger"
	clickyai "github.com/flanksource/gavel/ai"
	"github.com/flanksource/gavel/git"
	"github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/verify"
)

func generateCommitAnalysis(ctx context.Context, opts Options, diff string) (analysis commitAIAnalysis, err error) {
	if os.Getenv(testEnvVar) == "1" {
		logger.V(1).Infof("%s=1, returning stub commit analysis", testEnvVar)
		msg := strings.TrimSpace(opts.Message)
		if msg == "" {
			msg = stubMessage
		}
		return commitAIAnalysis{Message: msg}, nil
	}
	if explicitMessage := strings.TrimSpace(opts.Message); explicitMessage != "" {
		return commitAIAnalysis{Message: explicitMessage}, nil
	}

	model, err := opts.messageModel()
	if err != nil {
		return commitAIAnalysis{}, err
	}
	agent, err := BuildAgent(opts, model)
	if err != nil {
		return commitAIAnalysis{}, err
	}
	defer closeAgent(agent, &err)
	messagePrompt, err := opts.Config.Message.TemplateSource(opts.WorkDir, "")
	if err != nil {
		return commitAIAnalysis{}, err
	}
	return generateCommitAnalysisWithAgent(ctx, diff, agent, git.AnalyzeOptions{
		MessagePrompt:      messagePrompt,
		AllowedCommitTypes: opts.Config.Types,
	})
}

func generateCommitAnalysisWithAgent(ctx context.Context, diff string, agent clickyai.Agent, opts git.AnalyzeOptions) (commitAIAnalysis, error) {
	analysis := models.CommitAnalysis{Commit: models.Commit{Patch: diff}}
	opts.MaxBodyLines = maxBodyLinesForDiff(countDiffLines(diff))
	analyzed, err := analyzeCommitMessageWithAIFunc(ctx, analysis, agent, opts)
	if err != nil {
		return commitAIAnalysis{}, err
	}
	out := models.AIAnalysisOutput{
		Type:    analyzed.CommitType,
		Scope:   analyzed.Scope,
		Subject: analyzed.Subject,
		Body:    analyzed.Body,
	}
	return commitAIAnalysis{Message: strings.TrimSpace(out.String())}, nil
}

// countDiffLines counts changed content lines in a unified diff: lines starting
// with '+' or '-', excluding the '+++'/'---' file headers.
func countDiffLines(diff string) int {
	n := 0
	for _, line := range strings.Split(diff, "\n") {
		if strings.HasPrefix(line, "+++") || strings.HasPrefix(line, "---") {
			continue
		}
		if strings.HasPrefix(line, "+") || strings.HasPrefix(line, "-") {
			n++
		}
	}
	return n
}

// maxBodyLinesForDiff scales the commit-message body cap to the diff size:
// trivial diffs get a subject only (0), larger diffs allow a longer body.
func maxBodyLinesForDiff(changedLines int) int {
	switch {
	case changedLines <= 20:
		return 0
	case changedLines <= 100:
		return 3
	case changedLines <= 300:
		return 6
	case changedLines <= 800:
		return 10
	default:
		return 15
	}
}

// defaultGroupModel is the model used for AI commit grouping when neither
// --group-model nor commit.{groupModel,model} is set. Grouping reasons over the
// whole change set, so it defaults to a more capable tier than the haiku-class
// model used for message generation.
const defaultGroupModel = "claude-sonnet-4-5"

// defaultMessageModel is the model used for commit-message generation when
// nothing else selects one. Message writing is a summarisation task, so it
// defaults to a fast haiku-class tier.
const defaultMessageModel = "claude-haiku-4-5"

// modelFor is the single model ladder every commit AI operation runs through,
// highest precedence first:
//
//  1. CLI flags (--model/--mode/--backend/--effort)
//  2. the operation's own spec (commit.message / commit.grouping / pr.content)
//  3. the base `ai:` spec
//  4. the built-in default for that operation
//
// It exists once because it used to exist three times and disagreed with itself:
// messageModel read the --model flag first while groupModel read it last, behind
// the config — so any `ai: model:` in ~/.gavel.yaml silently disabled --model for
// grouping.
//
// Each layer is expanded before it is merged: a compact selector like
// "agent:sol:high" must set its own backend during the merge, or the base spec's
// backend survives and the run silently uses a different runtime.
//
// It selects but does not resolve — ai.NewProvider resolves the winner against
// the catalog exactly once, on the way to the provider. Keeping selection and
// resolution apart is what lets this ladder be about precedence alone.
func (opts Options) modelFor(op verify.PromptSpec, builtin string) (captainapi.Model, error) {
	spec := opts.AI.Merge(op.Spec)

	over, err := opts.Flags.ToModel()
	if err != nil {
		return captainapi.Model{}, err
	}
	out := spec.Model.Merge(over)
	if strings.TrimSpace(out.Name) == "" {
		out.Name = builtin
	}
	return out, nil
}

// messageModel resolves the LLM for commit-message generation.
func (opts Options) messageModel() (captainapi.Model, error) {
	return opts.modelFor(opts.Config.Message, defaultMessageModel)
}

// groupModel resolves the LLM for AI commit grouping. Grouping reasons over the
// whole change set, so it defaults to a more capable tier than message writing,
// and --group-model overrides the shared --model for this operation alone.
func (opts Options) groupModel() (captainapi.Model, error) {
	m, err := opts.modelFor(opts.Config.Grouping, defaultGroupModel)
	if err != nil || strings.TrimSpace(opts.GroupModel) == "" {
		return m, err
	}
	over, err := (captainapi.Model{Name: opts.GroupModel}).Expand()
	if err != nil {
		return captainapi.Model{}, fmt.Errorf("invalid --group-model %q: %w", opts.GroupModel, err)
	}
	return m.Merge(over), nil
}

// PRContentModel resolves the LLM for PR title/body/branch generation. Exported
// because `gavel pr create` builds its own Options and needs the same ladder.
func (opts Options) PRContentModel() (captainapi.Model, error) {
	return opts.modelFor(opts.PR.Content, defaultMessageModel)
}

// BuildAgent constructs an LLM agent for opts using an already-resolved model.
//
// It takes an api.Model, not a name: a name alone cannot carry the backend, mode
// or effort the ladder just resolved, and re-deriving them from the string is
// what made `--model agent:gpt-5.6-luna:medium` run as plain api:gpt-5.6-luna.
// It also builds on opts.AI.Budget rather than a hardcoded default config, which
// used to discard the configured budget entirely.
func BuildAgent(opts Options, model captainapi.Model) (clickyai.Agent, error) {
	cfg := clickyai.DefaultConfig()
	cfg.Model = model
	if opts.AI.Budget != (captainapi.Budget{}) {
		cfg.Budget = opts.AI.Budget
	}
	if opts.Flags.NoCache {
		cfg.NoCache = true
	}

	agent, err := newAgentFunc(cfg)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrLLMUnavailable, err)
	}
	if agent == nil {
		return nil, fmt.Errorf("%w: agent factory returned nil", ErrLLMUnavailable)
	}
	return agent, nil
}
