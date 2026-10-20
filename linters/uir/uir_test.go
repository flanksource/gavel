package uir

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	commonsContext "github.com/flanksource/commons/context"
	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestUIR(t *testing.T) { RegisterFailHandler(Fail); RunSpecs(t, "UIR adapter") }

var _ = Describe("UIR adapter protocol", func() {
	It("isolates mutable options and counters for every project invocation", func() {
		original := New("parent")
		first := original.NewProject("first").(*UIR)
		second := original.NewProject("second").(*UIR)
		Expect(first).NotTo(BeIdenticalTo(second))
		first.SetOptions(linters.RunOptions{WorkDir: "first", Executable: "first-uir"})
		Expect(second.WorkDir).To(Equal("second"))
	})
	It("executes the configured binary with JSON flags and preserves nonzero findings", func() {
		root := GinkgoT().TempDir()
		content, err := os.ReadFile("testdata/uir.sh")
		Expect(err).NotTo(HaveOccurred())
		binary := filepath.Join(root, "uir")
		Expect(os.WriteFile(binary, content, 0o700)).To(Succeed())
		GinkgoT().Setenv("UIR_TEST_REPORT", `[{"linter":"uir","success":true,"violations":[{"file":"sample.go","severity":"error"}]}]`)
		GinkgoT().Setenv("UIR_TEST_EXIT", "1")
		adapter := New(root)
		adapter.SetOptions(linters.RunOptions{WorkDir: root, Executable: binary, Files: []string{"sample.go"}})
		command, args := adapter.DryRunCommand()
		Expect(command).To(Equal(binary))
		Expect(args).To(Equal([]string{"lint", "--config", filepath.Join(root, ".uir.yaml"), "--format", "json", "--no-progress"}))
		findings, err := adapter.Run(commonsContext.NewContext(context.Background()), nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(findings).To(HaveLen(1))
		GinkgoT().Setenv("UIR_TEST_REPORT", "not JSON")
		_, err = adapter.Run(commonsContext.NewContext(context.Background()), nil)
		Expect(err).To(MatchError(ContainSubstring("decode lint report")))
	})
	DescribeTable("preserves findings and distinguishes execution errors", func(success bool, severity models.ViolationSeverity, exit int, valid bool) {
		expected := []models.Violation{{File: "sample.go", Severity: severity}}
		reports := []*linters.LinterResult{{Linter: "uir", Success: success, Violations: expected, FileCount: 1, RuleCount: 3}}
		if !success {
			reports[0].Error = "rule message failed"
		}
		data, err := json.Marshal(reports)
		Expect(err).NotTo(HaveOccurred())
		adapter := &UIR{}
		found, err := adapter.parse(data, exit)
		Expect(err == nil).To(Equal(valid))
		Expect(found).To(Equal(expected))
		Expect(adapter.fileCount).To(Equal(1))
		Expect(adapter.ruleCount).To(Equal(3))
	},
		Entry("error findings exit 1", true, models.SeverityError, 1, true),
		Entry("warnings exit 0", true, models.SeverityWarning, 0, true),
		Entry("execution failed", false, models.SeverityWarning, 1, false),
		Entry("unexpected exit", true, models.SeverityWarning, 2, false),
		Entry("missing failure exit", true, models.SeverityError, 0, false),
		Entry("warning failure exit", true, models.SeverityWarning, 1, false),
	)
	DescribeTable("rejects malformed reports", func(output string) {
		_, err := (&UIR{}).parse([]byte(output), 0)
		Expect(err).To(HaveOccurred())
	}, Entry("empty", ""), Entry("null", "null"), Entry("empty array", "[]"), Entry("wrong linter", `[{"linter":"other","success":true}]`), Entry("invalid JSON", "invalid"))
})
