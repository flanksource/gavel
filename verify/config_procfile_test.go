package verify

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	yamlv3 "gopkg.in/yaml.v3"
)

func TestRestartPolicyUnmarshal(t *testing.T) {
	jsonCases := map[string]RestartPolicy{
		`"on-failure"`: RestartOnFailure,
		`"always"`:     RestartAlways,
		`"no"`:         RestartNo,
		`true`:         RestartOnFailure,
		`false`:        RestartNo,
		`null`:         "",
	}
	for in, want := range jsonCases {
		var p RestartPolicy
		require.NoError(t, json.Unmarshal([]byte(in), &p), in)
		assert.Equal(t, want, p, in)
	}
	var bad RestartPolicy
	assert.Error(t, json.Unmarshal([]byte(`"maybe"`), &bad), "invalid enum is a loud error")

	yamlCases := map[string]RestartPolicy{
		"on-failure": RestartOnFailure,
		"always":     RestartAlways,
		"no":         RestartNo, // YAML 1.2 string "no"
		"true":       RestartOnFailure,
		"false":      RestartNo,
	}
	for in, want := range yamlCases {
		var p RestartPolicy
		require.NoError(t, yamlv3.Unmarshal([]byte(in), &p), in)
		assert.Equal(t, want, p, in)
	}
}

func TestMerge_ProcfileConfig(t *testing.T) {
	t.Run("scalar overrides win when set, omitted keep base", func(t *testing.T) {
		base := ProcfileConfig{AutoRestart: RestartNo, MaxRestarts: 3, Profile: "dev"}
		out := Merge(base, ProcfileConfig{AutoRestart: RestartAlways, MaxRestarts: 9})
		assert.Equal(t, RestartAlways, out.AutoRestart)
		assert.Equal(t, 9, out.MaxRestarts)
		assert.Equal(t, "dev", out.Profile, "omitted profile override keeps base")
	})

	t.Run("empty override preserves base scalars", func(t *testing.T) {
		base := ProcfileConfig{AutoRestart: RestartOnFailure, MaxRestarts: 2, Profile: "prod"}
		out := Merge(base, ProcfileConfig{})
		assert.Equal(t, RestartOnFailure, out.AutoRestart)
		assert.Equal(t, 2, out.MaxRestarts)
		assert.Equal(t, "prod", out.Profile)
	})

	t.Run("env merges key-by-key with override winning", func(t *testing.T) {
		base := ProcfileConfig{Env: map[string]string{"A": "1", "B": "2"}}
		override := ProcfileConfig{Env: map[string]string{"B": "20", "C": "3"}}
		out := Merge(base, override)
		assert.Equal(t, map[string]string{"A": "1", "B": "20", "C": "3"}, out.Env)
	})

	t.Run("resource limits + profile override when set and persist when omitted", func(t *testing.T) {
		base := ProcfileConfig{Mem: "256Mi", CPU: 50, Profile: "dev"}
		out := Merge(base, ProcfileConfig{Mem: "1Gi", Profile: "prod"})
		assert.Equal(t, "1Gi", out.Mem, "non-empty override wins")
		assert.Equal(t, 50.0, out.CPU, "omitted override keeps base")
		assert.Equal(t, "prod", out.Profile)

		kept := Merge(base, ProcfileConfig{})
		assert.Equal(t, "256Mi", kept.Mem)
		assert.Equal(t, 50.0, kept.CPU)
		assert.Equal(t, "dev", kept.Profile)
	})
}

func TestLoadGavelConfig_WithProcfile(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	dir := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(dir, ".git"), 0o755))

	cfgData := []byte(`procfile:
  profile: dev
  autoRestart: on-failure
  maxRestarts: 5
  mem: 512Mi
  env:
    RAILS_ENV: development
`)
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".gavel.yaml"), cfgData, 0o644))

	cfg, err := LoadGavelConfig(dir)
	require.NoError(t, err)
	assert.Equal(t, RestartOnFailure, cfg.Procfile.AutoRestart)
	assert.Equal(t, 5, cfg.Procfile.MaxRestarts)
	assert.Equal(t, "dev", cfg.Procfile.Profile)
	assert.Equal(t, "512Mi", cfg.Procfile.Mem)
	assert.Equal(t, "development", cfg.Procfile.Env["RAILS_ENV"])
}
