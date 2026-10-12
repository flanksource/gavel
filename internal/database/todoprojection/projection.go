// Package todoprojection keeps each TODO's activity watermark
// (todo_issues.updated_at) in step with the Captain rows it is projected from.
//
// Gavel never writes a Captain table, so it cannot hook Captain's writes with
// triggers. Captain instead notifies captain_row_change after every committed
// change to a session, prompt run, turn request or iteration; the Projection
// LISTENs there and re-runs the gavel-owned projection functions of
// schema/110_todo_projection_functions.sql, which read Captain rows and write
// only todo_issues. The watermark is therefore eventually consistent: it
// advances shortly after Captain commits, not inside Captain's transaction.
// Execution state is unaffected — todo_issue_runtime derives it at read time.
package todoprojection

import (
	"context"
	"fmt"
	"sync"

	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/commons/logger"
	"gorm.io/gorm"
)

// Projection is one LISTEN on Captain's row-change feed, projecting onto the
// Gavel tables of the same database.
type Projection struct {
	captain   *captaindb.DB
	db        *gorm.DB
	ready     chan struct{}
	readyOnce sync.Once
}

// New builds a Projection over a migrated shared pool; Captain's LISTEN takes
// one dedicated connection from it while Run is active.
func New(db *gorm.DB) (*Projection, error) {
	captain, err := captaindb.Use(db)
	if err != nil {
		return nil, fmt.Errorf("TODO projection: %w", err)
	}
	return &Projection{captain: captain, db: db, ready: make(chan struct{})}, nil
}

// Run blocks until ctx is cancelled (returning ctx.Err()) or the LISTEN fails:
// a projection error, a payload outside Captain's contract, or a connection
// Captain could not re-establish. Every (re)established LISTEN first resyncs
// every watermark, because nothing is delivered while nobody listens.
func (p *Projection) Run(ctx context.Context) error {
	return p.captain.ListenRowChanges(ctx, captaindb.RowChangeListener{
		OnChange: p.apply,
		OnResync: p.resync,
	})
}

// Ready is closed once the first resync has completed, i.e. the watermarks are
// current and every later change is being delivered. It stays closed across
// reconnects and restarts of Run.
func (p *Projection) Ready() <-chan struct{} { return p.ready }

func (p *Projection) resync(ctx context.Context) error {
	var advanced int
	if err := p.db.WithContext(ctx).Raw(`SELECT public.gavel_resync_todo_activity()`).Scan(&advanced).Error; err != nil {
		return fmt.Errorf("resync TODO activity from Captain: %w", err)
	}
	logger.V(1).Infof("TODO projection resynced from Captain: %d TODO(s) advanced", advanced)
	p.readyOnce.Do(func() { close(p.ready) })
	return nil
}

func (p *Projection) apply(ctx context.Context, change captaindb.RowChange) error {
	steps, err := stepsFor(change)
	if err != nil {
		return err
	}
	for _, step := range steps {
		if err := p.db.WithContext(ctx).Exec(stepSQL[step.kind], step.id).Error; err != nil {
			return fmt.Errorf("project Captain %s %s %s onto TODOs (%s %s): %w",
				change.Op, change.Table, change.ID, step.kind, step.id, err)
		}
	}
	return nil
}
