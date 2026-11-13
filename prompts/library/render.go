package library

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/flanksource/captain/pkg/ai/prompt"
	"github.com/flanksource/captain/pkg/api"
	captaincli "github.com/flanksource/captain/pkg/cli"
)

type RenderOptions struct {
	Directory string
	Name      string
	Data      map[string]any
	System    bool
}

// Render resolves editable prompt sources without selecting or invoking a model.
func Render(ctx context.Context, options RenderOptions) (api.Prompt, error) {
	directory, err := filepath.Abs(options.Directory)
	if err != nil {
		return api.Prompt{}, err
	}
	options.Directory = directory
	ctx, err = Context(ctx, options.Directory)
	if err != nil {
		return api.Prompt{}, err
	}
	detail, err := captaincli.InspectPrompt(ctx, options.Directory, options.Name)
	if err != nil {
		return api.Prompt{}, err
	}
	if detail.ParseError != "" {
		return api.Prompt{}, fmt.Errorf("prompt %s: %s", options.Name, detail.ParseError)
	}
	rendered, _, err := prompt.LoadNamed(options.Name, detail.Content).Render(prompt.RenderOptions{Data: options.Data, Declared: true, Dir: options.Directory})
	if err != nil {
		return api.Prompt{}, err
	}
	rendered.Prompt.User = strings.TrimSpace(rendered.Prompt.User)
	rendered.Prompt.System = strings.TrimSpace(rendered.Prompt.System)
	if rendered.Prompt.User == "" && rendered.Prompt.System == "" {
		return api.Prompt{}, fmt.Errorf("prompt %s rendered no instructions", options.Name)
	}
	if options.System && rendered.Prompt.System == "" {
		return api.Prompt{}, fmt.Errorf("prompt %s requires system instructions", options.Name)
	}
	if !options.System && rendered.Prompt.User == "" {
		return api.Prompt{}, fmt.Errorf("prompt %s requires user instructions", options.Name)
	}
	rendered.Prompt.Source = detail.ID
	return rendered.Prompt, nil
}
