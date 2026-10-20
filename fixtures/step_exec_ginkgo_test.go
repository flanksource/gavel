package fixtures

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/flanksource/clicky/task"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// fakeKind is a scriptable Kind: execute decides the outcome, observe (when
// set) makes it an Observer.
type fakeKind struct {
	name    string
	execute func(ctx context.Context, step *StepContext) (StepOutcome, error)
}

func (k fakeKind) Name() string            { return k.name }
func (k fakeKind) Schema() json.RawMessage { return json.RawMessage(`{"type":"object"}`) }
func (k fakeKind) Execute(ctx context.Context, step *StepContext) (StepOutcome, error) {
	return k.execute(ctx, step)
}

type observingKind struct {
	fakeKind
	observe func(ctx context.Context, step *StepContext, handle any) (map[string]any, error)
}

func (k observingKind) Observe(ctx context.Context, step *StepContext, handle any) (map[string]any, error) {
	return k.observe(ctx, step, handle)
}

type backgroundKind struct {
	fakeKind
	start func(ctx context.Context, step *StepContext) (BackgroundHandle, error)
}

func (k backgroundKind) Start(ctx context.Context, step *StepContext) (BackgroundHandle, error) {
	return k.start(ctx, step)
}

type handleFunc func(ctx context.Context) FixtureResult

func (h handleFunc) Stop(ctx context.Context) FixtureResult { return h(ctx) }

func useKind(kind Kind) {
	DeferCleanup(DefaultKinds.Replace(kind))
}

func passing(outputs map[string]any, exports map[string]string) func(context.Context, *StepContext) (StepOutcome, error) {
	return func(context.Context, *StepContext) (StepOutcome, error) {
		return StepOutcome{Result: FixtureResult{Status: task.StatusPASS}, Outputs: outputs, Exports: exports}, nil
	}
}

func stepNode(spec StepSpec) *FixtureNode {
	name := spec.ID
	if name == "" {
		name = spec.Kind
	}
	return &FixtureNode{Name: name, Type: TestNode, Test: &FixtureTest{Name: name, Step: &spec}}
}

func runSteps(specs ...StepSpec) []FixtureResult {
	GinkgoHelper()
	nodes := make([]*FixtureNode, 0, len(specs))
	for _, spec := range specs {
		nodes = append(nodes, stepNode(spec))
	}
	results, _, err := RunNodes(context.Background(), nodes, RunOptions{WorkDir: GinkgoT().TempDir()})
	Expect(err).ToNot(HaveOccurred())
	return results
}

func metadataJSON(result FixtureResult) map[string]any {
	GinkgoHelper()
	raw, err := json.Marshal(result.Metadata)
	Expect(err).ToNot(HaveOccurred())
	var decoded map[string]any
	Expect(json.Unmarshal(raw, &decoded)).To(Succeed())
	return decoded
}

