package git

import (
	"github.com/flanksource/captain/pkg/api"
	captaincli "github.com/flanksource/captain/pkg/cli"
	"github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/prompts"
)

func PrepareCommitMessage(commit models.CommitAnalysis, opts AnalyzeOptions) (captaincli.AIRuntimeResolved, error) {
	allowed, err := allowedCommitTypes(opts.AllowedCommitTypes)
	if err != nil {
		return captaincli.AIRuntimeResolved{}, err
	}
	options := opts.PromptOptions
	options.DefaultPrompt = commitMessagePrompt
	options.Name = prompts.CommitMessage
	options.RequireModel = true
	options.Saved = &opts.Saved.AI
	options.Normalize = func(spec api.Spec) (api.SpecNormalization, error) {
		return (captaincli.AIRuntimeOptions{}).Normalize(captaincli.AIRuntimeNormalizeOptions{Spec: spec, Saved: opts.Saved, Cwd: options.Dir})
	}
	options.Data = commit.AsMap()
	for key, value := range commitPromptData(opts.MaxBodyLines, allowed) {
		options.Data[key] = value
	}
	resolved, err := opts.Prompt.Resolve(options)
	if err != nil {
		return captaincli.AIRuntimeResolved{}, err
	}
	resolved.Spec.Prompt.SchemaJSON, err = enumerateCommitType(resolved.Spec.Prompt.SchemaJSON, allowed)
	if err != nil {
		return captaincli.AIRuntimeResolved{}, err
	}
	if resolved.Spec.Prompt.Source == "" {
		resolved.Spec.Prompt.Source = commitMessagePromptFile
	}
	return (captaincli.AIRuntimeOptions{}).Project(captaincli.AIRuntimeProjectOptions{Resolved: resolved, Saved: opts.Saved})
}
