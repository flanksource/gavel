// Package runtimetest admits todo runs for tests that need one in a given
// state without dispatching an agent. Captain files every captain_* row, as it
// does for a real dispatch: the runtime clears the run and links it, and
// promptrun.RecordCompleted admits, starts and settles it.
package runtimetest

import (
	"context"
	"errors"
	"fmt"

	"github.com/flanksource/captain/pkg/api"
	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/captain/pkg/promptrun"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
)

// ClaudeCLI is the runtime the admitted runs are filed under: one that leaves a
// transcript, so Captain binds an execution session to the run.
var ClaudeCLI = api.RuntimeOf(api.Anthropic, api.ModeCLI)

// Admit clears a dispatch through the runtime and has Captain admit and file
// the run the way promptrun.Run would, landing it in done's outcome.
func Admit(ctx context.Context, provider todos.RunLifecycleProvider, todo *types.TODO, preparation todos.RunPreparation, done promptrun.Completed) (todos.RunPreparationResult, error) {
	if provider == nil {
		return todos.RunPreparationResult{}, errors.New("runtimetest: a run lifecycle provider is required")
	}
	admission, err := provider.PrepareRun(ctx, todo, preparation)
	if err != nil {
		return todos.RunPreparationResult{}, err
	}
	if admission.Record == nil {
		return todos.RunPreparationResult{}, fmt.Errorf("runtimetest: run %s was cleared without a recording", admission.PromptRunID)
	}
	runID, err := promptrun.RecordCompleted(ctx, admission.Record, done)
	if err != nil {
		return todos.RunPreparationResult{}, err
	}
	if runID != admission.PromptRunID {
		return todos.RunPreparationResult{}, fmt.Errorf("runtimetest: cleared run %s but Captain admitted %s", admission.PromptRunID, runID)
	}
	if err := provider.RunAdmitted(ctx, todo, admission.RunPreparationResult); err != nil {
		return todos.RunPreparationResult{}, err
	}
	return admission.RunPreparationResult, nil
}

// Started is a run its agent is still working, bound to the provider session
// the agent reported.
func Started(providerSessionID string) promptrun.Completed {
	return promptrun.Completed{
		Runtime: ClaudeCLI, ProviderSessionID: providerSessionID,
		Outcome: promptrun.Outcome{State: captaindb.PromptRunStateRunning, Phase: captaindb.PromptRunPhaseGenerate},
	}
}

// Parked is a run that ended by asking: waiting on its answer, with the ask
// envelope as the run's output.
func Parked(providerSessionID string, envelope map[string]any) promptrun.Completed {
	summary, _ := envelope["summary"].(string)
	return promptrun.Completed{
		Runtime: ClaudeCLI, ProviderSessionID: providerSessionID,
		Outcome: promptrun.Outcome{
			State: captaindb.PromptRunStateWaiting, Phase: captaindb.PromptRunPhaseGenerate, Text: summary, JSON: envelope,
		},
	}
}
