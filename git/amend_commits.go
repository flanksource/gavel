package git

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/commons/logger"
	. "github.com/flanksource/gavel/models"

	"github.com/samber/lo"
)

type AmendCommitsOptions struct {
	HistoryOptions `json:",inline"`
	Threshold      float64 `flag:"threshold" help:"Quality score threshold - commits below this will be reviewed" default:"7.0"`
	Ref            string  `flag:"ref" help:"Base ref for commit range (e.g., origin/main, main)" default:""`
	DryRun         bool    `flag:"dry-run" help:"Show what would be changed without rebasing"`
}

type CommitReview struct {
	Commit       CommitAnalysis
	Score        int
	SuggestedMsg string
	UserDecision string // "accept", "skip", "cancel"
}

func (opt AmendCommitsOptions) Pretty() api.Text {
	t := opt.HistoryOptions.Pretty()
	t = t.Append(" threshold=", "text-muted").Append(fmt.Sprintf("%.1f", opt.Threshold), "font-mono")
	if opt.Ref != "" {
		t = t.Append(" ref=", "text-muted").Append(opt.Ref, "font-mono")
	}
	if opt.DryRun {
		t = t.Append(" dry-run=", "text-muted").Append("true", "font-mono")
	}
	return t
}

// AmendCommits analyzes commits, generates AI improvements, and interactively rewords them
func AmendCommits(ctx context.Context, options AmendCommitsOptions) error {
	logger.Infof("Starting amend-commits analysis")

	// Determine base ref
	baseRef := options.Ref
	if baseRef == "" {
		baseRef = getDefaultBaseRef()
	}

	// Validate ref exists
	if err := validateRef(baseRef); err != nil {
		return fmt.Errorf("invalid ref '%s': %w", baseRef, err)
	}

	// Get commits in range
	options.ShowPatch = true // Need patches for AI analysis
	commits, err := getCommitRange(baseRef, options.HistoryOptions)
	if err != nil {
		return fmt.Errorf("failed to get commit range: %w", err)
	}

	if len(commits) == 0 {
		clicky.Infof("No commits found in range %s..HEAD", baseRef)
		return nil
	}

	logger.Infof("Analyzing %d commits since %s", len(commits), baseRef)

	// Create analyzer context
	repoPath := options.Path
	if repoPath == "" {
		repoPath, _ = os.Getwd()
	}
	analyzerCtx, err := NewAnalyzerContext(ctx, repoPath)
	if err != nil {
		return fmt.Errorf("failed to create analyzer context: %w", err)
	}

	// Analyze commits
	analyzeOpts := AnalyzeOptions{
		HistoryOptions: options.HistoryOptions,
		AI:             true,
	}
	analyses, err := AnalyzeCommitHistory(analyzerCtx, commits, analyzeOpts)
	if err != nil {
		return fmt.Errorf("failed to analyze commits: %w", err)
	}

	analyses = lo.Filter(analyses, func(a CommitAnalysis, _ int) bool {
		return a.IsAnalyzed()
	})
	// Review each flagged commit
	reviews := make([]CommitReview, 0, len(analyses))
	for i, commit := range analyses {

		suggestedMsg := formatCommitMessage(commit)

		// Display commit for review
		displayCommitReview(i+1, len(analyses), commit, commit.GetQualityScore(), suggestedMsg)

		review := CommitReview{
			Commit:       commit,
			Score:        commit.GetQualityScore(),
			SuggestedMsg: suggestedMsg,
		}

		if options.DryRun {
			// In dry-run, don't prompt - just show
			review.UserDecision = "skip"
		} else {
			// Prompt user
			decision, err := promptUserDecision()
			if err != nil {
				return err
			}
			review.UserDecision = decision

			if decision == "cancel" {
				clicky.Warnf("User cancelled - no changes made")
				return nil
			}
		}

		reviews = append(reviews, review)
	}

	// Summary
	acceptedCount := 0
	for _, r := range reviews {
		if r.UserDecision == "accept" {
			acceptedCount++
		}
	}

	if options.DryRun {
		clicky.Infof("Dry-run complete: would amend %d commits", len(reviews))
		return nil
	}

	if acceptedCount == 0 {
		clicky.Infof("No commits accepted for amendment")
		return nil
	}

	clicky.Infof("Proceeding to rebase %d commits", acceptedCount)

	// Execute rebase
	return executeRebase(baseRef, reviews)
}

