package main

import "github.com/spf13/cobra"

func bindLintSourceFlags(command *cobra.Command) {
	command.Flags().Bool("generated", true, "Include generated-file findings; omit to preserve each linter's defaults")
}
