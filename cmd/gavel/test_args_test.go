package main

import (
	"reflect"
	"testing"

	"github.com/flanksource/gavel/testrunner"
	"github.com/spf13/cobra"
)

func TestSplitTestPassThroughArgs(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantPaths   []string
		wantThrough []string
	}{
		{name: "no separator", args: []string{"./pkg", "./cmd"}, wantPaths: []string{"./pkg", "./cmd"}},
		{name: "focus without paths", args: []string{"--", "--focus", "TestFoo"}, wantThrough: []string{"--focus", "TestFoo"}},
		{name: "paths and raw args", args: []string{"./pkg", "./cmd", "--", "--label-filter", "smoke"}, wantPaths: []string{"./pkg", "./cmd"}, wantThrough: []string{"--label-filter", "smoke"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{Use: "test"}
			cmd.Flags().SetInterspersed(true)
			if err := cmd.Flags().Parse(tt.args); err != nil {
				t.Fatalf("parse args: %v", err)
			}
			opts := testrunner.RunOptions{StartingPaths: append([]string(nil), cmd.Flags().Args()...)}
			if err := splitTestPassThroughArgs(cmd, &opts); err != nil {
				t.Fatalf("splitTestPassThroughArgs: %v", err)
			}
			if !reflect.DeepEqual(opts.StartingPaths, tt.wantPaths) {
				t.Fatalf("StartingPaths = %v, want %v", opts.StartingPaths, tt.wantPaths)
			}
			if !reflect.DeepEqual(opts.PassThroughArgs, tt.wantThrough) {
				t.Fatalf("PassThroughArgs = %v, want %v", opts.PassThroughArgs, tt.wantThrough)
			}
		})
	}
}

func TestFrameworkSubcommandsAreRegistered(t *testing.T) {
	for _, name := range []string{"go", "ginkgo", "jest", "vitest", "playwright"} {
		cmd, _, err := testCmd.Find([]string{name})
		if err != nil {
			t.Fatalf("find test framework subcommand %q: %v", name, err)
		}
		if cmd == nil || cmd.Name() != name {
			t.Fatalf("test framework subcommand %q = %#v", name, cmd)
		}
	}
}
