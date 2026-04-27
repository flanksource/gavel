package main

import (
	"context"
	"time"
)

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
