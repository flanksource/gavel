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
		report := func(violations ...map[string]any) (any, error) {
			items := make([]any, len(violations))
			for i, v := range violations {
				items[i] = v
			}
			return call(reportViolationTool, map[string]any{"violations": items})
		}

		It("records a violation with the file resolved against the work dir", func() {
			_, err := report(map[string]any{"file": "pkg/a.go", "line": 3, "severity": "warning", "message": "shadowed err"})
			Expect(err).NotTo(HaveOccurred())
			Expect(run.Reported()).To(Equal([]models.Violation{{
				File: filepath.Join(workDir, "pkg/a.go"), Line: 3, Severity: models.SeverityWarning,
				Message: models.StringPtr("shadowed err"), Source: "prompt",
			}}))
		})

		It("records several violations in one call", func() {
			out, err := report(
				map[string]any{"file": "pkg/a.go", "line": 1, "message": "first"},
				map[string]any{"file": "pkg/a.go", "line": 2, "message": "second"},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(Equal(reportOutput{Recorded: 2, Reported: 2}))
			Expect(run.Reported()).To(HaveLen(2))
		})

		It("keeps the quoted code and fills the columns from where it sits on the line", func() {
			_, err := report(map[string]any{"file": "pkg/a.go", "line": 3, "message": "typo", "code": "hre"})
			Expect(err).NotTo(HaveOccurred())
			Expect(run.Reported()).To(Equal([]models.Violation{{
				File: filepath.Join(workDir, "pkg/a.go"), Line: 3, Column: 2, EndColumn: 5,
				Message: models.StringPtr("typo"), Code: models.StringPtr("hre"), Source: "prompt",
			}}))
		})

		It("keeps the columns the model gave", func() {
			_, err := report(map[string]any{"file": "pkg/a.go", "line": 3, "column": 1, "end_column": 6, "message": "word", "code": "hre"})
			Expect(err).NotTo(HaveOccurred())
			Expect(run.Reported()[0].Column).To(Equal(1))
			Expect(run.Reported()[0].EndColumn).To(Equal(6))
		})

		It("accepts code spanning line to end_line", func() {
			_, err := report(map[string]any{"file": "pkg/a.go", "line": 2, "end_line": 3, "message": "pair", "code": "two\nthree"})
			Expect(err).NotTo(HaveOccurred())
			Expect(run.Reported()).To(HaveLen(1))
		})

		It("rejects code that is not on the stated line and says where it is", func() {
			out, err := report(
				map[string]any{"file": "pkg/a.go", "line": 1, "message": "misplaced", "code": "four"},
				map[string]any{"file": "pkg/a.go", "line": 2, "message": "ok", "code": "two"},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(Equal(reportOutput{Recorded: 1, Reported: 1, Rejected: []rejectedViolation{{
				Index: 0, Error: `code "four" is not on line 1 of pkg/a.go; it appears on line 4`,
			}}}))
			Expect(*run.Reported()[0].Message).To(Equal("ok"))
		})

		It("fails the call when every violation is rejected", func() {
			_, err := report(map[string]any{"file": "pkg/a.go", "line": 1, "message": "invented", "code": "nowhere"})
			Expect(err).To(MatchError(ContainSubstring(`code "nowhere" does not appear in pkg/a.go`)))
			Expect(run.Reported()).To(BeEmpty())
		})

		DescribeTable("rejects an invalid violation",
			func(input map[string]any, wantErr string) {
				_, err := report(input)
				Expect(err).To(MatchError(ContainSubstring(wantErr)))
				Expect(run.Reported()).To(BeEmpty())
			},
			Entry("blank message", map[string]any{"file": "pkg/a.go", "message": " "}, "nonempty message"),
			Entry("bad severity", map[string]any{"message": "x", "severity": "fatal"}, "invalid severity"),
			Entry("negative line", map[string]any{"message": "x", "line": -1}, "negative location"),
			Entry("code without a file and line", map[string]any{"message": "x", "code": "two"}, "code requires file and line"),
		)

		It("requires at least one violation", func() {
			_, err := call(reportViolationTool, map[string]any{"violations": []any{}})
			Expect(err).To(MatchError(ContainSubstring("at least one violation")))
		})
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

	It("keeps the tool-reported copy of a duplicate, which carries the checked location", func() {
		structured := []models.Violation{{File: "/r/a.go", Line: 3, Message: msg("typo"), Code: msg("hre")}}
		reported := []models.Violation{{File: "/r/a.go", Line: 3, Column: 2, EndColumn: 5, Message: msg("typo"), Code: msg("hre")}}
		Expect(mergeViolations(structured, reported)).To(Equal(reported))
	})
})

var _ = Describe("decodeViolations", func() {
	var workDir string

	BeforeEach(func() {
		workDir = GinkgoT().TempDir()
		Expect(os.WriteFile(filepath.Join(workDir, "a.md"), []byte("one\ntwo three\n"), 0o644)).To(Succeed())
	})

	answer := func(code string) map[string]any {
		return map[string]any{"violations": []any{map[string]any{"file": "a.md", "line": 2, "message": "m", "code": code}}}
	}

	It("fills the columns of a final-answer violation from its quoted code", func() {
		violations, err := decodeViolations(answer("three"), workDir)
		Expect(err).NotTo(HaveOccurred())
		Expect(violations).To(HaveLen(1))
		Expect([]int{violations[0].Column, violations[0].EndColumn}).To(Equal([]int{5, 10}))
	})

	It("fails a final answer whose quoted code is not in the file", func() {
		_, err := decodeViolations(answer("four"), workDir)
		Expect(err).To(MatchError(ContainSubstring(`prompt lint violation 1: code "four" does not appear in a.md`)))
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
