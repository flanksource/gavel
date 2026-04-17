package prwatch

import (
	"testing"

	"github.com/flanksource/gavel/github"
	"github.com/stretchr/testify/assert"
)

func TestIdentifyBot(t *testing.T) {
	tests := []struct {
		name     string
		author   string
		body     string
		expected string
	}{
		{"coderabbit [bot]", "coderabbitai[bot]", "review body", "coderabbit"},
		{"coderabbit graphql", "coderabbitai", "review body", "coderabbit"},
		{"vercel [bot]", "vercel[bot]", "deployment ready", "vercel"},
		{"vercel graphql", "vercel", "deployment ready", "vercel"},
		{"copilot github [bot]", "github-copilot[bot]", "suggestion", "copilot"},
		{"copilot github graphql", "github-copilot", "suggestion", "copilot"},
		{"copilot short", "copilot[bot]", "suggestion", "copilot"},
		{"gavel sticky comment", "github-actions[bot]", "<!-- sticky-comment:gavel -->\n## Results", "gavel"},
		{"gavel self-test", "github-actions[bot]", "<!-- sticky-comment:gavel-self-test -->\nSummary", "gavel"},
		{"unknown bot", "dependabot[bot]", "bump version", ""},
		{"human author", "octocat", "LGTM", ""},
		{"empty", "", "", ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, identifyBot(tc.author, tc.body))
		})
	}
}

func TestAnnotateBots(t *testing.T) {
	comments := []github.PRComment{
		{ID: 1, Author: "coderabbitai[bot]", Body: "review"},
		{ID: 2, Author: "vercel[bot]", Body: "preview ready"},
		{ID: 3, Author: "octocat", Body: "looks good"},
		{ID: 4, Author: "github-actions[bot]", Body: "<!-- sticky-comment:gavel -->\nresults"},
	}

	result := annotateBots(comments)
	assert.Equal(t, "coderabbit", result[0].BotType)
	assert.Equal(t, "vercel", result[1].BotType)
	assert.Equal(t, "", result[2].BotType)
	assert.Equal(t, "gavel", result[3].BotType)
}
