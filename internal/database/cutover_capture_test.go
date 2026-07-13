package database

import (
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	captainmigrations "github.com/flanksource/captain/migrations"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCaptureArtifactGenerationBindsFullReportAndSchema(t *testing.T) {
	report := validCapturedCutoverReport()
	schema := []byte(`{"tables":[{"name":"captain_sessions_legacy_v1"}]}`)

	_, canonicalReport, err := canonicalCapturedCutoverReport(report)
	require.NoError(t, err)
	baseline := captureArtifactGeneration(canonicalReport, schema)

	equivalentReport := *report
	equivalentReport.Details = json.RawMessage("{\n  \"validation\": \"complete\", \"count\": 2\n}")
	_, equivalentCanonicalReport, err := canonicalCapturedCutoverReport(&equivalentReport)
	require.NoError(t, err)
	assert.Equal(t, baseline, captureArtifactGeneration(equivalentCanonicalReport, schema), "equivalent JSONB formatting must not create a new generation")

	changedReport := *report
	changedReport.Details = json.RawMessage(`{"validation":"different"}`)
	_, changedCanonicalReport, err := canonicalCapturedCutoverReport(&changedReport)
	require.NoError(t, err)
	reportGeneration := captureArtifactGeneration(changedCanonicalReport, schema)
	assert.NotEqual(t, baseline, reportGeneration, "fields outside the archive checksums must bind the generation")

	schemaGeneration := captureArtifactGeneration(canonicalReport, []byte(`{"tables":[{"name":"captain_sessions_legacy_v1","constraints":["PRIMARY KEY (path)"]}]}`))
	assert.NotEqual(t, baseline, schemaGeneration, "the exact captured archive schema must bind the generation")
}

func TestReuseCapturedArtifactRequiresCurrentReportAndDatabaseSnapshots(t *testing.T) {
	tests := []struct {
		name      string
		mutate    func(*CutoverArtifactRequest)
		wantError bool
	}{
		{name: "exact current capture"},
		{
			name: "current report changed",
			mutate: func(request *CutoverArtifactRequest) {
				request.Report = map[string]any{"validated": true, "updatedAt": "later"}
			},
			wantError: true,
		},
		{
			name: "current archive snapshot changed",
			mutate: func(request *CutoverArtifactRequest) {
				request.Snapshots[0].Data = []byte("different archive rows")
			},
			wantError: true,
		},
		{
			name: "transcript freshness changed after cutover",
			mutate: func(request *CutoverArtifactRequest) {
				request.Snapshots[1].Data = []byte(`{"summary":{"stale":1}}`)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := CutoverArtifactRequest{
				Cutover:    legacyCaptainCutoverName,
				Generation: "sha256-current-capture",
				CreatedAt:  time.Date(2026, time.July, 13, 1, 2, 3, 0, time.UTC),
				Report:     map[string]any{"validated": true},
				Rollback:   map[string]any{"steps": []string{"restore archive"}},
				Snapshots: []CutoverArtifactSnapshot{
					{Name: "legacy-sessions.json", ContentType: "application/json", Data: []byte(`[{"path":"session.jsonl"}]`)},
					{Name: legacyTranscriptFreshnessSnapshot, ContentType: "application/json", Data: []byte(`{"summary":{"fresh":1}}`)},
				},
			}
			target := filepath.Join(t.TempDir(), request.Generation)
			_, err := WriteCutoverArtifact(target, request)
			require.NoError(t, err)
			if test.mutate != nil {
				test.mutate(&request)
			}

			result, err := reuseCapturedArtifact(target, request)
			if test.wantError {
				require.ErrorIs(t, err, ErrCutoverArtifactInvalid)
				assert.Nil(t, result)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, result)
			assert.NotEmpty(t, result.Checksums)
		})
	}
}

func validCapturedCutoverReport() *captainmigrations.LegacySessionCutoverReport {
	promptTable := "captain_session_prompts_legacy_v1"
	promptChecksum := "legacy-prompts"
	nativePromptChecksum := "native-prompts"
	return &captainmigrations.LegacySessionCutoverReport{
		CutoverKey:               legacyCaptainReportKey,
		LegacySessionsTable:      "captain_sessions_legacy_v1",
		LegacyPromptsTable:       &promptTable,
		LegacySessionRows:        2,
		LegacyPromptRows:         1,
		ImportedSessionRows:      2,
		ImportedPromptRunRows:    1,
		LegacySessionsChecksum:   "legacy-sessions",
		LegacyPromptsChecksum:    &promptChecksum,
		NativeSessionsChecksum:   "native-sessions",
		NativePromptRunsChecksum: &nativePromptChecksum,
		Details:                  json.RawMessage(`{"count":2,"validation":"complete"}`),
		StartedAt:                time.Date(2026, time.July, 13, 0, 59, 0, 0, time.UTC),
		CompletedAt:              time.Date(2026, time.July, 13, 1, 0, 0, 0, time.UTC),
		UpdatedAt:                time.Date(2026, time.July, 13, 1, 1, 0, 0, time.UTC),
	}
}
