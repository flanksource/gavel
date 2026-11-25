package main

import (
	"fmt"

	"github.com/flanksource/clicky"
	"github.com/flanksource/gavel/testrunner"
	"github.com/flanksource/gavel/testrunner/parsers"
	"github.com/spf13/cobra"
)

var testFrameworkHelp = map[parsers.Framework]struct{ title, scope, example string }{
	parsers.GoTest:     {"Go tests", "Pass Go package directories or a _test.go file.", "./pkg/parser -- -run TestParser -count=1"},
	parsers.Ginkgo:     {"Ginkgo specs", "Pass Go package directories or a Ginkgo spec file.", "./pkg/parser -- --label-filter integration --focus Parser"},
	parsers.Jest:       {"Jest tests", "Pass a Jest project directory or a test file.", "./web -- --testNamePattern parser"},
	parsers.Vitest:     {"Vitest tests", "Pass a Vitest project directory or a test file.", "./web -- --testNamePattern parser"},
	parsers.Playwright: {"Playwright tests", "Pass a Playwright project directory or a spec file.", "./web -- --project chromium --grep login"},
}

// registerTestFrameworkSubcommands wires `gavel test <framework>` commands.
// Each subcommand mirrors `gavel test`'s flag surface (clicky re-binds them
// against a fresh RunOptions) and pins Frameworks before delegating to
// runTests. Positional args remain starting paths while arguments after --
// pass through to the selected framework.
func registerTestFrameworkSubcommands() {
	if testCmd == nil {
		panic("testCmd must be initialized before registering framework subcommands")
	}
	for _, fw := range parsers.AllFrameworks {
		name := frameworkSubcommandName(fw)
		var flags testCommandFlags
		var sub *cobra.Command
		sub = clicky.AddNamedCommand(name, testCmd, testrunner.RunOptions{}, func(opts testrunner.RunOptions) (any, error) {
			opts, detach, err := finalizeTestCommandOptions(sub, opts, flags)
			if err != nil {
				return nil, err
			}
			opts.Frameworks = []string{string(fw)}
			return runTests(opts, detach)
		})
		help, ok := testFrameworkHelp[fw]
		if !ok {
			panic(fmt.Sprintf("missing framework help for %s", fw))
		}
		sub.Use = fmt.Sprintf("%s [flags] [path...] [-- %s args...]", name, fw)
		sub.Short = "Run " + help.title
		sub.Long = sub.Short + ".\n\n" + help.scope + " Use --focus for a test name pattern, or put runner arguments after --."
		sub.Example = "  " + sub.CommandPath() + " " + help.example
		bindTestCommandFlags(sub, &flags)
		if err := sub.Flags().MarkHidden("framework"); err != nil {
			panic(fmt.Sprintf("hide --framework on %s subcommand: %v", name, err))
		}
	}
}

// frameworkSubcommandName turns a Framework into the subcommand name. "go
// test" becomes "go" so users type `gavel test go` (space-free subcommands).
func frameworkSubcommandName(fw parsers.Framework) string {
	if fw == parsers.GoTest {
		return "go"
	}
	return string(fw)
}

func init() {
	// Subcommands must register after testCmd exists; test.go's init runs
	// first because files are initialized alphabetically within a package.
	// Rely on lexical ordering (test.go < test_framework_subcmds.go) here.
	registerTestFrameworkSubcommands()
}
