//go:build cli_smoke

package commands

import (
	"os"
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Jira ignore-field resync CLI", func() {
	describeIgnoreFieldResync(func(workDir string, args []string) {
		binary := os.Getenv("GAVEL_TEST_BINARY")
		Expect(binary).NotTo(BeEmpty(), "set GAVEL_TEST_BINARY to the binary produced by make build")
		command := exec.Command(binary, append(append([]string{"--cwd", workDir, "external-issues", "jira"}, args...), "--json", "--no-progress")...)
		command.Dir = GinkgoT().TempDir()
		output, err := command.CombinedOutput()
		Expect(err).NotTo(HaveOccurred(), string(output))
	})
})
