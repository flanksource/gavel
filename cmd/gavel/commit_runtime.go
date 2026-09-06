package main

import (
	"github.com/flanksource/captain/pkg/aiflags"
	"github.com/flanksource/captain/pkg/api"
	captaincli "github.com/flanksource/captain/pkg/cli"
	"github.com/flanksource/clicky"
	commitpkg "github.com/flanksource/gavel/commit"
	"github.com/spf13/cobra"
)

func registerCommitCommand(parent *cobra.Command, run func(CommitOptions) (any, error)) *cobra.Command {
	var cmd *cobra.Command
	cmd = clicky.AddNamedCommand("commit", parent, CommitOptions{}, func(opts CommitOptions) (any, error) {
		opts.ModelFlags = modelFlagsFromCommand(opts.ModelFlags, cmd)
		if cmd.Flags().Changed("group-model") {
			opts.groupModelFields = api.FieldPresence{"/model": true}
		}
		return run(opts)
	})
	cmd.Use = "commit [files...]"
	cmd.Args = cobra.ArbitraryArgs
	if f := cmd.Flags().Lookup("fixup"); f != nil {
		f.NoOptDefVal = commitpkg.FixupAuto
	}
	return cmd
}

func modelFlagsFromCommand(flags aiflags.ModelFlags, cmd *cobra.Command) aiflags.ModelFlags {
	return (captaincli.AIRuntimeOptions{AIProviderOptions: captaincli.AIProviderOptions{ModelFlags: flags}}).
		WithChangedFlags(cmd.Flags()).ModelFlags
}
