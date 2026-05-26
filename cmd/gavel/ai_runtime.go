package main

import (
	captainai "github.com/flanksource/captain/pkg/ai"
	captaincli "github.com/flanksource/captain/pkg/cli"
)

// defaultAIRuntimeOptions mirrors the boolean defaults clicky sets on
// LintOptions' embedded captaincli.AIRuntimeOptions via the `default:"true"`
// flag tags. Call sites that do NOT receive parsed CLI flags (e.g. the
// interactive commit prompt) use this to keep behaviour in sync with
// `gavel lint --ai-fix`. Captain owns the canonical defaults; if it changes
// any of them this helper must be revisited.
func defaultAIRuntimeOptions() captaincli.AIRuntimeOptions {
	return captaincli.AIRuntimeOptions{
		MCP:     true,
		Hooks:   true,
		Skills:  true,
		User:    true,
		Project: true,
		Memory:  true,
	}
}

// buildAIFixRequest produces the (Config, Request) pair every aifix caller
// in gavel uses. It honours ~/.captain.yaml + CLI overlays via opts and
// then forces the per-feature toggles aifix requires regardless of caller
// preferences:
//
//   - Edit: ai-fix must edit files in place. Without it codex-cli runs
//     read-only ("workspace is mounted read-only") and claude-cli refuses
//     the apply_patch tool.
//   - Verbose + StrictMCP: needed by claude-cli streaming, ignored by
//     codex-cli / gemini-cli.
func buildAIFixRequest(opts captaincli.AIRuntimeOptions) (captainai.Config, captainai.Request) {
	cfg := opts.ToConfig()
	req := opts.ToRequest("", "", "")
	req.Edit = true
	req.Verbose = true
	req.StrictMCP = true
	return cfg, req
}
