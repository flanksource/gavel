package todos

import (
	"context"
	"fmt"

	captainai "github.com/flanksource/captain/pkg/ai"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/ai"
	"github.com/flanksource/gavel/git"
	"github.com/flanksource/gavel/todos/types"
	"github.com/flanksource/gavel/verify"
)

// defaultTodoVerifyModel is the agentic captain backend used when a TODO has no
// ai.model front matter and no --model override.
const defaultTodoVerifyModel = "claude-code-sonnet"

// resolveTodoAgentConfig maps a TODO's ai: front matter onto an agent config for
// the captain-agentic verify run, with the --model flag taking precedence.
func resolveTodoAgentConfig(todo *types.TODO, model, backend string) ai.AgentConfig {
	cfg := todo.AI.ToAgentConfig(defaultTodoVerifyModel)
	if backend != "" {
		cfg.Backend = backend
		if model == "" {
			cfg.Model = defaultTodoVerifyModelForBackend(backend)
		}
	}
	if model != "" {
		cfg.Model = model
	}
	return cfg
}

func defaultTodoVerifyModelForBackend(backend string) string {
	switch captainai.Backend(backend) {
	case captainai.BackendClaudeAgent, captainai.BackendClaudeCLI:
		return "claude-agent-sonnet"
	case captainai.BackendCodexCLI:
		return "gpt-5-codex"
	default:
		return defaultTodoVerifyModel
	}
}

// defaultVerifyThreshold is the score (with implemented=true) at or above which
// an issue verification promotes the TODO to verified.
const defaultVerifyThreshold = 80

// VerifyOptions configures an issue verification run.
type VerifyOptions struct {
	WorkDir   string   // git root the commits live in
	Model     string   // verify model override (empty = config/default)
	Backend   string   // captain backend override (empty = infer from model/default)
	Threshold int      // promote to verified at this score (<=0 → defaultVerifyThreshold)
	Commits   []string // known commit SHAs (post-run); empty = discover by trailer
	// Prompt, when non-empty, overrides the rendered verify prompt verbatim — the
	// dashboard's editable prompt. The output schema is unchanged.
	Prompt string
}

func (o VerifyOptions) threshold() int {
	if o.Threshold > 0 {
		return o.Threshold
	}
	return defaultVerifyThreshold
}

// ResolveIssueCommits returns the commit SHAs implementing an issue: the known
// set (e.g. from the just-made post-run commit) when provided, otherwise the
// commits carrying the issue's Gavel-Issue-Id trailer, oldest-first.
func ResolveIssueCommits(workDir, issueID string, known []string) ([]string, error) {
	if len(known) > 0 {
		return known, nil
	}
	if issueID == "" {
		return nil, nil
	}
	commits, err := git.CommitsWithTrailer(workDir, git.TrailerIssueID, issueID)
	if err != nil {
		return nil, err
	}
	shas := make([]string, 0, len(commits))
	for i := len(commits) - 1; i >= 0; i-- { // newest-first → oldest-first
		shas = append(shas, commits[i].Hash)
	}
	return shas, nil
}

// BuildIssueContext maps a TODO and its commits into the verify issue context:
// selected static checks → CheckIDs, custom criteria → Criteria, comment events
// → Comments.
func BuildIssueContext(todo *types.TODO, shas []string) *verify.IssueContext {
	ic := &verify.IssueContext{
		ID:          todo.ID,
		Title:       todo.Title,
		Description: todo.MarkdownBody,
		CommitSHAs:  shas,
	}
	if todo.LLM != nil {
		ic.SessionID = todo.LLM.SessionId
	}
	for _, c := range todo.AcceptanceCriteria {
		if c.CheckID != "" {
			ic.CheckIDs = append(ic.CheckIDs, c.CheckID)
		} else if c.Text != "" {
			ic.Criteria = append(ic.Criteria, c.Text)
		}
	}
	for _, ev := range todo.ProviderEvents {
		if ev.Kind == "CommentAdded" && ev.Body != "" {
			ic.Comments = append(ic.Comments, verify.IssueComment{Author: ev.Actor, Body: ev.Body})
		}
	}
	return ic
}

