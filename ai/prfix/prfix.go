// Package prfix owns the overridable prompt behind `gavel pr status --ai-fix`:
// the agent reads a rendered PR status snapshot, edits files to fix the failing
// checks and unresolved comments, and the prompt's own workflow declares the
// verification loop (re-poll `gavel pr status`) and the per-turn commit policy.
package prfix

import (
	_ "embed"
	"strconv"
	"strings"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/prompts"
	"github.com/flanksource/gavel/verify"
)

//go:embed pr-status-fix.prompt
var prStatusFixPrompt string

// Prompts returns the configurable pr.fix prompt descriptor.
func Prompts() []prompts.Prompt {
	return []prompts.Prompt{{
		ID:          prompts.PRFix,
		Title:       "PR status AI fix",
		Description: "Fixes failing checks and unresolved comments for `gavel pr status --ai-fix`, and declares its own verify/commit workflow.",
		ConfigPath:  prompts.PRFix,
		Default:     prStatusFixPrompt,
	}}
}

// PRContext is the pull request identity the prompt renders. The caller builds it
// from a prwatch result so this package stays independent of prwatch (and of the
// GitHub models underneath it).
type PRContext struct {
	Number int
	Title  string
	URL    string
	Branch string
	// StatusText is the rendered `gavel pr status` snapshot handed to the agent.
	StatusText string
	// UnresolvedComments is the count of review comments still awaiting a reply.
	UnresolvedComments int
}

// ResolveSpec renders the pr.fix operation: the base `ai:` spec, the embedded
// default prompt, and any .gavel.yaml `pr.fix` override, layered in that
// precedence by verify.PromptSpec.Resolve.
func ResolveSpec(base api.Spec, override verify.PromptSpec, workDir string, pr PRContext) (api.Spec, error) {
	data := map[string]any{
		"workDir":    workDir,
		"number":     pr.Number,
		"title":      pr.Title,
		"url":        pr.URL,
		"branch":     pr.Branch,
		"statusText": strings.TrimRight(pr.StatusText, "\n"),
	}
	// Handlebars treats 0 as falsy, so an empty count simply drops its section.
	if pr.UnresolvedComments > 0 {
		data["unresolvedComments"] = strconv.Itoa(pr.UnresolvedComments)
	}
	return override.Resolve(base, prStatusFixPrompt, data, workDir)
}
