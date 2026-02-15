package claude

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/flanksource/clicky/ai"
	"github.com/flanksource/commons-db/llm"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/git"
	. "github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/todos/types"
)

func gitStash(workDir string) (restore func(), err error) {
	noop := func() {}

	// Check if working tree is dirty
	diffCmd := exec.Command("git", "diff", "--quiet")
	diffCmd.Dir = workDir
	diffErr := diffCmd.Run()

	cachedCmd := exec.Command("git", "diff", "--cached", "--quiet")
	cachedCmd.Dir = workDir
	cachedErr := cachedCmd.Run()

	untrackedCmd := exec.Command("git", "ls-files", "--others", "--exclude-standard")
	untrackedCmd.Dir = workDir
	untrackedOut, _ := untrackedCmd.Output()

	if diffErr == nil && cachedErr == nil && len(strings.TrimSpace(string(untrackedOut))) == 0 {
		return noop, nil
	}

	stashCmd := exec.Command("git", "stash", "push", "-m", "gavel-todos-run", "--include-untracked")
	stashCmd.Dir = workDir
	if out, err := stashCmd.CombinedOutput(); err != nil {
		return noop, fmt.Errorf("git stash failed: %w\n%s", err, out)
	}

	logger.Infof("Stashed dirty working tree")
	return func() {
		popCmd := exec.Command("git", "stash", "pop")
		popCmd.Dir = workDir
		if out, err := popCmd.CombinedOutput(); err != nil {
			logger.Warnf("git stash pop failed: %v\n%s", err, out)
		} else {
			logger.Infof("Restored stashed changes")
		}
	}, nil
}

func gitCommitChanges(ctx context.Context, workDir string, todo *types.TODO) (string, error) {
	// Check for unstaged changes
	diffCmd := exec.Command("git", "diff", "--quiet")
	diffCmd.Dir = workDir
	hasDiff := diffCmd.Run() != nil

	// Check for untracked files
	untrackedCmd := exec.Command("git", "ls-files", "--others", "--exclude-standard")
	untrackedCmd.Dir = workDir
	untrackedOut, _ := untrackedCmd.Output()
	hasUntracked := len(strings.TrimSpace(string(untrackedOut))) > 0

	if !hasDiff && !hasUntracked {
		logger.Infof("No changes to commit")
		return "", nil
	}

	// Stage all changes
	addCmd := exec.Command("git", "add", "-A")
	addCmd.Dir = workDir
	if out, err := addCmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("git add failed: %w\n%s", err, out)
	}

	// Fixup commit if working_commit is set
	if todo.WorkingCommit != "" {
		fixupCmd := exec.Command("git", "commit", "--fixup="+todo.WorkingCommit)
		fixupCmd.Dir = workDir
		if out, err := fixupCmd.CombinedOutput(); err != nil {
			return "", fmt.Errorf("git commit --fixup failed: %w\n%s", err, out)
		}
		logger.Infof("Created fixup commit for %s", todo.WorkingCommit)
		return gitRevParseHEAD(workDir)
	}

	msg, err := generateCommitMessage(ctx, workDir, todo)
	if err != nil {
		logger.Warnf("API commit message generation failed: %v, trying Claude CLI", err)
		msg, err = generateCommitMessageCLI(workDir)
		if err != nil {
			logger.Warnf("Claude CLI commit message failed: %v, using fallback", err)
			msg = fmt.Sprintf("fix: implement TODO %s", todo.Filename())
		}
	}

	commitCmd := exec.Command("git", "commit", "-m", msg)
	commitCmd.Dir = workDir
	if out, err := commitCmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("git commit failed: %w\n%s", err, out)
	}
	logger.Infof("Committed changes: %s", strings.Split(msg, "\n")[0])
	return gitRevParseHEAD(workDir)
}

func gitRevParseHEAD(workDir string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--short", "HEAD")
	cmd.Dir = workDir
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git rev-parse HEAD failed: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func generateCommitMessageCLI(workDir string) (string, error) {
	claudePath, err := exec.LookPath("claude")
	if err != nil {
		return "", fmt.Errorf("claude CLI not found: %w", err)
	}

	diffCmd := exec.Command("git", "diff", "--cached", "--stat")
	diffCmd.Dir = workDir
	stat, _ := diffCmd.Output()

	diffFullCmd := exec.Command("git", "diff", "--cached")
	diffFullCmd.Dir = workDir
	diff, err := diffFullCmd.Output()
	if err != nil {
		return "", fmt.Errorf("git diff --cached failed: %w", err)
	}

	// Truncate diff to avoid exceeding context
	diffStr := string(diff)
	if len(diffStr) > 8000 {
		diffStr = diffStr[:8000] + "\n... (truncated)"
	}

	prompt := fmt.Sprintf(`Generate a conventional commit message for these staged changes.
Output ONLY the commit message, nothing else.

Files changed:
%s

Diff:
%s`, string(stat), diffStr)

	cmd := exec.Command(claudePath, "-p", prompt)
	cmd.Dir = workDir
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("claude CLI failed: %w", err)
	}

	msg := strings.TrimSpace(string(out))
	if msg == "" {
		return "", fmt.Errorf("claude CLI returned empty message")
	}
	return msg, nil
}

func generateCommitMessage(ctx context.Context, workDir string, todo *types.TODO) (string, error) {
	diffCmd := exec.Command("git", "diff", "--cached")
	diffCmd.Dir = workDir
	diffOut, err := diffCmd.Output()
	if err != nil {
		return "", fmt.Errorf("git diff --cached failed: %w", err)
	}

	commit := CommitAnalysis{
		Commit: Commit{
			Subject: fmt.Sprintf("implement TODO %s", todo.Filename()),
			Patch:   string(diffOut),
		},
	}

	if changes, parseErr := git.ParsePatch(string(diffOut)); parseErr == nil {
		commit.Changes = Changes(changes)
	}

	agent, err := llm.NewLLMAgent(ai.DefaultConfig())
	if err != nil {
		return "", fmt.Errorf("failed to create AI agent: %w", err)
	}

	analyzed, err := git.AnalyzeWithAI(ctx, commit, agent, git.AnalyzeOptions{})
	if err != nil {
		return "", fmt.Errorf("AI analysis failed: %w", err)
	}

	return formatCommitMsg(analyzed), nil
}

func formatCommitMsg(analysis CommitAnalysis) string {
	firstLine := ""
	if analysis.CommitType != "" {
		firstLine = string(analysis.CommitType)
	}
	if analysis.Scope != "" {
		firstLine += fmt.Sprintf("(%s)", analysis.Scope)
	}
	if firstLine != "" {
		firstLine += ": "
	}
	firstLine += analysis.Subject

	if analysis.Body != "" {
		return firstLine + "\n\n" + analysis.Body
	}
	return firstLine
}
