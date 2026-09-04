package main

import (
	"context"
	"fmt"
	"os"

	"github.com/flanksource/clicky"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/ai"
	"github.com/flanksource/gavel/git"
	"github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/verify"
	"github.com/spf13/cobra"
)

type AnalysisResults struct {
	Summary  git.PathSummary         `json:"summary,omitempty"`
	Analyses []models.CommitAnalysis `json:"analyses,omitempty"`
}

// analyzeAI and amendAI hold the --ai-* flag values for `git analyze` and
// `git amend-commits`. One config per command, not the shared package default
// BindFlags used to write into, where whichever FlagSet parsed last decided the
// model for all of them.
//
// They are package scoped for the same reason statusAI is: so a test can reach
// what the flags actually parsed into. amendAI in particular used to be declared
// *after* the closure that would have read it, so every --ai-* flag on
// `git amend-commits` parsed into a struct no code ever read.
var (
	analyzeAI = ai.DefaultConfig()
	amendAI   = ai.DefaultConfig()
)

func init() {
	gitCmd := &cobra.Command{
		Use:   "git",
		Short: "Git related commands",
	}
	rootCmd.AddCommand(gitCmd)

	clicky.AddCommand(gitCmd, git.HistoryOptions{}, func(filter git.HistoryOptions) (any, error) {
		commits, err := git.GetCommitHistory(filter)
		if err != nil {
			logger.Errorf("git-history failed: %v", err)
			return nil, err
		}

		logger.Infof("git-history completed successfully: %d commits retrieved", len(commits))
		return commits, nil
	})

	analyze := clicky.AddCommand(gitCmd, git.AnalyzeOptions{}, func(options git.AnalyzeOptions) (any, error) {
		logger.Tracef("git-analyzer options: %+v", options)

		if options.Path == "" {
			options.Path = "."
		}

		// One agent for the whole command, resolved from .gavel.yaml with
		// --ai-model layered on top. The per-commit analysis used to build its own
		// from ai.DefaultConfig() deep inside AnalyzeCommitHistory, so it ran on
		// the hardcoded model while only the summary pass honoured the flag.
		cfg, err := verify.LoadGavelConfig(options.Path)
		if err != nil {
			return nil, fmt.Errorf("load .gavel.yaml for git analyze: %w", err)
		}
		if options.AI {
			agent, err := newAnalyzeAgent(cfg, analyzeAI)
			if err != nil {
				return nil, err
			}
			defer closeAgent(agent)
			options.Agent = agent
		}

		var analyses models.CommitAnalyses

		if len(options.Input) > 0 {
			logger.Infof("Loading analyses from %d input files", len(options.Input))

			analyses, err = git.LoadCommitAnalysesFromJSON(options.Input)
			if err != nil {
				logger.Errorf("git-analyzer: failed to load from JSON: %v", err)
				return nil, err
			}

			analyses = git.ApplyFilters(analyses, options.HistoryOptions)
			logger.Debugf("Applied filters, %d commits remaining", len(analyses))
		} else {
			if _, err := os.Stat(options.Path); os.IsNotExist(err) {
				logger.Errorf("git-analyzer: path '%s' does not exist", options.Path)
				return nil, fmt.Errorf("path '%s' does not exist", options.Path)
			}

			showPatches := options.ShowPatch
			options.ShowPatch = true
			commits, err := git.GetCommitHistory(options.HistoryOptions)
			if err != nil {
				logger.Errorf("git-analyzer: failed to get commit history: %v", err)
				return nil, err
			}

			analyzerCtx, err := git.NewAnalyzerContext(context.Background(), options.Path)
			if err != nil {
				logger.Errorf("git-analyzer: failed to create analyzer context: %v", err)
				return nil, err
			}

			// Rewritten messages come from the same prompt as `gavel commit`, so
			// they answer to the same .gavel.yaml commit.types vocabulary.
			options.AllowedCommitTypes = cfg.Commit.Types

			logger.Debugf("git-analyzer: retrieved %d commits, starting analysis", len(commits))
			analyses, err = git.AnalyzeCommitHistory(analyzerCtx, commits, options)
			if err != nil {
				logger.Errorf("git-analyzer: failed to analyze commits: %v", err)
				return nil, err
			}

			if !showPatches {
				for i := range analyses {
					analyses[i].Patch = ""
				}
			}
		}

		clicky.WaitForGlobalCompletion()

		if options.Summary {
			clicky.Infof("Generating summary for %d analyzed commits", len(analyses))
			opts := git.SummaryOptions{
				Window:        options.SummaryWindow,
				MaxCategories: 7,
				MaxWorkers:    options.MaxConcurrent,
			}
			summaryPrompt, err := cfg.Commit.Summary.TemplateSource(options.Path, "")
			if err != nil {
				return nil, fmt.Errorf("resolve commit.summary prompt override: %w", err)
			}
			opts.SummaryPrompt = summaryPrompt

			if options.Agent != nil {
				clicky.Infof("Summarizing using AI %s", options.Agent)
				opts.Agent = options.Agent
				opts.Context = context.Background()
			}
			return git.Summarize(analyses, opts)
		}

		return analyses, nil
	})

	ai.BindFlags(analyze.Flags(), &analyzeAI)

	amendCommits := clicky.AddCommand(gitCmd, git.AmendCommitsOptions{}, func(options git.AmendCommitsOptions) (any, error) {
		logger.Tracef("git-amend-commits options: %+v", options)

		if options.Path == "" {
			options.Path = "."
		}

		if _, err := os.Stat(options.Path); os.IsNotExist(err) {
			logger.Errorf("git-amend-commits: path '%s' does not exist", options.Path)
			return nil, fmt.Errorf("path '%s' does not exist", options.Path)
		}

		// Amend is unconditionally an AI command, so it always needs an agent.
		// It used to get one from ai.DefaultConfig() inside AnalyzeCommitHistory,
		// which is why its --ai-* flags and .gavel.yaml were both ignored.
		cfg, err := verify.LoadGavelConfig(options.Path)
		if err != nil {
			return nil, fmt.Errorf("load .gavel.yaml for git amend-commits: %w", err)
		}
		agent, err := newAnalyzeAgent(cfg, amendAI)
		if err != nil {
			return nil, err
		}
		defer closeAgent(agent)
		options.Agent = agent
		options.AllowedCommitTypes = cfg.Commit.Types

		if err := git.AmendCommits(context.Background(), options); err != nil {
			logger.Errorf("git-amend-commits failed: %v", err)
			return nil, err
		}

		logger.Infof("git-amend-commits completed successfully")
		return nil, nil
	})

	ai.BindFlags(amendCommits.Flags(), &amendAI)

	clicky.AddNamedCommand("summary", gitCmd, git.SummaryByTypeOptions{}, func(opts git.SummaryByTypeOptions) (any, error) {
		repoArgs, otherArgs, err := git.SplitRepoPathArgs(opts.Args)
		if err != nil {
			return nil, err
		}
		opts.Args = otherArgs
		opts.RepoPaths = repoArgs
		if len(opts.RepoPaths) == 0 {
			if opts.Path == "" {
				opts.Path = "."
			}
			if _, err := os.Stat(opts.Path); os.IsNotExist(err) {
				return nil, fmt.Errorf("path %q does not exist", opts.Path)
			}
		}
		return git.GetCommitGroupSummaries(opts)
	})

	clicky.AddCommand(gitCmd, git.InitConfigOptions{}, func(opts git.InitConfigOptions) (any, error) {
		configPath, err := git.InitConfig(opts)
		if err != nil {
			return nil, err
		}
		return configPath, nil
	})
}

// newAnalyzeAgent builds the agent for `git analyze`/`git amend-commits` from the
// commit.message model ladder with the command's --ai-* flags layered on top.
//
// Both commands rewrite commit messages with the same prompt as `gavel commit`,
// so they resolve the same operation spec rather than inventing their own.
func newAnalyzeAgent(cfg verify.GavelConfig, flags ai.AgentConfig) (ai.Agent, error) {
	runCfg := flags
	runCfg.Model = cfg.ModelFor(cfg.Commit.Message, flags.Model)
	agent, err := ai.NewAgent(runCfg)
	if err != nil {
		return nil, fmt.Errorf("build AI agent for git analysis: %w", err)
	}
	return agent, nil
}

// closeAgent releases a process-backed agent, warning rather than failing: the
// analysis itself has already succeeded by the time this runs.
func closeAgent(agent ai.Agent) {
	if err := agent.Close(); err != nil {
		logger.Warnf("failed to close AI agent: %v", err)
	}
}
