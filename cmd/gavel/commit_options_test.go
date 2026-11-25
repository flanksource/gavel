package main

import (
	"context"
	"testing"

	commitpkg "github.com/flanksource/gavel/commit"
	"github.com/flanksource/gavel/verify"
	"github.com/spf13/cobra"
)

func TestBuildCommitOptionsTreeAliasEnablesInteractive(t *testing.T) {
	got := buildCommitOptions(CommitOptions{Tree: true}, "/repo", verify.GavelConfig{}, nil)
	if !got.Interactive {
		t.Fatalf("Tree alias did not enable interactive mode")
	}
}

func TestParsePushFlag(t *testing.T) {
	cases := []struct {
		value string
		push  bool
		mode  commitpkg.PushMode
	}{
		{value: "", push: false, mode: commitpkg.PushModeCommits},
		{value: "false", push: false, mode: commitpkg.PushModeCommits},
		{value: "true", push: true, mode: commitpkg.PushModeCommits},
		{value: "branch", push: true, mode: commitpkg.PushModeBranch},
	}
	for _, c := range cases {
		push, mode, err := parsePushFlag(c.value)
		if err != nil || push != c.push || mode != c.mode {
			t.Errorf("parsePushFlag(%q) = %v, %q, %v; want %v, %q, nil", c.value, push, mode, err, c.push, c.mode)
		}
	}
	if _, _, err := parsePushFlag("main"); err == nil {
		t.Errorf("parsePushFlag(%q) accepted an unknown mode", "main")
	}
}

func TestCommitPushFlagParsesBareAndBranchValues(t *testing.T) {
	cases := map[string][]string{
		"true":   {"-p", "file.go"},
		"branch": {"--push=branch"},
		"":       {},
	}
	for want, args := range cases {
		var got string
		cmd := registerCommitCommand(&cobra.Command{Use: "gavel"}, func(opts CommitOptions) (any, error) {
			got = opts.Push
			return nil, nil
		})
		if err := cmd.ParseFlags(args); err != nil {
			t.Fatalf("parse %v: %v", args, err)
		}
		cmd.SetContext(context.Background())
		if err := cmd.RunE(cmd, cmd.Flags().Args()); err != nil {
			t.Fatalf("run %v: %v", args, err)
		}
		if got != want {
			t.Errorf("args %v: Push = %q, want %q", args, got, want)
		}
	}
}

func TestBuildCommitOptionsPassesInteractiveSummary(t *testing.T) {
	got := buildCommitOptions(CommitOptions{Interactive: true, Summary: true}, "/repo", verify.GavelConfig{}, nil)
	if !got.Interactive || !got.Summary {
		t.Fatalf("interactive summary flags were not propagated: %+v", got)
	}
}
