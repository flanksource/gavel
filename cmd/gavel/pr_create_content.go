package main

import (
	"fmt"
	"strings"

	"github.com/flanksource/captain/pkg/aiflags"
	commitpkg "github.com/flanksource/gavel/commit"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/verify"
	"github.com/spf13/cobra"
)

func defaultPRCreateDeps() prCreateDeps {
	return prCreateDeps{
		createPR: github.CreatePR, openBrowser: openBrowser,
		generateContent: commitpkg.GeneratePRContent,
	}
}

func runPRCreate(cmd *cobra.Command, args []string) error {
	cfg, err := verify.LoadGavelConfig(".")
	if err != nil {
		return fmt.Errorf("load PR configuration: %w", err)
	}
	base := prCreateBase
	if !cmd.Flags().Changed("base") && cfg.PR.Base != "" {
		base = cfg.PR.Base
	}
	return runPRCreateWithDeps(cmd.Context(), args[0], prCreateOptions{
		Base: base, Draft: prCreateDraft, Mainline: prCreateMainline, Repo: prCreateRepo,
		Flags: modelFlagsFromCommand(aiflags.ModelFlags{Model: prCreateModel, NoCache: prCreateNoCache}, cmd),
	}, defaultPRCreateDeps())
}

func prContentInputForSHA(wtPath string, options prCreateOptions) (commitpkg.PRContentInput, error) {
	msg, err := captureGit(wtPath, "show", "-s", "--format=%B", "HEAD")
	if err != nil {
		return commitpkg.PRContentInput{}, fmt.Errorf("read PR commit message: %w", err)
	}
	rawFiles, err := captureGit(wtPath, "show", "--name-only", "--format=", "HEAD")
	if err != nil {
		return commitpkg.PRContentInput{}, fmt.Errorf("read PR commit files: %w", err)
	}
	var files []string
	for _, line := range strings.Split(strings.TrimSpace(rawFiles), "\n") {
		if s := strings.TrimSpace(line); s != "" {
			files = append(files, s)
		}
	}
	runtimeOptions, err := loadPRContentOptions(wtPath, options)
	if err != nil {
		return commitpkg.PRContentInput{}, err
	}
	return commitpkg.PRContentInput{
		Commits: []commitpkg.PRCommitInput{{Message: strings.TrimSpace(msg), Files: files}},
		Options: runtimeOptions,
	}, nil
}

func loadPRContentOptions(dir string, options prCreateOptions) (commitpkg.Options, error) {
	cfg, err := verify.LoadGavelConfig(dir)
	if err != nil {
		return commitpkg.Options{}, fmt.Errorf("load .gavel.yaml for PR content: %w", err)
	}
	return commitpkg.Options{WorkDir: dir, AI: cfg.AI, PR: cfg.PR, Flags: options.Flags, Saved: options.Saved}, nil
}
