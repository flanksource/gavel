package verifier_test

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/clicky/task"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flanksource/gavel/fixtures"
	_ "github.com/flanksource/gavel/fixtures/kinds/services"
	"github.com/flanksource/gavel/fixtures/verifier"
)

// pollingKind is a typed kind that fails its first Execute (to produce a retry
// attempt), then reports ready from observation readyAt on.
type pollingKind struct {
	executed, observed *int
	readyAt            int
}

func (pollingKind) Name() string            { return "detail-probe" }
func (pollingKind) Schema() json.RawMessage { return json.RawMessage(`{"type":"object"}`) }

func (k pollingKind) Execute(context.Context, *fixtures.StepContext) (fixtures.StepOutcome, error) {
	*k.executed++
	if *k.executed == 1 {
		return fixtures.StepOutcome{Result: fixtures.FixtureResult{Status: task.StatusFAIL, Error: "not yet"}}, nil
	}
	return fixtures.StepOutcome{Result: fixtures.FixtureResult{Status: task.StatusPASS},
		Outputs: map[string]any{"url": "http://127.0.0.1:1"}, Exports: map[string]string{"PROBE_URL": "http://127.0.0.1:1"}}, nil
}

func (k pollingKind) Observe(context.Context, *fixtures.StepContext, any) (map[string]any, error) {
	*k.observed++
	state := "starting"
	if *k.observed >= k.readyAt {
		state = "ready"
	}
	return map[string]any{"state": state}, nil
}

