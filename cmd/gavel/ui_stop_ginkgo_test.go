package main

import (
	"context"

	"github.com/flanksource/clicky"
	clickytask "github.com/flanksource/clicky/task"
	flanksourceContext "github.com/flanksource/commons/context"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("standalone UI stop", func() {
	It("cancels the run context and every clicky global task", func() {
		started := make(chan struct{})
		handle := clicky.StartTask("standalone ui stop target", func(ctx flanksourceContext.Context, _ *clickytask.Task) (struct{}, error) {
			close(started)
			<-ctx.Done()
			return struct{}{}, ctx.Err()
		})
		DeferCleanup(func() {
			clicky.CancelAllGlobalTasks()
			clicky.WaitForGlobalCompletionSilent()
			clicky.ClearGlobalTasks()
		})
		Eventually(started).Should(BeClosed())

		runCtx, cancelRun := context.WithCancel(context.Background())
		standaloneUIStop(cancelRun)()

		Expect(runCtx.Err()).To(MatchError(context.Canceled))
		Expect(handle.WaitFor().Status).To(Equal(clickytask.StatusCancelled))
	})
})
