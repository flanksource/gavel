package verify

import (
	"fmt"
	"strings"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/clicky/api/icons"
	"github.com/flanksource/commons/logger"
)

type AutoFixOptions struct {
	FixModel       string
	MaxTurns       int
	ScoreThreshold int
	PatchOnly      bool
}

type TrackedFinding struct {
	ID             string // check ID or "rating:<dim>" or "completeness"
	Evidence       string // summary of evidence
	DiscoveredTurn int
	FixedTurn      int // 0 = unfixed
}

type TurnResult struct {
	Turn   int
	Score  int
	Result *VerifyResult
}

type FixLoopResult struct {
	Turns    []TurnResult
	Findings []TrackedFinding
	Passed   bool
	Reason   string
}

func RunAutoFix(verifyOpts RunOptions, fixOpts AutoFixOptions) (*FixLoopResult, error) {
	if fixOpts.MaxTurns <= 0 {
		fixOpts.MaxTurns = 3
	}
	if fixOpts.ScoreThreshold <= 0 {
		fixOpts.ScoreThreshold = 80
	}

	fixModel := fixOpts.FixModel
	if fixModel == "" {
		fixModel = verifyOpts.Config.Model
	}

	loop := &FixLoopResult{}

	// Initial verify
	result, err := RunVerify(verifyOpts)
	if err != nil {
		return nil, fmt.Errorf("initial verify failed: %w", err)
	}

	loop.Turns = append(loop.Turns, TurnResult{Turn: 1, Score: result.Score, Result: result})
	loop.Findings = extractFindings(result, 1)

	if checkConverged(result, fixOpts) {
		loop.Passed = true
		loop.Reason = fmt.Sprintf("score %d >= %d on initial verify", result.Score, fixOpts.ScoreThreshold)
		return loop, nil
	}

	for turn := 1; turn <= fixOpts.MaxTurns; turn++ {
		logger.Infof("Auto-fix turn %d/%d (score: %d)", turn, fixOpts.MaxTurns, result.Score)

		prompt := buildFixPrompt(result, verifyOpts, loop, turn)

		fixTool, fixModelResolved := ResolveCLI(fixModel)
		err := executeFix(fixTool, fixModelResolved, prompt, verifyOpts.RepoPath, fixOpts.PatchOnly)
		if err != nil {
			logger.Warnf("Fix turn %d failed: %v", turn, err)
			continue
		}

		// Re-verify
		result, err = RunVerify(verifyOpts)
		if err != nil {
			logger.Warnf("Re-verify after turn %d failed: %v", turn, err)
			continue
		}

		loop.Turns = append(loop.Turns, TurnResult{Turn: turn + 1, Score: result.Score, Result: result})
		updateFindings(loop, result, turn+1)

		if checkConverged(result, fixOpts) {
			loop.Passed = true
			loop.Reason = fmt.Sprintf("score %d >= %d after turn %d", result.Score, fixOpts.ScoreThreshold, turn)
			return loop, nil
		}
	}

	lastScore := loop.Turns[len(loop.Turns)-1].Score
	loop.Reason = fmt.Sprintf("score %d < %d after %d turns", lastScore, fixOpts.ScoreThreshold, fixOpts.MaxTurns)
	return loop, nil
}

func checkConverged(r *VerifyResult, opts AutoFixOptions) bool {
	if r.Score >= opts.ScoreThreshold {
		return true
	}
	for _, cr := range r.Checks {
		if !cr.Pass {
			return false
		}
	}
	return true // all checks pass
}

func extractFindings(r *VerifyResult, turn int) []TrackedFinding {
	var findings []TrackedFinding
	for id, cr := range r.Checks {
		if cr.Pass {
			continue
		}
		evidence := formatEvidence(cr.Evidence)
		findings = append(findings, TrackedFinding{ID: id, Evidence: evidence, DiscoveredTurn: turn})
	}
	for dim, rr := range r.Ratings {
		if rr.Score >= 80 {
			continue
		}
		evidence := fmt.Sprintf("score %d/100", rr.Score)
		if len(rr.Findings) > 0 {
			evidence += ": " + formatEvidence(rr.Findings)
		}
		findings = append(findings, TrackedFinding{
			ID: "rating:" + dim, Evidence: evidence, DiscoveredTurn: turn,
		})
	}
	if !r.Completeness.Pass {
		evidence := r.Completeness.Summary
		if evidence == "" && len(r.Completeness.Evidence) > 0 {
			evidence = formatEvidence(r.Completeness.Evidence)
		}
		findings = append(findings, TrackedFinding{
			ID: "completeness", Evidence: evidence, DiscoveredTurn: turn,
		})
	}
	return findings
}

