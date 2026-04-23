package commit

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/cmd/gavel/choose"
)

var ErrCompatibilityCancelled = errors.New("commit cancelled by compatibility checks")

type CompatibilityDecision int

const (
	CompatibilityDecisionCancel CompatibilityDecision = iota
	CompatibilityDecisionContinue
)

type CompatibilityDecider func(ctx context.Context, commit CommitResult) (CompatibilityDecision, error)

type CompatibilityParams struct {
	Commit  CommitResult
	Mode    string
	Decider CompatibilityDecider
}

type CompatibilityOutcome struct {
	Cancelled bool
}

var interactiveCompatibilityDecider = runChooseCompatibilityDecider

func RunCompatibilityCheck(ctx context.Context, p CompatibilityParams) (CompatibilityOutcome, error) {
	mode, err := normalizeCheckMode(p.Mode, "--compat")
	if err != nil {
		return CompatibilityOutcome{}, err
	}
	if mode == CheckModeSkip {
		return CompatibilityOutcome{}, nil
	}
	if !hasCompatibilityFindings(p.Commit) {
		return CompatibilityOutcome{}, nil
	}

	if mode == IgnoreCheckModePrompt && p.Decider == nil && !stdinIsTerminal() {
		logger.Warnf("compatibility check: stdin is not a terminal; escalating to --compat=fail")
		mode = IgnoreCheckModeFail
	}

	switch mode {
	case IgnoreCheckModeFail:
		return CompatibilityOutcome{}, formatCompatibilityError(p.Commit)
	case IgnoreCheckModePrompt:
	default:
		return CompatibilityOutcome{}, fmt.Errorf("unknown --compat mode: %q", mode)
	}

	decider := p.Decider
	if decider == nil {
		decider = interactiveCompatibilityDecider
	}

	decision, err := decider(ctx, p.Commit)
	if err != nil {
		return CompatibilityOutcome{}, fmt.Errorf("compatibility prompt: %w", err)
	}
	if decision == CompatibilityDecisionCancel {
		return CompatibilityOutcome{Cancelled: true}, nil
	}

	return CompatibilityOutcome{}, nil
}

func applyCompatibilityCheck(ctx context.Context, opts Options, commit CommitResult) error {
	if !shouldRunCompatibilityAnalysis(opts.CompatMode) {
		return nil
	}

	outcome, err := RunCompatibilityCheck(ctx, CompatibilityParams{
		Commit:  commit,
		Mode:    opts.CompatMode,
		Decider: nil,
	})
	if err != nil {
		return err
	}
	if outcome.Cancelled {
		return ErrCompatibilityCancelled
	}
	return nil
}

func hasCompatibilityFindings(commit CommitResult) bool {
	return len(commit.FunctionalityRemoved) > 0 || len(commit.CompatibilityIssues) > 0
}

func compatibilityTarget(commit CommitResult) string {
	if commit.Label != "" {
		return fmt.Sprintf("commit group %q", commit.Label)
	}
	return "generated commit"
}

func formatCompatibilityError(commit CommitResult) error {
	return errors.New(strings.TrimSpace(fmt.Sprintf("%s has compatibility warnings:\n%s", compatibilityTarget(commit), formatCompatibilityFindings(commit))))
}

func formatCompatibilityFindings(commit CommitResult) string {
	var lines []string
	if strings.TrimSpace(commit.Message) != "" {
		lines = append(lines, fmt.Sprintf("commit: %s", firstLine(commit.Message)))
	}
	if len(commit.FunctionalityRemoved) > 0 {
		lines = append(lines, "functionality removed:")
		for _, item := range commit.FunctionalityRemoved {
			lines = append(lines, fmt.Sprintf("  - %s", item))
		}
	}
	if len(commit.CompatibilityIssues) > 0 {
		lines = append(lines, "compatibility issues:")
		for _, item := range commit.CompatibilityIssues {
			lines = append(lines, fmt.Sprintf("  - %s", item))
		}
	}
	return strings.Join(lines, "\n")
}

func formatCompatibilityAnalysisFailure(err error) string {
	return fmt.Sprintf("AI compatibility analysis failed: %s", err)
}

func runChooseCompatibilityDecider(_ context.Context, commit CommitResult) (CompatibilityDecision, error) {
	header := clicky.Text("Compatibility warning", "text-yellow-600 font-bold").NewLine().
		Append(formatCompatibilityFindings(commit), "").
		ANSI()

	items := []string{
		"Continue commit",
		"Cancel commit",
	}
	idx, err := choose.Run(items, choose.WithHeader(header), choose.WithLimit(1))
	if err != nil {
		return CompatibilityDecisionCancel, err
	}
	if len(idx) == 0 || idx[0] == 1 {
		return CompatibilityDecisionCancel, nil
	}
	return CompatibilityDecisionContinue, nil
}

func appendCompatibilityPreview(t api.Text, functionalityRemoved, compatibilityIssues []string) api.Text {
	t = appendCompatibilitySection(t, "Functionality removed", functionalityRemoved)
	t = appendCompatibilitySection(t, "Compatibility issues", compatibilityIssues)
	return t
}

func appendCompatibilitySection(t api.Text, title string, items []string) api.Text {
	if len(items) == 0 {
		return t
	}

	t = t.Append("    ", "").
		Append(title+":", "text-yellow-600").
		NewLine()
	for _, item := range items {
		t = t.Append("      - ", "text-yellow-600").
			Append(item).
			NewLine()
	}
	return t
}
