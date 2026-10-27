package prompt

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/flanksource/captain/pkg/api"
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

	result := func(input summaryInput) *linters.LinterResult {
		return (&reviewRun{WorkDir: "/workspace"}).result(input)
	}
	succeeded := func(output map[string]any) captaincli.PromptRunSummary {
		return captaincli.PromptRunSummary{Success: true, StructuredOutput: output}
	}

	It("passes an explicitly empty violations array", func() {
		got := result(summaryInput{Name: "proposal-review", Summary: succeeded(map[string]any{"violations": []any{}})})
		Expect(got).To(SatisfyAll(HaveField("Linter", "proposal-review"), HaveField("Success", true), HaveField("Violations", BeEmpty())))
		Expect(linters.HasFailures([]*linters.LinterResult{got})).To(BeFalse())
	})

	It("reports cancellation as a failed lint execution", func() {
		got := result(summaryInput{Err: context.Canceled})
		Expect(got).To(SatisfyAll(HaveField("Success", false), HaveField("Error", context.Canceled.Error())))
	})

	It("reports a run that did not succeed with Captain's reason", func() {
		got := result(summaryInput{Summary: captaincli.PromptRunSummary{Error: "verification failed"}})
		Expect(got).To(SatisfyAll(HaveField("Success", false), HaveField("Error", "verification failed")))
	})

	DescribeTable("keeps standard lint verdicts", func(severity string, fails bool) {
		got := result(summaryInput{Summary: succeeded(map[string]any{"violations": []any{
			map[string]any{"file": "pkg/service.go", "line": 12, "message": "Return the failure to the caller", "severity": severity},
		}})})
		Expect(got.Success).To(BeTrue())
		Expect(got.Violations).To(HaveLen(1))
		Expect(got.Violations[0].File).To(Equal("/workspace/pkg/service.go"))
		Expect(linters.HasFailures([]*linters.LinterResult{got})).To(Equal(fails))
	}, Entry("error", "error", true), Entry("warning", "warning", false), Entry("info", "info", false), Entry("unspecified", "", true))

	DescribeTable("rejects invalid output", func(output map[string]any) {
		got := result(summaryInput{Summary: succeeded(output)})
		Expect(got.Success).To(BeFalse())
		Expect(got.Error).NotTo(BeEmpty())
	}, Entry("missing", nil), Entry("missing violations", map[string]any{}),
		Entry("null", map[string]any{"violations": nil}),
		Entry("blank message", map[string]any{"violations": []any{map[string]any{"message": " "}}}),
		Entry("unknown severity", map[string]any{"violations": []any{map[string]any{"message": "Failure", "severity": "critical"}}}))

	It("keeps violations reported through the tool when the run fails", func() {
		run := &reviewRun{WorkDir: "/workspace"}
		_, err := run.reportViolation(context.Background(), map[string]any{"violations": []any{map[string]any{"file": "a.go", "line": 1, "message": "leak"}}})
		Expect(err).NotTo(HaveOccurred())
		got := run.result(summaryInput{Err: errors.New("provider unavailable")})
		Expect(got.Success).To(BeFalse())
		Expect(got.Violations).To(HaveExactElements(HaveField("File", "/workspace/a.go")))
	})

	DescribeTable("gives each path its own run with that path as the files variable", func(paths, vars []string, wantNames []string, wantVars [][]string) {
		opts := Options{Prompt: captaincli.AIPromptOptions{File: "review.prompt", Var: vars}}
		reviews := opts.reviews(paths)
		names, gotVars := []string{}, [][]string{}
		for _, review := range reviews {
			names, gotVars = append(names, review.Name), append(gotVars, review.Prompt.Var)
		}
		Expect(names).To(Equal(wantNames))
		Expect(gotVars).To(Equal(wantVars))
		Expect(opts.Prompt.Var).To(Equal(vars), "the shared options are not mutated")
	},
		Entry("no scope", nil, []string{"focus=errors"}, []string{"review"}, [][]string{{"focus=errors"}}),
		Entry("paths", []string{"pkg", "cmd/x"}, []string{"focus=errors"}, []string{"review:pkg", "review:cmd/x"},
			[][]string{{"focus=errors", "files=pkg"}, {"focus=errors", "files=cmd/x"}}),
	)

	DescribeTable("rejects a files variable that conflicts with path arguments", func(vars []string, jsonVars string) {
		dir := GinkgoT().TempDir()
		opts := Options{Paths: []string{dir}, Prompt: captaincli.AIPromptOptions{Var: vars, Vars: jsonVars}}
		_, _, err := opts.scope(context.Background(), dir)
		Expect(err).To(MatchError(ContainSubstring("files")))
	},
		Entry("--var", []string{"files=other.md"}, ""),
		Entry("--vars", nil, `{"files":"other.md"}`),
	)

	It("rejects a path that does not exist", func() {
		dir := GinkgoT().TempDir()
		_, _, err := (Options{Paths: []string{"missing"}}).scope(context.Background(), dir)
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
			paths, skip, err := (Options{Changed: true, Paths: []string{"pkg"}}).scope(context.Background(), dir)
			Expect(err).NotTo(HaveOccurred())
			Expect(skip).To(BeFalse())
			Expect(paths).To(Equal([]string{"pkg/a.go"}))
		})

		It("skips the prompt when nothing changed", func() {
			_, skip, err := (Options{Changed: true}).scope(context.Background(), dir)
			Expect(err).NotTo(HaveOccurred())
			Expect(skip).To(BeTrue())
		})
	})
})

