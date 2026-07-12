package migrategrite

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLiveExporterCopiesAndValidatesExport(t *testing.T) {
	source := []byte(`{
		"meta":{"schema_version":1,"generated_ts":10,"event_count":1},
		"issues":[{"issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","title":"x","state":"open"}],
		"events":[{"event_id":"event","issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","ts_unix_ms":5,"kind":{"IssueCreated":{"title":"x","body":"body"}}}]
	}`)
	var gotArgs []string
	exporter := LiveExporter{
		WorkDir: "/repo", Binary: "/bin/grite",
		Runner: func(_ context.Context, workDir, binary string, args ...string) ([]byte, error) {
			assert.Equal(t, "/repo", workDir)
			assert.Equal(t, "/bin/grite", binary)
			gotArgs = append([]string(nil), args...)
			return []byte(`{"ok":true,"data":{"output_path":".grite/export.json","event_count":1}}`), nil
		},
		ReadFile: func(path string) ([]byte, error) {
			assert.Equal(t, "/repo/.grite/export.json", path)
			return source, nil
		},
	}

	snapshot, copied, result, err := exporter.Export(context.Background(), 41)
	require.NoError(t, err)
	assert.Equal(t, 1, result.EventCount)
	assert.Len(t, snapshot.Events, 1)
	assert.Equal(t, source, copied)
	assert.Equal(t, []string{"--no-daemon", "export", "--format", "json", "--since", "41", "--json"}, gotArgs)

	copied[0] = 'x'
	assert.True(t, strings.HasPrefix(string(source), "{"), "returned source must be owned")
}

func TestLiveExporterRejectsEnvelopeCountMismatch(t *testing.T) {
	exporter := LiveExporter{
		WorkDir: "/repo", Binary: "/bin/grite",
		Runner: func(context.Context, string, string, ...string) ([]byte, error) {
			return []byte(`{"ok":true,"data":{"output_path":"export.json","event_count":2}}`), nil
		},
		ReadFile: func(string) ([]byte, error) {
			return []byte(`{"meta":{"schema_version":1,"event_count":1},"issues":[],"events":[]}`), nil
		},
	}
	_, _, _, err := exporter.Export(context.Background(), 0)
	require.ErrorContains(t, err, "reports")
}
