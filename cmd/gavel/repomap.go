package main

import (
	"github.com/spf13/cobra"
)

var repomapCmd = &cobra.Command{
	Use:   "repomap",
	Short: "View and query repository mapping configuration",
}

func init() {
	rootCmd.AddCommand(repomapCmd)
}
