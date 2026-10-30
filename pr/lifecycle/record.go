package lifecycle

import (
	"context"
	"fmt"
	"slices"

	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/captain/pkg/promptrun"
	workflow "github.com/flanksource/gavel/lifecycle"
	"github.com/google/uuid"
)

func (h Host) record(snapshot Snapshot, action Action, runs []captaindb.PromptRun) *promptrun.Recording {
	session := SessionID(snapshot.PR.URL)
	return &promptrun.Recording{
		DB: h.DB, PromptRunID: uuid.New(), Origin: "gavel.pr.lifecycle", SpecProfile: action.Step.Name,
		Placement: promptrun.Placement{Sessions: []captaindb.CreateSessionInput{{ID: session, Source: "gavel.pr.lifecycle", Title: snapshot.PR.Title, CWD: h.Dir, Project: h.Repo.FullName(), Metadata: map[string]any{"prURL": snapshot.PR.URL}}}},
		Link: func(ctx context.Context, tx *captaindb.DB, run *captaindb.PromptRun) error {
			metadata := map[string]any{"lifecycle": h.Definition, "context": action.Event, "subject": snapshot.Subject(), "headSHA": snapshot.PR.HeadSHA}
			updated, err := tx.UpdatePromptRun(ctx, captaindb.UpdatePromptRunInput{ID: run.ID, ExpectedVersion: run.Version, Metadata: &metadata})
			if err == nil {
				*run = *updated
			}
			return err
		},
		Outcome: func(result promptrun.Result, runErr error, stopped bool) (promptrun.Outcome, error) {
			out, err := promptrun.DefaultOutcome(result, runErr, stopped)
			if err != nil || runErr != nil || stopped {
				return out, err
			}
			success, ok := out.JSON["success"].(bool)
			if !ok {
				return out, fmt.Errorf("PR lifecycle output needs boolean success")
			}
			event, ok := out.JSON["event"].(string)
			if !ok || !slices.Contains(action.Step.Events, event) {
				return out, fmt.Errorf("PR lifecycle output event %q is not declared", event)
			}
			comment, ok := out.JSON["comment"].(string)
			if !ok {
				return out, fmt.Errorf("PR lifecycle output needs string comment")
			}
			engine, err := workflow.New(h.Definition)
			if err != nil {
				return out, err
			}
			state, err := engine.Outcome(action.Step, workflow.Context{Subject: snapshot.Subject(), Event: action.Event.Variables(), Runs: stepRuns(runs)}, workflow.StepResult{Run: workflow.RunFacts{State: string(out.State)}, Verify: result.Report, Envelope: workflow.Envelope{Success: success, Event: event, Comment: comment, Extra: out.JSON}})
			if err != nil {
				return out, err
			}
			if out.State == captaindb.PromptRunStateSucceeded {
				out.State = captaindb.PromptRunState(state)
			}
			out.Text = comment
			if out.State == captaindb.PromptRunStateFailed && out.Error == "" {
				out.Error = comment
			}
			return out, nil
		},
	}
}
