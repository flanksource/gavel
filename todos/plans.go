package todos

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	captaincli "github.com/flanksource/captain/pkg/cli"
	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/gavel/todos/types"
)

// ResolveSessionPlan recovers the plan from the todo's agent session via
// captain's canonical, read-only plan resolver (the same one the dashboard's
// Plan tab uses), for when the envelope's reported path/content is missing or
// wrong. It never ingests transcripts: a read must not write Captain's tables.
// File-backed sessions return path+content; inline-only sessions return content
// with an empty path. A session with no plan keeps Captain's
// captaincli.ErrNoPlan in the error chain, for a caller that tolerates exactly
// that; every other resolver error, and a resolved plan with no content, is a
// plain error.
func ResolveSessionPlan(ctx context.Context, captain *captaindb.DB, sessionID string) (path string, content string, err error) {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return "", "", fmt.Errorf("resolve session plan: a session ID is required")
	}
	if captain == nil {
		return "", "", fmt.Errorf("resolve plan of session %s: Captain database handle is nil", sessionID)
	}
	res, err := captaincli.ResolvePlan(ctx, captain, captaincli.PlanOptions{SessionID: sessionID, Source: "all"})
	if err != nil {
		return "", "", fmt.Errorf("resolve plan of session %s: %w", sessionID, err)
	}
	if strings.TrimSpace(res.Content) == "" {
		return "", "", fmt.Errorf("resolve plan of session %s: the resolved plan is empty", sessionID)
	}
	if res.OnDisk {
		path = res.Path
	}
	return path, res.Content, nil
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
