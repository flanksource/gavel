package verifier_test

import (
	"context"
	"encoding/json"
	"time"

	"github.com/flanksource/clicky/task"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/fixtures/verifier"
)

// pollingKind is a typed kind that fails its first Execute (to produce a retry
// attempt), then reports ready on its second observation.
type pollingKind struct{ executed, observed *int }

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
	if *k.observed >= 2 {
		state = "ready"
	}
	return map[string]any{"state": state}, nil
}

var _ = Describe("verifier.Report typed step detail", func() {
	It("carries step, outputs, exports, wait and attempts in the exact shape the step renderer parses", func() {
		executed, observed := 0, 0
		DeferCleanup(fixtures.DefaultKinds.Replace(pollingKind{executed: &executed, observed: &observed}))
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
		observations := wait["observations"].([]any)
		Expect(observations).To(HaveLen(2))
		for _, raw := range observations {
			observation := raw.(map[string]any)
			_, parseErr := time.Parse(time.RFC3339, observation["at"].(string))
			Expect(parseErr).ToNot(HaveOccurred())
			Expect(observation["delayMs"]).To(BeAssignableToTypeOf(float64(0)))
			Expect(observation["until"]).To(BeAssignableToTypeOf(true))
			Expect(observation["failWhen"]).To(BeAssignableToTypeOf(true))
			Expect(observation["outputs"]).To(BeAssignableToTypeOf(map[string]any{}))
		}
		Expect(observations[1]).To(HaveKeyWithValue("until", true))
	})
})
