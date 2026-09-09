package outline

import (
	"fmt"
	"os"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/captain/pkg/captainconfig"
	captaincli "github.com/flanksource/captain/pkg/cli"
	"github.com/flanksource/gavel/prompts"
	"github.com/flanksource/gavel/verify"
)

// Prompts returns the overridable prompt templates owned by the outline package:
// the per-test AI summary used by `gavel test outline --ai-summary`. The override
// is the typed test.outlineSummaryPrompt field, resolved against Default at the
// call site.
func Prompts() []prompts.Prompt {
	return []prompts.Prompt{{
		ID:          prompts.TestOutlineSummary,
		Title:       "Test outline summary",
		Description: "One-line AI summary of what each test verifies for `gavel test outline --ai-summary`. Variables: {{ids}} (the test ids), {{file}}, {{source}}. Output schema is fixed (tests[]).",
		ConfigPath:  "test.outlineSummary",
		Default:     testSummaryPromptTemplate,
		UsedBy:      []string{"gavel test outline --ai-summary"},
	}}
}

type summaryPrompt struct {
	dir        string
	base       api.Spec
	override   verify.PromptSpec
	saved      captainconfig.Config
	fileSource *string
}

func resolveSummaryPrompt(workDir string) (*summaryPrompt, error) {
	cfg, err := verify.LoadGavelConfig(workDir)
	if err != nil {
		return nil, fmt.Errorf("load .gavel.yaml for test outline summary prompt: %w", err)
	}
	saved, _, err := captainconfig.Load()
	if err != nil {
		return nil, fmt.Errorf("load saved AI defaults: %w", err)
	}
	prepared := &summaryPrompt{dir: workDir, base: cfg.AI, override: cfg.Test.OutlineSummary, saved: saved}
	if err := api.ValidateSpecLayers(api.PromptSpecLayer("ai", cfg.AI), api.PromptSpecLayer("test.outlineSummary", cfg.Test.OutlineSummary.Spec)); err != nil {
		return nil, err
	}
	if cfg.Test.OutlineSummary.File != "" {
		raw, err := os.ReadFile(cfg.Test.OutlineSummary.ResolvedFilePath(workDir))
		if err != nil {
			return nil, fmt.Errorf("read test outline summary prompt: %w", err)
		}
		source := string(raw)
		prepared.fileSource = &source
	}
	return prepared, nil
}

func (p *summaryPrompt) resolve(data map[string]any) (captaincli.AIRuntimeResolved, error) {
	resolved, err := p.override.Resolve(verify.PromptResolveOptions{
		Base: p.base, DefaultPrompt: testSummaryPromptTemplate, Data: data, Dir: p.dir, FileSource: p.fileSource,
		Saved: &p.saved.AI, Name: "test.outlineSummary", RequireModel: true,
		Normalize: func(spec api.Spec) (api.SpecNormalization, error) {
			return (captaincli.AIRuntimeOptions{}).Normalize(captaincli.AIRuntimeNormalizeOptions{Spec: spec, Saved: p.saved, Cwd: p.dir})
		},
	})
	if err != nil {
		return captaincli.AIRuntimeResolved{}, err
	}
	return (captaincli.AIRuntimeOptions{}).Project(captaincli.AIRuntimeProjectOptions{Resolved: resolved, Saved: p.saved})
}
