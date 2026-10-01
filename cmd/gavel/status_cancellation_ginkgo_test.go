package main

import (
	"context"
	"errors"

	"github.com/flanksource/clicky"
	clickytask "github.com/flanksource/clicky/task"
	flanksourceContext "github.com/flanksource/commons/context"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("status command cancellation", func() {
	It("cancels global tasks when its command context ends", func() {
		started := make(chan struct{})
		stopped := make(chan struct{})
		handle := clicky.StartTask("status cancellation target", func(ctx flanksourceContext.Context, _ *clickytask.Task) (struct{}, error) {
			close(started)
			<-ctx.Done()
			close(stopped)
			return struct{}{}, ctx.Err()
		}, clickytask.WithCancellationDrain())
		DeferCleanup(func() {
			clicky.CancelAllGlobalTasks()
			clicky.WaitForGlobalCompletionSilent()
			clicky.ClearGlobalTasks()
		})
		Eventually(started).Should(BeClosed())

		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := runStatus(ctx, StatusOptions{WorkDir: GinkgoT().TempDir(), NoRepomap: true})

		Expect(errors.Is(err, context.Canceled)).To(BeTrue())
		Eventually(stopped).Should(BeClosed())
		_, taskErr := handle.GetResult()
		Expect(errors.Is(taskErr, context.Canceled)).To(BeTrue())
	})
})
