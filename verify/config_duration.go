package verify

import (
	"fmt"
	"strings"
	"time"
)

// Timeouts resolves the configured deadlines. A malformed value is an error
// rather than a silent fall back to the default: a repo that asked for a longer
// deadline and quietly kept the short one fails as a killed suite much later,
// where the cause is no longer visible.
func (t TestConfig) Timeouts() (global, test, lint time.Duration, err error) {
	if global, err = parseConfiguredDuration("test.timeout", t.Timeout); err != nil {
		return 0, 0, 0, err
	}
	if test, err = parseConfiguredDuration("test.testTimeout", t.TestTimeout); err != nil {
		return 0, 0, 0, err
	}
	if lint, err = parseConfiguredDuration("test.lintTimeout", t.LintTimeout); err != nil {
		return 0, 0, 0, err
	}
	return global, test, lint, nil
}

// parseConfiguredDuration reads one duration field, treating empty as unset.
func parseConfiguredDuration(field, value string) (time.Duration, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return 0, nil
	}
	parsed, err := time.ParseDuration(trimmed)
	if err != nil {
		return 0, fmt.Errorf("%s: %q is not a duration (use a Go duration such as 20m or 90s)", field, value)
	}
	if parsed <= 0 {
		return 0, fmt.Errorf("%s: %q must be positive", field, value)
	}
	return parsed, nil
}
