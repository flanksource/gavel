package ui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/flanksource/gavel/lifecycle"
	"github.com/flanksource/gavel/todos"
	lifecycle1 "github.com/flanksource/gavel/todos/lifecycle"
	"github.com/flanksource/gavel/todos/types"
)

// todoCreateStepsResponse is what a create dialog can start on the todo it is
// about to save.
type todoCreateStepsResponse struct {
	Steps []todoCreateStepOption `json:"steps"`
}

// todoCreateStepOption is one lifecycle step that applies to a new todo.
type todoCreateStepOption struct {
	Name      string `json:"name"`
	Label     string `json:"label"`
	Prompt    string `json:"prompt"`
	Auxiliary bool   `json:"auxiliary"`
}

func (s *Server) handleTodoNewSteps(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	status := types.Status(strings.TrimSpace(r.URL.Query().Get("status")))
	if status == "" {
		status = types.StatusDraft
	}
	if err := types.ValidateAssignableStatus(status); err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	provider, source, err := s.todoProviderContext(r.Context(), todoSourceFromRequest(r))
	if err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	steps, err := todoCreateSteps(provider, source.Dir, status)
	if err != nil {
		writeTodoError(w, http.StatusInternalServerError, err)
		return
	}
	response := todoCreateStepsResponse{Steps: make([]todoCreateStepOption, 0, len(steps))}
	for _, step := range steps {
		response.Steps = append(response.Steps, todoCreateStepOption{
			Name: step.Name, Label: stepLabel(step.Name), Prompt: step.Prompt, Auxiliary: step.Auxiliary,
		})
	}
	json.NewEncoder(w).Encode(response) //nolint:errcheck
}

// todoCreateSteps evaluates the workspace's lifecycle against a todo that does
// not exist yet: the given status, medium priority, no labels, body, attempts,
// plan, verification or runs, projected the way the lifecycle projects a saved one.
func todoCreateSteps(provider todos.Provider, dir string, status types.Status) ([]lifecycle.Step, error) {
	host, err := lifecycle1.NewHost(provider, dir, lifecycle1.HostDashboard)
	if err != nil {
		return nil, err
	}
	subject, err := host.UnsavedSubject(&types.TODO{TODOFrontmatter: types.TODOFrontmatter{Status: status, Priority: types.PriorityMedium}})
	if err != nil {
		return nil, err
	}
	return host.Def.Applicable(lifecycle.Context{Subject: subject})
}

// validateTodoCreateStep rejects a step the lifecycle does not apply to a new
// todo in this status, naming the steps it does.
func validateTodoCreateStep(provider todos.Provider, dir string, status types.Status, name string) error {
	steps, err := todoCreateSteps(provider, dir, status)
	if err != nil {
		return err
	}
	names := make([]string, 0, len(steps))
	for _, step := range steps {
		names = append(names, step.Name)
	}
	if !slices.Contains(names, name) {
		return fmt.Errorf("step %q does not apply to a new %s todo; applicable steps: %s",
			name, status, strings.Join(names, ", "))
	}
	return nil
}
