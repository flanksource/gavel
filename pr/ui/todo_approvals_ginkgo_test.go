package ui

import (
	"context"
	"fmt"

	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// promptRunRow is an in-memory prompt run row with captain's optimistic
// versioning: an update on a stale version conflicts, and each committed
// update advances the version.
type promptRunRow struct {
	run captaindb.PromptRun
	// conflictsLeft makes that many updates fail as if another writer had moved
	// the row first, advancing the version each time so a re-read sees it.
	conflictsLeft int
	reads         int
	written       []captaindb.PromptRunState
}

func (r *promptRunRow) GetPromptRun(context.Context, uuid.UUID) (*captaindb.PromptRun, error) {
	r.reads++
	snapshot := r.run
	return &snapshot, nil
}

func (r *promptRunRow) UpdatePromptRun(_ context.Context, input captaindb.UpdatePromptRunInput) (*captaindb.PromptRun, error) {
	if input.ExpectedVersion != r.run.Version {
		return nil, fmt.Errorf("%w: prompt run %s is no longer at version %d", captaindb.ErrPromptRunConflict, input.ID, input.ExpectedVersion)
	}
	if r.conflictsLeft > 0 {
		r.conflictsLeft--
		r.run.Version++
		return nil, fmt.Errorf("%w: prompt run %s is no longer at version %d", captaindb.ErrPromptRunConflict, input.ID, input.ExpectedVersion)
	}
	r.run.State = *input.State
	r.run.Version++
	r.written = append(r.written, *input.State)
	return &r.run, nil
}

func (r *promptRunRow) ListTurnRequests(context.Context, captaindb.TurnRequestFilter) ([]captaindb.TurnRequest, error) {
	panic("not used")
}
func (r *promptRunRow) GetTurnRequest(context.Context, uuid.UUID) (*captaindb.TurnRequest, error) {
	panic("not used")
}
func (r *promptRunRow) ResolveToolApprovalRequest(context.Context, captaindb.ResolveToolApprovalRequestInput) (*captaindb.TurnRequest, error) {
	panic("not used")
}
func (r *promptRunRow) CancelPendingTurnRequests(context.Context, uuid.UUID, uuid.UUID, string) error {
	panic("not used")
}

var _ = Describe("setPromptRunState", func() {
	var (
		ctx   context.Context
		runID uuid.UUID
		row   *promptRunRow
	)

	BeforeEach(func() {
		ctx = context.Background()
		runID = uuid.New()
		row = &promptRunRow{run: captaindb.PromptRun{ID: runID, State: captaindb.PromptRunStateRunning, Version: 3}}
	})

	It("moves a live run on the version it read", func() {
		Expect(setPromptRunState(row, runID, captaindb.PromptRunStateWaiting)(ctx)).To(Succeed())

		Expect(row.written).To(Equal([]captaindb.PromptRunState{captaindb.PromptRunStateWaiting}))
		Expect(row.run.Version).To(Equal(int64(4)))
	})

	It("is a no-op when the run already is in the state", func() {
		row.run.State = captaindb.PromptRunStateWaiting

		Expect(setPromptRunState(row, runID, captaindb.PromptRunStateWaiting)(ctx)).To(Succeed())

		Expect(row.written).To(BeEmpty())
	})

	It("retries once on the fresh version when the row moved underneath it", func() {
		row.conflictsLeft = 1

		Expect(setPromptRunState(row, runID, captaindb.PromptRunStateWaiting)(ctx)).To(Succeed())

		Expect(row.reads).To(Equal(2), "the retry must re-read rather than replay the stale version")
		Expect(row.written).To(Equal([]captaindb.PromptRunState{captaindb.PromptRunStateWaiting}))
	})

	It("reports a second conflict instead of fighting for the row", func() {
		row.conflictsLeft = 2

		err := setPromptRunState(row, runID, captaindb.PromptRunStateWaiting)(ctx)

		Expect(err).To(MatchError(captaindb.ErrPromptRunConflict))
		Expect(row.written).To(BeEmpty())
	})

	DescribeTable("never resurrects a finished run",
		func(terminal captaindb.PromptRunState) {
			row.run.State = terminal

			err := setPromptRunState(row, runID, captaindb.PromptRunStateRunning)(ctx)

			Expect(err).To(MatchError(ContainSubstring(string(terminal))))
			Expect(row.written).To(BeEmpty())
			Expect(row.run.State).To(Equal(terminal))
		},
		Entry("succeeded", captaindb.PromptRunStateSucceeded),
		Entry("failed", captaindb.PromptRunStateFailed),
		Entry("cancelled", captaindb.PromptRunStateCancelled),
	)
})
