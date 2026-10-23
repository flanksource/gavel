package lifecycle

import (
	"fmt"
	"slices"
	"strings"

	"github.com/flanksource/gavel/lifecycle"
	"github.com/flanksource/gavel/todos/tools"
	"github.com/flanksource/gavel/todos/types"
)

const StepVerify = "verify"

var VerifyEvents = []string{"passed", "failed"}

var validStatuses = map[string]bool{
	string(types.StatusDraft): true, string(types.StatusPending): true, string(types.StatusInProgress): true,
	string(types.StatusReview): true, string(types.StatusVerified): true,
	string(types.StatusUnverified): true, string(types.StatusCompleted): true, string(types.StatusFailed): true,
	lifecycle.OutcomeKeep: true,
}

func Validate(def lifecycle.Lifecycle) error {
	if err := def.Validate(); err != nil {
		return err
	}
	step, ok := def.Step(StepVerify)
	if !ok {
		return fmt.Errorf("lifecycle %s has no %q step; verification must be a step of every lifecycle", def.Name, StepVerify)
	}
	if def.Disabled != nil && slices.Contains(*def.Disabled, StepVerify) {
		return fmt.Errorf("lifecycle %s: cannot disable required step %q", def.Name, StepVerify)
	}
	if step.Prompt != promptRefPrefix+StepVerify {
		return fmt.Errorf("step verify: prompt %q is not supported; the verify step runs the definition of done and must reference todos.verify", step.Prompt)
	}
	if !slices.Equal(step.Events, VerifyEvents) {
		return fmt.Errorf("step verify: events must be %v, the verdicts the definition of done reports", VerifyEvents)
	}
	if len(step.Tools) != 0 {
		return fmt.Errorf("step verify: the verify step runs no agent, so it takes no tools")
	}
	return validateSteps(def.Steps)
}

func validateSteps(steps []lifecycle.Step) error {
	sets := tools.Sets()
	for _, step := range steps {
		for _, set := range step.Tools {
			if _, ok := sets[set]; !ok {
				return fmt.Errorf("step %s: tool set %q is not one of %s", step.Name, set, strings.Join(tools.SetNames(), ", "))
			}
		}
		for i, outcome := range step.Outcomes {
			if !validStatuses[outcome.Status] {
				return fmt.Errorf("step %s: outcome %d status %q is not a todo status", step.Name, i, outcome.Status)
			}
		}
	}
	return nil
}
