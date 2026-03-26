package fixtures

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExpandVars(t *testing.T) {
	vars := map[string]any{
		"GIT_ROOT_DIR": "/home/user/project",
		"GOOS":         "linux",
		"count":        42,
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple $VAR", "$GIT_ROOT_DIR/testdata", "/home/user/project/testdata"},
		{"braced ${VAR}", "${GIT_ROOT_DIR}/testdata", "/home/user/project/testdata"},
		{"unknown var passthrough", "$HOME/files", "$HOME/files"},
		{"mixed known and unknown", "$GIT_ROOT_DIR/$HOME", "/home/user/project/$HOME"},
		{"go template preserved", "{{.GIT_ROOT_DIR}}/foo", "{{.GIT_ROOT_DIR}}/foo"},
		{"mixed $VAR and template", "$GOOS/{{.GIT_ROOT_DIR}}", "linux/{{.GIT_ROOT_DIR}}"},
		{"empty string", "", ""},
		{"no vars", "plain text", "plain text"},
		{"numeric value", "$count items", "42 items"},
		{"multiple same var", "$GOOS-$GOOS", "linux-linux"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, ExpandVars(tt.input, vars))
		})
	}
}
