package lifecycle

import (
	"context"
	"errors"
	"fmt"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/captain/pkg/promptrun"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
	"github.com/google/uuid"
)

// progressSink persists the definition of done's in-flight reports. The first
// write that fails is remembered and fails the run once it is over, rather than
// aborting the verification hook mid-tree over a persistence error.
type progressSink struct {
	ctx      context.Context
	provider todos.RunProgressProvider
	todo     *types.TODO
	err      error
}

func (h *Host) progressSink(ctx context.Context, todo *types.TODO) *progressSink {
	sink := &progressSink{ctx: ctx, todo: todo}
	sink.provider, _ = h.Provider.(todos.RunProgressProvider)
	return sink
}

func (s *progressSink) record(report api.VerifyReport) {
	if s.provider == nil || s.err != nil {
		return
	}
	persistCtx, cancel := todos.PersistenceContext(s.ctx)
	defer cancel()
	if err := s.provider.RecordRunProgress(persistCtx, s.todo, report); err != nil {
		s.err = fmt.Errorf("record verification progress: %w", err)
	}
}

// recordIterations files the finished run's per-turn rows — and with them the
// verification report the attempt listing, phase index and run history read —
// under the prompt run it was admitted as. A verify-only step generated
// nothing, but its verdict is filed all the same: that row is the only place
// its report lives. A write that fails is an error on the run, like a progress
// write that failed, and is logged where the run's other persistence failures
// are.
func (h *Host) recordIterations(exec *todos.ExecutorContext, promptRunID uuid.UUID, d *dispatched) {
	provider, ok := h.Provider.(todos.RunIterationProvider)
	if !ok {
		return
	}
	records := promptrun.IterationRecords(d.out, d.cancelled || d.timedOut)
	if len(records) == 0 {
		return
	}
	persistCtx, cancel := todos.PersistenceContext(exec)
	defer cancel()
	if err := provider.RecordRunIterations(persistCtx, promptRunID, records); err != nil {
		exec.Logger.Errorf("failed to record run iterations: %v", err)
		d.err = errors.Join(d.err, fmt.Errorf("record run iterations: %w", err))
	}
}

func setSessionID(todo *types.TODO, sessionID string) {
	if todo == nil || sessionID == "" {
		return
	}
	if todo.LLM == nil {
		todo.LLM = &types.LLM{}
	}
	todo.LLM.SessionId = sessionID
}

// priorSessionID is the todo's recorded agent session, the id a resume reuses.
func priorSessionID(todo *types.TODO) string {
	if todo != nil && todo.LLM != nil {
		return todo.LLM.SessionId
	}
	return ""
}
