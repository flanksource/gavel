package prompt

import (
	"context"
	"os"
	"path/filepath"

	captaincli "github.com/flanksource/captain/pkg/cli"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("file scope", func() {
	const fileA, fileB, dir = "pkg/a.go", "pkg/b.go", "cmd"
	var workDir string

	BeforeEach(func() {
		workDir = GinkgoT().TempDir()
		writeFiles(workDir, map[string]string{fileA: "package pkg\n", fileB: "package pkg\n", dir + "/main.go": "package main\n"})
	})

	reportSchema := func(run *reviewRun) map[string]any {
		for _, tool := range run.tools() {
			if tool.Name == reportViolationTool {
				return tool.InputSchema["properties"].(map[string]any)["violations"].(map[string]any)["items"].(map[string]any)
			}
		}
		Fail("no " + reportViolationTool + " tool")
		return nil
	}
	scopeOf := func(paths ...string) fileScope {
		scope, err := newFileScope(workDir, paths)
		Expect(err).NotTo(HaveOccurred())
		return scope
	}

	It("enumerates the files under review as report_violation's required file", func() {
		items := reportSchema(&reviewRun{WorkDir: workDir, Scope: scopeOf(fileA, fileB)})
		Expect(items["properties"].(map[string]any)["file"]).To(HaveKeyWithValue("enum", []string{fileA, fileB}))
		Expect(items["required"]).To(ConsistOf("message", "file"))
	})

	DescribeTable("leaves file a free string when it cannot be enumerated", func(paths ...string) {
		items := reportSchema(&reviewRun{WorkDir: workDir, Scope: scopeOf(paths...)})
		Expect(items["properties"].(map[string]any)["file"]).NotTo(HaveKey("enum"))
		Expect(items["required"]).To(ConsistOf("message"))
	},
		Entry("a directory in scope", fileA, dir),
		Entry("no path scope"),
	)

	It("rejects a reported violation for a file that is not under review and records the rest", func() {
		run := &reviewRun{WorkDir: workDir, Scope: scopeOf(fileA)}
		out, err := run.reportViolation(context.Background(), map[string]any{"violations": []any{
			map[string]any{"file": fileB, "line": 1, "message": "elsewhere"},
			map[string]any{"file": fileA, "line": 1, "message": "here"},
		}})
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(Equal(reportOutput{Recorded: 1, Reported: 1, Rejected: []rejectedViolation{{
			Index: 0, Error: "file " + filepath.Join(workDir, fileB) + " is not under review; want one of " + fileA,
		}}}))
	})

	DescribeTable("checks a final answer's files against the scope", func(paths []string, file string, ok bool) {
		got := (&reviewRun{WorkDir: workDir, Scope: scopeOf(paths...)}).result(summaryInput{Summary: captaincli.PromptRunSummary{
			Success: true, StructuredOutput: map[string]any{"violations": []any{map[string]any{"file": file, "message": "m"}}},
		}})
		Expect(got.Success).To(Equal(ok), got.Error)
	},
		Entry("a listed file", []string{fileA, fileB}, fileB, true),
		Entry("an unlisted file", []string{fileA}, fileB, false),
		Entry("no file while files are under review", []string{fileA}, "", false),
		Entry("a file under a directory in scope", []string{dir}, dir+"/main.go", true),
		Entry("a file outside every directory in scope", []string{dir}, fileA, false),
		Entry("any file without a path scope", nil, fileA, true),
	)

	It("keeps a path that no longer exists, as a deleted changed file, enumerable", func() {
		Expect(os.Remove(filepath.Join(workDir, fileB))).To(Succeed())
		Expect(scopeOf(fileA, fileB)).To(Equal(fileScope{Paths: []string{fileA, fileB}, Exact: true}))
	})
})
