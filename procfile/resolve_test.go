package procfile

import (
	"testing"

	cexec "github.com/flanksource/clicky/exec"
	"github.com/flanksource/gavel/verify"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolvePolicy(t *testing.T) {
	cfg := verify.ProcfileConfig{AutoRestart: verify.RestartAlways, MaxRestarts: 3}

	t.Run("entry overrides global", func(t *testing.T) {
		ten := 10
		p, maxR := resolvePolicy(cfg, Entry{AutoRestart: verify.RestartOnFailure, MaxRestarts: &ten})
		assert.Equal(t, cexec.RestartOnFailure, p)
		assert.Equal(t, 10, maxR)
	})

	t.Run("falls back to global when entry omits them", func(t *testing.T) {
		p, maxR := resolvePolicy(cfg, Entry{})
		assert.Equal(t, cexec.RestartAlways, p)
		assert.Equal(t, 3, maxR)
	})
}

func TestResolveLimits(t *testing.T) {
	cfg := verify.ProcfileConfig{Mem: "256Mi", CPU: 50}

	t.Run("entry overrides global", func(t *testing.T) {
		l, err := resolveLimits(cfg, Entry{Mem: "512Mi", CPU: 100})
		require.NoError(t, err)
		assert.Equal(t, uint64(512*1024*1024), l.MaxRSSBytes)
		assert.Equal(t, 100.0, l.MaxCPUPercent)
	})

	t.Run("falls back to global when entry omits them", func(t *testing.T) {
		l, err := resolveLimits(cfg, Entry{})
		require.NoError(t, err)
		assert.Equal(t, uint64(256*1024*1024), l.MaxRSSBytes)
		assert.Equal(t, 50.0, l.MaxCPUPercent)
	})

	t.Run("invalid mem is a loud error", func(t *testing.T) {
		_, err := resolveLimits(verify.ProcfileConfig{}, Entry{Name: "x", Mem: "notabyte"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid mem")
	})
}
