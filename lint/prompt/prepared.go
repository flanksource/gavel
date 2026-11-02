package prompt

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"

	"github.com/flanksource/captain/pkg/api"
	captaincli "github.com/flanksource/captain/pkg/cli"
	"github.com/flanksource/commons-db/shell"
	"github.com/flanksource/gavel/linters"
)

type PrepareOptions struct {
	Directory string
	Reference string
	Version   string
	Variables map[string]any
	Presets   []string
	Spec      *api.Spec
	Paths     []string
	Env       []string
}

type Prepared struct{ input reviewsInput }

func Prepare(ctx context.Context, options PrepareOptions) (*Prepared, error) {
	detail, err := captaincli.InspectPrompt(ctx, options.Directory, options.Reference)
	if err != nil {
		return nil, err
	}
	if detail.Version != options.Version {
		return nil, captaincli.ErrPromptVersion
	}
	if detail.Run.Spec != nil && detail.Run.Spec.Setup != nil {
		return nil, errors.New("prompt setup conflicts with the prepared launch workspace; configure setup on workspace")
	}
	if len(options.Paths) > 0 {
		if _, exists := options.Variables["files"]; exists {
			return nil, errors.New("path scope conflicts with variables.files")
		}
	}
	scope, err := newFileScope(options.Directory, options.Paths)
	if err != nil {
		return nil, err
	}
	presets := slices.Clone(options.Presets)
	if options.Presets == nil {
		presets = slices.Clone(detail.Presets)
	}
	if !slices.Contains(presets, "lint") {
		presets = append(presets, "lint")
	}
	spec := api.Spec{}
	if options.Spec != nil {
		spec = *options.Spec
	}
	if spec.Setup != nil {
		return nil, errors.New("prompt configuration must use the prepared launch workspace; configure setup on workspace")
	}
	spec.Setup = &shell.Setup{Cwd: options.Directory, Env: options.Env}
	prepared := &Prepared{input: reviewsInput{WorkDir: options.Directory}}
	variables := maps.Clone(options.Variables)
	if variables == nil {
		variables = map[string]any{}
	}
	if len(options.Paths) > 0 {
		variables["files"] = options.Paths
	}
	run := &reviewRun{WorkDir: options.Directory, Scope: scope, Name: detail.Name}
	invocation, err := captaincli.PreparePrompt(ctx, captaincli.PreparePromptOptions{
		Reference: detail.ID, Version: options.Version, Directory: options.Directory, Tools: run.tools(),
		Request: captaincli.PromptRenderRequest{Variables: variables, Presets: presets, Spec: &spec},
	})
	if err != nil {
		return nil, fmt.Errorf("prepare %s: %w", detail.Name, err)
	}
	prepared.input.Reviews = []review{{Name: run.Name, Scope: scope, Prepared: invocation, Run: run}}
	return prepared, nil
}

func (p *Prepared) Run(ctx context.Context, publisher Publisher) []*linters.LinterResult {
	p.input.Publisher = publisher
	for _, review := range p.input.Reviews {
		review.Run.Publisher = publisher
	}
	return runReviews(ctx, p.input)
}
