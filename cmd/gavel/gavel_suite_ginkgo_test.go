package main

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
)

func TestGavelCLI(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Gavel CLI Suite")
}

// Specs here load .gavel.yaml, which layers ~/.gavel.yaml under the workspace's,
// so without an isolated HOME they assert against the developer's own config.
var _ = BeforeEach(func() {
	GinkgoT().Setenv("HOME", GinkgoT().TempDir())
})

// withRootContext gives cmd the context the root command resolves from --cwd,
// as it does before any subcommand runs.
func withRootContext(cmd *cobra.Command) *cobra.Command {
	cmd.SetContext(context.Background())
	Expect(rootCmd.PersistentPreRunE(cmd, nil)).To(Succeed())
	return cmd
}
