package ui

import (
	"context"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("todo run control", func() {
	It("cancels an owned headless run once and reports its stopping state", func() {
		registry := newTodoRunRegistry()
		issueID := uuid.New()
		runContext, cancel := context.WithCancelCause(context.Background())

		cleanup, err := registry.register([]uuid.UUID{issueID}, true, cancel)
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(cleanup)
		Expect(registry.status(issueID)).To(Equal(todoRunControlStatus{CanStop: true}))

		Expect(registry.stop(issueID)).To(Succeed())
		Expect(context.Cause(runContext)).To(MatchError(errTodoRunStopped))
		Expect(registry.status(issueID)).To(Equal(todoRunControlStatus{Stopping: true}))
		Expect(registry.stop(issueID)).To(MatchError(errTodoRunStopping))
	})

	It("does not advertise a run whose driver cannot be interrupted", func() {
		registry := newTodoRunRegistry()
		issueID := uuid.New()
		_, cancel := context.WithCancelCause(context.Background())

		cleanup, err := registry.register([]uuid.UUID{issueID}, false, cancel)
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(cleanup)
		Expect(registry.status(issueID)).To(Equal(todoRunControlStatus{}))
		Expect(registry.stop(issueID)).To(MatchError(errTodoRunNotStoppable))
	})
})
