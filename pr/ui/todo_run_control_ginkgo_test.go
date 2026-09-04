package ui

import (
	"context"

	"github.com/flanksource/gavel/todos/run"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("todo run control", func() {
	It("cancels an owned headless run once and reports its stopping state", func() {
		registry := run.NewRegistry()
		issueID, promptRunID := uuid.New(), uuid.New()
		runContext, cancel := context.WithCancelCause(context.Background())

		handle, err := registry.Register(run.RegisterOptions{
			IssueIDs: []uuid.UUID{issueID}, Stoppable: true, Cancel: cancel,
		})
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(handle.Release)
		handle.BindPromptRun(promptRunID)
		Expect(registry.Status(issueID)).To(Equal(run.ControlStatus{CanStop: true}))

		Expect(registry.Stop(promptRunID)).To(Succeed())
		Expect(context.Cause(runContext)).To(MatchError(run.ErrStopped))
		Expect(registry.Status(issueID)).To(Equal(run.ControlStatus{Stopping: true}))
		Expect(registry.Stop(promptRunID)).To(MatchError(run.ErrStopping))
	})

	It("does not advertise a run whose driver cannot be interrupted", func() {
		registry := run.NewRegistry()
		issueID, promptRunID := uuid.New(), uuid.New()
		_, cancel := context.WithCancelCause(context.Background())

		handle, err := registry.Register(run.RegisterOptions{
			IssueIDs: []uuid.UUID{issueID}, Stoppable: false, Cancel: cancel,
		})
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(handle.Release)
		handle.BindPromptRun(promptRunID)
		Expect(registry.Status(issueID)).To(Equal(run.ControlStatus{}))
		Expect(registry.Stop(promptRunID)).To(MatchError(run.ErrNotStoppable))
	})

	It("refuses a second run on one todo but admits a confirmed concurrent one", func() {
		registry := run.NewRegistry()
		issueID := uuid.New()
		_, cancel := context.WithCancelCause(context.Background())
		opts := run.RegisterOptions{IssueIDs: []uuid.UUID{issueID}, Stoppable: true, Cancel: cancel}

		first, err := registry.Register(opts)
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(first.Release)

		_, err = registry.Register(opts)
		Expect(err).To(MatchError(run.ErrAlreadyOwned))

		opts.Concurrent = true
		second, err := registry.Register(opts)
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(second.Release)
	})

	It("stops only the run named by the prompt run id", func() {
		registry := run.NewRegistry()
		issueID := uuid.New()
		firstRunID, secondRunID := uuid.New(), uuid.New()
		firstCtx, cancelFirst := context.WithCancelCause(context.Background())
		secondCtx, cancelSecond := context.WithCancelCause(context.Background())

		first, err := registry.Register(run.RegisterOptions{
			IssueIDs: []uuid.UUID{issueID}, Stoppable: true, Cancel: cancelFirst,
		})
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(first.Release)
		first.BindPromptRun(firstRunID)

		second, err := registry.Register(run.RegisterOptions{
			IssueIDs: []uuid.UUID{issueID}, Stoppable: true, Concurrent: true, Cancel: cancelSecond,
		})
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(second.Release)
		second.BindPromptRun(secondRunID)

		Expect(registry.Stop(secondRunID)).To(Succeed())
		Expect(context.Cause(secondCtx)).To(MatchError(run.ErrStopped))
		Expect(firstCtx.Err()).NotTo(HaveOccurred())
	})
})
