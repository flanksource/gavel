package main

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
)

var _ = Describe("Configured command precedence", func() {
	It("retains configured paths when only native arguments follow the separator", func() {
		cmd := &cobra.Command{Use: "test"}
		Expect(cmd.Flags().Parse([]string{"--", "-count=1"})).To(Succeed())
		fields := configuredFlagFields(cmd)
		Expect(fields).NotTo(HaveKey("paths"))
		Expect(fields).NotTo(HaveKey("files"))
	})
	It("marks positional paths before native arguments as explicit", func() {
		cmd := &cobra.Command{Use: "test"}
		Expect(cmd.Flags().Parse([]string{"./services", "--", "-count=1"})).To(Succeed())
		Expect(configuredFlagFields(cmd)).To(HaveKeyWithValue("paths", true))
	})
	DescribeTable("pins the selected tool over configured command defaults", func(path []string, option string) {
		cmd, _, err := rootCmd.Find(path)
		Expect(err).NotTo(HaveOccurred())
		Expect(configuredFlagFields(cmd)).To(HaveKeyWithValue(option, true))
	}, Entry("Go", []string{"test", "go"}, "framework"), Entry("Ruff", []string{"lint", "ruff"}, "linters"))
})
