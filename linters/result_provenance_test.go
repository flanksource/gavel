package linters

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/flanksource/clicky"
	commonsContext "github.com/flanksource/commons/context"
	"github.com/flanksource/gavel/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// cachingLinter answers from its own cache and says so, as an in-process
// linter over database entities does when every entity was a hit.
type cachingLinter struct {
	dryRunLinter
	cache *ResultProvenance
}

func (l cachingLinter) Name() string { return "oipa-transaction" }

func (l cachingLinter) RunResult(commonsContext.Context, *clicky.Task) (*LinterResult, error) {
	return &LinterResult{
		Violations: []models.Violation{{File: "transaction/Fee", Line: 3, Severity: models.SeverityWarning}},
		Cache:      l.cache,
	}, nil
}

func TestLinterResultKeepsTheProvenanceItsLinterReports(t *testing.T) {
	stored := time.Now().Add(-3 * time.Hour).UTC().Truncate(time.Second)
	cache := &ResultProvenance{Source: "build", Fingerprint: "lint+v7+transaction:fee+fp:build+static", StoredAt: stored, Hit: true}

	result := RunLinterWithTask(context.Background(), nil, cachingLinter{cache: cache}, RunOptions{WorkDir: t.TempDir()})
	require.Equal(t, cache, result.Cache)

	encoded, err := json.Marshal(result)
	require.NoError(t, err)
	var decoded LinterResult
	require.NoError(t, json.Unmarshal(encoded, &decoded))
	assert.Equal(t, cache, decoded.Cache, "a saved snapshot keeps where the result came from")
}

func TestLinterResultPrettyShowsTheAgeOfACachedResult(t *testing.T) {
	cases := map[string]struct {
		cache *ResultProvenance
		want  string
	}{
		"a full hit names its age": {
			cache: &ResultProvenance{Source: "build", StoredAt: time.Now().Add(-3*time.Hour - time.Minute), Hit: true},
			want:  "(cached, linted 3h ago)",
		},
		"a partial hit is a fresh run": {
			cache: &ResultProvenance{Source: "build", StoredAt: time.Now().Add(-3 * time.Hour), Hit: false},
			want:  "(1 violations,",
		},
		"no provenance is a fresh run": {want: "(1 violations,"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			result := &LinterResult{Linter: "oipa-transaction", Success: true, Cache: tc.cache,
				Violations: []models.Violation{{File: "transaction/Fee", Severity: models.SeverityWarning}}}
			text := result.Pretty().String()
			assert.True(t, strings.Contains(text, tc.want), text)
		})
	}
}
