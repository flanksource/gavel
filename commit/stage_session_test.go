package commit

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/flanksource/gavel/verify"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var sessionEnvironmentMarkers = []string{
	"CODEX_THREAD_ID", "CODEX_SESSION_ID", "CODEX_SANDBOX",
	"CLAUDE_CODE_SESSION_ID", "CLAUDE_SESSION_ID", "CLAUDECODE",
	"GEMINI_SESSION_ID", "GEMINI_CLI", "CAPTAIN_SESSION_ID",
}

// clearSessionEnv blanks every session-id env var so a test inheriting a real
// agent session (e.g. running inside Claude/Codex) starts from a clean slate.
func clearSessionEnv(t *testing.T) {
	t.Helper()
	t.Setenv(EnvSessionID, "")
	for _, marker := range sessionEnvironmentMarkers {
		t.Setenv(marker, "")
	}
}

func TestResolveEnvSessionIDUsesCaptainMarkers(t *testing.T) {
	clearSessionEnv(t)
	t.Setenv("CODEX_THREAD_ID", "codex-thread")

	assert.Equal(t, "codex-thread", resolveEnvSessionID())
}

// TestResolveEnvSessionIDPrecedence pins GAVEL_SESSION_ID as the override before
// Captain's provider-marker precedence.
func TestResolveEnvSessionIDPrecedence(t *testing.T) {
	tests := []struct {
		name       string
		gavel      string
		claudeCode string
		claude     string
		codex      string
		want       string
	}{
		{"gavel wins", "gav-1", "cc-1", "cla-1", "cod-1", "gav-1"},
		{"codex when gavel unset", "", "cc-1", "cla-1", "cod-1", "cod-1"},
		{"claude code when codex unset", "", "cc-1", "cla-1", "", "cc-1"},
		{"legacy claude when claude-code unset", "", "", "cla-1", "", "cla-1"},
		{"empty when none set", "", "", "", "", ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			clearSessionEnv(t)
			t.Setenv(EnvSessionID, tc.gavel)
			t.Setenv("CLAUDE_CODE_SESSION_ID", tc.claudeCode)
			t.Setenv("CLAUDE_SESSION_ID", tc.claude)
			t.Setenv("CODEX_SESSION_ID", tc.codex)
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
	t.Setenv("CLAUDE_SESSION_ID", sessionID)

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
// Codex session id from CODEX_THREAD_ID, locates the rollout, and stages exactly
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
	t.Setenv("CODEX_THREAD_ID", sessionID)

	err := stageFiles(dir, StageSession, verify.CommitConfig{})
	require.NoError(t, err)

	assert.Equal(t, []string{"app.go"}, mustStagedFiles(t, dir))
}

// writeCodexRollout lays down a Codex rollout under the fake HOME whose filename
// carries sessionID and whose custom tool call applies a patch to relPaths.
func writeCodexRollout(t *testing.T, home, sessionID, cwd string, relPaths []string) {
	t.Helper()
	sessionsDir := filepath.Join(home, ".codex", "sessions", "2026", "06", "29")
	require.NoError(t, os.MkdirAll(sessionsDir, 0o755))

	patch := "*** Begin Patch\n"
	for _, p := range relPaths {
		patch += "*** Update File: " + p + "\n@@\n-old\n+new\n"
	}
	patch += "*** End Patch\n"
	input := "const patch = " + strconv.Quote(patch) + ";\ntext(await tools.apply_patch(patch));\n"

	lines := []map[string]any{
		{"timestamp": "2026-06-29T10:00:00.000Z", "type": "session_meta", "payload": map[string]any{
			"id": sessionID, "cwd": cwd, "cli_version": "0.1.0", "model_provider": "openai",
		}},
		{"timestamp": "2026-06-29T10:00:01.000Z", "type": "response_item", "payload": map[string]any{
			"type": "custom_tool_call", "name": "exec", "input": input, "call_id": "call-1",
		}},
		{"timestamp": "2026-06-29T10:00:02.000Z", "type": "response_item", "payload": map[string]any{
			"type": "custom_tool_call_output", "call_id": "call-1", "output": "Success",
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
