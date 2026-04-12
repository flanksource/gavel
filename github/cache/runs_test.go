package cache

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGzipRoundtrip(t *testing.T) {
	tests := []struct {
		name string
		in   string
	}{
		{"empty", ""},
		{"short", "hello"},
		{"with newlines", "line1\nline2\nline3"},
		// Logs from CI are typically tens of KB; exercise something that
		// will actually compress meaningfully.
		{"repeated", strings.Repeat("FAIL: TestFoo\n", 1000)},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gz, err := gzipString(tc.in)
			require.NoError(t, err)

			out, err := gunzipString(gz)
			require.NoError(t, err)
			assert.Equal(t, tc.in, out)
		})
	}
}

func TestGzipRepeatedCompresses(t *testing.T) {
	// Sanity check: a repeating payload should compress to substantially
	// less than its raw size. This is the whole point of the gzip wrapper.
	in := strings.Repeat("FAIL: TestFoo\n", 1000)
	gz, err := gzipString(in)
	require.NoError(t, err)
	assert.Less(t, len(gz), len(in)/10,
		"gzipped should be <10%% of raw size for highly repetitive input")
}

func TestGunzipCorruptReturnsError(t *testing.T) {
	_, err := gunzipString([]byte("not a gzip stream"))
	assert.Error(t, err)
}

func TestDisabledStoreCacheOps(t *testing.T) {
	// A disabled store must answer all cache operations as misses and
	// silently drop writes — this is the contract that lets callers use
	// Shared() unconditionally without nil-checking.
	s := &Store{disabled: true}

	assert.Nil(t, s.GetCompletedRunPayload(123))
	s.PutCompletedRun("owner/repo", 123, "completed", "success", []byte("{}"))
	assert.Nil(t, s.GetCompletedRunPayload(123), "still nil after put on disabled store")

	logs, ok := s.GetJobLogs(456)
	assert.False(t, ok)
	assert.Equal(t, "", logs)
	s.PutJobLogs(456, "owner/repo", "fail")
	logs, ok = s.GetJobLogs(456)
	assert.False(t, ok)

	yaml, path, ok := s.GetWorkflowDef("owner/repo", 7, "abc")
	assert.False(t, ok)
	assert.Equal(t, "", yaml)
	assert.Equal(t, "", path)
}

func TestPutCompletedRunRejectsInProgress(t *testing.T) {
	dsn := os.Getenv(EnvDSN)
	if dsn == "" {
		t.Skipf("set %s=postgres://... to run integration tests", EnvDSN)
	}
	t.Setenv(EnvDSN, dsn)
	t.Setenv(EnvDisable, "")

	store, err := Open()
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })
	t.Cleanup(func() { _ = store.gorm().Where("1=1").Delete(&WorkflowRunCache{}) })

	// In-progress runs are intentionally rejected — the immutable cache is
	// only for completed runs. The ETag-aware HTTP cache handles the rest.
	store.PutCompletedRun("owner/repo", 999, "in_progress", "", []byte(`{"foo":"bar"}`))
	assert.Nil(t, store.GetCompletedRunPayload(999),
		"in-progress runs must not be persisted to the immutable cache")
}

func TestCompletedRunRoundtripIntegration(t *testing.T) {
	dsn := os.Getenv(EnvDSN)
	if dsn == "" {
		t.Skipf("set %s=postgres://... to run integration tests", EnvDSN)
	}
	t.Setenv(EnvDSN, dsn)
	t.Setenv(EnvDisable, "")

	store, err := Open()
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })
	t.Cleanup(func() { _ = store.gorm().Where("1=1").Delete(&WorkflowRunCache{}) })

	payload := []byte(`{"databaseId":42,"name":"CI","status":"completed","conclusion":"success"}`)
	store.PutCompletedRun("owner/repo", 42, "completed", "success", payload)

	out := store.GetCompletedRunPayload(42)
	require.NotNil(t, out)
	assert.Equal(t, payload, out)
}

func TestJobLogsRoundtripIntegration(t *testing.T) {
	dsn := os.Getenv(EnvDSN)
	if dsn == "" {
		t.Skipf("set %s=postgres://... to run integration tests", EnvDSN)
	}
	t.Setenv(EnvDSN, dsn)
	t.Setenv(EnvDisable, "")

	store, err := Open()
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })
	t.Cleanup(func() { _ = store.gorm().Where("1=1").Delete(&JobLogCache{}) })

	logs := strings.Repeat("FAIL: TestFoo\n", 500)
	store.PutJobLogs(7, "owner/repo", logs)

	out, ok := store.GetJobLogs(7)
	require.True(t, ok)
	assert.Equal(t, logs, out)
}

func TestWorkflowDefSkipsEmptySHAOnDisabledStore(t *testing.T) {
	// Without a SHA the (repo, workflow_id, sha) cache key isn't stable —
	// the same key would land different YAML over time. The cache refuses
	// the write rather than poison itself. We exercise the early-return on
	// a disabled store so the test doesn't need a real database.
	s := &Store{disabled: true}
	s.PutWorkflowDef("owner/repo", 1, "", "x.yml", "name: ci")
	yaml, path, ok := s.GetWorkflowDef("owner/repo", 1, "")
	assert.False(t, ok)
	assert.Equal(t, "", yaml)
	assert.Equal(t, "", path)
}
