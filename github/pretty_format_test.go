package github

import (
	"strings"
	"testing"

	"github.com/flanksource/clicky/api"
	"github.com/stretchr/testify/assert"
)

// styledSpan is one rendered text node flattened out of the api.Text tree.
type styledSpan struct {
	content string
	style   string
}

// flattenText walks an api.Text tree depth-first into a flat list of spans so a
// test can assert which styles apply to which content without parsing ANSI/HTML.
func flattenText(t api.Text) []styledSpan {
	spans := []styledSpan{{content: t.Content, style: t.Style}}
	for _, child := range t.Children {
		if ct, ok := child.(api.Text); ok {
			spans = append(spans, flattenText(ct)...)
		}
	}
	return spans
}

// styleFor returns the style of the first span whose content contains substr.
func styleFor(spans []styledSpan, substr string) (string, bool) {
	for _, s := range spans {
		if strings.Contains(s.content, substr) {
			return s.style, true
		}
	}
	return "", false
}

func TestWorkflowRunPrettyActionNameNotBold(t *testing.T) {
	run := WorkflowRun{Name: "CI Build", Status: "completed", Conclusion: "success"}
	spans := flattenText(run.Pretty())

	style, ok := styleFor(spans, "CI Build")
	assert.True(t, ok, "action name should be rendered")
	assert.NotContains(t, style, "font-bold", "action/workflow name must not be bold")
}

func TestPRInfoPrettyTitleNotBold(t *testing.T) {
	pr := PRInfo{Number: 7, Title: "Fix the thing"}
	spans := flattenText(pr.Pretty())

	style, ok := styleFor(spans, "Fix the thing")
	assert.True(t, ok, "PR title should be rendered")
	assert.NotContains(t, style, "font-bold", "PR title must not be bold")
}

func TestPRListItemPrettyRepoBold(t *testing.T) {
	item := PRListItem{Number: 1, Title: "first", Repo: "flanksource/gavel", State: "OPEN", Source: "br"}
	spans := flattenText(item.Pretty())

	style, ok := styleFor(spans, "gavel")
	assert.True(t, ok, "repo name should be rendered")
	assert.Contains(t, style, "font-bold", "repo name should be the bold standout")

	titleStyle, ok := styleFor(spans, "first")
	assert.True(t, ok, "PR title should be rendered")
	assert.NotContains(t, titleStyle, "font-bold", "PR title must not be bold")
}

func TestPRSearchResultsPrettyDividerBetweenItems(t *testing.T) {
	results := PRSearchResults{
		{Number: 1, Title: "first", Repo: "flanksource/a", State: "OPEN", Source: "br1"},
		{Number: 2, Title: "second", Repo: "flanksource/a", State: "OPEN", Source: "br2"},
		{Number: 3, Title: "third", Repo: "flanksource/a", State: "OPEN", Source: "br3"},
	}
	spans := flattenText(results.Pretty())

	dividers := 0
	for _, s := range spans {
		if strings.Contains(s.content, prListDivider) {
			dividers++
		}
	}
	// 3 single-repo items => dividers appear between them only: exactly 2.
	assert.Equal(t, 2, dividers, "divider should appear between items, not after the last")
}