func updateFindings(loop *FixLoopResult, r *VerifyResult, turn int) {
	currentFailing := failingIDs(r)

	// Mark fixed findings
	for i := range loop.Findings {
		if loop.Findings[i].FixedTurn > 0 {
			continue
		}
		if !currentFailing[loop.Findings[i].ID] {
			loop.Findings[i].FixedTurn = turn
		}
	}

	// Track new findings
	existing := map[string]bool{}
	for _, f := range loop.Findings {
		if f.FixedTurn == 0 {
			existing[f.ID] = true
		}
	}
	for _, nf := range extractFindings(r, turn) {
		if !existing[nf.ID] {
			loop.Findings = append(loop.Findings, nf)
		}
	}
}

func failingIDs(r *VerifyResult) map[string]bool {
	ids := map[string]bool{}
	for id, cr := range r.Checks {
		if !cr.Pass {
			ids[id] = true
		}
	}
	for dim, rr := range r.Ratings {
		if rr.Score < 80 {
			ids["rating:"+dim] = true
		}
	}
	if !r.Completeness.Pass {
		ids["completeness"] = true
	}
	return ids
}

func formatEvidence(evs []Evidence) string {
	var parts []string
	for _, e := range evs {
		if e.Line > 0 {
			parts = append(parts, fmt.Sprintf("%s:%d — %s", e.File, e.Line, e.Message))
		} else if e.File != "" {
			parts = append(parts, fmt.Sprintf("%s — %s", e.File, e.Message))
		} else {
			parts = append(parts, e.Message)
		}
	}
	return strings.Join(parts, "; ")
}

func buildFixPrompt(r *VerifyResult, verifyOpts RunOptions, loop *FixLoopResult, turn int) string {
	var b strings.Builder

	b.WriteString("You are fixing code review findings. Apply the minimal changes needed to resolve each issue.\n")
	b.WriteString("After making changes, run tests to ensure nothing breaks.\n\n")

	b.WriteString("## Findings to fix\n\n")

	for id, cr := range r.Checks {
		if cr.Pass {
			continue
		}
		b.WriteString(fmt.Sprintf("### Check: %s (FAILED)\n", id))
		for _, e := range cr.Evidence {
			if e.Line > 0 {
				b.WriteString(fmt.Sprintf("- %s:%d — %s\n", e.File, e.Line, e.Message))
			} else if e.File != "" {
				b.WriteString(fmt.Sprintf("- %s — %s\n", e.File, e.Message))
			} else {
				b.WriteString(fmt.Sprintf("- %s\n", e.Message))
			}
		}
		b.WriteString("\n")
	}

	for dim, rr := range r.Ratings {
		if rr.Score >= 80 {
			continue
		}
		b.WriteString(fmt.Sprintf("### Rating: %s (score: %d/100)\n", dim, rr.Score))
		for _, f := range rr.Findings {
			if f.Line > 0 {
				b.WriteString(fmt.Sprintf("- %s:%d — %s\n", f.File, f.Line, f.Message))
			} else {
				b.WriteString(fmt.Sprintf("- %s\n", f.Message))
			}
		}
		b.WriteString("\n")
	}

	if !r.Completeness.Pass {
		b.WriteString("### Completeness (FAILED)\n")
		if r.Completeness.Summary != "" {
			b.WriteString(fmt.Sprintf("- %s\n", r.Completeness.Summary))
		}
		for _, e := range r.Completeness.Evidence {
			b.WriteString(fmt.Sprintf("- %s — %s\n", e.File, e.Message))
		}
		b.WriteString("\n")
	}

	if turn > 1 && len(loop.Turns) > 1 {
		b.WriteString("## Previous attempts\n\n")
		for _, tr := range loop.Turns {
			b.WriteString(fmt.Sprintf("- Turn %d: score %d/100\n", tr.Turn, tr.Score))
		}
		b.WriteString("\nThe above issues persist after previous fix attempts. Try a different approach.\n\n")

		var fixed []string
		for _, f := range loop.Findings {
			if f.FixedTurn > 0 {
				fixed = append(fixed, fmt.Sprintf("%s (fixed in turn %d)", f.ID, f.FixedTurn))
			}
		}
		if len(fixed) > 0 {
			b.WriteString("Already fixed: " + strings.Join(fixed, ", ") + "\n\n")
		}
	}

	if verifyOpts.Config.Prompt != "" {
		b.WriteString("## Additional context\n\n")
		b.WriteString(verifyOpts.Config.Prompt + "\n")
	}

	return b.String()
}

