package fixtures

import (
	"regexp"
	"strings"
)

var (
	colorCodeRegex   = regexp.MustCompile(`\x1b\[(3[0-7]|38;[25];|4[0-7]|48;[25];)`)

	cursorUpdateCode = regexp.MustCompile(`\x1b\[([0-9]*[ABCDHJ]|[0-9]*K|2J|\?25[hl])`)
)

func hasAnyANSI(s string) bool {
	return strings.Contains(s, "\x1b[")
}

func hasColorCodes(s string) bool {
	return colorCodeRegex.MatchString(s)
}

func hasCursorUpdates(s string) bool {
	return cursorUpdateCode.MatchString(s)
}

func BuildANSIContext(stdout, stderr string) map[string]bool {
	combined := stdout + stderr
	return map[string]bool{
		"has_color":   hasColorCodes(combined),
		"has_any":     hasAnyANSI(combined),
		"has_updates": hasCursorUpdates(combined),
	}
}
