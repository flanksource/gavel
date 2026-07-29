package types

import (
	"strings"
	"testing"
)

func TestIsAssignableStatus(t *testing.T) {
	cases := []struct {
		status Status
		want   bool
	}{
		{status: StatusDraft, want: true},
		{status: StatusPending, want: true},
		{status: StatusVerified, want: true},
		{status: StatusCompleted, want: true},
		{status: StatusSkipped, want: true},
		{status: StatusInProgress, want: false},
		{status: StatusReview, want: false},
		{status: StatusAsk, want: false},
		{status: StatusFailed, want: false},
		{status: StatusUnverified, want: false},
		{status: Status("bogus"), want: false},
		{status: Status(""), want: false},
	}
	for _, tc := range cases {
		t.Run(string(tc.status), func(t *testing.T) {
			if got := IsAssignableStatus(tc.status); got != tc.want {
				t.Fatalf("IsAssignableStatus(%q) = %v, want %v", tc.status, got, tc.want)
			}
		})
	}
}

// Every assignable status must round-trip through the known set, otherwise a
// caller-writable status would be rejected by parsers and filters.
func TestAssignableStatusesAreKnown(t *testing.T) {
	for _, status := range AssignableStatuses() {
		if !IsKnownStatus(status) {
			t.Fatalf("assignable status %q is not a known status", status)
		}
	}
}

// A projected status must fail loudly rather than being accepted and dropped by
// the storage layer.
func TestValidateAssignableStatus(t *testing.T) {
	if err := ValidateAssignableStatus(StatusPending); err != nil {
		t.Fatalf("ValidateAssignableStatus(pending): %v", err)
	}

	err := ValidateAssignableStatus(StatusFailed)
	if err == nil {
		t.Fatal("ValidateAssignableStatus(failed) = nil, want error")
	}
	for _, want := range []string{"failed", "draft", "pending", "verified", "completed", "skipped"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("ValidateAssignableStatus(failed) = %q, want it to mention %q", err, want)
		}
	}

	if err := ValidateAssignableStatus(Status("bogus")); err == nil {
		t.Fatal("ValidateAssignableStatus(bogus) = nil, want error")
	}
}

func TestValidatePriority(t *testing.T) {
	for _, priority := range KnownPriorities() {
		if err := ValidatePriority(priority); err != nil {
			t.Fatalf("ValidatePriority(%q): %v", priority, err)
		}
	}
	for _, priority := range []Priority{"critical", "", "urgent"} {
		if err := ValidatePriority(priority); err == nil {
			t.Fatalf("ValidatePriority(%q) = nil, want error", priority)
		}
	}
}
