package main

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func configuredFlagFields(cmd *cobra.Command) map[string]bool {
	fields := map[string]bool{}
	cmd.Flags().Visit(func(flag *pflag.Flag) { fields[flag.Name] = true })
	cmd.InheritedFlags().Visit(func(flag *pflag.Flag) { fields[flag.Name] = true })
	if len(cmd.Flags().Args()) > 0 {
		fields["paths"], fields["files"] = true, true
	}
	if fields["work-dir"] || fields["cwd"] {
		fields["work-dir"] = true
	}
	return fields
}
