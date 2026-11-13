// Package lint exposes Gavel's native prompt lint pipeline to comments embedders.
package lint

import (
	"context"
	"fmt"

	"github.com/flanksource/captain/pkg/api"
	captain "github.com/flanksource/captain/pkg/cli"
	prompt "github.com/flanksource/gavel/lint/prompt"
	"github.com/flanksource/gavel/linters"
)

type Selection struct {
	Directory string         `json:"directory"`
	Prompt    string         `json:"prompt"`
	Version   string         `json:"version"`
	Paths     []string       `json:"paths"`
	Variables map[string]any `json:"variables,omitempty"`
	Presets   []string       `json:"presets,omitempty"`
	Runtime   *api.Spec      `json:"runtime,omitempty"`
}

func Discover(ctx context.Context, directory string) ([]captain.PromptSummary, error) {
	if directory == "" {
		return nil, fmt.Errorf("lint directory is required")
	}
	return captain.DiscoverPrompts(ctx, directory, captain.PromptListOptions{})
}
func Inspect(ctx context.Context, selection Selection) (captain.PromptDetail, error) {
	detail, err := captain.InspectPrompt(ctx, selection.Directory, selection.Prompt)
	if err != nil {
		return detail, err
	}
	if detail.Version != selection.Version {
		return detail, captain.ErrPromptVersion
	}
	return detail, nil
}

// Prepare validates the selected prompt/version and path scope without calling a model.
func Prepare(ctx context.Context, selection Selection) (*prompt.Prepared, error) {
	if selection.Directory == "" || selection.Prompt == "" || selection.Version == "" {
		return nil, fmt.Errorf("directory, prompt and version are required")
	}
	return prompt.Prepare(ctx, prompt.PrepareOptions{Directory: selection.Directory, Reference: selection.Prompt, Version: selection.Version, Paths: selection.Paths, Variables: selection.Variables, Presets: selection.Presets, Spec: selection.Runtime})
}
func Run(ctx context.Context, selection Selection, publisher prompt.Publisher) ([]*linters.LinterResult, error) {
	prepared, err := Prepare(ctx, selection)
	if err != nil {
		return nil, err
	}
	return prepared.Run(ctx, publisher), nil
}
func Preview(ctx context.Context, selection Selection) ([]captain.PromptRenderResult, error) {
	prepared, err := Prepare(ctx, selection)
	if err != nil {
		return nil, err
	}
	return prepared.Preview()
}
