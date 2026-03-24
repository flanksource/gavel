package fixtures

import (
	"testing"
)

func TestANSIDetection(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		hasAny     bool
		hasColor   bool
		hasUpdates bool
	}{
		{name: "plain text", input: "Hello, world!"},
		{name: "bold only", input: "\x1b[1mBold\x1b[0m", hasAny: true},
		{name: "foreground color 8-color", input: "\x1b[31mRed\x1b[0m", hasAny: true, hasColor: true},
		{name: "foreground 24-bit RGB", input: "\x1b[38;2;239;68;68mRed\x1b[0m", hasAny: true, hasColor: true},
		{name: "foreground 256-color", input: "\x1b[38;5;196mRed\x1b[0m", hasAny: true, hasColor: true},
		{name: "background color", input: "\x1b[42mGreen BG\x1b[0m", hasAny: true, hasColor: true},
		{name: "cursor up", input: "\x1b[2A", hasAny: true, hasUpdates: true},
		{name: "clear screen", input: "\x1b[2J", hasAny: true, hasUpdates: true},
		{name: "erase line", input: "\x1b[K", hasAny: true, hasUpdates: true},
		{name: "hide cursor", input: "\x1b[?25l", hasAny: true, hasUpdates: true},
		{name: "mixed color and cursor", input: "\x1b[31mRed\x1b[0m\x1b[2J", hasAny: true, hasColor: true, hasUpdates: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasAnyANSI(tt.input); got != tt.hasAny {
				t.Errorf("hasAnyANSI: got %v, want %v", got, tt.hasAny)
			}
			if got := hasColorCodes(tt.input); got != tt.hasColor {
				t.Errorf("hasColorCodes: got %v, want %v", got, tt.hasColor)
			}
			if got := hasCursorUpdates(tt.input); got != tt.hasUpdates {
				t.Errorf("hasCursorUpdates: got %v, want %v", got, tt.hasUpdates)
			}
		})
	}
}
