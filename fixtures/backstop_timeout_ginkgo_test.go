package fixtures

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/flanksource/clicky/task"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// stubbornKind ignores its context and holds for a fixed time, so only the
// fixture task's backstop deadline can cut it off.
type stubbornKind struct {
	name string
	hold time.Duration
}

func (k stubbornKind) Name() string          { return k.name }
func (stubbornKind) Schema() json.RawMessage { return json.RawMessage(`{"type":"object"}`) }

func (k stubbornKind) Execute(context.Context, *StepContext) (StepOutcome, error) {
	time.Sleep(k.hold)
	return StepOutcome{Result: FixtureResult{Status: task.StatusPASS}}, nil
}

type fixtureVerdict struct {
	Name     string
	Status   task.Status
	TimedOut bool
	Error    string
}

const cutOffError = `hang exceeded its backstop deadline: task "hang" timed out after 200ms: context deadline exceeded`

// runBackstopDocuments runs each document as its own fixture file with a 150ms
// backstop grace, and returns the run error, every node's verdict and the stats.
func runBackstopDocuments(documents map[string]string) (error, []fixtureVerdict, Stats) {
	DeferCleanup(DefaultKinds.Replace(stubbornKind{name: "backstop-hang", hold: time.Second}))
	DeferCleanup(DefaultKinds.Replace(stubbornKind{name: "backstop-pass"}))
	dir := GinkgoT().TempDir()
	var paths []string
	for name, body := range documents {
		path := filepath.Join(dir, name)
		Expect(os.WriteFile(path, []byte(body), 0o600)).To(Succeed())
		paths = append(paths, path)
	}
	runner, err := NewRunner(RunnerOptions{Paths: paths, WorkDir: dir})
	Expect(err).NotTo(HaveOccurred())
	runner.backstopGrace = 150 * time.Millisecond

	tree, runErr := runner.Run()

	var verdicts []fixtureVerdict
	tree.Walk(func(node *FixtureNode) {
		Expect(node.Results).NotTo(BeNil(), node.Name)
		verdicts = append(verdicts, fixtureVerdict{node.Results.Name, node.Results.Status, node.Results.TimedOut, node.Results.Error})
	})
	return runErr, verdicts, tree.GetStats()
}

var _ = Describe("a fixture cut off by its backstop deadline", func() {
	It("fails that fixture as timed out and keeps every other fixture's result", func() {
		runErr, verdicts, stats := runBackstopDocuments(map[string]string{
			"hang.fixture.yaml": "kind: Fixture\nsteps:\n  - { id: hang, kind: backstop-hang, timeout: 50ms }\n",
			"pass.fixture.yaml": "kind: Fixture\nsteps:\n  - { id: pass, kind: backstop-pass }\n",
		})

		Expect(runErr).To(MatchError("fixture tests failed"))
		Expect(verdicts).To(ConsistOf(
			fixtureVerdict{Name: "hang", Status: task.StatusFAIL, TimedOut: true, Error: cutOffError},
			fixtureVerdict{Name: "pass", Status: task.StatusPASS},
		))
		Expect(stats).To(Equal(Stats{Total: 2, Passed: 1, Failed: 1, TimedOut: 1}))
	})

	It("skips the steps after it, naming the cut-off step, and counts them", func() {
		runErr, verdicts, stats := runBackstopDocuments(map[string]string{
			"chain.fixture.yaml": "kind: Fixture\nsteps:\n  - { id: hang, kind: backstop-hang, timeout: 50ms }\n  - { id: after, kind: backstop-pass }\n",
		})

		Expect(runErr).To(MatchError("fixture tests failed"))
		Expect(verdicts).To(Equal([]fixtureVerdict{
			{Name: "hang", Status: task.StatusFAIL, TimedOut: true, Error: cutOffError},
			{Name: "after", Status: task.StatusSKIP, Error: "skipped: step hang failed"},
		}))
		Expect(stats).To(Equal(Stats{Total: 2, Failed: 1, TimedOut: 1, Skipped: 1}))
	})
})
