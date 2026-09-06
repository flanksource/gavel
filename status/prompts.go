package status

import (
	"fmt"
	"os"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/captain/pkg/captainconfig"
	captaincli "github.com/flanksource/captain/pkg/cli"
	"github.com/flanksource/gavel/prompts"
	"github.com/flanksource/gavel/verify"
)

// Prompts returns the overridable prompt templates owned by the status package:
// the per-file AI summary used by `gavel status --ai`. The override is the typed
// status.summaryPrompt field, resolved against Default at the call site.
func Prompts() []prompts.Prompt {
	return []prompts.Prompt{{
		ID:          prompts.StatusSummary,
		Title:       "Status file summary",
		Description: "One-line AI summary of each changed file for `gavel status --ai`. Variable: {{details}} (the staged/unstaged diff or file contents). Output schema is fixed ({summary}).",
		ConfigPath:  "status.summary",
		Default:     fileSummaryPromptTemplate,
		UsedBy:      []string{"gavel status --ai"},
	}}
}

type SummaryPromptOptions struct {
	Dir      string
	Base     api.Spec
	Override verify.PromptSpec
	Request  api.Spec
	Saved    captainconfig.Config
	Runtime  captaincli.AIRuntimeOptions
}

type SummaryPrompt struct {
	options    SummaryPromptOptions
	fileSource *string
}

func ResolveSummaryPrompt(options SummaryPromptOptions) (*SummaryPrompt, error) {
	layers := []api.SpecLayer{api.PromptSpecLayer("ai", options.Base),
		api.PromptSpecLayer("status.summary", options.Override.Spec), api.RequestSpecLayer("request", options.Request)}
	if err := api.ValidateSpecLayers(layers...); err != nil {
		return nil, err
	}
	prepared := &SummaryPrompt{options: options}
	if options.Override.File != "" {
		raw, err := os.ReadFile(options.Override.ResolvedFilePath(options.Dir))
		if err != nil {
			return nil, fmt.Errorf("read status summary prompt: %w", err)
		}
		source := string(raw)
		prepared.fileSource = &source
	}
	return prepared, nil
}

func (p *SummaryPrompt) resolve(details string) (captaincli.AIRuntimeResolved, error) {
	resolved, err := p.options.Override.Resolve(verify.PromptResolveOptions{
		Base: p.options.Base, DefaultPrompt: fileSummaryPromptTemplate,
		Data: map[string]any{"details": details}, Dir: p.options.Dir, FileSource: p.fileSource,
		Request: p.options.Request, Saved: &p.options.Saved.AI, Name: "status.summary", RequireModel: true,
		Normalize: func(spec api.Spec) (api.SpecNormalization, error) {
			return p.options.Runtime.Normalize(captaincli.AIRuntimeNormalizeOptions{Spec: spec, Saved: p.options.Saved, Cwd: p.options.Dir})
		},
	})
	if err != nil {
		return captaincli.AIRuntimeResolved{}, err
	}
	return p.options.Runtime.Project(captaincli.AIRuntimeProjectOptions{Resolved: resolved, Saved: p.options.Saved})
}
