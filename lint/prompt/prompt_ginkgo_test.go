package prompt

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	captaincli "github.com/flanksource/captain/pkg/cli"
	"github.com/flanksource/gavel/linters"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestPrompt(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Prompt lint")
}

var _ = Describe("Prompt lint results", func() {
	DescribeTable("uses the prompt reference as the lint name", func(file, ref, want string) {
		opts := Options{Prompt: captaincli.AIPromptOptions{File: file, PromptRef: ref}}
		Expect(opts.name()).To(Equal(want))
	},
		Entry("named prompt", "proposal-review", "", "proposal-review"),
		Entry("prompt file", "reviews/proposal-review.prompt", "", "proposal-review"),
		Entry("file flag", "", "@reviews/proposal-review.prompt", "proposal-review"),
		Entry("positional reference takes precedence", "proposal-review", "@other.prompt", "proposal-review"),
		Entry("stdin", "-", "", "prompt"),
		Entry("inline", "", "Review the change", "prompt"),
	)

	It("labels findings with the requested prompt name", func() {
		results := resultsFromRun(resultInput{Name: "proposal-review", Run: captaincli.PromptRunResult{StructuredOutput: map[string]any{"violations": []any{}}}})
		Expect(results).To(HaveExactElements(HaveField("Linter", "proposal-review")))
	})

	It("includes the prompt name in each model result and invocation failure", func() {
		results := resultsFromRun(resultInput{Name: "proposal-review", Error: errors.New("recording failed"), Run: captaincli.PromptRunResult{Runs: []captaincli.PromptRunItem{
			{Selector: "api:reviewer", StructuredOutput: map[string]any{"violations": []any{}}},
		}}})
		Expect(results).To(HaveExactElements(HaveField("Linter", "proposal-review:api:reviewer"), HaveField("Linter", "proposal-review")))
	})

	It("passes an explicitly empty violations array", func() {
		results := resultsFromRun(resultInput{Run: captaincli.PromptRunResult{StructuredOutput: map[string]any{"violations": []any{}}}})
		Expect(results).To(HaveExactElements(SatisfyAll(HaveField("Success", true), HaveField("Violations", BeEmpty()))))
		Expect(linters.HasFailures(results)).To(BeFalse())
	})

	It("retains an invocation error even when a completed model reported an error finding", func() {
		results := resultsFromRun(resultInput{Error: errors.New("recording failed"), Run: captaincli.PromptRunResult{Runs: []captaincli.PromptRunItem{
			{Selector: "api:reviewer", StructuredOutput: map[string]any{"violations": []any{map[string]any{"message": "Missing validation"}}}},
		}}})
		Expect(results).To(HaveLen(2))
		Expect(results[1].Error).To(Equal("recording failed"))
	})

	It("reports cancellation as a failed lint execution", func() {
		results := resultsFromRun(resultInput{Error: context.Canceled})
		Expect(results).To(HaveExactElements(SatisfyAll(HaveField("Success", false), HaveField("Error", context.Canceled.Error()))))
	})

	DescribeTable("keeps standard lint verdicts", func(severity string, fails bool) {
		result := captaincli.PromptRunResult{StructuredOutput: map[string]any{"violations": []any{
			map[string]any{"file": "pkg/service.go", "line": 12, "message": "Return the failure to the caller", "severity": severity},
		}}}
		results := resultsFromRun(resultInput{Run: result, WorkDir: "/workspace"})
		Expect(results).To(HaveLen(1))
		Expect(results[0].Success).To(BeTrue())
		Expect(results[0].Violations).To(HaveLen(1))
		Expect(results[0].Violations[0].File).To(Equal("/workspace/pkg/service.go"))
		Expect(linters.HasFailures(results)).To(Equal(fails))
	}, Entry("error", "error", true), Entry("warning", "warning", false), Entry("info", "info", false), Entry("unspecified", "", true))

	DescribeTable("rejects invalid output", func(output map[string]any) {
		results := resultsFromRun(resultInput{Run: captaincli.PromptRunResult{StructuredOutput: output}})
		Expect(results).To(HaveLen(1))
		Expect(results[0].Success).To(BeFalse())
		Expect(results[0].Error).NotTo(BeEmpty())
	}, Entry("missing", nil), Entry("missing violations", map[string]any{}),
		Entry("null", map[string]any{"violations": nil}),
		Entry("blank message", map[string]any{"violations": []any{map[string]any{"message": " "}}}),
		Entry("unknown severity", map[string]any{"violations": []any{map[string]any{"message": "Failure", "severity": "critical"}}}))

	DescribeTable("injects the review scope as the files variable", func(paths, vars []string, jsonVars string, want []string) {
		dir := GinkgoT().TempDir()
		for _, path := range paths {
			Expect(os.MkdirAll(filepath.Join(dir, path), 0o755)).To(Succeed())
		}
		opts := Options{Paths: paths, Prompt: captaincli.AIPromptOptions{Var: vars, Vars: jsonVars}}
		scoped, skip, err := opts.scoped(context.Background(), dir)
		Expect(err).NotTo(HaveOccurred())
		Expect(skip).To(BeFalse())
		Expect(scoped.Prompt.Var).To(Equal(want))
	},
		Entry("no scope", nil, []string{"focus=errors"}, "", []string{"focus=errors"}),
		Entry("paths", []string{"pkg", "cmd/x"}, []string{"focus=errors"}, `{"rfp":"a.md"}`, []string{"focus=errors", "files=pkg cmd/x"}),
	)

	DescribeTable("rejects a files variable that conflicts with path arguments", func(vars []string, jsonVars string) {
		dir := GinkgoT().TempDir()
		opts := Options{Paths: []string{dir}, Prompt: captaincli.AIPromptOptions{Var: vars, Vars: jsonVars}}
		_, _, err := opts.scoped(context.Background(), dir)
		Expect(err).To(MatchError(ContainSubstring("files")))
	},
		Entry("--var", []string{"files=other.md"}, ""),
		Entry("--vars", nil, `{"files":"other.md"}`),
	)

	It("rejects a path that does not exist", func() {
		dir := GinkgoT().TempDir()
		_, _, err := (Options{Paths: []string{"missing"}}).scoped(context.Background(), dir)
		Expect(err).To(MatchError(ContainSubstring("missing")))
	})

	Describe("change selection", func() {
		var dir string
		BeforeEach(func() {
			dir = GinkgoT().TempDir()
			git := func(args ...string) {
				cmd := exec.Command("git", args...)
				cmd.Dir = dir
				out, err := cmd.CombinedOutput()
				Expect(err).NotTo(HaveOccurred(), string(out))
			}
			git("init", "-q")
			git("config", "user.email", "test@example.com")
			git("config", "user.name", "Test")
			for _, path := range []string{"pkg/a.go", "pkg/b.go", "cmd/main.go"} {
				Expect(os.MkdirAll(filepath.Dir(filepath.Join(dir, path)), 0o755)).To(Succeed())
				Expect(os.WriteFile(filepath.Join(dir, path), []byte("package x\n"), 0o600)).To(Succeed())
			}
			git("add", ".")
			git("-c", "commit.gpgsign=false", "commit", "-q", "-m", "init")
		})

		It("narrows changed files to the path arguments", func() {
			Expect(os.WriteFile(filepath.Join(dir, "pkg/a.go"), []byte("package y\n"), 0o600)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(dir, "cmd/main.go"), []byte("package y\n"), 0o600)).To(Succeed())
			scoped, skip, err := (Options{Changed: true, Paths: []string{"pkg"}}).scoped(context.Background(), dir)
			Expect(err).NotTo(HaveOccurred())
			Expect(skip).To(BeFalse())
			Expect(scoped.Prompt.Var).To(Equal([]string{"files=pkg/a.go"}))
		})

		It("skips the prompt when nothing changed", func() {
			_, skip, err := (Options{Changed: true}).scoped(context.Background(), dir)
			Expect(err).NotTo(HaveOccurred())
			Expect(skip).To(BeTrue())
		})
	})

	It("preserves findings and failure when one model fails", func() {
		results := resultsFromRun(resultInput{Run: captaincli.PromptRunResult{Runs: []captaincli.PromptRunItem{
			{Selector: "api:reviewer", StructuredOutput: map[string]any{"violations": []any{map[string]any{"message": "Missing validation"}}}},
			{Selector: "api:critic", Status: "failed", Error: "provider unavailable"},
		}}})
		Expect(results).To(HaveLen(2))
		Expect(results[0].Violations).To(HaveLen(1))
		Expect(results[1].Error).To(Equal("provider unavailable"))
		Expect(linters.HasFailures(results)).To(BeTrue())
	})
})