var _ = Describe("verifier.Report typed step detail", func() {
	It("projects stored attempt evidence after a fixture-worker JSON round trip", func() {
		original := fixtures.FixtureResult{Name: "probe", Type: "detail-probe", Status: task.StatusPASS, Metadata: map[string]any{
			"step": fixtures.StepRef{ID: "probe", Kind: "detail-probe"},
			"attempts": []fixtures.AttemptReport{{Attempt: 1, Status: fixtures.AttemptFailed, Result: &fixtures.FixtureResult{
				Name: "probe", Type: "detail-probe", Status: task.StatusFAIL, Stdout: "refused launch",
				Metadata: map[string]any{"step": fixtures.StepRef{ID: "probe", Kind: "detail-probe"}, "outputs": map[string]any{"phase": "refused"}},
			}}},
		}}
		raw, err := json.Marshal(original)
		Expect(err).NotTo(HaveOccurred())
		var restored fixtures.FixtureResult
		Expect(json.Unmarshal(raw, &restored)).To(Succeed())
		report := verifier.Report([]fixtures.FixtureResult{restored}, nil)
		var detail struct {
			Attempts []struct {
				Result api.VerifyNode `json:"result"`
			} `json:"attempts"`
		}
		Expect(json.Unmarshal(report.Tests[0].Detail, &detail)).To(Succeed())
		Expect(detail.Attempts).To(HaveLen(1))
		result, evidence := splitDetail(detail.Attempts[0].Result)
		Expect(result).To(Equal(api.VerifyNode{Name: "probe", Framework: "detail-probe", Failed: true, Stdout: "refused launch"}))
		Expect(evidence).To(MatchJSON(`{"step":{"id":"probe","kind":"detail-probe"},"outputs":{"phase":"refused"}}`))
	})

	It("carries step, outputs, exports, wait and attempts in the exact shape the step renderer parses", func() {
		executed, observed := 0, 0
		DeferCleanup(fixtures.DefaultKinds.Replace(pollingKind{executed: &executed, observed: &observed, readyAt: 2}))
		interval, timeout := fixtures.Duration(5*time.Millisecond), fixtures.Duration(5*time.Second)
		spec := fixtures.StepSpec{ID: "probe", Kind: "detail-probe", Retry: &fixtures.RetryPolicy{Attempts: 2, When: "true"},
			Wait: &fixtures.WaitPolicy{Until: "state == 'ready'", Interval: &interval, Timeout: &timeout, Jitter: "10%"}}
		node := &fixtures.FixtureNode{Name: "probe", Type: fixtures.TestNode, Test: &fixtures.FixtureTest{Name: "probe", Step: &spec}}

		results, snapshot, err := fixtures.RunNodes(context.Background(), []*fixtures.FixtureNode{node}, fixtures.RunOptions{WorkDir: GinkgoT().TempDir()})
		Expect(err).ToNot(HaveOccurred())
		report := verifier.Report(results, snapshot)

		var detail map[string]any
		Expect(json.Unmarshal(report.Tests[0].Detail, &detail)).To(Succeed())
		Expect(report.Tests[0].Framework).To(Equal("detail-probe"))
		Expect(detail["step"]).To(Equal(map[string]any{"id": "probe", "kind": "detail-probe"}))
		Expect(detail["outputs"]).To(Equal(map[string]any{"state": "ready"}))
		Expect(detail["exports"]).To(Equal(map[string]any{"PROBE_URL": map[string]any{"value": "http://127.0.0.1:1", "step": "probe"}}))
		Expect(detail["background"]).To(BeFalse())

		attempts := detail["attempts"].([]any)
		Expect(attempts).To(HaveLen(2))
		Expect(attempts[0]).To(HaveKeyWithValue("attempt", float64(1)))
		Expect(attempts[0]).To(HaveKeyWithValue("status", "failed"))
		Expect(attempts[0]).To(HaveKeyWithValue("error", "not yet"))
		Expect(attempts[0]).To(HaveKeyWithValue("durationMs", BeAssignableToTypeOf(float64(0))))
		Expect(attempts[1]).To(HaveKeyWithValue("status", "passed"))

		wait := detail["wait"].(map[string]any)
		Expect(wait["outcome"]).To(BeElementOf("satisfied", "failed", "timeout"))
		Expect(wait["outcome"]).To(Equal("satisfied"))
		Expect(wait["policy"]).To(Equal(map[string]any{
			"until": "state == 'ready'", "interval": "5ms", "timeout": "5s", "jitter": "10%",
			"backoff": map[string]any{"factor": 1.5, "max": "10s"},
		}))
		Expect(wait).NotTo(HaveKey("observations"))
		Expect(wait["durationMs"]).To(BeNumerically(">=", 4))
		Expect(wait["outputs"]).To(Equal(map[string]any{"state": "ready"}))
		first := attempts[0].(map[string]any)["result"].(map[string]any)
		last := attempts[1].(map[string]any)["result"].(map[string]any)
		Expect(first).To(HaveKeyWithValue("failed", true))
		Expect(first["detail"]).NotTo(HaveKey("wait"))
		Expect(last).To(HaveKeyWithValue("passed", true))
		Expect(last["detail"]).To(HaveKeyWithValue("outputs", map[string]any{"state": "ready"}))
		Expect(last["detail"]).NotTo(HaveKey("attempts"))
		Expect(string(report.Tests[0].Detail)).NotTo(ContainSubstring("starting"))
	})

	It("reports a background watcher's verdict with its step record, events and background flag", func() {
		events := []map[string]any{{"type": "console", "text": "boom", "url": "http://127.0.0.1:1/"}}
		DeferCleanup(fixtures.DefaultKinds.Replace(watchingKind{events: events}))
		spec := fixtures.StepSpec{ID: "watch", Kind: "detail-watcher", Background: true}

		report := runSteps(spec)

		Expect(report.Tests).To(HaveLen(2), "the started step and its teardown verdict")
		node, detail := splitDetail(report.Tests[1])
		Expect(node).To(Equal(api.VerifyNode{Name: "watch", Framework: "detail-watcher", Message: "console error: boom", Failed: true}))
		Expect(detail).To(MatchJSON(`{
			"step": {"id": "watch", "kind": "detail-watcher"},
			"outputs": {}, "exports": {}, "attempts": [], "background": true,
			"events": [{"type": "console", "text": "boom", "url": "http://127.0.0.1:1/"}]
		}`))
	})

	It("reports a step whose wait timed out as failed and timed out", func() {
		executed, observed := 0, 0
		DeferCleanup(fixtures.DefaultKinds.Replace(pollingKind{executed: &executed, observed: &observed, readyAt: math.MaxInt}))
		interval, timeout := fixtures.Duration(5*time.Millisecond), fixtures.Duration(30*time.Millisecond)
		spec := fixtures.StepSpec{ID: "probe", Kind: "detail-probe", Retry: &fixtures.RetryPolicy{Attempts: 2, When: "true"},
			Wait: &fixtures.WaitPolicy{Until: "state == 'ready'", Interval: &interval, Timeout: &timeout}}

		report := runSteps(spec)

		Expect(report.State).To(Equal(api.VerifyStateTimedOut))
		Expect(report.Validate()).To(Succeed())
		node, detail := splitDetail(report.Tests[0])
		Expect(node).To(Equal(api.VerifyNode{
			Name: "probe", Framework: "detail-probe", Failed: true, TimedOut: true,
			Message: "wait until state == 'ready' did not hold within 30ms",
			Context: &api.VerifyNodeContext{CELExpression: "state == 'ready'"},
		}))
		Expect(detail).To(ContainSubstring(`"outcome":"timeout"`))
	})

	It("records a services entry's readiness probes as a wait report under its outputs", func() {
		spec := fixtures.StepSpec{ID: "stack", Kind: "services", Config: json.RawMessage(`{"services": {"api": {
			"exec": "sleep 30",
			"ready": {"exec": "test -f ready || { touch ready; exit 1; }"},
			"wait": {"interval": "10ms", "timeout": "5s"}
		}}}`)}

		report := runSteps(spec)

		node, detail := splitDetail(report.Tests[0])
		Expect(node).To(Equal(api.VerifyNode{Name: "stack", Framework: "services", Passed: true}))
		entry := `{"type": "exec", "pid": "<volatile>", "log": "<volatile>", "ready": true, "env": {}, "wait": {
			"policy": {"interval": "10ms", "timeout": "5s", "backoff": {"factor": 1.5, "max": "2s"}},
			"outcome": "satisfied", "durationMs": "<volatile>", "outputs": {"ready": true}}}`
		var decoded map[string]any
		Expect(json.Unmarshal([]byte(detail), &decoded)).To(Succeed())
		attempts := decoded["attempts"].([]any)
		Expect(attempts).To(HaveLen(1))
		attemptResult := attempts[0].(map[string]any)["result"].(map[string]any)
		Expect(attemptResult["detail"]).To(HaveKeyWithValue("outputs", decoded["outputs"]))
		delete(decoded, "attempts")
		raw, err := json.Marshal(decoded)
		Expect(err).NotTo(HaveOccurred())
		Expect(detail).NotTo(Or(ContainSubstring("observations"), ContainSubstring("exit status 1")))
		Expect(scrubVolatile(string(raw))).To(MatchJSON(`{
			"step": {"id": "stack", "kind": "services"},
			"outputs": {"services": {"api": ` + entry + `}, "api": ` + entry + `, "ready": true},
			"exports": {}, "background": false
		}`))
	})
})

