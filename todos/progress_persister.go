package todos

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/flanksource/captain/pkg/api"

	"github.com/flanksource/gavel/todos/types"
)

const progressPersistenceInterval = 500 * time.Millisecond

type progressPersister struct {
	mu        sync.Mutex
	ctx       context.Context
	provider  RunProgressProvider
	todo      *types.TODO
	lastWrite time.Time
	pending   *api.VerifyReport
	timer     *time.Timer
	writeErr  error
	closed    bool
}

func newProgressPersister(ctx context.Context, provider RunProgressProvider, todo *types.TODO) *progressPersister {
	return &progressPersister{ctx: ctx, provider: provider, todo: todo}
}

// Sink is the capverify.Options.Progress callback: every in-flight verification
// report the fixture verifier publishes is persisted, rate-limited to one write
// per interval with the last one guaranteed.
func (p *progressPersister) Sink(report api.VerifyReport) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return fmt.Errorf("verification progress persister is closed")
	}
	if p.writeErr != nil {
		return p.writeErr
	}
	if p.lastWrite.IsZero() || time.Since(p.lastWrite) >= progressPersistenceInterval {
		p.stopTimerLocked()
		return p.writeLocked(report)
	}
	p.pending = &report
	if p.timer == nil {
		delay := progressPersistenceInterval - time.Since(p.lastWrite)
		p.timer = time.AfterFunc(delay, p.flush)
	}
	return nil
}

func (p *progressPersister) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.stopTimerLocked()
	if p.pending != nil && p.writeErr == nil {
		report := *p.pending
		p.pending = nil
		_ = p.writeLocked(report)
	}
	p.closed = true
	return p.writeErr
}

func (p *progressPersister) flush() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.timer = nil
	if p.closed || p.pending == nil || p.writeErr != nil {
		return
	}
	report := *p.pending
	p.pending = nil
	_ = p.writeLocked(report)
}

func (p *progressPersister) writeLocked(report api.VerifyReport) error {
	persistCtx, cancel := PersistenceContext(p.ctx)
	defer cancel()
	if err := p.provider.RecordRunProgress(persistCtx, p.todo, report); err != nil {
		p.writeErr = fmt.Errorf("persist verification progress: %w", err)
		return p.writeErr
	}
	p.lastWrite = time.Now()
	return nil
}

func (p *progressPersister) stopTimerLocked() {
	if p.timer != nil {
		p.timer.Stop()
		p.timer = nil
	}
}
