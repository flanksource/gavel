package main

import (
	"fmt"

	"github.com/flanksource/clicky"
	"github.com/flanksource/gavel/testrunner"
	"github.com/flanksource/gavel/testrunner/parsers"
)

// registerTestFrameworkSubcommands wires `gavel test <framework>` commands.
// Each subcommand mirrors `gavel test`'s flag surface (clicky re-binds them
// against a fresh RunOptions) and pins Frameworks before delegating to
// runTests. Positional args pass through as starting paths, matching the
// parent command.
func registerTestFrameworkSubcommands() {
	if testCmd == nil {
		panic("testCmd must be initialized before registering framework subcommands")
	}
	for _, fw := range parsers.AllFrameworks {
		name := frameworkSubcommandName(fw)
		sub := clicky.AddNamedCommand(name, testCmd, testrunner.RunOptions{}, func(opts testrunner.RunOptions) (any, error) {
			opts.Frameworks = []string{string(fw)}
			return runTests(opts)
		})
		sub.Short = fmt.Sprintf("Run only %s tests", fw)
		sub.Flags().SetInterspersed(true)
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
