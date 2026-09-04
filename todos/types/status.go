package types

import (
	"fmt"
	"strings"
)

// AssignableStatuses returns the statuses a human, CLI, or API caller may write
// directly. The remaining known statuses (in_progress, review, ask, failed,
// unverified) are projected from the execution state of the last run; storage
// owns them, so accepting one from a caller would change nothing.
func AssignableStatuses() []Status {
	return []Status{
		StatusDraft,
		StatusPending,
		StatusVerified,
		StatusCompleted,
		StatusSkipped,
	}
}

// IsAssignableStatus reports whether a caller may write status directly.
func IsAssignableStatus(status Status) bool {
	for _, assignable := range AssignableStatuses() {
		if status == assignable {
			return true
		}
	}
	return false
}

// ValidateAssignableStatus rejects a status a caller cannot write, separating
// "you meant a run projection" from "you typed something unknown" so the write
// fails loudly instead of being silently declined by storage.
func ValidateAssignableStatus(status Status) error {
	if IsAssignableStatus(status) {
		return nil
	}
	if IsKnownStatus(status) {
		return fmt.Errorf("status %q is projected from the last run and cannot be assigned; assignable statuses: %s",
			status, joinStatuses(AssignableStatuses()))
	}
	return fmt.Errorf("unknown status %q; assignable statuses: %s", status, joinStatuses(AssignableStatuses()))
}

// KnownPriorities returns the priorities accepted by parsers, the CLI, and the API.
func KnownPriorities() []Priority {
	return []Priority{PriorityHigh, PriorityMedium, PriorityLow}
}

func IsKnownPriority(priority Priority) bool {
	for _, known := range KnownPriorities() {
		if priority == known {
			return true
		}
	}
	return false
}

func ValidatePriority(priority Priority) error {
	if IsKnownPriority(priority) {
		return nil
	}
	names := make([]string, 0, len(KnownPriorities()))
	for _, known := range KnownPriorities() {
		names = append(names, string(known))
	}
	return fmt.Errorf("unknown priority %q; valid priorities: %s", priority, strings.Join(names, ", "))
}

func joinStatuses(statuses []Status) string {
	names := make([]string, 0, len(statuses))
	for _, status := range statuses {
		names = append(names, string(status))
	}
	return strings.Join(names, ", ")
}
