package main

import (
	"fmt"

	"github.com/flanksource/clicky"
	"github.com/flanksource/gavel/testrunner"
	"github.com/flanksource/gavel/testrunner/parsers"
	"github.com/spf13/cobra"
)

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
		fw := fw
		name := frameworkSubcommandName(fw)
		var sub *cobra.Command
		sub = clicky.AddNamedCommand(name, testCmd, testrunner.RunOptions{}, func(opts testrunner.RunOptions) (any, error) {
			if err := splitTestPassThroughArgs(sub, &opts); err != nil {
				return nil, err
			}
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
