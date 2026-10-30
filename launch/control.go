package launch

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/flanksource/gavel/resultstore"
	"github.com/flanksource/gavel/todos/native"
)

func (s *Service) persistActive(ctx context.Context, active *activeExecution) (resultstore.Run, error) {
	for range 4 {
		stored, err := s.options.Store.Get(ctx, active.run.Metadata.ExecutionID)
		if err != nil {
			return resultstore.Run{}, err
		}
		active.run.StopRequested = stored.StopRequested
		if stored.StopRequested {
			active.cancel()
			if active.run.Terminal() {
				active.run.Status.State = "cancelled"
			}
		}
		updated, err := s.options.Store.Update(ctx, active.run, stored.Metadata.Revision)
		if !errors.Is(err, resultstore.ErrConflict) {
			return updated, err
		}
	}
	return resultstore.Run{}, fmt.Errorf("persist execution after concurrent control updates: %w", resultstore.ErrConflict)
}

func (s *Service) watch(ctx context.Context, active *activeExecution, publish Publish) {
	active.mu.Lock()
	id := active.run.Metadata.ExecutionID
	active.mu.Unlock()
	control := time.NewTicker(500 * time.Millisecond)
	heartbeat := time.NewTicker(native.OwnerHeartbeatInterval)
	defer control.Stop()
	defer heartbeat.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-heartbeat.C:
			if err := publish(resultstore.Run{}); err != nil {
				return
			}
		case <-control.C:
			run, err := s.options.Store.Get(ctx, id)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				active.mu.Lock()
				active.fault = fmt.Errorf("read execution control: %w", err)
				active.mu.Unlock()
				active.cancel()
				return
			}
			if run.StopRequested {
				active.cancel()
				return
			}
		}
	}
}
