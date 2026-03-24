package fixtures

import (
	"testing"
)

func TestBuildANSIContext(t *testing.T) {
	tests := []struct {
		name       string
		stdout     string
		hasAny     bool
		hasColor   bool
		hasUpdates bool
	}{
		{
			name:   "plain text",
			stdout: "Hello, world!",
		},
		{
			name:   "bold only",
			stdout: "\x1b[1mBold\x1b[0m",
			hasAny: true,
		},
		{
			name:     "foreground color 8-color",
			stdout:   "\x1b[31mRed\x1b[0m",
			hasAny:   true,
			hasColor: true,
		},
		{
			name:     "foreground 24-bit RGB",
			stdout:   "\x1b[38;2;239;68;68mRed\x1b[0m",
			hasAny:   true,
			hasColor: true,
		},
		{
			name:     "foreground 256-color",
			stdout:   "\x1b[38;5;196mRed\x1b[0m",
			hasAny:   true,
			hasColor: true,
		},
		{
			name:     "background color",
			stdout:   "\x1b[42mGreen BG\x1b[0m",
			hasAny:   true,
			hasColor: true,
		},
		{
			name:       "cursor up",
			stdout:     "\x1b[2A",
			hasAny:     true,
			hasUpdates: true,
		},
		{
			name:       "clear screen",
			stdout:     "\x1b[2J",
			hasAny:     true,
			hasUpdates: true,
		},
		{
			name:       "erase line",
			stdout:     "\x1b[K",
			hasAny:     true,
			hasUpdates: true,
		},
		{
			name:       "hide cursor",
			stdout:     "\x1b[?25l",
			hasAny:     true,
			hasUpdates: true,
		},
		{
			name:     "mixed color and cursor",
			stdout:   "\x1b[31mRed\x1b[0m\x1b[2J",
			hasAny:   true,
			hasColor: true,
			hasUpdates: true,
		},
		{
			name:     "color in stderr only",
			stdout:   "plain",
			hasAny:   true,
			hasColor: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stderr := ""
			if tt.name == "color in stderr only" {
				stderr = "\x1b[31mError\x1b[0m"
			}
			ctx := BuildANSIContext(tt.stdout, stderr)

			if ctx["has_any"] != tt.hasAny {
				t.Errorf("has_any: got %v, want %v", ctx["has_any"], tt.hasAny)
			}
			if ctx["has_color"] != tt.hasColor {
				t.Errorf("has_color: got %v, want %v", ctx["has_color"], tt.hasColor)
			}
			if ctx["has_updates"] != tt.hasUpdates {
				t.Errorf("has_updates: got %v, want %v", ctx["has_updates"], tt.hasUpdates)
			}
		})
	}
}
