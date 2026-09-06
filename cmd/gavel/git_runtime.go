package main

import (
	"github.com/flanksource/captain/pkg/captainconfig"
	"github.com/flanksource/gavel/ai"
	"github.com/flanksource/gavel/git"
	"github.com/flanksource/gavel/verify"
	"github.com/spf13/pflag"
)

type gitAIOptions struct {
	Config  verify.GavelConfig
	Flags   ai.AgentConfig
	FlagSet *pflag.FlagSet
}

func configureGitAI(options *git.AnalyzeOptions, input gitAIOptions) error {
	saved, _, err := captainconfig.Load()
	if err != nil {
		return err
	}
	options.Prompt = input.Config.Commit.Message
	options.PromptOptions = verify.PromptResolveOptions{Base: input.Config.AI, Request: ai.FlagSpec(input.Flags, input.FlagSet), Dir: options.Path}
	options.Saved = saved
	options.AllowedCommitTypes = input.Config.Commit.Types
	options.AgentFactory = func(resolved ai.AgentConfig) (ai.Agent, error) {
		resolved.CacheTTL = input.Flags.CacheTTL
		resolved.CacheDBPath = input.Flags.CacheDBPath
		resolved.ProjectName = input.Flags.ProjectName
		resolved.MaxConcurrent = input.Flags.MaxConcurrent
		return ai.NewAgent(resolved)
	}
	return nil
}
