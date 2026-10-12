package main

import (
	"context"
	"time"

	"github.com/flanksource/clicky"
)

// standaloneUIStop is the stop callback for gavel's own CLI UI. That process
// hosts exactly one run, so stopping it also cancels every clicky global task
// (the per-package test/lint tasks); testui.Server never does that itself.
func standaloneUIStop(cancel context.CancelFunc) func() {
	return func() {
		cancel()
		clicky.CancelAllGlobalTasks()
	}
}

func newStopContext(parent context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if parent == nil {
		parent = context.Background()
	}

	baseCtx, cancelBase := context.WithCancel(parent)
	if timeout <= 0 {
		return baseCtx, cancelBase
	}

	timeoutCtx, cancelTimeout := context.WithTimeout(baseCtx, timeout)
	return timeoutCtx, func() {
		cancelTimeout()
		cancelBase()
	}
}
