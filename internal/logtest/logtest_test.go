// Tests that Capture collects the global logger's lines and restores the logger afterwards.
package logtest

import (
	"testing"

	"github.com/flanksource/commons/logger"
	"github.com/stretchr/testify/assert"
)

func TestCaptureCollectsLinesAndRestoresTheLogger(t *testing.T) {
	previous := logger.StandardLogger()
	t.Run("captured", func(t *testing.T) {
		logged := Capture(t)
		logger.Warnf("credential fetch refused")
		assert.Contains(t, logged.String(), "credential fetch refused")
	})
	assert.Same(t, previous, logger.StandardLogger())
}
