package main

import (
	"os"
	"strconv"

	captainai "github.com/flanksource/captain/pkg/ai"
	"github.com/flanksource/captain/pkg/api"
	captaincli "github.com/flanksource/captain/pkg/cli"
)

// defaultAIRuntimeOptions mirrors the boolean defaults clicky sets on
// LintOptions' embedded captaincli.AIRuntimeOptions via the `default:"true"`
// flag tags. Call sites that do NOT receive parsed CLI flags (e.g. the
// interactive commit prompt) use this to keep behaviour in sync with
// `gavel lint --ai-fix`. Captain owns the canonical defaults; if it changes
// any of them this helper must be revisited.
func defaultAIRuntimeOptions() captaincli.AIRuntimeOptions {
	// The zero value enables everything: the ambient-context toggles are now
	// negative flags (No*), all defaulting to false.
	return captaincli.AIRuntimeOptions{}
}

// buildAIFixRequest produces the (Config, Request) pair every aifix caller
// in gavel uses. Precedence is explicit CLI flags, the Gavel operation spec,
// then ~/.captain.yaml defaults. It also forces the per-feature toggles aifix
// requires regardless of caller preferences:
//
//   - Cwd: Captain launches the provider from the target repository even when
//     Gavel was invoked elsewhere with --cwd.
//   - Edit: ai-fix must edit files in place. Without it codex-cli runs
//     read-only ("workspace is mounted read-only") and claude-cli refuses
//     the apply_patch tool. Captain's ToRequest converts opts.Edit into the
//     PresetEdit permission preset, so it is set before building the request.
func buildAIFixRequest(opts captaincli.AIRuntimeOptions, operation api.Spec, workDir string) (captainai.Config, captainai.Request, error) {
	if opts.Model == "" {
		opts.Model = operation.Name
	}
	if opts.Backend == "" {
		opts.Backend = string(operation.Backend)
	}
	if opts.Effort == "" {
		opts.Effort = string(operation.Effort)
	}
	if opts.MaxTokens == 0 {
		opts.MaxTokens = operation.Budget.MaxTokens
	}
	if opts.MaxTurns == 0 {
		opts.MaxTurns = operation.Budget.MaxTurns
	}
	if operation.Budget.Cost > 0 && (opts.Budget == "" || opts.Budget == "0") {
		opts.Budget = strconv.FormatFloat(operation.Budget.Cost, 'f', -1, 64)
	}
	opts.Edit = true
	cfg, err := opts.ToConfig()
	if err != nil {
		return captainai.Config{}, captainai.Request{}, err
	}
	req, err := opts.ToRequest("", "", "")
	if err != nil {
		return captainai.Config{}, captainai.Request{}, err
	}
	req.SetCwd(workDir)
	if len(opts.Fallback) == 0 && len(operation.Fallbacks) > 0 {
		cfg.Model.Fallbacks = operation.Fallbacks
		cfg.Model, err = captainai.ResolveModelSelectors(cfg.Model)
		if err != nil {
			return captainai.Config{}, captainai.Request{}, err
		}
	}
	req.Name = cfg.Model.Name
	req.Backend = cfg.Model.Backend
	req.Fallbacks = cfg.Model.Fallbacks
	if operation.NoCache {
		req.NoCache = true
	}
	if operation.Budget.Timeout != "" {
		req.Budget.Timeout = operation.Budget.Timeout
	}
	// The rendered prompt and the loop the operation declares (verify commands,
	// commit policy). Callers that re-render per iteration (aifix) overwrite
	// Prompt on their own turn clones; nothing overwrites Workflow.
	if operation.Prompt.User != "" {
		req.Prompt.User = operation.Prompt.User
	}
	if operation.Prompt.System != "" {
		req.Prompt.System = operation.Prompt.System
	}
	if operation.Workflow != nil {
		req.Workflow = operation.Workflow
	}
	req.Budget.Cost = cfg.Budget.Cost
	cfg.Model = req.Model
	cfg.Budget = req.Budget
	return cfg, req, nil
}

func newAIFixRenderer() *captaincli.EventRenderer {
	return captaincli.NewEventRenderer(os.Stderr)
}
