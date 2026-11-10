package lint

import (
	"fmt"
	"sync"
)

// lintProgress counts a run's linter invocations as they are scheduled and
// finish, telling Options.Progress done of total after each and once more when
// the run is done. A report that fails is kept and fails the run.
type lintProgress struct {
	report func(done, total int) error

	mu                    sync.Mutex
	scheduledN, finishedN int
	err                   error
}

func (p *lintProgress) scheduled() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.scheduledN++
}

func (p *lintProgress) finished() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.finishedN++
	p.reportLocked()
}

// done reports the final count and returns the first report that failed.
func (p *lintProgress) done() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.reportLocked()
	return p.err
}

func (p *lintProgress) reportLocked() {
	if p.report == nil || p.err != nil {
		return
	}
	if err := p.report(p.finishedN, p.scheduledN); err != nil {
		p.err = fmt.Errorf("report lint progress: %w", err)
	}
}
