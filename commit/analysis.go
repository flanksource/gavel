package commit

import (
	"context"
	"os"
	"strings"

	"github.com/flanksource/commons/logger"
	clickyai "github.com/flanksource/gavel/ai"
	"github.com/flanksource/gavel/git"
	"github.com/flanksource/gavel/internal/prompting"
	"github.com/flanksource/gavel/models"
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

	promptOptions, saved, err := opts.promptOptions()
	if err != nil {
		return commitAIAnalysis{}, err
	}
	analyzeOptions := git.AnalyzeOptions{
		Prompt: opts.Config.Message, PromptOptions: promptOptions, Saved: saved,
		AllowedCommitTypes: opts.Config.Types,
		MaxBodyLines:       maxBodyLinesForDiff(countDiffLines(diff)),
	}
	prepared, err := git.PrepareCommitMessage(models.CommitAnalysis{Commit: models.Commit{Patch: diff}}, analyzeOptions)
	if err != nil {
		return commitAIAnalysis{}, err
	}
	agent, err := BuildAgent(prepared.Config)
	if err != nil {
		return commitAIAnalysis{}, err
	}
	defer closeAgent(agent, &err)
	analyzeOptions.Prepared = &prepared
	return generateCommitAnalysisWithAgent(ctx, diff, agent, analyzeOptions)
}

func generateCommitAnalysisWithAgent(ctx context.Context, diff string, agent clickyai.Agent, opts git.AnalyzeOptions) (commitAIAnalysis, error) {
	analysis := models.CommitAnalysis{Commit: models.Commit{Patch: diff}}
	opts.MaxBodyLines = maxBodyLinesForDiff(countDiffLines(diff))
	// Stop the task renderer before the AI prompt takes over the terminal. This
	// lives here rather than inside git.AnalyzeWithAI because the git analyze path
	// calls that from inside a clicky batch, where waiting for global completion
	// deadlocks.
	prompting.Prepare()
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