// RunIssueVerification scores a TODO's commits against its issue spec and stored
// acceptance criteria, records the verdict (status + persistent comment), and
// returns the result. A confirmed verdict (implemented and score >= threshold)
// promotes the TODO to verified; otherwise it is left open (pending). It is
// advisory: it never marks the TODO failed.
func RunIssueVerification(ctx context.Context, provider Provider, todo *types.TODO, opts VerifyOptions) (*verify.VerifyResult, error) {
	shas, err := ResolveIssueCommits(opts.WorkDir, todo.ID, opts.Commits)
	if err != nil {
		return nil, fmt.Errorf("resolve issue commits: %w", err)
	}
	if len(shas) == 0 {
		return nil, fmt.Errorf("no commits found for issue %s; nothing to verify", todo.ID)
	}

	cfg, err := verify.LoadConfig(opts.WorkDir)
	if err != nil {
		logger.Warnf("verify: failed to load config: %v", err)
	}
	if opts.Model != "" {
		cfg.Model = opts.Model
	}
	applyTodoVerifyPrompt(&cfg, opts.WorkDir)

	agentCfg := resolveTodoAgentConfig(todo, opts.Model, opts.Backend)
	result, err := verify.RunVerify(verify.RunOptions{
		Config:         cfg,
		RepoPath:       opts.WorkDir,
		AgentConfig:    &agentCfg,
		Issue:          BuildIssueContext(todo, shas),
		PromptOverride: opts.Prompt,
	})
	if err != nil {
		return nil, err
	}

	persistVerificationVerdict(ctx, provider, todo, result, opts.threshold())
	return result, nil
}

// PreviewIssueVerification renders the verify prompt a RunIssueVerification would
// send for a todo, without executing it, so the dashboard can seed an editable
// prompt. It mirrors RunIssueVerification's scope/issue/config construction.
func PreviewIssueVerification(todo *types.TODO, opts VerifyOptions) (string, error) {
	shas, err := ResolveIssueCommits(opts.WorkDir, todo.ID, opts.Commits)
	if err != nil {
		return "", fmt.Errorf("resolve issue commits: %w", err)
	}
	if len(shas) == 0 {
		return "", fmt.Errorf("no commits found for issue %s; nothing to verify", todo.ID)
	}
	cfg, err := verify.LoadConfig(opts.WorkDir)
	if err != nil {
		logger.Warnf("verify: failed to load config: %v", err)
	}
	if opts.Model != "" {
		cfg.Model = opts.Model
	}
	applyTodoVerifyPrompt(&cfg, opts.WorkDir)
	return verify.PreviewPrompt(verify.RunOptions{
		Config:   cfg,
		RepoPath: opts.WorkDir,
		Issue:    BuildIssueContext(todo, shas),
	})
}

// applyTodoVerifyPrompt overlays the todos.verifyPrompt override onto the
// verify config so TODO verification uses the todo-specific template while
// `gavel verify` keeps verify.promptTemplate.
func applyTodoVerifyPrompt(cfg *verify.VerifyConfig, workDir string) {
	gcfg, err := verify.LoadGavelConfig(workDir)
	if err != nil {
		logger.Warnf("verify: failed to load .gavel.yaml for todos.verifyPrompt: %v", err)
		return
	}
	if !gcfg.Todos.VerifyPrompt.IsZero() {
		cfg.PromptTemplate = gcfg.Todos.VerifyPrompt
	}
}

// persistVerificationVerdict saves the verdict comment and transitions the TODO:
// verified when confirmed, pending otherwise. Persistence failures are logged,
// not fatal — the verdict is still returned to the caller.
func persistVerificationVerdict(ctx context.Context, provider Provider, todo *types.TODO, result *verify.VerifyResult, threshold int) {
	if provider == nil {
		return
	}
	if err := provider.SaveVerification(ctx, todo, result); err != nil {
		logger.Warnf("verify: failed to save verification: %v", err)
	}

	confirmed := result.Implemented != nil && *result.Implemented && result.Score >= threshold
	status := types.StatusPending
	if confirmed {
		status = types.StatusVerified
	}
	todo.Status = status
	if err := provider.UpdateState(ctx, todo, StateUpdate{Status: &status}); err != nil {
		logger.Warnf("verify: failed to update status: %v", err)
	}
}
