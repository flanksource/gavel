package main

import (
	"github.com/spf13/cobra"
)

var (
	fixSyncDir string
	fixRepo    string
)

var prFixCmd = &cobra.Command{
	Use:   "fix [pr-number]",
	Short: "Retired file-backed PR TODO workflow",
	Long: `The former PR fix workflow wrote PR failures to .todos files, discovered
them with the file provider, and executed them as runtime TODOs. Runtime TODOs
now live only in PostgreSQL, so this compatibility command returns migration
guidance without fetching a PR or reading/writing .todos.`,
	SilenceUsage: true,
	Args:         cobra.MaximumNArgs(1),
	RunE:         runPRFix,
}

func runPRFix(_ *cobra.Command, _ []string) error {
	return retiredTODOFileRuntimeError("gavel pr fix", "file-backed PR syncing")
}

func init() {
	prCmd.AddCommand(prFixCmd)
	prFixCmd.Flags().StringVarP(&fixRepo, "repo", "R", "", "GitHub repository (owner/repo)")
	prFixCmd.Flags().StringVar(&fixSyncDir, "dir", "", "Retired .todos compatibility option (runtime TODOs use PostgreSQL)")
	prFixCmd.Flags().IntVar(&maxRetries, "max-retries", 3, "Maximum retry attempts")
	prFixCmd.Flags().Float64Var(&maxBudget, "max-budget", 0, "Maximum budget in USD")
	prFixCmd.Flags().IntVar(&maxTurns, "max-turns", 0, "Maximum conversation turns")
	prFixCmd.Flags().StringVar(&groupBy, "group-by", "", "Group TODOs by: file, directory, all, or none")
	prFixCmd.Flags().BoolVar(&dirty, "dirty", false, "Skip git stash/checkout")
	prFixCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print commands without executing")
}
