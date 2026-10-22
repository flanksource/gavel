// Package retention deletes run data older than its configured window: rows and the blobs they
// name. Windows are properties (retention.<category>) accepting durations such as 60d.
package retention

import (
	"fmt"
	"strings"
	"time"

	"github.com/flanksource/commons/duration"
	"github.com/flanksource/commons/properties"
)

// Window reads the duration property key, or returns def when it is unset. A value that isn't a
// positive duration is an error rather than a silent default.
func Window(key string, def time.Duration) (time.Duration, error) {
	value := strings.TrimSpace(properties.String("", key))
	if value == "" {
		return def, nil
	}
	parsed, err := duration.ParseDuration(value)
	if err != nil || time.Duration(parsed) <= 0 {
		return 0, fmt.Errorf("property %s=%s: want a positive duration such as 60d or 36h", key, value)
	}
	return time.Duration(parsed), nil
}
