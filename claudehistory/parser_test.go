package claudehistory

import "testing"

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple path with slashes",
			input:    "/Users/moshe/projects",
			expected: "-Users-moshe-projects",
		},
		{
			name:     "path with dots in directory names",
			input:    "/Users/moshe/go/src/github.com/flanksource/arch-unit-todos",
			expected: "-Users-moshe-go-src-github-com-flanksource-arch-unit-todos",
		},
		{
			name:     "path with multiple dots",
			input:    "/home/user/.config/app.json",
			expected: "-home-user--config-app-json",
		},
		{
			name:     "empty path",
			input:    "",
			expected: "",
		},
		{
			name:     "path with consecutive dots",
			input:    "/path/to/file..txt",
			expected: "-path-to-file--txt",
		},
		{
			name:     "relative path",
			input:    "./relative/path.go",
			expected: "--relative-path-go",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizePath(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizePath(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
