package main

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func configuredFlagFields(cmd *cobra.Command) map[string]bool {
	fields := map[string]bool{}
	cmd.Flags().Visit(func(flag *pflag.Flag) { fields[flag.Name] = true })
	cmd.InheritedFlags().Visit(func(flag *pflag.Flag) { fields[flag.Name] = true })
	paths := len(cmd.Flags().Args())
	if separator := cmd.ArgsLenAtDash(); separator >= 0 {
		paths = separator
	}
	if paths > 0 {
		fields["paths"], fields["files"] = true, true
	}
	for _, name := range []string{"framework", "linters"} {
		if flag := cmd.Flags().Lookup(name); flag != nil && flag.Hidden {
			fields[name] = true
		}
	}
	if fields["work-dir"] || fields["cwd"] {
		fields["work-dir"] = true
	}
	return fields
}
