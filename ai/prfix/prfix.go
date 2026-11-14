// Package prfix owns the overridable prompt behind `gavel pr status --ai-fix`:
// the agent reads a rendered PR status snapshot, edits files to fix the failing
// checks and unresolved comments, and the prompt's own workflow declares the
// verification loop (re-poll `gavel pr status`) and the per-turn commit policy.
package prfix

import (
	_ "embed"
	"fmt"
	"strconv"
	"strings"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/prompts"
	"github.com/flanksource/gavel/verify"
)

//go:embed pr-status-fix.prompt
var prStatusFixPrompt string

//go:embed pr-review.prompt
var prReviewPrompt string

// Prompts returns the configurable pr.fix prompt descriptor.
func Prompts() []prompts.Prompt {
	return []prompts.Prompt{{
		ID:          prompts.PRFix,
		Title:       "PR status AI fix",
		Description: "Fixes failing checks and unresolved comments for `gavel pr status --ai-fix`, and declares its own verify/commit workflow.",
		ConfigPath:  prompts.PRFix,
		Default:     prStatusFixPrompt,
		UsedBy:      []string{"gavel pr status --ai-fix"},
	}, {ID: prompts.PRReview, Title: "PR review", Description: "Reviews a PR from a manually selected lifecycle context.", ConfigPath: prompts.PRReview, Default: prReviewPrompt, UsedBy: []string{"PR lifecycle"}}}
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
	// BaseBranch is the branch the PR merges into.
	BaseBranch string
	// BaseSHA is the BaseBranch commit the conflict rebase replays onto.
	BaseSHA string
	// Conflicts are the paths GitHub's merge of BaseBranch conflicts on. Non-empty
	// means gavel started rebasing the agent's tree onto BaseSHA before the first
	// turn, and finishing that rebase comes before any CI fix.
	Conflicts []Conflict
}

// Conflict is one conflicted path and the kind git reported for it.
type Conflict struct {
	Path string `json:"path"`
	Kind string `json:"kind,omitempty"`
}

type ResolveOptions struct {
	Base      api.Spec
	Context   *api.Spec
	Prompt    verify.PromptSpec
	Dir       string
	PR        PRContext
	Reference string
	Inputs    map[string]any
}

// Layers renders the operation before CLI flags and saved defaults are resolved.
func Layers(options ResolveOptions) ([]api.SpecLayer, error) {
	reference := options.Reference
	if reference == "" {
		reference = prompts.PRFix
	}
	defaultPrompt := prStatusFixPrompt
	switch {
	case reference == prompts.PRFix:
	case reference == prompts.PRReview:
		defaultPrompt = prReviewPrompt
	case strings.HasPrefix(reference, "file:"):
		options.Prompt.File = strings.TrimPrefix(reference, "file:")
	default:
		return nil, fmt.Errorf("unknown PR prompt reference %q", reference)
	}
	pr := options.PR
	data := map[string]any{
		"workDir":    options.Dir,
		"number":     pr.Number,
		"title":      pr.Title,
		"url":        pr.URL,
		"branch":     pr.Branch,
		"baseBranch": pr.BaseBranch,
		"statusText": strings.TrimRight(pr.StatusText, "\n"),
	}
	for key, value := range options.Inputs {
		data[key] = value
	}
	if len(pr.Conflicts) > 0 {
		conflicts := make([]map[string]any, 0, len(pr.Conflicts))
		for _, c := range pr.Conflicts {
			conflicts = append(conflicts, map[string]any{"path": c.Path, "kind": c.Kind})
		}
		data["conflicts"] = conflicts
		data["baseSha"] = pr.BaseSHA
	}
	// Handlebars treats 0 as falsy, so an empty count simply drops its section.
	if pr.UnresolvedComments > 0 {
		data["unresolvedComments"] = strconv.Itoa(pr.UnresolvedComments)
	}
	return options.Prompt.Layers(verify.PromptResolveOptions{Base: options.Base, Context: options.Context, DefaultPrompt: defaultPrompt, Data: data, Dir: options.Dir, Name: reference})
}
