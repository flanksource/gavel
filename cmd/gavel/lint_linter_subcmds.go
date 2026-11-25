package main

import (
	"fmt"

	"github.com/flanksource/clicky"
	"github.com/spf13/cobra"
)

// lintSubcommands enumerates every linter that gavel registers in
// lint.Execute. Each entry becomes `gavel lint <name>` with the Linters
// filter pinned. Aliases (e.g. secrets → betterleaks) get their own
// subcommand too so `gavel lint secrets` stays discoverable.
var lintSubcommands = []struct {
	subcmd     string
	linterName string
	short      string
	example    string
}{
	{"golangci-lint", "golangci-lint", "Check Go code with golangci-lint", "--tags integration ./pkg"},
	{"golangci", "golangci-lint", "Check Go code (alias for golangci-lint)", "--changed"},
	{"ruff", "ruff", "Check Python code with Ruff", "--exclude ruff:F401 ./src"},
	{"eslint", "eslint", "Check JavaScript and TypeScript with ESLint", "--changed ./web"},
	{"oxlint", "oxlint", "Check JavaScript and TypeScript with Oxlint", "--dry-run ./web"},
	{"react-doctor", "react-doctor", "Check React projects with React Doctor", "--changed ./web"},
	{"pyright", "pyright", "Check Python types with Pyright", "--dry-run ./src"},
	{"tsc", "tsc", "Check TypeScript types with the TypeScript compiler", "--dry-run ./web"},
	{"typescript", "tsc", "Check TypeScript types (alias for tsc)", "--dry-run ./web"},
	{"markdownlint", "markdownlint", "Check Markdown formatting and structure", "./docs"},
	{"vale", "vale", "Check prose against Vale style rules", "./docs"},
	{"jscpd", "jscpd", "Detect duplicate code with jscpd", "--summary ./src"},
	{"uir", "uir", "Check architecture rules from .uir.yaml and .uir.d", "--exclude 'small-*'"},
	{"sql", "sql", "Validate migration scripts with SQLGlot and native UIR rules", "migrations -- --dialect tsql --compatibility-level 130"},
	{"json", "json", "Validate JSON, YAML, NDJSON and Markdown frontmatter", "--schema schema.json config.json"},
	{"github", "github", "Audit GitHub PR health, Dependabot security alerts and CodeQL findings", "-- --repo acme/service --stale-after 720h"},
	{"yaml", "json", "Validate YAML and other structured files (alias for json)", "--schema schema.json config.yaml"},
	{"betterleaks", "betterleaks", "Scan for secrets with Betterleaks", "--changed"},
	{"secrets", "betterleaks", "Scan for secrets (alias for betterleaks)", "--dry-run"},
}

func registerLintLinterSubcommands(parent *cobra.Command) {
	for _, entry := range lintSubcommands {
		var sub *cobra.Command
		sub = clicky.AddNamedCommand(entry.subcmd, parent, LintOptions{}, func(opts LintOptions) (any, error) {
			opts, err := splitLintPassThroughArgs(sub, opts)
			if err != nil {
				return nil, err
			}
			opts.Linters = []string{entry.linterName}
			return runLint(sub.Context(), opts)
		})
		sub.Use = entry.subcmd + " [flags] [path...] [-- linter args...]"
		bindLintSourceFlags(sub)
		sub.Short = entry.short
		sub.Long = entry.short + ".\n\nPass files or directories to scope findings. Arguments after -- go to the linter."
		sub.Example = "  " + sub.CommandPath() + " " + entry.example
		switch entry.linterName {
		case "json":
			sub.Use = entry.subcmd + " [flags] [path...]"
			sub.Long = entry.short + ".\n\nPass files or directories to validate. Use --schema to apply a JSON Schema to the selected files."
		case "github":
			sub.Use = "github [flags] [-- github args...]"
			sub.Long = entry.short + ".\n\nThis audit is repository-wide and runs when explicitly selected. Use --cwd to select a project. After --, use --repo owner/repo to override configured repository selection and --stale-after to set the PR inactivity threshold (default 720h)."
		}
		if err := sub.Flags().MarkHidden("linters"); err != nil {
			panic(fmt.Sprintf("hide --linters on %s: %v", entry.subcmd, err))
		}
	}
}