// recordingPublisher records every publish as "<linter>=<status>" in order.
type recordingPublisher struct {
	mu     sync.Mutex
	events []string
	latest map[string]*linters.LinterResult
}

func (p *recordingPublisher) UpsertLintResult(result *linters.LinterResult) {
	p.mu.Lock()
	defer p.mu.Unlock()
	status := string(result.Status)
	if status == "" {
		status = "final"
	}
	p.events = append(p.events, result.Linter+"="+status)
	if p.latest == nil {
		p.latest = map[string]*linters.LinterResult{}
	}
	p.latest[result.Linter] = result
}

func (p *recordingPublisher) RemoveLintResult(linter, _ string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.events = append(p.events, linter+"=removed")
	delete(p.latest, linter)
}

func (p *recordingPublisher) Events() []string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return slices.Clone(p.events)
}

var _ = Describe("runReviews", func() {
	var (
		publisher *recordingPublisher
		launched  []captaincli.PromptInvocation
		mu        sync.Mutex
		stopped   []string
	)

	BeforeEach(func() {
		publisher, launched, stopped = &recordingPublisher{}, nil, nil
		DeferCleanup(func(l, a, s any) {
			launchPrompt = l.(func(context.Context, captaincli.PromptInvocation) (captaincli.PromptRunResult, error))
			awaitPromptRun = a.(func(context.Context, string) (captaincli.PromptRunSummary, error))
			stopPromptRun = s.(func(string) error)
		}, launchPrompt, awaitPromptRun, stopPromptRun)
		launchPrompt = func(_ context.Context, inv captaincli.PromptInvocation) (captaincli.PromptRunResult, error) {
			mu.Lock()
			defer mu.Unlock()
			launched = append(launched, inv)
			files := inv.Options.Var[len(inv.Options.Var)-1]
			return captaincli.PromptRunResult{RunID: "run-" + strings.TrimPrefix(files, "files=")}, nil
		}
		awaitPromptRun = func(_ context.Context, runID string) (captaincli.PromptRunSummary, error) {
			return captaincli.PromptRunSummary{RunID: runID, Success: true, StructuredOutput: map[string]any{"violations": []any{
				map[string]any{"file": "a.go", "message": "from " + runID},
			}}}, nil
		}
		stopPromptRun = func(runID string) error {
			mu.Lock()
			defer mu.Unlock()
			stopped = append(stopped, runID)
			return nil
		}
	})

	reviewsFor := func(paths ...string) []review {
		return (Options{Prompt: captaincli.AIPromptOptions{File: "review.prompt"}}).reviews(paths)
	}

	It("publishes every run as pending before any launches, then running and final per run", func() {
		results := runReviews(context.Background(), reviewsInput{Reviews: reviewsFor("a.go", "b.go"), WorkDir: "/w", Concurrency: 1, Publisher: publisher})

		Expect(publisher.Events()).To(Equal([]string{
			"review:a.go=pending", "review:b.go=pending",
			"review:a.go=running", "review:a.go=final",
			"review:b.go=running", "review:b.go=final",
		}))
		Expect(results).To(HaveExactElements(
			SatisfyAll(HaveField("Linter", "review:a.go"), HaveField("RunID", "run-a.go"), HaveField("Success", true), HaveField("Violations", HaveExactElements(HaveField("File", "/w/a.go")))),
			SatisfyAll(HaveField("Linter", "review:b.go"), HaveField("RunID", "run-b.go"), HaveField("Success", true)),
		))
		Expect(launched).To(HaveLen(2))
		for _, inv := range launched {
			Expect(inv.Options.RequiredPresets).To(ContainElement("lint"))
			Expect(inv.Tools).To(HaveLen(3))
		}
	})

	It("stops the Captain run when the caller is cancelled", func() {
		ctx, cancel := context.WithCancel(context.Background())
		awaitPromptRun = func(ctx context.Context, _ string) (captaincli.PromptRunSummary, error) {
			cancel()
			<-ctx.Done()
			return captaincli.PromptRunSummary{}, ctx.Err()
		}

		results := runReviews(ctx, reviewsInput{Reviews: reviewsFor("a.go"), WorkDir: "/w", Concurrency: 1})

		Expect(stopped).To(Equal([]string{"run-a.go"}))
		Expect(results).To(HaveExactElements(SatisfyAll(HaveField("Success", false), HaveField("Error", context.Canceled.Error()))))
	})

	It("forwards rejected response events and keeps the terminal schema failure", func() {
		rejected := api.Event{Kind: api.EventText, Text: "# Findings\n\n- Define a baseline.", Rejected: true}
		failure := api.Event{Kind: api.EventError, Error: "schema validation failed: expected JSON"}
		var observed []api.Event
		launch := launchPrompt
		launchPrompt = func(ctx context.Context, inv captaincli.PromptInvocation) (captaincli.PromptRunResult, error) {
			inv.OnEvent(0, rejected)
			inv.OnEvent(0, failure)
			return launch(ctx, inv)
		}
		awaitPromptRun = func(context.Context, string) (captaincli.PromptRunSummary, error) {
			return captaincli.PromptRunSummary{Text: rejected.Text, Error: failure.Error}, errors.New(failure.Error)
		}
		results := runReviews(context.Background(), reviewsInput{Reviews: reviewsFor("a.go"), WorkDir: "/w", Concurrency: 1,
			OnEvent: func(_ int, event api.Event) { observed = append(observed, event) },
		})
		Expect(observed).To(Equal([]api.Event{rejected, failure}))
		Expect(results).To(HaveExactElements(And(HaveField("Success", false), HaveField("Error", failure.Error))))
	})

	It("reports a launch failure as that run's error", func() {
		launchPrompt = func(context.Context, captaincli.PromptInvocation) (captaincli.PromptRunResult, error) {
			return captaincli.PromptRunResult{}, fmt.Errorf("runtime cannot expose caller tools")
		}

		results := runReviews(context.Background(), reviewsInput{Reviews: reviewsFor("a.go"), WorkDir: "/w", Concurrency: 1, Publisher: publisher})

		Expect(results).To(HaveExactElements(HaveField("Error", "runtime cannot expose caller tools")))
		Expect(publisher.Events()).To(Equal([]string{"review:a.go=pending", "review:a.go=final"}))
	})

	It("replaces the pending entry with one result per model for a multi-model prompt", func() {
		launchPrompt = func(context.Context, captaincli.PromptInvocation) (captaincli.PromptRunResult, error) {
			return captaincli.PromptRunResult{Runs: []captaincli.PromptRunItem{{RunID: "r1", Selector: "api:sonnet"}, {RunID: "r2", Selector: "api:haiku"}}}, nil
		}

		results := runReviews(context.Background(), reviewsInput{Reviews: reviewsFor("a.go"), WorkDir: "/w", Concurrency: 1, Publisher: publisher})

		Expect(results).To(HaveExactElements(HaveField("Linter", "review:a.go:api:sonnet"), HaveField("Linter", "review:a.go:api:haiku")))
		Expect(publisher.Events()[:4]).To(Equal([]string{"review:a.go=pending", "review:a.go=removed", "review:a.go:api:sonnet=running", "review:a.go:api:haiku=running"}))
		Expect(publisher.latest).To(HaveLen(2))
	})

	It("publishes violations reported through the tool while the run is still going", func() {
		awaitPromptRun = func(_ context.Context, runID string) (captaincli.PromptRunSummary, error) {
			mu.Lock()
			tools := launched[0].Tools
			mu.Unlock()
			for _, tool := range tools {
				if tool.Name == reportViolationTool {
					_, err := tool.Handler(context.Background(), map[string]any{"violations": []any{map[string]any{"file": "b.go", "line": 2, "message": "early"}}})
					Expect(err).NotTo(HaveOccurred())
				}
			}
			publisher.mu.Lock()
			live := publisher.latest["review:a.go"]
			publisher.mu.Unlock()
			Expect(live).To(SatisfyAll(HaveField("Status", linters.LinterStatusRunning), HaveField("Violations", HaveExactElements(HaveField("File", "/w/b.go")))))
			return captaincli.PromptRunSummary{RunID: runID, Success: true, StructuredOutput: map[string]any{"violations": []any{}}}, nil
		}

		results := runReviews(context.Background(), reviewsInput{Reviews: reviewsFor("a.go"), WorkDir: "/w", Concurrency: 1, Publisher: publisher})

		Expect(results).To(HaveExactElements(HaveField("Violations", HaveExactElements(HaveField("File", "/w/b.go")))))
	})
})
