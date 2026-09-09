package ui

import (
	"fmt"
	"strings"

	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
)

// activeStepFor is the lifecycle step a todo's current attempt runs — the step
// an answer resumes. The provider's phase index marks it: the phase whose latest
// run is the todo's active pointer, or one parked at waiting, which is where an
// ask outcome leaves its run. Exactly one phase must match. None means there is
// no turn to resume; two means the pointer and the index disagree, and picking
// one — as iterating the map used to, in a different order per request — would
// resume a step by chance. Neither is a choice this handler may make on its
// own, so both are errors rather than a fallback to the class the last run
// reported, which is not a step name a project's lifecycle need declare.
func activeStepFor(todo *types.TODO) (string, error) {
	var candidates []string
	for _, phase := range todo.PhaseRuns.Ordered() {
		if phase.Active || phase.State == string(captaindb.PromptRunStateWaiting) {
			candidates = append(candidates, string(phase.Phase))
		}
	}
	switch len(candidates) {
	case 1:
		return candidates[0], nil
	case 0:
		return "", fmt.Errorf("todo %s has no active or waiting step run to resume", todos.TODOReference(todo))
	default:
		return "", fmt.Errorf("todo %s has %d step runs that could be resumed (%s); stop the ones that should not continue",
			todos.TODOReference(todo), len(candidates), strings.Join(candidates, ", "))
	}
}
