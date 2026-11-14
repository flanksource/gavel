package fixtures

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/flanksource/clicky/task"
	flanksourceContext "github.com/flanksource/commons/context"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flanksource/gavel/models"
)

var _ = Describe("a step's recorded evidence", func() {
	const (
		printed  = "listing the looping groups"
		warning  = "Warning: Null value is eliminated by an aggregate"
		command  = "SELECT TOP (@p1) guid FROM activities"
		workDir  = "/src/app"
		failure  = "mssql: The number of rows provided for a TOP clause must be an integer"
		template = "SELECT TOP ($(.params.max)) guid FROM activities"
	)
	record := map[string]any{"template": template, "arguments": []any{nil}}

	It("survives an Execute that fails: stdout, stderr, details and command land on the errored result", func() {
		useKind(fakeKind{name: "recorder", execute: func(_ context.Context, step *StepContext) (StepOutcome, error) {
			fmt.Fprintln(step.Stdout(), printed)
			fmt.Fprintln(step.Stderr(), warning)
			step.SetCommand(command, workDir)
			Expect(step.SetDetail(QueryMetadataKey, record)).To(Succeed())
			return StepOutcome{}, errors.New(failure)
		}})

		result := runSteps(StepSpec{ID: "resume", Kind: "recorder"})[0]

		Expect(result.Status).To(Equal(task.StatusERR))
		Expect(result.Error).To(Equal(failure))
		Expect([]string{result.Stdout, result.Stderr, result.Command, result.CWD}).To(Equal([]string{printed + "\n", warning + "\n", command, workDir}))
		Expect(metadataJSON(result)[QueryMetadataKey]).To(Equal(map[string]any{"template": template, "arguments": []any{nil}}))
		Expect(result.Duration).To(BeNumerically(">", 0))
	})

	It("leaves what the kind set on its result as the kind set it", func() {
		useKind(fakeKind{name: "recorder", execute: func(_ context.Context, step *StepContext) (StepOutcome, error) {
			fmt.Fprint(step.Stdout(), "streamed")
			step.SetCommand("recorded", "")
			Expect(step.SetDetail("view", "recorded view")).To(Succeed())
			return StepOutcome{Result: FixtureResult{Status: task.StatusPASS, Stdout: "kind's own", Command: "kind's command",
				Metadata: map[string]any{"view": "kind's view"}}}, nil
		}})

		result := runSteps(StepSpec{ID: "inspect", Kind: "recorder"})[0]

		Expect([]string{result.Stdout, result.Command}).To(Equal([]string{"kind's own", "kind's command"}))
		Expect(result.Metadata["view"]).To(Equal("kind's view"))
	})

	It("starts each retried attempt with a clean record", func() {
		attempt := 0
		useKind(fakeKind{name: "flaky", execute: func(_ context.Context, step *StepContext) (StepOutcome, error) {
			attempt++
			fmt.Fprintf(step.Stdout(), "attempt %d\n", attempt)
			if attempt == 1 {
				Expect(step.SetDetail("first", true)).To(Succeed())
				return StepOutcome{}, errors.New("connection refused")
			}
			return StepOutcome{Result: FixtureResult{Status: task.StatusPASS}}, nil
		}})

		result := runSteps(StepSpec{ID: "flaky", Kind: "flaky",
			Retry: &RetryPolicy{Attempts: 2, When: "error.contains('refused')", Delay: dur(time.Millisecond)}})[0]

		Expect(result.Status).To(Equal(task.StatusPASS), result.Error)
		Expect(result.Stdout).To(Equal("attempt 2\n"))
		Expect(result.Metadata).NotTo(HaveKey("first"))
	})

	It("streams what it records to the live progress node while the step runs", func() {
		var mu sync.Mutex
		var outputs []string
		var details []map[string]any
		release := make(chan struct{})
		useKind(fakeKind{name: "recorder", execute: func(_ context.Context, step *StepContext) (StepOutcome, error) {
			fmt.Fprintln(step.Stdout(), printed)
			Expect(step.SetDetail(QueryMetadataKey, record)).To(Succeed())
			<-release
			return StepOutcome{}, errors.New(failure)
		}})
		sink := func(_ context.Context, snapshot ExecutionSnapshot) error {
			mu.Lock()
			defer mu.Unlock()
			if node := snapshot.Root.Children[0]; node.FinishedAt == nil {
				outputs, details = append(outputs, node.Output), append(details, node.Metadata)
			}
			return nil
		}
		done := make(chan []FixtureResult)
		go func() {
			defer GinkgoRecover()
			results, _, err := RunNodes(WithProgressSink(context.Background(), sink), []*FixtureNode{stepNode(StepSpec{ID: "resume", Kind: "recorder"})},
				RunOptions{WorkDir: GinkgoT().TempDir()})
			Expect(err).NotTo(HaveOccurred())
			done <- results
		}()
		Eventually(func() []map[string]any { mu.Lock(); defer mu.Unlock(); return details }).Should(ContainElement(HaveKey(QueryMetadataKey)))
		close(release)
		Expect((<-done)[0].Stdout).To(Equal(printed + "\n"))
		mu.Lock()
		defer mu.Unlock()
		Expect(outputs).To(ContainElement(printed+"\n"), "the running node showed the output before the step finished")
	})

	It("mirrors what it records onto the step's own task, when it runs as one", func() {
		snapshots := make(chan task.TaskSnapshot, 1)
		running := task.StartTask("evidence", func(ctx flanksourceContext.Context, t *task.Task) (bool, error) {
			step := &StepContext{evidence: newStepEvidence(withFixtureTask(ctx, t), nil)}
			fmt.Fprintln(step.Stdout(), printed)
			step.SetCommand(command, workDir)
			Expect(step.SetDetail(QueryMetadataKey, record)).To(Succeed())
			Expect(step.Progress(2, 5)).To(Succeed())
			snapshots <- task.SnapshotTask(t, nil)
			return true, nil
		})
		_, err := running.GetResult()
		Expect(err).NotTo(HaveOccurred())

		snapshot := <-snapshots
		Expect([]any{snapshot.Stdout, snapshot.Progress, snapshot.MaxValue}).To(Equal([]any{printed + "\n", 2, 5}))
		Expect(snapshot.Details).To(Equal(map[string]any{QueryMetadataKey: map[string]any{"template": template, "arguments": []any{nil}},
			"command": command, "cwd": workDir}))
	})

	Describe("workflow commands", func() {
		printing := func(status task.Status, chunks ...string) FixtureResult {
			useKind(fakeKind{name: "printer", execute: func(_ context.Context, step *StepContext) (StepOutcome, error) {
				for _, chunk := range chunks {
					fmt.Fprint(step.Stdout(), chunk)
				}
				return StepOutcome{Result: FixtureResult{Status: status}}, nil
			}})
			return runSteps(StepSpec{ID: "print", Kind: "printer"})[0]
		}
		findings := func(result FixtureResult) []string {
			var out []string
			for _, child := range result.Children {
				out = append(out, string(child.Results.Status)+" "+child.Name)
			}
			return out
		}

		It("keeps output and group markers, applies commands split across writes, and drops their lines", func() {
			result := printing(task.StatusPASS,
				"::group::Bu", "ild\ncompiling\n::desc", "ription::copying files\n",
				"::status::2 of 3\n::notice title=Cache::hit\n::endgroup::\n", "done")

			Expect(result.Status).To(Equal(task.StatusPASS), result.Error)
			Expect(result.Stdout).To(Equal("::group::Build\ncompiling\n::endgroup::\ndone"))
			Expect([]any{result.Metadata["description"], result.Metadata["status"]}).To(Equal([]any{"copying files", "2 of 3"}))
			Expect(findings(result)).To(Equal([]string{string(task.StatusPASS) + " Cache: hit"}))
			Expect(*result.Children[0].Results.Metadata[ViolationMetadataKey].(models.Violation).Code).To(Equal("Build"))
		})

		It("fails a step that passed when it printed an ::error", func() {
			result := printing(task.StatusPASS, "::warning::slow\n::error file=a.go,line=3::boom\n")

			Expect(result.Status).To(Equal(task.StatusFAIL))
			Expect(result.Error).To(Equal("1 error finding(s), 1 other: ::error: boom"))
			Expect(findings(result)).To(Equal([]string{string(task.StatusWarning) + " ::warning: slow", string(task.StatusFAIL) + " ::error: boom"}))
		})

		It("warns a step that passed when it printed only a ::warning", func() {
			result := printing(task.StatusPASS, "::warning::deprecated flag\n")

			Expect([]string{string(result.Status), result.Error}).To(Equal([]string{string(task.StatusWarning), "1 warning finding(s)"}))
		})

		It("keeps the kind's own failure when the step also printed an ::error", func() {
			result := printing(task.StatusFAIL, "::error::boom\n")

			Expect(result.Status).To(Equal(task.StatusFAIL))
			Expect(result.Error).To(BeEmpty())
		})

		It("reports a malformed command as an error finding instead of dropping it", func() {
			result := printing(task.StatusPASS, "::progress::half\n")

			Expect(result.Status).To(Equal(task.StatusFAIL))
			Expect(findings(result)).To(Equal([]string{string(task.StatusFAIL) + " ::invalid: ::progress \"half\": want a <done>/<total> message or done and total properties"}))
		})

		It("streams a partial plain line at once, and holds one that may be a command until its newline", func() {
			step := &StepContext{}
			fmt.Fprint(step.Stdout(), "downloading 40%")
			fmt.Fprint(step.Stdout(), "\n::progr")
			Expect(step.evidence.output().Stdout).To(Equal("downloading 40%\n"))

			fmt.Fprint(step.Stdout(), "ess::1/2\n")
			Expect(step.evidence.output().Stdout).To(Equal("downloading 40%\n"))
		})
	})

	It("records without a runner — a kind's unit test — and publishes nowhere", func() {
		step := &StepContext{}
		fmt.Fprint(step.Stdout(), printed)
		Expect(step.SetDetail(QueryMetadataKey, record)).To(Succeed())
		Expect(step.Progress(1, 1)).To(Succeed())
		Expect(step.evidence.fill(FixtureResult{}).Stdout).To(Equal(printed))
	})
})
