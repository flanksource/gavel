package commit

import (
	"fmt"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/captain/pkg/captainconfig"
	captaincli "github.com/flanksource/captain/pkg/cli"
	"github.com/flanksource/commons/logger"
	clickyai "github.com/flanksource/gavel/ai"
	"github.com/flanksource/gavel/verify"
)

func (opts Options) savedConfig() (captainconfig.Config, error) {
	if opts.Saved != nil {
		return *opts.Saved, nil
	}
	saved, _, err := captainconfig.Load()
	return saved, err
}

// LoadAIConfig captures one snapshot for all AI operations and retries in this invocation.
func (opts *Options) LoadAIConfig() error {
	if opts.Saved != nil || (!opts.Push && !opts.Summary &&
		(opts.Message != "" || (opts.Fixup != "" && opts.Fixup != FixupAuto))) {
		return nil
	}
	saved, err := opts.savedConfig()
	if err != nil {
		return fmt.Errorf("load Captain configuration: %w", err)
	}
	opts.Saved = &saved
	return nil
}

func (opts Options) promptOptions() (verify.PromptResolveOptions, captainconfig.Config, error) {
	saved, err := opts.savedConfig()
	if err != nil {
		return verify.PromptResolveOptions{}, saved, err
	}
	model, err := opts.Flags.ToModel()
	if err != nil {
		return verify.PromptResolveOptions{}, saved, err
	}
	return verify.PromptResolveOptions{
		Base: opts.AI, Dir: opts.WorkDir, Request: api.Spec{Model: model}, RequireModel: true,
	}, saved, nil
}

type commitPromptOptions struct {
	Prompt  verify.PromptSpec
	Name    string
	Default string
	Data    map[string]any
	Request api.Spec
}

func (opts Options) resolvePrompt(in commitPromptOptions) (captaincli.AIRuntimeResolved, error) {
	options, saved, err := opts.promptOptions()
	if err != nil {
		return captaincli.AIRuntimeResolved{}, err
	}
	options.Name, options.DefaultPrompt, options.Data = in.Name, in.Default, in.Data
	options.Request = options.Request.Merge(in.Request)
	options.Saved = &saved.AI
	options.Normalize = func(spec api.Spec) (api.SpecNormalization, error) {
		return (captaincli.AIRuntimeOptions{}).Normalize(captaincli.AIRuntimeNormalizeOptions{
			Spec: spec, Saved: saved, Cwd: opts.WorkDir,
		})
	}
	resolution, err := in.Prompt.Resolve(options)
	if err != nil {
		return captaincli.AIRuntimeResolved{}, err
	}
	resolved, err := (captaincli.AIRuntimeOptions{}).Project(captaincli.AIRuntimeProjectOptions{
		Resolved: resolution, Saved: saved,
	})
	if err != nil {
		return captaincli.AIRuntimeResolved{}, fmt.Errorf("resolve %s: %w", in.Name, err)
	}
	for _, warning := range resolved.Resolution.Warnings {
		logger.Warnf("%s: %s", in.Name, warning)
	}
	return resolved, nil
}

// BuildAgent uses the configuration projected from the complete resolved prompt.
func BuildAgent(cfg clickyai.AgentConfig) (clickyai.Agent, error) {
	defaults := clickyai.DefaultConfig()
	if cfg.MaxConcurrent == 0 {
		cfg.MaxConcurrent = defaults.MaxConcurrent
	}
	if cfg.CacheTTL == 0 {
		cfg.CacheTTL = defaults.CacheTTL
	}
	agent, err := newAgentFunc(cfg)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrLLMUnavailable, err)
	}
	if agent == nil {
		return nil, fmt.Errorf("%w: agent factory returned nil", ErrLLMUnavailable)
	}
	return agent, nil
}
