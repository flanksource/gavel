package todos

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	captaincli "github.com/flanksource/captain/pkg/cli"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/todos/types"
)

// ResolveSessionPlan recovers the plan from the todo's agent session via
// captain's canonical plan resolver (the same one the dashboard's Plan tab
// uses), for when the envelope's reported path/content is missing or wrong.
// File-backed sessions return path+content; inline-only sessions return content
// with an empty path.
func ResolveSessionPlan(sessionID string) (path string, content string) {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return "", ""
	}
	// Configure Captain with Gavel's process-owned pool only when a path that
	// can consult Captain persistence is actually used. Unconfigured databases
	// remain disabled and preserve Captain's file-backed plan resolution.
	if _, err := database.Shared(context.Background()); err != nil {
		return "", ""
	}
	res, err := captaincli.RunPlan(captaincli.PlanOptions{SessionID: sessionID})
	if err != nil {
		return "", ""
	}
	if res.OnDisk {
		path = res.Path
	}
	return path, res.Content
}

// ValidatePlanFile verifies the agent-reported native plan file exists and is
// non-empty — the contract a plan run's envelope must satisfy for new/updated
// plans. The path is the agent's own plan-mode file (e.g. under
// ~/.claude/plans/); gavel never writes it.
func ValidatePlanFile(path string) error {
	path = strings.TrimSpace(path)
	if path == "" {
		return fmt.Errorf("plan file path is empty")
	}
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("plan file %s: %w", path, err)
	}
	if info.IsDir() {
		return fmt.Errorf("plan file %s is a directory", path)
	}
	if info.Size() == 0 {
		return fmt.Errorf("plan file %s is empty", path)
	}
	return nil
}

// HasPlan reports whether a todo has a plan worth surfacing in the listing:
// a validated on-disk plan file, or the todo is awaiting review — reached
// even when the plan is inline-only with no on-disk path (see
// applyPlanOutcome's PlanNew/PlanUpdated branch, outcome.go).
func HasPlan(todo *types.TODO) bool {
	if todo == nil {
		return false
	}
	if todo.Status == types.StatusReview {
		return true
	}
	if todo.Provider == ProviderDB && todo.PlanStatus != "" {
		return true
	}
	return ValidatePlanFile(todo.PlanPath) == nil
}

// ReadPlanFile reads the todo's recorded plan file. exists is false (with no
// error) when the todo has no recorded plan or the file is gone — callers
// render "no plan" rather than failing a page load over a deleted file.
func ReadPlanFile(path string) (content string, modTime time.Time, exists bool, err error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", time.Time{}, false, nil
	}
	info, statErr := os.Stat(path)
	if os.IsNotExist(statErr) {
		return "", time.Time{}, false, nil
	}
	if statErr != nil {
		return "", time.Time{}, false, fmt.Errorf("plan file %s: %w", path, statErr)
	}
	data, readErr := os.ReadFile(path)
	if readErr != nil {
		return "", time.Time{}, false, fmt.Errorf("plan file %s: %w", path, readErr)
	}
	return string(data), info.ModTime(), true, nil
}
