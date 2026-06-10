package status

import (
	"strings"
	"testing"

	clickytext "github.com/flanksource/clicky/text"
	"github.com/stretchr/testify/assert"
)

func TestStatusRendererReflectsAppliedUpdates(t *testing.T) {
	result := &Result{
		Branch: "main",
		Files: []FileStatus{
			{Path: "a.go", State: StateStaged, StagedKind: KindModified},
		},
	}
	result.PrepareAISummaries()

	renderer := NewStatusRenderer(result)

	// Before any update the file is pending.
	pending := clickytext.StripANSI(renderer.RenderLive(nil).String())
	assert.Contains(t, pending, "⏳ ai")

	renderer.Apply(AISummaryUpdate{Index: 0, Status: AISummaryStatusDone, Summary: "refactor handler flow"})

	done := clickytext.StripANSI(renderer.RenderLive(nil).String())
	assert.Contains(t, done, "refactor handler flow")
	assert.NotContains(t, done, "⏳ ai")
	assert.NotContains(t, done, "⟳ ai")
	assert.Equal(t, 1, strings.Count(done, "a.go"))
}

func TestStatusRendererFinalMatchesLive(t *testing.T) {
	result := &Result{
		Branch: "main",
		Files:  []FileStatus{{Path: "b.go", State: StateStaged, StagedKind: KindModified}},
	}
	result.PrepareAISummaries()
	renderer := NewStatusRenderer(result)
	renderer.Apply(AISummaryUpdate{Index: 0, Status: AISummaryStatusDone, Summary: "add cache layer"})

	live := clickytext.StripANSI(renderer.RenderLive(nil).String())
	final := clickytext.StripANSI(renderer.RenderFinal(nil).String())
	assert.Equal(t, live, final)
	assert.Contains(t, final, "add cache layer")
}
