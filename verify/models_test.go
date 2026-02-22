package verify

import (
	"fmt"
	"testing"
)

type mockAdapter struct {
	Claude
	models []string
	err    error
}

func (m mockAdapter) ListModels() ([]string, error) { return m.models, m.err }

func TestFormatModelHint(t *testing.T) {
	tests := []struct {
		name     string
		models   []string
		err      error
		expected string
	}{
		{
			name:     "with models",
			models:   []string{"claude-sonnet-4", "claude-haiku-4"},
			expected: "Available claude models: claude-sonnet-4, claude-haiku-4",
		},
		{
			name:     "empty list",
			models:   nil,
			expected: "",
		},
		{
			name:     "error returns empty",
			err:      fmt.Errorf("api error"),
			expected: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			adapter := mockAdapter{models: tc.models, err: tc.err}
			got := formatModelHint(adapter)
			if got != tc.expected {
				t.Errorf("got %q, want %q", got, tc.expected)
			}
		})
	}
}

func TestContainsModelError(t *testing.T) {
	tests := []struct {
		msg      string
		expected bool
	}{
		{"model not found: claude-4-5-haiku", true},
		{"Model does not exist", true},
		{"invalid_model_id", true},
		{"not_found_error", true},
		{"connection timed out", false},
		{"permission denied", false},
		{"rate limit exceeded", false},
	}

	for _, tc := range tests {
		t.Run(tc.msg, func(t *testing.T) {
			if got := containsModelError(tc.msg); got != tc.expected {
				t.Errorf("containsModelError(%q) = %v, want %v", tc.msg, got, tc.expected)
			}
		})
	}
}
