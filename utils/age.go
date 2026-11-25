package utils

import (
	"fmt"
	"time"
)

// HumanAge returns a Starship-style relative age ("3m", "5h", "2d") for the
// given duration. Returns "" for zero or negative durations so callers can
// skip rendering when the time is unavailable.
func HumanAge(d time.Duration) string {
	if d <= 0 {
		return ""
	}
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	}
}
