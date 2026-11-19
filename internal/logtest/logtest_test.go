// Tests that Capture collects the global logger's lines, from any goroutine, and restores the
// logger afterwards.
package logtest

import (
	"sync"
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

func TestCaptureCanBeReadWhileAnotherGoroutineLogs(t *testing.T) {
	logged := Capture(t)
	var logging sync.WaitGroup
	logging.Go(func() {
		for range 200 {
			logger.Warnf("background sweep failed")
		}
	})
	for range 200 {
		_ = logged.String()
	}
	logging.Wait()
	assert.Contains(t, logged.String(), "background sweep failed")
}
