package runtime

import (
	"context"
	"fmt"

	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/gavel/todos"
	"github.com/google/uuid"
)

var _ todos.RunIterationProvider = (*Provider)(nil)

// RecordRunIterations writes a run's per-turn rows into
// captain_prompt_run_iterations, the record captain_prompt_run_overview's
// latest_verification_result — and so the attempt listing, the phase index and
// the run history — is read from. Each row is a full statement of its turn
// (the store is last-write-wins), so a replay restates every verdict.
func (p *Provider) RecordRunIterations(ctx context.Context, promptRunID uuid.UUID, records []captaindb.UpsertPromptRunIterationInput) error {
	if promptRunID == uuid.Nil {
		return fmt.Errorf("record run iterations: prompt run ID is required")
	}
	for _, record := range records {
		record.PromptRunID = promptRunID
		if _, err := p.captain.UpsertPromptRunIteration(ctx, record); err != nil {
			return fmt.Errorf("record run iteration %d of %s: %w", record.Iteration, promptRunID, err)
		}
	}
	return nil
}
