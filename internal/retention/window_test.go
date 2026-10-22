// Tests reading a retention window from properties: the default when unset, day and hour
// durations, and a descriptive error for a value that isn't a positive duration.
package retention

import (
	"testing"
	"time"

	"github.com/flanksource/commons/properties"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWindow(t *testing.T) {
	const key = "retention.window_test"
	t.Cleanup(func() { properties.Set(key, "") })

	window, err := Window(key, 60*24*time.Hour)
	require.NoError(t, err)
	assert.Equal(t, 60*24*time.Hour, window, "unset is the default")

	for value, want := range map[string]time.Duration{"7d": 7 * 24 * time.Hour, "36h": 36 * time.Hour, "1w": 7 * 24 * time.Hour} {
		properties.Set(key, value)
		window, err = Window(key, time.Hour)
		require.NoError(t, err, value)
		assert.Equal(t, want, window, value)
	}

	for _, value := range []string{"sixty days", "0d", "-1d"} {
		properties.Set(key, value)
		_, err = Window(key, time.Hour)
		require.Error(t, err, value)
		assert.Contains(t, err.Error(), key+"="+value)
		assert.Contains(t, err.Error(), "positive duration such as 60d or 36h")
	}
}

// setProperty sets a property for the rest of the test.
func setProperty(t *testing.T, key, value string) {
	t.Helper()
	properties.Set(key, value)
	t.Cleanup(func() { properties.Set(key, "") })
}
