package main

import (
	"os"

	captainai "github.com/flanksource/captain/pkg/ai"
	"github.com/flanksource/captain/pkg/ai/pricing"
	captaincli "github.com/flanksource/captain/pkg/cli"
	"github.com/flanksource/gavel/ai/aifix"
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
// in gavel uses. It honours ~/.captain.yaml + CLI overlays via opts and
// then forces the per-feature toggles aifix requires regardless of caller
// preferences:
//
//   - Edit: ai-fix must edit files in place. Without it codex-cli runs
//     read-only ("workspace is mounted read-only") and claude-cli refuses
//     the apply_patch tool. Captain's ToRequest converts opts.Edit into the
//     PresetEdit permission preset, so it is set before building the request.
func buildAIFixRequest(opts captaincli.AIRuntimeOptions) (captainai.Config, captainai.Request, error) {
	opts.Edit = true
	cfg, err := opts.ToConfig()
	if err != nil {
		return captainai.Config{}, captainai.Request{}, err
	}
	req, err := opts.ToRequest("", "", "")
	if err != nil {
		return captainai.Config{}, captainai.Request{}, err
	}
	return cfg, req, nil
}

// newAIFixRenderer builds the stderr event renderer for an ai-fix run, prefixing
// each line with `[<model> <pct>%]`. The context window is looked up once from
// captain's pricing registry; an unknown model yields a model-only prefix.
func newAIFixRenderer(aiCfg captainai.Config) func(int, captainai.Event) {
	contextWindow := 0
	if info, ok := pricing.GetModelInfo(aiCfg.Model.Name); ok {
		contextWindow = info.ContextWindow
	}
	return aifix.NewStderrRenderer(os.Stderr, aiCfg.Model.Name, contextWindow)
}
