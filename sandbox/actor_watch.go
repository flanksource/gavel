// Watches a provisioning run's substrate actor for the worker substrate places it on, recording
// that as a substrate event of the run: substrate reports it only when asked.
package sandbox

import (
	"context"
	"fmt"
	"time"

	"github.com/flanksource/captain/pkg/substrate"
	"github.com/flanksource/commons/logger"
	"github.com/google/uuid"
)

// ActorReader is an executor that reads the substrate actor of a run's environment.
type ActorReader interface {
	GetActor(ctx context.Context, environment string) (substrate.Actor, error)
}

// GetActor reads the actor of environment over the executor's connection.
func (e *substrateExecutor) GetActor(ctx context.Context, environment string) (substrate.Actor, error) {
	return e.client.GetActor(ctx, environment)
}

// actorPollInterval is how often a provisioning run's actor is read.
const actorPollInterval = 3 * time.Second

// watchActor records, as an "assigned" substrate event of run id, each worker substrate places
// the run's actor on, until ctx ends: once the run starts or ends. It does nothing when the
// dispatcher's executor can't read actors.
func (d *Dispatcher) watchActor(ctx context.Context, id uuid.UUID) {
	reader, ok := d.Executor.(ActorReader)
	if !ok {
		return
	}
	ticker := time.NewTicker(actorPollInterval)
	defer ticker.Stop()
	worker := ""
	for {
		actor, err := reader.GetActor(ctx, JobName(id))
		switch {
		case ctx.Err() != nil:
			return
		case err != nil:
			// The actor may not be created yet, or substrate may be briefly unreachable.
			logger.Debugf("sandbox run %s: %v", id, err)
		case actor.Worker != "" && actor.Worker != worker:
			worker = actor.Worker
			d.substrateEvent(context.Background(), id, "assigned",
				fmt.Sprintf("placed on worker %s in pool %s on node %s (actor %s)", actor.Worker, actor.Pool, actor.Node, actor.State))
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

// substrateEvent records what substrate reported of run id's environment while the run still
// provisions: holding the run's row, so an event never lands after the run has ended.
func (d *Dispatcher) substrateEvent(ctx context.Context, id uuid.UUID, kind, message string) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := d.DB.WithContext(ctx).Exec(`INSERT INTO sandbox_run_events (run_id, source, kind, message)
		SELECT id, 'substrate', ?, ? FROM sandbox_runs WHERE id = ? AND status = 'provisioning' FOR UPDATE`,
		kind, message, id).Error; err != nil {
		logger.Warnf("record substrate event %s of sandbox run %s: %v", kind, id, err)
	}
}