var _ = Describe("generic step execution", func() {
	It("exports a later step can read from its env and templates, recording where each came from", func() {
		useKind(fakeKind{name: "producer", execute: passing(map[string]any{"url": "http://127.0.0.1:9"}, map[string]string{"PRODUCER_URL": "http://127.0.0.1:9"})})
		var seen []string
		useKind(fakeKind{name: "consumer", execute: func(_ context.Context, step *StepContext) (StepOutcome, error) {
			rendered, err := step.Expand("{{.env.API}} $PRODUCER_URL {{.steps.server.url}}")
			seen = append(seen, rendered)
			return StepOutcome{Result: FixtureResult{Status: task.StatusPASS}}, err
		}})

		results := runSteps(
			StepSpec{ID: "server", Kind: "producer", Export: map[string]string{"API": "{{.outputs.url}}/api"}},
			StepSpec{ID: "client", Kind: "consumer"},
		)

		Expect(results[1].Status).To(Equal(task.StatusPASS), results[1].Error)
		Expect(seen).To(Equal([]string{"http://127.0.0.1:9/api http://127.0.0.1:9 http://127.0.0.1:9"}))
		meta := metadataJSON(results[0])
		Expect(meta).To(HaveKeyWithValue("step", map[string]any{"id": "server", "kind": "producer"}))
		Expect(meta).To(HaveKeyWithValue("outputs", map[string]any{"url": "http://127.0.0.1:9"}))
		Expect(meta).To(HaveKeyWithValue("exports", map[string]any{
			"API":          map[string]any{"value": "http://127.0.0.1:9/api", "step": "server"},
			"PRODUCER_URL": map[string]any{"value": "http://127.0.0.1:9", "step": "server"},
		}))
		Expect(meta).To(HaveKeyWithValue("background", false))
		Expect(meta["attempts"]).To(HaveLen(1))
		Expect(results[0].Type).To(Equal("producer"))
	})

	It("exports every entry of an output map with \"*\" and suppresses a default export set to empty", func() {
		useKind(fakeKind{name: "producer", execute: passing(map[string]any{"env": map[string]any{"A": "1", "B": "2"}}, map[string]string{"DEFAULT": "x"})})

		results := runSteps(StepSpec{ID: "svc", Kind: "producer", Export: map[string]string{"*": "{{ .outputs.env }}", "DEFAULT": ""}})

		Expect(results[0].Status).To(Equal(task.StatusPASS), results[0].Error)
		Expect(metadataJSON(results[0])["exports"]).To(Equal(map[string]any{
			"A": map[string]any{"value": "1", "step": "svc"},
			"B": map[string]any{"value": "2", "step": "svc"},
		}))
	})

	It("fails the second step that re-exports a name without override", func() {
		useKind(fakeKind{name: "producer", execute: passing(map[string]any{}, map[string]string{"SHARED": "v"})})

		results := runSteps(StepSpec{ID: "one", Kind: "producer"}, StepSpec{ID: "two", Kind: "producer"})

		Expect(results[0].Status).To(Equal(task.StatusPASS))
		Expect(results[1].Status).To(Equal(task.StatusERR))
		Expect(results[1].Error).To(ContainSubstring(`SHARED was already exported by step "one"`))
	})

	It("skips every typed step after the first failure, naming it, without executing them", func() {
		useKind(fakeKind{name: "ok", execute: passing(nil, nil)})
		useKind(fakeKind{name: "broken", execute: func(context.Context, *StepContext) (StepOutcome, error) {
			return StepOutcome{Result: FixtureResult{Status: task.StatusFAIL}}, errors.New("service exited before it was ready")
		}})
		executed := 0
		useKind(fakeKind{name: "after", execute: func(context.Context, *StepContext) (StepOutcome, error) {
			executed++
			return StepOutcome{Result: FixtureResult{Status: task.StatusPASS}}, nil
		}})

		results := runSteps(StepSpec{ID: "first", Kind: "ok"}, StepSpec{ID: "stack", Kind: "broken"},
			StepSpec{ID: "todo", Kind: "after"}, StepSpec{ID: "plan", Kind: "after"})

		Expect([]task.Status{results[0].Status, results[2].Status, results[3].Status}).To(Equal(
			[]task.Status{task.StatusPASS, task.StatusSKIP, task.StatusSKIP}))
		Expect(results[1].Status).To(BeElementOf(task.StatusFAIL, task.StatusERR))
		Expect([]string{results[2].Error, results[3].Error}).To(Equal([]string{"skipped: step stack failed", "skipped: step stack failed"}))
		Expect(results[3].Type).To(Equal("after"))
		Expect(executed).To(BeZero())
	})

	It("waits on an Observer until its CEL verdict, recording only final outputs", func() {
		calls := 0
		useKind(observingKind{
			fakeKind: fakeKind{name: "poller", execute: passing(map[string]any{"state": "starting"}, nil)},
			observe: func(context.Context, *StepContext, any) (map[string]any, error) {
				calls++
				state := "starting"
				if calls >= 2 {
					state = "ready"
				}
				return map[string]any{"state": state}, nil
			},
		})

		results := runSteps(StepSpec{ID: "svc", Kind: "poller", CEL: "state == 'ready'",
			Wait: &WaitPolicy{Until: "state == 'ready'", Interval: dur(5 * time.Millisecond), Timeout: dur(5 * time.Second)}})

		Expect(results[0].Status).To(Equal(task.StatusPASS), results[0].Error)
		wait := metadataJSON(results[0])["wait"].(map[string]any)
		Expect(wait["outcome"]).To(Equal(WaitSatisfied))
		Expect(wait).NotTo(HaveKey("observations"))
		Expect(wait["outputs"]).To(Equal(map[string]any{"state": "ready"}))
		Expect(wait["policy"]).To(HaveKeyWithValue("until", "state == 'ready'"))
		Expect(metadataJSON(results[0])["outputs"]).To(Equal(map[string]any{"state": "ready"}))
	})

	It("fails a wait on a timeout with the last observed outputs", func() {
		useKind(observingKind{
			fakeKind: fakeKind{name: "poller", execute: passing(nil, nil)},
			observe: func(context.Context, *StepContext, any) (map[string]any, error) {
				return map[string]any{"state": "starting"}, nil
			},
		})

		results := runSteps(StepSpec{ID: "svc", Kind: "poller",
			Wait: &WaitPolicy{Until: "state == 'ready'", Interval: dur(5 * time.Millisecond), Timeout: dur(30 * time.Millisecond)}})

		Expect(results[0].Status).To(Equal(task.StatusFAIL))
		Expect(results[0].Error).To(ContainSubstring("did not hold within"))
		Expect(metadataJSON(results[0])["wait"]).To(HaveKeyWithValue("outcome", WaitTimeout))
	})

	It("rejects a wait on a kind that cannot be observed", func() {
		useKind(fakeKind{name: "blind", execute: passing(nil, nil)})

		results := runSteps(StepSpec{ID: "x", Kind: "blind",
			Wait: &WaitPolicy{Until: "true", Interval: dur(time.Millisecond), Timeout: dur(time.Second)}})

		Expect(results[0].Status).To(Equal(task.StatusERR))
		Expect(results[0].Error).To(ContainSubstring(`kind "blind" cannot wait`))
	})

	It("re-runs a failed Execute while retry.when holds and records every attempt", func() {
		calls := 0
		useKind(fakeKind{name: "flaky", execute: func(context.Context, *StepContext) (StepOutcome, error) {
			calls++
			if calls < 3 {
				return StepOutcome{}, errors.New("connection refused")
			}
			return StepOutcome{Result: FixtureResult{Status: task.StatusPASS}}, nil
		}})

		results := runSteps(StepSpec{ID: "x", Kind: "flaky",
			Retry: &RetryPolicy{Attempts: 3, When: "error.contains('refused')", Delay: dur(time.Millisecond)}})

		Expect(results[0].Status).To(Equal(task.StatusPASS), results[0].Error)
		attempts := metadataJSON(results[0])["attempts"].([]any)
		Expect(attempts).To(HaveLen(3))
		Expect(attempts[0]).To(HaveKeyWithValue("error", "connection refused"))
		Expect(attempts[2]).To(HaveKeyWithValue("status", AttemptPassed))
	})

	It("fails a step whose cel is false, with the CEL trace", func() {
		useKind(fakeKind{name: "producer", execute: passing(map[string]any{"count": 2}, nil)})

		results := runSteps(StepSpec{ID: "x", Kind: "producer", CEL: "count == 3"})

		Expect(results[0].Status).To(Equal(task.StatusFAIL))
		Expect(results[0].CELExpression).To(Equal("count == 3"))
		Expect(results[0].CELTrace).ToNot(BeEmpty())
		attempts := results[0].Metadata["attempts"].([]AttemptReport)
		Expect(attempts[0].Status).To(Equal(AttemptFailed))
		Expect(attempts[0].Result.CELTrace).To(Equal(results[0].CELTrace))
		Expect(attempts[0].Result.Metadata).NotTo(HaveKey("attempts"))
	})

	It("renders config templates before Execute, unless the kind templates its own", func() {
		var configs []string
		capture := func(_ context.Context, step *StepContext) (StepOutcome, error) {
			configs = append(configs, string(step.Spec.Config))
			return StepOutcome{Result: FixtureResult{Status: task.StatusPASS}, Outputs: map[string]any{"v": "seen"}}, nil
		}
		useKind(fakeKind{name: "rendered", execute: capture})
		useKind(selfTemplatingKind{fakeKind{name: "raw", execute: capture}})

		unset := runSteps(StepSpec{ID: "a", Kind: "rendered", Config: json.RawMessage(`{"value":"{{.steps.b.v}}","n":3}`)})
		Expect(unset[0].Status).To(Equal(task.StatusERR), "a reference to a step that has not run is not set")
		Expect(unset[0].Error).To(ContainSubstring("{{.steps.b.v}}"))

		results := runSteps(
			StepSpec{ID: "b", Kind: "raw", Config: json.RawMessage(`{"value":"{{.port}}"}`)},
			StepSpec{ID: "c", Kind: "rendered", Config: json.RawMessage(`{"z":"{{.steps.b.v}}","a":[1,{"y":2,"x":"$KEEP"}]}`)},
		)
		Expect([]task.Status{results[0].Status, results[1].Status}).To(Equal([]task.Status{task.StatusPASS, task.StatusPASS}))
		Expect(configs).To(Equal([]string{`{"value":"{{.port}}"}`, `{"z":"seen","a":[1,{"y":2,"x":"$KEEP"}]}`}))
	})

	It("fails a step killed by its own timeout as timed out", func() {
		useKind(fakeKind{name: "slow", execute: func(ctx context.Context, _ *StepContext) (StepOutcome, error) {
			<-ctx.Done()
			return StepOutcome{}, ctx.Err()
		}})

		results := runSteps(StepSpec{ID: "x", Kind: "slow", Timeout: dur(20 * time.Millisecond)})

		Expect([]any{results[0].Status, results[0].TimedOut, results[0].Error}).To(Equal([]any{
			task.StatusFAIL, true, "step x timed out after 20ms: context deadline exceeded",
		}))
		attempts := results[0].Metadata["attempts"].([]AttemptReport)
		Expect(attempts[0].Status).To(Equal(AttemptFailed))
		Expect(attempts[0].Result.TimedOut).To(BeTrue())
		Expect(attempts[0].Result.Error).To(Equal(results[0].Error))
	})

	It("does not report a step cancelled with its run as timed out", func() {
		ctx, cancel := context.WithCancel(context.Background())
		useKind(fakeKind{name: "slow", execute: func(ctx context.Context, _ *StepContext) (StepOutcome, error) {
			cancel()
			<-ctx.Done()
			return StepOutcome{}, ctx.Err()
		}})

		results, _, err := RunNodes(ctx, []*FixtureNode{stepNode(StepSpec{ID: "x", Kind: "slow", Timeout: dur(time.Minute)})},
			RunOptions{WorkDir: GinkgoT().TempDir()})

		Expect(err).ToNot(HaveOccurred())
		Expect([]any{results[0].Status, results[0].TimedOut}).To(Equal([]any{task.StatusERR, false}))
	})

	It("reports an unknown kind as an errored result", func() {
		results := runSteps(StepSpec{ID: "x", Kind: "no-such-kind"})

		Expect(results[0].Status).To(Equal(task.StatusERR))
		Expect(results[0].Error).To(ContainSubstring(`unknown fixture step kind "no-such-kind"`))
	})

	It("collects a background step's verdict and runs cleanups at teardown, surfacing their errors", func() {
		var mu sync.Mutex
		var order []string
		note := func(s string) { mu.Lock(); defer mu.Unlock(); order = append(order, s) }
		useKind(backgroundKind{
			fakeKind: fakeKind{name: "watcher"},
			start: func(_ context.Context, step *StepContext) (BackgroundHandle, error) {
				note("start")
				return handleFunc(func(context.Context) FixtureResult {
					note("stop")
					return FixtureResult{Name: "watcher verdict", Status: task.StatusFAIL, Error: "saw an alert"}
				}), nil
			},
		})
		useKind(fakeKind{name: "service", execute: func(context.Context, *StepContext) (StepOutcome, error) {
			return StepOutcome{Result: FixtureResult{Status: task.StatusPASS}, Cleanup: func(context.Context) error {
				note("cleanup")
				return errors.New("port still bound")
			}}, nil
		}})

		results := runSteps(StepSpec{ID: "svc", Kind: "service"}, StepSpec{ID: "watch", Kind: "watcher", Background: true})

		Expect(order).To(Equal([]string{"start", "stop", "cleanup"}))
		Expect(namesOfResults(results)).To(Equal([]string{"svc", "watch", "watcher verdict", "teardown"}))
		Expect(results[1].Status).To(Equal(task.StatusPASS))
		Expect(results[2].Status).To(Equal(task.StatusFAIL))
		Expect(results[2].Metadata).To(HaveKeyWithValue("background", true))
		Expect(results[3].Status).To(Equal(task.StatusERR))
		Expect(results[3].Error).To(ContainSubstring("teardown svc: port still bound"))
	})
})

