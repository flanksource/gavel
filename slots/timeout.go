package slots

import (
	"context"
	"sync"
	"time"
)

type timeoutKey struct{}

type timeoutBudget struct {
	sync.Mutex
	timer      *time.Timer
	remaining  time.Duration
	started    time.Time
	paused     int
	expired    bool
	generation uint64
	cancel     context.CancelCauseFunc
	cause      error
	ctx        context.Context
}

type timeoutContext struct {
	context.Context
	budget *timeoutBudget
	done   <-chan struct{}
}

func (c timeoutContext) Done() <-chan struct{} { return c.done }

func (c timeoutContext) Err() error {
	c.budget.Lock()
	expired := c.budget.expired
	c.budget.Unlock()
	if expired {
		return context.DeadlineExceeded
	}
	return c.Context.Err()
}

// Timeout budgets execution time. Acquire pauses all enclosing slot
// budgets while queued; cancellation and the caller's whole-run deadline remain active.
type TimeoutOptions struct {
	Limits   Limits
	Duration time.Duration
	Cause    error
}

func Timeout(parent context.Context, options TimeoutOptions) (context.Context, context.CancelFunc) {
	cause := options.Cause
	if cause == nil {
		cause = context.DeadlineExceeded
	}
	if !options.Limits.Enabled() {
		return context.WithTimeoutCause(parent, options.Duration, cause)
	}
	options.Cause = cause
	ctx, cancel, start := DeferredTimeout(parent, options)
	start()
	return ctx, cancel
}

// DeferredTimeout arms its budget only when start is called by the task worker.
func DeferredTimeout(parent context.Context, options TimeoutOptions) (context.Context, context.CancelFunc, func()) {
	ctx, cancel := context.WithCancelCause(parent)
	cause := options.Cause
	if cause == nil {
		cause = context.DeadlineExceeded
	}
	b := &timeoutBudget{remaining: options.Duration, cancel: cancel, cause: cause, ctx: ctx}
	budgets, _ := parent.Value(timeoutKey{}).([]*timeoutBudget)
	ctx = context.WithValue(ctx, timeoutKey{}, append(append([]*timeoutBudget{}, budgets...), b))
	// A distinct Done channel makes child contexts consult Err instead of
	// inheriting the underlying cancel context's context.Canceled error.
	done := make(chan struct{})
	context.AfterFunc(ctx, func() { close(done) })
	stop := func() {
		b.Lock()
		b.generation++
		if b.timer != nil {
			b.timer.Stop()
		}
		cancel(context.Canceled)
		b.Unlock()
	}
	return timeoutContext{Context: ctx, budget: b, done: done}, stop, func() {
		b.Lock()
		if ctx.Err() == nil {
			b.start()
		}
		b.Unlock()
	}
}

func (b *timeoutBudget) start() {
	b.generation++
	generation := b.generation
	b.started = time.Now()
	b.timer = time.AfterFunc(b.remaining, func() {
		b.Lock()
		defer b.Unlock()
		if b.generation == generation && b.ctx.Err() == nil {
			b.expired = true
			b.cancel(b.cause)
		}
	})
}

func pauseTimeouts(ctx context.Context) func() {
	enclosing, _ := ctx.Value(timeoutKey{}).([]*timeoutBudget)
	var budgets []*timeoutBudget
	for _, b := range enclosing {
		b.Lock()
		if b.timer != nil {
			budgets = append(budgets, b)
		}
		b.Unlock()
	}
	for _, b := range budgets {
		b.Lock()
		if b.paused == 0 {
			b.generation++
			b.timer.Stop()
			b.remaining -= time.Since(b.started)
		}
		b.paused++
		b.Unlock()
	}
	return func() {
		for _, b := range budgets {
			b.Lock()
			b.paused--
			if b.paused == 0 {
				b.start()
			}
			b.Unlock()
		}
	}
}
