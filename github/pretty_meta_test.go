package github

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// GitHub reports UNKNOWN mergeability once a PR is merged or closed. Rendered
// bare, beside the state, it reads as an error rather than an absent answer.
func TestMergeableIsLabelledAndOnlyShownWhileOpen(t *testing.T) {
	for _, tc := range []struct {
		name      string
		pr        PRInfo
		wantShown bool
	}{
		{"open and mergeable", PRInfo{State: "OPEN", Mergeable: "MERGEABLE"}, true},
		{"open and conflicting", PRInfo{State: "OPEN", Mergeable: "CONFLICTING"}, true},
		{"open but unknown", PRInfo{State: "OPEN", Mergeable: "UNKNOWN"}, false},
		{"merged", PRInfo{State: "MERGED", Mergeable: "UNKNOWN"}, false},
		{"closed", PRInfo{State: "CLOSED", Mergeable: "MERGEABLE"}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tc.pr.Number = 61
			rendered := tc.pr.Pretty().String()

			assert.NotContains(t, rendered, "UNKNOWN", "an unknown answer is not worth a column")
			if tc.wantShown {
				assert.Contains(t, rendered, "Mergeable: "+tc.pr.Mergeable, "the value needs a label to be readable")
			} else {
				assert.NotContains(t, rendered, "Mergeable:")
			}
		})
	}
}

// A job that never ran reports equal start and end times; "(0s)" reads as a
// real measurement of work that did not happen.
func TestFormatDurationOmitsJobsThatNeverRan(t *testing.T) {
	at := time.Date(2026, 8, 10, 12, 0, 0, 0, time.UTC)

	assert.Equal(t, "", FormatDuration(Job{Status: "COMPLETED", Conclusion: "SKIPPED", StartedAt: at, CompletedAt: at}))
	assert.Equal(t, "", FormatDuration(Job{Status: "COMPLETED", Conclusion: "SKIPPED"}))
	assert.Equal(t, "(17s)", FormatDuration(Job{Status: "COMPLETED", Conclusion: "SUCCESS", StartedAt: at, CompletedAt: at.Add(17 * time.Second)}))
	assert.Equal(t, "(2m 32s)", FormatDuration(Job{Status: "COMPLETED", Conclusion: "SUCCESS", StartedAt: at, CompletedAt: at.Add(152 * time.Second)}))
}

func TestJobPrettyDropsEmptyDuration(t *testing.T) {
	at := time.Date(2026, 8, 10, 12, 0, 0, 0, time.UTC)
	skipped := Job{Name: "Delete Storybook Preview", Status: "COMPLETED", Conclusion: "SKIPPED", StartedAt: at, CompletedAt: at}

	rendered := skipped.Pretty().String()

	assert.Contains(t, rendered, "Delete Storybook Preview")
	assert.NotContains(t, rendered, "(0s)")
	assert.False(t, strings.HasSuffix(rendered, " "), "no dangling separator where the duration used to be")
}

func TestWorkflowRunPrettyAsOverridesTheHeading(t *testing.T) {
	run := WorkflowRun{DatabaseID: 27, Name: "Storybook", Status: "COMPLETED", Conclusion: "SUCCESS"}

	assert.Contains(t, run.Pretty().String(), "Storybook")
	assert.Contains(t, run.PrettyAs("Storybook (run 27)").String(), "Storybook (run 27)")
}
