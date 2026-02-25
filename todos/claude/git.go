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

func gitCheckoutBranch(workDir, branch string) (restore func(), err error) {
	noop := func() {}

	currentCmd := exec.Command("git", "branch", "--show-current")
	currentCmd.Dir = workDir
	currentOut, err := currentCmd.Output()
	if err != nil {
		return noop, fmt.Errorf("failed to get current branch: %w", err)
	}
	currentBranch := strings.TrimSpace(string(currentOut))

	if currentBranch == branch {
		return noop, nil
	}

	// Stash any dirty changes before switching
	stashRestore, err := gitStash(workDir, false)
	if err != nil {
		return noop, fmt.Errorf("failed to stash before branch switch: %w", err)
	}

	logger.Infof("Switching from %s to %s", currentBranch, branch)
	checkoutCmd := exec.Command("git", "checkout", branch)
	checkoutCmd.Dir = workDir
	if out, err := checkoutCmd.CombinedOutput(); err != nil {
		stashRestore()
		return noop, fmt.Errorf("git checkout %s failed: %w\n%s", branch, err, out)
	}

	return func() {
		restoreCmd := exec.Command("git", "checkout", currentBranch)
		restoreCmd.Dir = workDir
		if out, err := restoreCmd.CombinedOutput(); err != nil {
			logger.Warnf("git checkout %s failed: %v\n%s", currentBranch, err, out)
		} else {
			logger.Infof("Restored branch %s", currentBranch)
		}
		stashRestore()
	}, nil
}

func gitStash(workDir string, dirty bool) (restore func(), err error) {
	noop := func() {}

	if dirty {
		return noop, nil
	}

	statusCmd := exec.Command("git", "status", "--porcelain", "--", ".", ":!.todos")
	statusCmd.Dir = workDir
	statusOut, _ := statusCmd.Output()

	untrackedCmd := exec.Command("git", "ls-files", "--others", "--exclude-standard", "--", ".", ":!.todos")
	untrackedCmd.Dir = workDir
	untrackedOut, _ := untrackedCmd.Output()

	var files []string
	for _, line := range strings.Split(strings.TrimSpace(string(statusOut)), "\n") {
		if f := strings.TrimSpace(line); f != "" {
			files = append(files, strings.TrimSpace(f[2:]))
		}
	}
	for _, line := range strings.Split(strings.TrimSpace(string(untrackedOut)), "\n") {
		if f := strings.TrimSpace(line); f != "" {
			files = append(files, f)
		}
	}

	if len(files) == 0 {
		return noop, nil
	}

	logger.Infof("Stashing %d files:", len(files))
	for _, f := range files {
		logger.Infof("  %s", f)
	}

	stashCmd := exec.Command("git", "stash", "push", "-m", "gavel-todos-run", "--include-untracked", "--", ".", ":!.todos")
	stashCmd.Dir = workDir
	if out, err := stashCmd.CombinedOutput(); err != nil {
		return noop, fmt.Errorf("git stash failed: %w\n%s", err, out)
	}

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

func gitCommitGroupChanges(ctx context.Context, workDir string, todos []*types.TODO) (string, error) {
	diffCmd := exec.Command("git", "diff", "--quiet")
	diffCmd.Dir = workDir
	hasDiff := diffCmd.Run() != nil

	untrackedCmd := exec.Command("git", "ls-files", "--others", "--exclude-standard")
	untrackedCmd.Dir = workDir
	untrackedOut, _ := untrackedCmd.Output()
	hasUntracked := len(strings.TrimSpace(string(untrackedOut))) > 0

	if !hasDiff && !hasUntracked {
		logger.Infof("No changes to commit")
		return "", nil
	}

	addCmd := exec.Command("git", "add", "-A")
	addCmd.Dir = workDir
	if out, err := addCmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("git add failed: %w\n%s", err, out)
	}

	var names []string
	for _, t := range todos {
		names = append(names, t.Filename())
	}
	msg := fmt.Sprintf("fix: implement TODOs %s", strings.Join(names, ", "))

	commitCmd := exec.Command("git", "commit", "-m", msg)
	commitCmd.Dir = workDir
	if out, err := commitCmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("git commit failed: %w\n%s", err, out)
	}
	logger.Infof("Committed changes: %s", msg)
	return gitRevParseHEAD(workDir)
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
