package main

import (
	"fmt"

	"github.com/flanksource/clicky"
	"github.com/spf13/cobra"
)

// lintSubcommands enumerates every linter that gavel registers in
// executeLinters. Each entry becomes `gavel lint <name>` with the Linters
// filter pinned. Aliases (e.g. secrets → betterleaks) get their own
// subcommand too so `gavel lint secrets` stays discoverable.
var lintSubcommands = []struct {
	subcmd     string
	linterName string
	short      string
}{
	{"golangci-lint", "golangci-lint", "Run only golangci-lint"},
	{"golangci", "golangci-lint", "Alias for golangci-lint"},
	{"ruff", "ruff", "Run only ruff (Python)"},
	{"eslint", "eslint", "Run only eslint"},
	{"pyright", "pyright", "Run only pyright (Python types)"},
	{"tsc", "tsc", "Run only tsc (TypeScript compile check)"},
	{"typescript", "tsc", "Alias for tsc"},
	{"markdownlint", "markdownlint", "Run only markdownlint"},
	{"vale", "vale", "Run only vale (prose)"},
	{"jscpd", "jscpd", "Run only jscpd (duplicate-code detector)"},
	{"betterleaks", "betterleaks", "Run only betterleaks (secret scanner)"},
	{"secrets", "betterleaks", "Alias for betterleaks"},
}

func registerLintLinterSubcommands(parent *cobra.Command) {
	for _, entry := range lintSubcommands {
		sub := clicky.AddNamedCommand(entry.subcmd, parent, LintOptions{}, func(opts LintOptions) (any, error) {
			opts.Linters = []string{entry.linterName}
			return runLint(opts)
		})
		sub.Short = entry.short
		if err := sub.Flags().MarkHidden("linters"); err != nil {
			panic(fmt.Sprintf("hide --linters on %s: %v", entry.subcmd, err))
		}
	}
}
