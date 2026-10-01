package merge

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/captain/pkg/captainconfig"
	captaincli "github.com/flanksource/captain/pkg/cli"
	"github.com/flanksource/gavel/prompts"
	"github.com/flanksource/gavel/verify"
)

//go:embed todos-merge.prompt
var mergeTemplate string

// Prompts returns the overridable prompt template owned by this package. The
// merge prompt is registered here rather than in todos/prompt's catalog because
// that catalog is the set of prompts a LIFECYCLE STEP may run — each bound to a
// behaviour class and a run envelope — and merge records no run.
func Prompts() []prompts.Prompt {
	return []prompts.Prompt{{
		ID:    prompts.TodosMerge,
		Title: "Todo merge prompt",
		Description: "The one-shot prompt behind `gavel todos merge`: folds several TODOs into one title, " +
			"body, verification fixture and plan. Variables: {{count}}, {{{body}}} (the TODO sections, " +
			"fixtures verbatim), {{{plans}}} (their existing plans).",
		ConfigPath: "todos.merge",
		Default:    mergeTemplate,
		UsedBy:     []string{"gavel todos merge"},
	}}
}

// Template is the built-in merge prompt source, for callers that render it
// themselves (the settings preview).
func Template() string { return mergeTemplate }

// resolvePrompt layers the merge prompt the way every other one-shot gavel
// prompt is layered: the base ai: spec, then the todos.merge override, then the
// caller's request (--model/--effort).
func (opts Options) resolvePrompt(data map[string]any) (captaincli.AIRuntimeResolved, error) {
	var fileSource *string
	if opts.Override.File != "" {
		raw, err := os.ReadFile(opts.Override.ResolvedFilePath(opts.WorkDir))
		if err != nil {
			return captaincli.AIRuntimeResolved{}, fmt.Errorf("read todos merge prompt: %w", err)
		}
		source := string(raw)
		fileSource = &source
	}
	resolution, err := opts.Override.Resolve(verify.PromptResolveOptions{
		Base: opts.Base, DefaultPrompt: mergeTemplate, Data: data, Dir: opts.WorkDir,
		FileSource: fileSource, Request: opts.Request, Saved: &opts.Saved.AI,
		Name: prompts.TodosMerge, RequireModel: true,
		Normalize: func(spec api.Spec) (api.SpecNormalization, error) {
			return opts.Runtime.Normalize(captaincli.AIRuntimeNormalizeOptions{
				Spec: spec, Saved: opts.Saved, Cwd: opts.WorkDir,
			})
		},
	})
	if err != nil {
		return captaincli.AIRuntimeResolved{}, err
	}
	resolved, err := opts.Runtime.Project(captaincli.AIRuntimeProjectOptions{
		Resolved: resolution, Saved: opts.Saved,
	})
	if err != nil {
		return captaincli.AIRuntimeResolved{}, fmt.Errorf("resolve %s: %w", prompts.TodosMerge, err)
	}
	return resolved, nil
}

// savedConfig loads the captain configuration once per merge, so the prompt and
// the agent resolve against the same snapshot.
func loadSavedConfig() (captainconfig.Config, error) {
	saved, _, err := captainconfig.Load()
	if err != nil {
		return captainconfig.Config{}, fmt.Errorf("load Captain configuration: %w", err)
	}
	return saved, nil
}
