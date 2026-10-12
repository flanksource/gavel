package lifecycle

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/gavel/todos/triagenew"
	"github.com/flanksource/gavel/todos/types"
)

func (h *Host) newTriageReview(input *stepInput, todo *types.TODO) *triagenew.Review {
	review := &triagenew.Review{Provider: h.Provider, SourceID: todo.ID}
	review.Approved = func(ctx context.Context, proposed map[string]any) error {
		record := input.input.Record
		if record == nil || record.DB == nil {
			return fmt.Errorf("triage.new requires a recorded Captain approval")
		}
		run, err := record.DB.GetPromptRun(ctx, record.PromptRunID)
		if err != nil {
			return err
		}
		requests, err := record.DB.ListTurnRequests(ctx, database.TurnRequestFilter{SessionID: run.SessionID, PromptRunID: &run.ID})
		if err != nil {
			return err
		}
		want, err := json.Marshal(proposed)
		if err != nil {
			return err
		}
		for _, request := range requests {
			if request.State != database.TurnRequestStateApproved || request.Request["tool"] != "triage_review" {
				continue
			}
			got, err := json.Marshal(request.Request["input"])
			if err != nil {
				return err
			}
			if string(got) != string(want) || request.Response["interrupt"] == true {
				continue
			}
			if updated, ok := request.Response["updatedInput"]; ok {
				changed, err := json.Marshal(updated)
				if err != nil {
					return err
				}
				if string(changed) != string(want) {
					continue
				}
			}
			return nil
		}
		return fmt.Errorf("triage.new proposal has no matching durable approval")
	}
	return review
}
