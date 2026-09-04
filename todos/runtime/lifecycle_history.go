package runtime

import (
	"context"
	"fmt"

	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/lifecycle"
	"github.com/flanksource/gavel/todos/types"
)

var _ todos.RunHistoryProvider = (*Provider)(nil)

// RunHistory lists the todo's linked Captain prompt runs oldest first, each
// under the step it was dispatched as and with the status its lifecycle outcome
// recorded. One query per todo: it is read where one todo is being decided
// about, never across a backlog.
func (p *Provider) RunHistory(ctx context.Context, todo *types.TODO) ([]todos.StepRunRecord, error) {
	issueID, err := p.todoID(todo)
	if err != nil {
		return nil, err
	}
	rows, err := p.repository.ListIssueRunHistory(ctx, issueID, lifecycle.EventLifecycleOutcome)
	if err != nil {
		return nil, fmt.Errorf("list run history for issue %s: %w", issueID, err)
	}
	records := make([]todos.StepRunRecord, 0, len(rows))
	for _, row := range rows {
		records = append(records, todos.StepRunRecord{
			Step: string(row.Phase), PromptRunID: row.PromptRunID.String(), State: row.State, Outcome: row.Outcome,
			StartedAt: row.StartedAt, FinishedAt: row.FinishedAt,
		})
	}
	return records, nil
}