func getDefaultBaseRef() string {
	// Try origin/main first, fall back to main
	if validateRef("origin/main") == nil {
		return "origin/main"
	}
	if validateRef("main") == nil {
		return "main"
	}
	return "HEAD~10" // Last 10 commits as fallback
}

func validateRef(ref string) error {
	cmd := exec.Command("git", "rev-parse", "--verify", ref)
	return cmd.Run()
}

func getCommitRange(baseRef string, options HistoryOptions) ([]Commit, error) {
	// Get commits in range baseRef..HEAD
	cmd := exec.Command("git", "rev-list", fmt.Sprintf("%s..HEAD", baseRef))
	cmd.Dir = options.Path

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	commitHashes := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(commitHashes) == 1 && commitHashes[0] == "" {
		return []Commit{}, nil
	}

	// Get full commit info
	return GetCommitHistory(options)
}

func displayCommitReview(index, total int, commit CommitAnalysis, score int, suggestedMsg string) {
	separator := strings.Repeat("─", 60)

	fmt.Fprintf(os.Stderr, "\n%s\n", separator)
	fmt.Fprintf(os.Stderr, "Commit %d/%d: %s (Score: %d)\n", index, total, commit.Hash[:8], score)
	fmt.Fprintf(os.Stderr, "%s\n\n", separator)

	// Score reasoning
	fmt.Fprintf(os.Stderr, "📊 Flagged because:\n")
	if commit.CommitType == "" {
		fmt.Fprintf(os.Stderr, "  - Missing commit type (feat/fix/docs/etc)\n")
	}
	if commit.Scope == "" {
		fmt.Fprintf(os.Stderr, "  - Missing scope\n")
	}
	if len(commit.Subject) < 10 {
		fmt.Fprintf(os.Stderr, "  - Subject too short\n")
	}
	if commit.Body == "" {
		fmt.Fprintf(os.Stderr, "  - Missing body/description\n")
	}

	// Files changed
	fmt.Fprintf(os.Stderr, "\n📁 Files changed (%d files):\n", len(commit.Changes))
	for _, change := range commit.Changes {
		fmt.Fprintf(os.Stderr, "  %s (+%d -%d)\n", change.File, change.Adds, change.Dels)
	}

	// Current message
	fullMsg := commit.Subject
	if commit.Body != "" {
		fullMsg += "\n\n" + commit.Body
	}
	fmt.Fprintf(os.Stderr, "\n📝 Current message:\n")
	fmt.Fprintf(os.Stderr, "  %s\n", fullMsg)

	// Suggested message
	fmt.Fprintf(os.Stderr, "\n✨ Suggested message:\n")
	for _, line := range strings.Split(suggestedMsg, "\n") {
		fmt.Fprintf(os.Stderr, "  %s\n", line)
	}

	// Diff preview (condensed)
	if len(commit.Changes) > 0 {
		summary := commit.Changes.Summary()
		fmt.Fprintf(os.Stderr, "\n🔍 Changes: +%d -%d\n", summary.Adds, summary.Dels)
	}

	fmt.Fprintf(os.Stderr, "\n")
}

func formatCommitMessage(analysis CommitAnalysis) string {
	var parts []string

	// First line: type(scope): subject
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
	parts = append(parts, firstLine)

	// Body
	if analysis.Body != "" {
		parts = append(parts, "", analysis.Body)
	}

	return strings.Join(parts, "\n")
}

func promptUserDecision() (string, error) {
	fmt.Fprintf(os.Stderr, "[A]ccept | [S]kip | [C]ancel: ")

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	input = strings.ToLower(strings.TrimSpace(input))

	switch input {
	case "a", "accept":
		return "accept", nil
	case "s", "skip":
		return "skip", nil
	case "c", "cancel":
		return "cancel", nil
	default:
		fmt.Fprintf(os.Stderr, "Invalid input '%s', please enter a/s/c\n", input)
		return promptUserDecision()
	}
}

