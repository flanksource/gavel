package main

import (
	"bytes"
	"strings"
	"testing"

	clickytext "github.com/flanksource/clicky/text"
	"github.com/flanksource/gavel/status"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderStatusOutputNonInteractivePrintsFinalState(t *testing.T) {
	result := &status.Result{
		Branch: "main",
		Files: []status.FileStatus{
			{
				Path:       "a.go",
				State:      status.StateStaged,
				StagedKind: status.KindModified,
			},
		},
	}
	result.PrepareAISummaries()

	updates := make(chan status.AISummaryUpdate, 2)
	updates <- status.AISummaryUpdate{Index: 0, Status: status.AISummaryStatusRunning}
	updates <- status.AISummaryUpdate{Index: 0, Status: status.AISummaryStatusDone, Summary: "refactor handler flow"}
	close(updates)

	var output bytes.Buffer
	err := renderStatusOutput(&output, result, updates, false)
	require.NoError(t, err)

	clean := clickytext.StripANSI(output.String())
	assert.Contains(t, clean, "refactor handler flow")
	assert.NotContains(t, clean, "⏳ ai")
	assert.NotContains(t, clean, "⟳ ai")
	assert.Equal(t, 1, strings.Count(clean, "a.go"))
}
