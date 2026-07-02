package todos

import (
	"fmt"
	"os"
	"strings"
	"time"
)

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
