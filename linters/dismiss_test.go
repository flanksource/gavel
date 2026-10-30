package linters

import (
	"testing"

	"github.com/flanksource/gavel/models"
	"github.com/stretchr/testify/assert"
)

func TestDismissViolations(t *testing.T) {
	const (
		kept      = "fp-kept"
		dismissed = "fp-dismissed"
	)
	fresh := func() []*LinterResult {
		return []*LinterResult{{Linter: "vale", Violations: []models.Violation{{Fingerprint: kept}, {Fingerprint: dismissed}}}}
	}
	fingerprints := func(violations []models.Violation) []string {
		var out []string
		for _, v := range violations {
			out = append(out, v.Fingerprint)
		}
		return out
	}

	results := fresh()
	moved := DismissViolations(results, map[string]bool{dismissed: true})
	afterDismiss := map[string][]string{"violations": fingerprints(results[0].Violations), "dismissed": fingerprints(results[0].Dismissed)}
	restored := RestoreViolations(results, map[string]bool{dismissed: true})

	assert.Equal(t, map[string]any{
		"moved":        1,
		"afterDismiss": map[string][]string{"violations": {kept}, "dismissed": {dismissed}},
		"restored":     1,
		"afterRestore": []string{kept, dismissed},
		"dismissedNow": 0,
	}, map[string]any{
		"moved":        moved,
		"afterDismiss": afterDismiss,
		"restored":     restored,
		"afterRestore": fingerprints(results[0].Violations),
		"dismissedNow": len(results[0].Dismissed),
	})
}
