package todos

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/todos/types"
)

const progressPersistenceInterval = 500 * time.Millisecond

type progressPersister struct {
	mu        sync.Mutex
	ctx       context.Context
	provider  RunProgressProvider
	todo      *types.TODO
	lastWrite time.Time
	pending   *fixtures.ExecutionSnapshot
	timer     *time.Timer
	writeErr  error
	closed    bool
}

func newProgressPersister(ctx context.Context, provider RunProgressProvider, todo *types.TODO) *progressPersister {
	return &progressPersister{ctx: ctx, provider: provider, todo: todo}
}

func (p *progressPersister) Sink(_ context.Context, snapshot fixtures.ExecutionSnapshot) error {
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
		return p.writeLocked(snapshot)
	}
	p.pending = &snapshot
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
		snapshot := *p.pending
		p.pending = nil
		_ = p.writeLocked(snapshot)
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
	snapshot := *p.pending
	p.pending = nil
	_ = p.writeLocked(snapshot)
}

func (p *progressPersister) writeLocked(snapshot fixtures.ExecutionSnapshot) error {
	persistCtx, cancel := providerPersistenceContext(p.ctx)
	defer cancel()
	if err := p.provider.RecordRunProgress(persistCtx, p.todo, snapshot); err != nil {
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