// watchingKind is a background kind whose watcher saw events and fails on them.
type watchingKind struct{ events []map[string]any }

func (watchingKind) Name() string            { return "detail-watcher" }
func (watchingKind) Schema() json.RawMessage { return json.RawMessage(`{"type":"object"}`) }

func (watchingKind) Execute(context.Context, *fixtures.StepContext) (fixtures.StepOutcome, error) {
	return fixtures.StepOutcome{}, fmt.Errorf("detail-watcher only runs in the background")
}

func (k watchingKind) Start(context.Context, *fixtures.StepContext) (fixtures.BackgroundHandle, error) {
	return k, nil
}

func (k watchingKind) Stop(context.Context) fixtures.FixtureResult {
	return fixtures.FixtureResult{Status: task.StatusFAIL, Error: "console error: boom", Metadata: map[string]any{"events": k.events}}
}

// runSteps runs typed steps as one fixture and maps the results to a report.
func runSteps(specs ...fixtures.StepSpec) api.VerifyReport {
	nodes := make([]*fixtures.FixtureNode, 0, len(specs))
	for i := range specs {
		nodes = append(nodes, &fixtures.FixtureNode{Name: specs[i].ID, Type: fixtures.TestNode,
			Test: &fixtures.FixtureTest{Name: specs[i].ID, Step: &specs[i]}})
	}
	results, snapshot, err := fixtures.RunNodes(context.Background(), nodes, fixtures.RunOptions{WorkDir: GinkgoT().TempDir()})
	Expect(err).ToNot(HaveOccurred())
	return verifier.Report(results, snapshot)
}

// splitDetail separates a node's wall-clock duration and JSON detail so the
// rest of it is compared whole.
func splitDetail(node api.VerifyNode) (api.VerifyNode, string) {
	detail := string(node.Detail)
	node.Detail, node.Duration = nil, 0
	return node, detail
}

// scrubVolatile replaces the values that differ run to run (timestamps,
// durations, pids, log paths) with "<volatile>".
func scrubVolatile(detail string) string {
	var decoded any
	Expect(json.Unmarshal([]byte(detail), &decoded)).To(Succeed())
	var scrub func(any)
	scrub = func(value any) {
		switch v := value.(type) {
		case map[string]any:
			for key, child := range v {
				if volatileKeys[key] {
					v[key] = "<volatile>"
					continue
				}
				scrub(child)
			}
		case []any:
			for _, child := range v {
				scrub(child)
			}
		}
	}
	scrub(decoded)
	raw, err := json.Marshal(decoded)
	Expect(err).ToNot(HaveOccurred())
	return string(raw)
}

var volatileKeys = map[string]bool{"at": true, "durationMs": true, "pid": true, "log": true}
