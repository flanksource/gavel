package ui

import (
	"context"
	"errors"
	"time"

	"github.com/flanksource/captain/pkg/ai/approval"
	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// The window a tool approval gets used to be a compile-time constant, so an
// unattended CI run and a dashboard somebody is watching waited the same 24
// hours — well past the point their own budget had already killed the run. It is
// now read off the spec Captain admitted the run with, which is where the
// existing .gavel.yaml → frontmatter → request layering has already been applied.
var _ = Describe("approvalWindow", func() {
	var promptRunID uuid.UUID

	BeforeEach(func() { promptRunID = uuid.New() })

	It("uses the window the run's spec declares", func(ctx SpecContext) {
		store := &fakeApprovalStore{run: promptRunWithSpec(promptRunID, map[string]any{
			"permissions": map[string]any{"mode": "acceptEdits", "approvalTimeout": "20m"},
		})}
		Expect(approvalWindow(ctx, store, promptRunID)).To(Equal(20 * time.Minute))
	})

	It("falls back to Captain's provider ceiling when the spec declares no window", func(ctx SpecContext) {
		store := &fakeApprovalStore{run: promptRunWithSpec(promptRunID, map[string]any{
			"permissions": map[string]any{"mode": "acceptEdits"},
		})}
		Expect(approvalWindow(ctx, store, promptRunID)).To(Equal(approval.ProviderTimeout))
	})

	It("falls back to the ceiling when the run carries no rendered spec at all", func(ctx SpecContext) {
		store := &fakeApprovalStore{run: &captaindb.PromptRun{ID: promptRunID}}
		Expect(approvalWindow(ctx, store, promptRunID)).To(Equal(approval.ProviderTimeout))
	})

	It("refuses a declared window that would never bound anything", func(ctx SpecContext) {
		store := &fakeApprovalStore{run: promptRunWithSpec(promptRunID, map[string]any{
			"permissions": map[string]any{"approvalTimeout": "soon"},
		})}
		_, err := approvalWindow(ctx, store, promptRunID)
		Expect(err).To(MatchError(ContainSubstring("soon")),
			"widening a window somebody tried to narrow is the failure this field exists to prevent")
	})

	It("reports a store it could not read rather than inventing a window", func(ctx SpecContext) {
		unreachable := errors.New("captain database is gone")
		store := &fakeApprovalStore{err: unreachable}
		_, err := approvalWindow(ctx, store, promptRunID)
		Expect(err).To(MatchError(unreachable))
	})
})

func promptRunWithSpec(id uuid.UUID, spec map[string]any) *captaindb.PromptRun {
	return &captaindb.PromptRun{ID: id, RenderedSpec: spec}
}

// fakeApprovalStore answers the one method approvalWindow uses; the rest of the
// interface is present because the seam is the interface, not this spec's needs.
type fakeApprovalStore struct {
	run *captaindb.PromptRun
	err error
}

func (f *fakeApprovalStore) GetPromptRun(context.Context, uuid.UUID) (*captaindb.PromptRun, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.run, nil
}

func (f *fakeApprovalStore) ListTurnRequests(context.Context, captaindb.TurnRequestFilter) ([]captaindb.TurnRequest, error) {
	return nil, errors.New("not used by these specs")
}

func (f *fakeApprovalStore) GetTurnRequest(context.Context, uuid.UUID) (*captaindb.TurnRequest, error) {
	return nil, errors.New("not used by these specs")
}

func (f *fakeApprovalStore) ResolveToolApprovalRequest(
	context.Context, captaindb.ResolveToolApprovalRequestInput,
) (*captaindb.TurnRequest, error) {
	return nil, errors.New("not used by these specs")
}

func (f *fakeApprovalStore) CancelPendingTurnRequests(context.Context, uuid.UUID, uuid.UUID, string) error {
	return errors.New("not used by these specs")
}

func (f *fakeApprovalStore) UpdatePromptRun(
	context.Context, captaindb.UpdatePromptRunInput,
) (*captaindb.PromptRun, error) {
	return nil, errors.New("not used by these specs")
}
