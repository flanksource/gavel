package prwatch

import (
	"strings"

	"github.com/flanksource/gavel/github"
)

var botAuthors = map[string]string{
	// GraphQL returns login without [bot]; REST may include it.
	"coderabbitai[bot]":   "coderabbit",
	"coderabbitai":        "coderabbit",
	"vercel[bot]":         "vercel",
	"vercel":              "vercel",
	"github-copilot[bot]": "copilot",
	"github-copilot":      "copilot",
	"copilot[bot]":        "copilot",
	"copilot":             "copilot",
	"github-actions[bot]": "",
	"github-actions":      "",
	"dependabot[bot]":     "",
	"dependabot":          "",
	"renovate[bot]":       "",
	"renovate":            "",
}

func identifyBot(author, body string) string {
	if bt, ok := botAuthors[author]; ok && bt != "" {
		return bt
	}
	if strings.Contains(body, "<!-- sticky-comment:gavel") {
		return "gavel"
	}
	return ""
}

func annotateBots(comments []github.PRComment) []github.PRComment {
	for i := range comments {
		comments[i].BotType = identifyBot(comments[i].Author, comments[i].Body)
	}
	return comments
}
