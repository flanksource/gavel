package main

import (
	"context"
	"fmt"
	"os"

	"github.com/flanksource/gavel/git"
	"github.com/flanksource/gavel/models"
	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/ai"
	"github.com/flanksource/commons-db/llm"
	"github.com/flanksource/commons/logger"
	"github.com/spf13/cobra"
)

type AnalysisResults struct {
	Summary  git.PathSummary         `json:"summary,omitempty"`
	Analyses []models.CommitAnalysis `json:"analyses,omitempty"`
}

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

		var analyses models.CommitAnalyses
		var err error

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
			if options.Path == "" {
				options.Path = "."
			}

			if _, err := os.Stat(options.Path); os.IsNotExist(err) {
				logger.Errorf("git-analyzer: path '%s' does not exist", options.Path)
				return nil, fmt.Errorf("path '%s' does not exist", options.Path)
			}

			showPatches := options.HistoryOptions.ShowPatch
			options.HistoryOptions.ShowPatch = true
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

			logger.Debugf("git-analyzer: retrieved %d commits, starting analysis", len(commits))
			analyses, err = git.AnalyzeCommitHistory(analyzerCtx, commits, options)
			if err != nil {
				logger.Errorf("git-analyzer: failed to analyze commits: %v", err)
				return nil, err
			}

			if !showPatches {
				for i := range analyses {
					analyses[i].Commit.Patch = ""
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
			}

			if options.AI {
				agent, err := llm.NewLLMAgent(ai.DefaultConfig())
				if err != nil {
					return nil, fmt.Errorf("failed to get default AI agent for summary: %w", err)
				}
				clicky.Infof("Summarizing using AI %s", agent)
				opts.Agent = agent
				opts.Context = context.Background()
			}
			return git.Summarize(analyses, opts)
		}

		return analyses, nil
	})

	ai.BindFlags(analyze.Flags())

	amendCommits := clicky.AddCommand(gitCmd, git.AmendCommitsOptions{}, func(options git.AmendCommitsOptions) (any, error) {
		logger.Tracef("git-amend-commits options: %+v", options)

		if options.Path == "" {
			options.Path = "."
		}

		if _, err := os.Stat(options.Path); os.IsNotExist(err) {
			logger.Errorf("git-amend-commits: path '%s' does not exist", options.Path)
			return nil, fmt.Errorf("path '%s' does not exist", options.Path)
		}

		err := git.AmendCommits(context.Background(), options)
		if err != nil {
			logger.Errorf("git-amend-commits failed: %v", err)
			return nil, err
		}

		logger.Infof("git-amend-commits completed successfully")
		return nil, nil
	})

	ai.BindFlags(amendCommits.Flags())
}
