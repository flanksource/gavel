package run

import (
	"errors"

	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/lifecycle"
	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// The report a CLI prints is built from the run's own result and must survive the
// write that follows it — a triage run whose SaveAttempt failed used to print
// nothing but "failed".
var _ = ginkgo.Describe("settling a finished run", func() {
	var (
		steps     []string
		outcome   *lifecycle.StepOutcome
		reported  string
		reportErr error
	)
	ginkgo.BeforeEach(func() {
		steps, reported, reportErr = nil, "", nil
		outcome = &lifecycle.StepOutcome{
			Status:    string(lifecycle.OutcomeKeep),
			Execution: &todos.ExecutionResult{Summary: "Triaged it."},
		}
	})

	report := func(_ *lifecycle.StepOutcome, status string, err error) {
		steps = append(steps, "report")
		reported, reportErr = status, err
	}

	ginkgo.It("reports before it persists", func() {
		Expect(settleOutcome(outcome, nil, report, func(*lifecycle.StepOutcome, string) error {
			steps = append(steps, "persist")
			return nil
		})).To(Succeed())

		Expect(steps).To(Equal([]string{"report", "persist"}))
		Expect(reported).To(Equal(lifecycle.OutcomeKeep))
		Expect(reportErr).To(BeNil())
	})

	ginkgo.It("still reports when persistence fails, and joins the write's error", func() {
		writeFailed := errors.New("persist TODO attempt: connection reset")

		err := settleOutcome(outcome, nil, report, func(*lifecycle.StepOutcome, string) error { return writeFailed })

		Expect(steps).To(Equal([]string{"report"}))
		Expect(err).To(MatchError(writeFailed))
	})

	ginkgo.It("reports the run's own error, and keeps an unclassified status", func() {
		outcome.Status = ""
		runFailed := errors.New("agent exited 1")

		err := settleOutcome(outcome, runFailed, report, func(*lifecycle.StepOutcome, string) error { return nil })

		Expect(reported).To(Equal(lifecycle.OutcomeKeep), "an unclassified failure keeps the todo's status")
		Expect(reportErr).To(MatchError(runFailed))
		Expect(err).To(MatchError(runFailed))
	})

	ginkgo.It("persists a run with no reporter", func() {
		persisted := false

		Expect(settleOutcome(outcome, nil, nil, func(*lifecycle.StepOutcome, string) error {
			persisted = true
			return nil
		})).To(Succeed())

		Expect(persisted).To(BeTrue())
	})
})
