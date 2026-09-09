package entity

import (
	"testing"

	"github.com/flanksource/clicky"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
)

func TestEntityRuntimeProfile(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Entity Runtime Profile Suite")
}

var _ = Describe("Generated TODO runtime-profile flags", func() {
	DescribeTable("publishes runtime-profile on run-shaped bulk commands", func(action string) {
		root := &cobra.Command{Use: "gavel"}
		clicky.GenerateCLI(root)
		cmd, _, err := root.Find([]string{"todo", action})
		Expect(err).NotTo(HaveOccurred())
		Expect(cmd.Name()).To(Equal(action))
		flag := cmd.Flags().Lookup("runtime-profile")
		Expect(flag).NotTo(BeNil())
		Expect(flag.DefValue).To(BeEmpty())
	}, Entry("run", "run"), Entry("plan", "plan"), Entry("triage", "triage"))
})
