package prompt

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("review tools", func() {
	var workDir string
	var run *reviewRun

	BeforeEach(func() {
		workDir = GinkgoT().TempDir()
		outside := GinkgoT().TempDir()
		Expect(os.MkdirAll(filepath.Join(workDir, "pkg"), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(workDir, "pkg", "a.go"), []byte("one\ntwo\nthree\nfour\n"), 0o644)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(outside, "secret.txt"), []byte("secret"), 0o644)).To(Succeed())
		Expect(os.Symlink(filepath.Join(outside, "secret.txt"), filepath.Join(workDir, "link.txt"))).To(Succeed())
		Expect(os.WriteFile(filepath.Join(workDir, "blob.bin"), []byte("a\x00b"), 0o644)).To(Succeed())
		run = &reviewRun{WorkDir: workDir}
	})

	call := func(name string, input map[string]any) (any, error) {
		for _, tool := range run.tools() {
			if tool.Name == name {
				return tool.Handler(context.Background(), input)
			}
		}
		Fail("no tool named " + name)
		return nil, nil
	}

	It("names every tool the agent gets", func() {
		names := []string{}
		for _, tool := range run.tools() {
			names = append(names, tool.Name)
			Expect(tool.InputSchema).NotTo(BeNil(), tool.Name)
		}
		Expect(names).To(Equal([]string{readFileTool, lintTool, reportViolationTool}))
	})

	Describe(readFileTool, func() {
		It("reads a window of lines", func() {
			out, err := call(readFileTool, map[string]any{"path": "pkg/a.go", "offset": 2, "limit": 2})
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(Equal(readFileOutput{Path: "pkg/a.go", StartLine: 2, EndLine: 3, TotalLines: 4, Content: "two\nthree"}))
		})

		DescribeTable("refuses paths outside the work dir or unreadable content",
			func(path, wantErr string) {
				_, err := call(readFileTool, map[string]any{"path": path})
				Expect(err).To(MatchError(ContainSubstring(wantErr)))
			},
			Entry("parent escape", "../etc/passwd", "relative"),
			Entry("absolute path", "/etc/passwd", "relative"),
			Entry("symlink escape", "link.txt", "path escapes"),
			Entry("binary", "blob.bin", "binary"),
			Entry("missing", "pkg/none.go", "no such file"),
		)
	})

	Describe(lintTool, func() {
		It("rejects paths outside the work dir before running anything", func() {
			_, err := call(lintTool, map[string]any{"paths": []any{"../other"}})
			Expect(err).To(MatchError(ContainSubstring("relative")))
		})

		It("requires at least one path", func() {
			_, err := call(lintTool, map[string]any{})
			Expect(err).To(MatchError(ContainSubstring("paths")))
		})
	})

	Describe(reportViolationTool, func() {
		It("records a violation with the file resolved against the work dir", func() {
			_, err := call(reportViolationTool, map[string]any{"file": "pkg/a.go", "line": 3, "severity": "warning", "message": "shadowed err"})
			Expect(err).NotTo(HaveOccurred())
			Expect(run.Reported()).To(Equal([]models.Violation{{
				File: filepath.Join(workDir, "pkg/a.go"), Line: 3, Severity: models.SeverityWarning,
				Message: models.StringPtr("shadowed err"), Source: "prompt",
			}}))
		})

		DescribeTable("rejects an invalid violation",
			func(input map[string]any, wantErr string) {
				_, err := call(reportViolationTool, input)
				Expect(err).To(MatchError(ContainSubstring(wantErr)))
				Expect(run.Reported()).To(BeEmpty())
			},
			Entry("blank message", map[string]any{"file": "pkg/a.go", "message": " "}, "nonempty message"),
			Entry("bad severity", map[string]any{"message": "x", "severity": "fatal"}, "invalid severity"),
			Entry("negative line", map[string]any{"message": "x", "line": -1}, "negative location"),
		)
	})
})

var _ = Describe("mergeViolations", func() {
	msg := func(s string) *string { return &s }
	rule := &models.Rule{Method: "errcheck"}

	It("unions structured and reported violations, dropping duplicates", func() {
		structured := []models.Violation{{File: "/r/a.go", Line: 1, Message: msg("a"), Rule: rule}}
		reported := []models.Violation{
			{File: "/r/a.go", Line: 1, Message: msg("a"), Rule: &models.Rule{Method: "errcheck"}},
			{File: "/r/a.go", Line: 2, Message: msg("b")},
		}
		Expect(mergeViolations(structured, reported)).To(Equal([]models.Violation{
			{File: "/r/a.go", Line: 1, Message: msg("a"), Rule: rule},
			{File: "/r/a.go", Line: 2, Message: msg("b")},
		}))
	})
})

var _ = Describe("tool definitions", func() {
	It("marks the read tools read-only and the report tool as not", func() {
		run := &reviewRun{WorkDir: "/r"}
		readOnly := map[string]bool{}
		for _, tool := range run.tools() {
			readOnly[tool.Name] = tool.ReadOnlyHint != nil && *tool.ReadOnlyHint
			Expect(tool.DefaultPermission).To(Equal(api.ToolPolicyAllow), tool.Name)
			Expect(strings.TrimSpace(tool.Description)).NotTo(BeEmpty())
		}
		Expect(readOnly).To(Equal(map[string]bool{readFileTool: true, lintTool: true, reportViolationTool: false}))
	})
})
