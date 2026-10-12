package ui

import (
	"context"
	"errors"

	"github.com/flanksource/captain/pkg/api"
	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type verificationSessionStore struct {
	fakeCaptainSessionStore
	report          *api.VerifyReport
	iteration       int
	verificationErr error
	requestedRun    uuid.UUID
}

func (s *verificationSessionStore) LatestPromptRunVerification(_ context.Context, runID uuid.UUID) (*api.VerifyReport, int, error) {
	s.requestedRun = runID
	return s.report, s.iteration, s.verificationErr
}

var _ = Describe("TODO session failure classification", func() {
	const iteration = 3
	const reason = "10 definition-of-done checks: 10 failed"
	var store *verificationSessionStore
	var resolved captainSessionResolution

	BeforeEach(func() {
		root := &captaindb.Session{ID: uuid.New(), Source: "gavel"}
		transcript := &captaindb.Session{ID: uuid.New(), Source: "claude", StateReason: reason}
		store = &verificationSessionStore{
			fakeCaptainSessionStore: fakeCaptainSessionStore{
				run: root, transcript: transcript,
				overview: &captaindb.SessionOverview{LifecycleStatus: "failed", HealthState: "healthy", ActivityState: "idle"},
				runs:     []captaindb.PromptRun{{ID: uuid.New(), SessionID: root.ID, State: captaindb.PromptRunStateFailed, Phase: captaindb.PromptRunPhaseVerify, CurrentIteration: iteration}},
			},
			report:    &api.VerifyReport{Kind: "todo", Ran: true, State: api.VerifyStateFailed, Reason: reason},
			iteration: iteration,
		}
		resolved = captainSessionResolution{run: root, transcript: transcript, known: true}
	})

	It("returns verification failure evidence without a session error", func(ctx SpecContext) {
		stats, err := captainSessionStats(ctx, store, "provider-session", resolved)
		Expect(err).NotTo(HaveOccurred())
		Expect(stats.State).To(Equal("verification_failed"))
		Expect(stats.Error).To(BeEmpty())
		Expect(stats.Verification).To(Equal(store.report))
		Expect(stats.InProgress).To(BeFalse())
		Expect(store.requestedRun).To(Equal(store.runs[0].ID))
	})

	DescribeTable("preserves operational failures",
		func(ctx SpecContext, change func()) {
			change()
			stats, err := captainSessionStats(ctx, store, "provider-session", resolved)
			Expect(err).NotTo(HaveOccurred())
			Expect(stats.State).To(Equal("error"))
			Expect(stats.Error).To(Equal(reason))
			Expect(stats.Verification).To(BeNil())
		},
		Entry("an agent generation failure", func() { store.runs[0].Phase = captaindb.PromptRunPhaseGenerate }),
		Entry("a post-run failure after failed checks", func() { store.runs[0].Phase = captaindb.PromptRunPhasePostRun }),
		Entry("an earlier verification iteration", func() { store.iteration = iteration - 1 }),
		Entry("no verification report", func() { store.report = nil }),
		Entry("a verifier execution error", func() { store.report.State = api.VerifyStateErrored }),
		Entry("a passing verification before an operational failure", func() { store.report.State = api.VerifyStatePassed; store.report.Passed = true }),
		Entry("a zombie session with failed checks", func() { store.overview.HealthState = "zombie" }),
		Entry("a stalled failed session with failed checks", func() { store.overview.HealthState = "stalled" }),
	)

	It("surfaces verification storage errors", func(ctx SpecContext) {
		store.verificationErr = errors.New("verification storage unavailable")
		_, err := captainSessionStats(ctx, store, "provider-session", resolved)
		Expect(err).To(MatchError(store.verificationErr))
	})

	It("classifies failed checks before a transcript has been ingested", func(ctx SpecContext) {
		resolved.transcript = nil
		stats, err := captainSessionStats(ctx, store, "provider-session", resolved)
		Expect(err).NotTo(HaveOccurred())
		Expect(stats.State).To(Equal("verification_failed"))
		Expect(stats.Error).To(BeEmpty())
		Expect(stats.Verification).To(Equal(store.report))
	})
})
