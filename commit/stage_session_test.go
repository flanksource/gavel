package commit

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/flanksource/gavel/verify"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// clearSessionEnv blanks every session-id env var so a test inheriting a real
// agent session (e.g. running inside Claude/Codex) starts from a clean slate.
func clearSessionEnv(t *testing.T) {
	t.Helper()
	t.Setenv(EnvSessionID, "")
	t.Setenv(EnvClaudeSessionID, "")
	t.Setenv(EnvCodexSessionID, "")
}

// TestResolveEnvSessionIDPrecedence pins the env lookup order that backs
// --stage=session: GAVEL_SESSION_ID over CLAUDE_SESSION_ID over CODEX_SESSION_ID.
func TestResolveEnvSessionIDPrecedence(t *testing.T) {
	tests := []struct {
		name   string
		gavel  string
		claude string
		codex  string
		want   string
	}{
		{"gavel wins", "gav-1", "cla-1", "cod-1", "gav-1"},
		{"claude when gavel unset", "", "cla-1", "cod-1", "cla-1"},
		{"codex when others unset", "", "", "cod-1", "cod-1"},
		{"empty when none set", "", "", "", ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv(EnvSessionID, tc.gavel)
			t.Setenv(EnvClaudeSessionID, tc.claude)
			t.Setenv(EnvCodexSessionID, tc.codex)
			assert.Equal(t, tc.want, resolveEnvSessionID())
		})
	}
}

// TestStageSessionModeResolvesClaudeEnvID confirms --stage=session reads the
// Claude session id from CLAUDE_SESSION_ID and stages only that session's edits.
func TestStageSessionModeResolvesClaudeEnvID(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	clearSessionEnv(t)

	dir := initCommitRepo(t)
	writeFile(t, dir, "app.go", "package app\n")        // edited by session
	writeFile(t, dir, "unrelated.go", "package main\n") // NOT edited

	sessionID := "claude-sess-1"
	writeSessionLog(t, home, sessionID, []string{filepath.Join(dir, "app.go")})
	t.Setenv(EnvClaudeSessionID, sessionID)

	err := stageFiles(dir, StageSession, verify.CommitConfig{})
	require.NoError(t, err)

	assert.Equal(t, []string{"app.go"}, mustStagedFiles(t, dir))
}

// TestStageSessionModeFallsBackToStagedWithoutEnv confirms that with no session
// id in the environment, --stage=session behaves like --stage=staged: it adds
// nothing on its own, so only files the caller already staged get committed.
func TestStageSessionModeFallsBackToStagedWithoutEnv(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	clearSessionEnv(t)

	dir := initCommitRepo(t)
	writeFile(t, dir, "app.go", "package app\n")

	err := stageFiles(dir, StageSession, verify.CommitConfig{})
	require.NoError(t, err)

	assert.Empty(t, mustStagedFiles(t, dir), "session fallback must not auto-stage; staged set was empty")
}

// TestStageSessionModeResolvesCodexRollout confirms --stage=session resolves a
// Codex session id from CODEX_SESSION_ID, locates the rollout, and stages exactly
// the files its apply_patch touched.
func TestStageSessionModeResolvesCodexRollout(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	clearSessionEnv(t)

	dir := initCommitRepo(t)
	writeFile(t, dir, "app.go", "package app\n")        // patched by session
	writeFile(t, dir, "unrelated.go", "package main\n") // NOT patched

	sessionID := "019eedeb-7bda-75b1-abdd-86a2c48cd1d5"
	writeCodexRollout(t, home, sessionID, dir, []string{"app.go"})
	t.Setenv(EnvCodexSessionID, sessionID)

	err := stageFiles(dir, StageSession, verify.CommitConfig{})
	require.NoError(t, err)

	assert.Equal(t, []string{"app.go"}, mustStagedFiles(t, dir))
}

// TestParseApplyPatchPaths verifies the apply_patch path scanner across the shell
// shapes Codex emits: JSON-escaped \n inside a command array, real newlines, a
// rename (Update + Move to), and a path terminated by a closing JSON quote.
func TestParseApplyPatchPaths(t *testing.T) {
	tests := []struct {
		name    string
		command string
		want    []string
	}{
		{
			name:    "json escaped newlines",
			command: `{"command":["apply_patch","*** Begin Patch\n*** Update File: a/b.go\n*** Add File: c.go\n*** End Patch\n"]}`,
			want:    []string{"a/b.go", "c.go"},
		},
		{
			name:    "real newlines",
			command: "*** Begin Patch\n*** Update File: x.go\n*** End Patch\n",
			want:    []string{"x.go"},
		},
		{
			name:    "rename captures old and new",
			command: "*** Update File: old.go\n*** Move to: new.go\n",
			want:    []string{"old.go", "new.go"},
		},
		{
			name:    "delete terminated by quote",
			command: `["apply_patch","*** Delete File: gone.go"]`,
			want:    []string{"gone.go"},
		},
		{
			name:    "no markers",
			command: `{"command":["go","test","./..."]}`,
			want:    nil,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, parseApplyPatchPaths(tc.command))
		})
	}
}

// writeCodexRollout lays down a Codex rollout under the fake HOME whose filename
// carries sessionID and whose single apply_patch shell call updates relPaths.
func writeCodexRollout(t *testing.T, home, sessionID, cwd string, relPaths []string) {
	t.Helper()
	sessionsDir := filepath.Join(home, ".codex", "sessions", "2026", "06", "29")
	require.NoError(t, os.MkdirAll(sessionsDir, 0o755))

	patch := "*** Begin Patch\n"
	for _, p := range relPaths {
		patch += "*** Update File: " + p + "\n@@\n-old\n+new\n"
	}
	patch += "*** End Patch\n"
	argsBytes, err := json.Marshal(map[string]any{"command": []string{"apply_patch", patch}})
	require.NoError(t, err)

	lines := []map[string]any{
		{"timestamp": "2026-06-29T10:00:00.000Z", "type": "session_meta", "payload": map[string]any{
			"id": sessionID, "cwd": cwd, "cli_version": "0.1.0", "model_provider": "openai",
		}},
		{"timestamp": "2026-06-29T10:00:01.000Z", "type": "response_item", "payload": map[string]any{
			"type": "function_call", "name": "shell", "arguments": string(argsBytes), "call_id": "call-1",
		}},
		{"timestamp": "2026-06-29T10:00:02.000Z", "type": "response_item", "payload": map[string]any{
			"type": "function_call_output", "call_id": "call-1", "output": "Success",
		}},
	}

	var content []byte
	for _, l := range lines {
		b, err := json.Marshal(l)
		require.NoError(t, err)
		content = append(content, b...)
		content = append(content, '\n')
	}

	name := "rollout-2026-06-29T10-00-00-" + sessionID + ".jsonl"
	require.NoError(t, os.WriteFile(filepath.Join(sessionsDir, name), content, 0o644))
}