func executeRebase(baseRef string, reviews []CommitReview) error {
	if 1 == 1 {
		return fmt.Errorf("wait")
	}
	// Create a map of hash -> new message for accepted commits
	newMessages := make(map[string]string)
	for _, review := range reviews {
		if review.UserDecision == "accept" {
			newMessages[review.Commit.Hash] = review.SuggestedMsg
		}
	}

	if len(newMessages) == 0 {
		return nil
	}

	// Generate rebase todo script
	todoScript, err := generateRebaseTodo(baseRef, newMessages)
	if err != nil {
		return fmt.Errorf("failed to generate rebase todo: %w", err)
	}

	// Create temporary files for the rebase process
	todoFile, err := os.CreateTemp("", "git-rebase-todo-*")
	if err != nil {
		return fmt.Errorf("failed to create todo file: %w", err)
	}
	defer func() {
		_ = os.Remove(todoFile.Name())
	}()

	if _, err := todoFile.WriteString(todoScript); err != nil {
		return fmt.Errorf("failed to write todo file: %w", err)
	}
	if err := todoFile.Close(); err != nil {
		return fmt.Errorf("failed to close todo file: %w", err)
	}

	// Create editor script that will provide commit messages
	editorScript, err := createEditorScript(newMessages)
	if err != nil {
		return fmt.Errorf("failed to create editor script: %w", err)
	}
	defer func() {
		_ = os.Remove(editorScript)
	}()

	// Execute git rebase
	cmd := exec.Command("git", "rebase", "-i", baseRef)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("GIT_SEQUENCE_EDITOR=cp %s", todoFile.Name()),
		fmt.Sprintf("GIT_EDITOR=%s", editorScript),
	)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		// Check if rebase is in progress
		if isRebaseInProgress() {
			return fmt.Errorf("rebase failed - conflicts detected. Resolve conflicts and run 'git rebase --continue' or 'git rebase --abort'")
		}
		return fmt.Errorf("rebase failed: %w", err)
	}

	clicky.Infof("✓ Successfully rebased %d commits", len(newMessages))
	return nil
}

func generateRebaseTodo(baseRef string, newMessages map[string]string) (string, error) {
	// Get list of commits in range
	cmd := exec.Command("git", "log", "--reverse", "--format=%H %s", fmt.Sprintf("%s..HEAD", baseRef))
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	var todoLines []string
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, " ", 2)
		if len(parts) < 2 {
			continue
		}

		hash := parts[0]
		subject := parts[1]

		// Check if this commit should be reworded
		if _, shouldReword := newMessages[hash]; shouldReword {
			todoLines = append(todoLines, fmt.Sprintf("reword %s %s", hash[:8], subject))
		} else {
			todoLines = append(todoLines, fmt.Sprintf("pick %s %s", hash[:8], subject))
		}
	}

	return strings.Join(todoLines, "\n") + "\n", nil
}

func createEditorScript(newMessages map[string]string) (string, error) {
	// Create a script that will replace commit messages
	// This script is invoked by git for each "reword" command

	scriptFile, err := os.CreateTemp("", "git-editor-*")
	if err != nil {
		return "", err
	}
	defer func() {
		_ = scriptFile.Close()
	}()

	// Write the editor script
	script := `#!/bin/bash
# Git editor script for amending commit messages

COMMIT_MSG_FILE="$1"

# Get the current commit hash
CURRENT_HASH=$(git rev-parse HEAD)

# Check each hash we have a message for
`

	for hash, msg := range newMessages {
		escapedMsg := strings.ReplaceAll(msg, `"`, `\"`)
		escapedMsg = strings.ReplaceAll(escapedMsg, "\n", "\\n")
		script += fmt.Sprintf(`
if [ "$CURRENT_HASH" = "%s" ]; then
    echo -e "%s" > "$COMMIT_MSG_FILE"
    exit 0
fi
`, hash, escapedMsg)
	}

	script += `
# If we don't have a message for this commit, leave it unchanged
exit 0
`

	if _, err := scriptFile.WriteString(script); err != nil {
		return "", err
	}

	scriptPath := scriptFile.Name()

	// Make script executable
	if err := os.Chmod(scriptPath, 0755); err != nil {
		return "", err
	}

	return scriptPath, nil
}

func isRebaseInProgress() bool {
	// Check if .git/rebase-merge or .git/rebase-apply exists
	if _, err := os.Stat(".git/rebase-merge"); err == nil {
		return true
	}
	if _, err := os.Stat(".git/rebase-apply"); err == nil {
		return true
	}
	return false
}
