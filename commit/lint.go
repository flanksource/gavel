package commit

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/flanksource/clicky/api"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/verify"
)

// ErrLintFindings is returned when lint runs against the staged file set and
// reports at least one violation. The commit is aborted; callers (CLI) can
// surface the violations and offer interactive triage.
var ErrLintFindings = errors.New("commit blocked: lint reported violations")

// LintRunner abstracts the one bit the commit package needs from
// cmd/gavel/lint.go's executeLinters: run a chosen set of linters on a chosen
// set of files at a working directory. We keep it as a function variable so
// the CLI wires its full implementation in and tests can swap in a fake.
type LintRunner func(ctx context.Context, workDir string, linterNames, files []string) ([]*linters.LinterResult, error)

// SetLintRunner is wired from main during init() to break the dependency
// cycle (cmd/gavel imports commit, but commit can't import cmd/gavel).
func SetLintRunner(r LintRunner) { lintRunnerImpl = r }

var lintRunnerImpl LintRunner

// LintGates is the resolved on/off state for the two independent lint switches
// after merging CLI flags with .gavel.yaml commit.lint.*. Both default to the
// "secrets-only on, full-lint off" posture decided in design.
type LintGates struct {
	// FullLint runs every detected non-secrets linter.
	FullLint bool
	// Secrets runs the betterleaks/secrets linter.
	Secrets bool
}

// resolveLintGates merges three sources of truth in priority order:
//  1. explicit CLI flag (rawLint, rawSecrets) — non-nil wins,
//  2. .gavel.yaml commit.lint.{enabled,secrets} — non-nil wins,
//  3. defaults: FullLint=false, Secrets=true.
//
// rawLint / rawSecrets carry the *string form of the flag so an unset flag is
// distinguishable from an explicit "false" — clicky binds bool flags as a
// plain bool, so callers translate "user passed --lint-secrets=false" into
// rawSecrets="false" and "user did not pass the flag" into rawSecrets="".
func resolveLintGates(rawLint, rawSecrets string, cfg verify.CommitLintConfig) (LintGates, error) {
	gates := LintGates{FullLint: false, Secrets: true}

	if cfg.Enabled != nil {
		gates.FullLint = *cfg.Enabled
	}
	if cfg.Secrets != nil {
		gates.Secrets = *cfg.Secrets
	}

	if v, set, err := parseTriBool(rawLint, "--lint"); err != nil {
		return LintGates{}, err
	} else if set {
		gates.FullLint = v
	}
	if v, set, err := parseTriBool(rawSecrets, "--lint-secrets"); err != nil {
		return LintGates{}, err
	} else if set {
		gates.Secrets = v
	}

	return gates, nil
}

// parseTriBool returns (value, set, err). An empty string means the flag was
// not provided; "true"/"false" (case-insensitive) and "1"/"0" produce the
// matching bool. Any other value is a hard error so typos fail loudly.
func parseTriBool(raw, name string) (bool, bool, error) {
	s := strings.ToLower(strings.TrimSpace(raw))
	if s == "" {
		return false, false, nil
	}
	switch s {
	case "true", "1", "yes", "on":
		return true, true, nil
	case "false", "0", "no", "off":
		return false, true, nil
	default:
		return false, false, fmt.Errorf("invalid %s value: %q (expected true|false)", name, raw)
	}
}

// applyLintGate runs the configured linters against `files` (the staged file
// set for the in-flight commit). When violations are found it returns
// ErrLintFindings wrapping a result the caller can render and triage.
//
// No-op when both gates are off.
func applyLintGate(ctx context.Context, workDir string, files []string, gates LintGates) (*LintGateResult, error) {
	if !gates.FullLint && !gates.Secrets {
		return nil, nil
	}
	if lintRunnerImpl == nil {
		logger.V(2).Infof("commit lint gate skipped: no LintRunner registered (test binary or unwired build)")
		return nil, nil
	}
	if len(files) == 0 {
		return nil, nil
	}

	var requested []string
	switch {
	case gates.FullLint && gates.Secrets:
		// Pass nil = run every detected linter (lint command default).
	case gates.FullLint:
		// Run every detected linter EXCEPT betterleaks. The lint registry
		// doesn't expose negative selection, so this is implemented as a
		// post-filter inside the CLI runner. We hand the gate state along by
		// passing the special sentinel "!betterleaks".
		requested = []string{"!betterleaks"}
	case gates.Secrets:
		requested = []string{"betterleaks"}
	}

	results, err := lintRunnerImpl(ctx, workDir, requested, files)
	if err != nil {
		return nil, fmt.Errorf("run commit lint: %w", err)
	}

	violationCount := 0
	for _, r := range results {
		if r == nil || r.Skipped {
			continue
		}
		violationCount += len(r.Violations)
	}
	out := &LintGateResult{Results: results, Violations: violationCount, Gates: gates}
	if violationCount == 0 {
		return out, nil
	}
	return out, ErrLintFindings
}

// LintGateResult is the public payload returned to the CLI when the
// pre-commit lint pass produces findings (or runs cleanly). The CLI uses
// Results to render output and feed the existing triage flow.
type LintGateResult struct {
	Results    []*linters.LinterResult
	Violations int
	Gates      LintGates
}

// Pretty renders a one-line summary, used by the CLI in addition to the
// per-violation rendering the lint command already provides.
func (r *LintGateResult) Pretty() api.Text {
	if r == nil {
		return api.Text{}
	}
	logger.V(2).Infof("LintGateResult: violations=%d gates=%+v", r.Violations, r.Gates)
	return api.Text{}
}