func executeFix(tool CLITool, model, prompt, workDir string, patchOnly bool) error {
	args := buildFixArgs(tool, model, prompt, patchOnly)

	logger.V(1).Infof("fix exec: %s %s", tool.Binary, strings.Join(args, " "))

	proc := clicky.Exec(tool.Binary, args...).WithCwd(workDir)
	if logger.V(2).Enabled() {
		proc = proc.Debug()
	}

	result := proc.Run().Result()
	if result.Error != nil {
		return fmt.Errorf("%s fix failed: %w\nstderr: %s", tool.Binary, result.Error, result.Stderr)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("%s fix exited with code %d\nstderr: %s", tool.Binary, result.ExitCode, result.Stderr)
	}
	return nil
}

func buildFixArgs(tool CLITool, model, prompt string, patchOnly bool) []string {
	switch tool.Binary {
	case "claude":
		if patchOnly {
			args := []string{"-p", prompt, "--output-format", "json"}
			if model != "" && model != "claude" {
				args = append(args, "--model", model)
			}
			return args
		}
		// Interactive: claude with tool-use (no -p for interactive, use --prompt)
		args := []string{"-p", prompt, "--allowedTools", "Edit,Write,Bash,Read,Glob,Grep"}
		if model != "" && model != "claude" {
			args = append(args, "--model", model)
		}
		return args

	case "codex":
		args := []string{"exec"}
		if !patchOnly {
			args = append(args, "--full-auto")
		}
		if model != "" && model != "codex" {
			args = append(args, "-m", model)
		}
		args = append(args, "--", prompt)
		return args

	case "gemini":
		args := []string{"-p", prompt}
		if model != "" && model != "gemini" {
			args = append(args, "-m", model)
		}
		return args

	default:
		return tool.BuildArgs(prompt, model, "", false)
	}
}

func (r FixLoopResult) Pretty() api.Text {
	lastTurn := r.Turns[len(r.Turns)-1]
	firstTurn := r.Turns[0]

	statusIcon := icons.Check.WithStyle("text-green-600")
	statusLabel := "PASSED"
	if !r.Passed {
		statusIcon = icons.Cross.WithStyle("text-red-600")
		statusLabel = "FAILED"
	}

	text := clicky.Text("Auto-Fix", "font-bold").
		Append(" ").Add(statusIcon).
		Append(fmt.Sprintf(" %s", statusLabel), "font-bold").
		Append(fmt.Sprintf("  %d → %d/100", firstTurn.Score, lastTurn.Score), ratingColor(lastTurn.Score)).
		Append(fmt.Sprintf("  (%d turns)", len(r.Turns)-1), "text-gray-500")

	// Score progression
	text = text.NewLine().NewLine().Append("Score progression: ", "font-bold")
	for i, tr := range r.Turns {
		if i > 0 {
			text = text.Append(" → ", "text-gray-400")
		}
		label := fmt.Sprintf("Turn %d: %d", tr.Turn, tr.Score)
		text = text.Append(label, ratingColor(tr.Score))
	}

	// Findings grouped by status
	var fixed, unfixed []TrackedFinding
	for _, f := range r.Findings {
		if f.FixedTurn > 0 {
			fixed = append(fixed, f)
		} else {
			unfixed = append(unfixed, f)
		}
	}

	if len(fixed) > 0 {
		text = text.NewLine().NewLine().Append(fmt.Sprintf("Fixed (%d)", len(fixed)), "font-bold text-green-600")
		for _, f := range fixed {
			text = text.NewLine().Append("  ", "").
				Add(icons.Check.WithStyle("text-green-600")).
				Append(fmt.Sprintf(" %s", f.ID), "").
				Append(fmt.Sprintf("  discovered turn %d, fixed turn %d", f.DiscoveredTurn, f.FixedTurn), "text-gray-500")
		}
	}

	if len(unfixed) > 0 {
		text = text.NewLine().NewLine().Append(fmt.Sprintf("Unfixed (%d)", len(unfixed)), "font-bold text-red-600")
		for _, f := range unfixed {
			text = text.NewLine().Append("  ", "").
				Add(icons.Cross.WithStyle("text-red-600")).
				Append(fmt.Sprintf(" %s", f.ID), "").
				Append(fmt.Sprintf("  discovered turn %d", f.DiscoveredTurn), "text-gray-500")
			if f.Evidence != "" {
				text = text.NewLine().Append(fmt.Sprintf("    %s", f.Evidence), "text-gray-500")
			}
		}
	}

	if r.Reason != "" {
		text = text.NewLine().NewLine().Append(r.Reason, "text-gray-500 italic")
	}

	return text
}
