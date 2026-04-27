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
	"github.com/flanksource/gavel/internal/prompting"
	"github.com/flanksource/gavel/models"

	"github.com/samber/lo"
)

type AmendCommitsOptions struct {
	HistoryOptions `json:",inline"`
	Threshold      float64 `flag:"threshold" help:"Quality score threshold - commits below this will be reviewed" default:"7.0"`
	Ref            string  `flag:"ref" help:"Base ref for commit range (e.g., origin/main, main)" default:""`
	DryRun         bool    `flag:"dry-run" help:"Show what would be changed without rebasing"`
}

type CommitReview struct {
	Commit       models.CommitAnalysis
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

	analyses = lo.Filter(analyses, func(a models.CommitAnalysis, _ int) bool {
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

func getCommitRange(baseRef string, options HistoryOptions) ([]models.Commit, error) {
	// Get commits in range baseRef..HEAD
	cmd := exec.Command("git", "rev-list", fmt.Sprintf("%s..HEAD", baseRef))
	cmd.Dir = options.Path

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	commitHashes := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(commitHashes) == 1 && commitHashes[0] == "" {
		return []models.Commit{}, nil
	}

	// Get full commit info
	return GetCommitHistory(options)
}

func displayCommitReview(index, total int, commit models.CommitAnalysis, score int, suggestedMsg string) {
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

func formatCommitMessage(analysis models.CommitAnalysis) string {
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
	prompting.Prepare()
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

func executeRebase(_ string, _ []CommitReview) error {
	// FIXME: rebase not yet implemented
	return fmt.Errorf("rebase not yet implemented")
}