type defaultWaitingKind struct{ observingKind }

func (defaultWaitingKind) DefaultWait(StepSpec) *WaitPolicy {
	return &WaitPolicy{Until: "state == 'ready'"}
}

var _ = Describe("kind default waits", func() {
	It("fills a DefaultWait's missing timing from the document defaults, then the built-in values", func() {
		useKind(defaultWaitingKind{observingKind{
			fakeKind: fakeKind{name: "service", execute: passing(nil, nil)},
			observe: func(context.Context, *StepContext, any) (map[string]any, error) {
				return map[string]any{"state": "ready"}, nil
			},
		}})

		results := runSteps(StepSpec{ID: "svc", Kind: "service", Defaults: &StepDefaults{Wait: &WaitPolicy{Jitter: "10%"}}})

		Expect(results[0].Status).To(Equal(task.StatusPASS), results[0].Error)
		policy := metadataJSON(results[0])["wait"].(map[string]any)["policy"]
		Expect(policy).To(Equal(map[string]any{
			"until": "state == 'ready'", "interval": "1s", "jitter": "10%", "timeout": "2m0s",
			"backoff": map[string]any{"factor": 1.5, "max": "10s"},
		}))
	})
})

type selfTemplatingKind struct{ fakeKind }

func (selfTemplatingKind) TemplatesOwnConfig() {}

func namesOfResults(results []FixtureResult) []string {
	names := make([]string, 0, len(results))
	for _, r := range results {
		names = append(names, r.Name)
	}
	return names
}
