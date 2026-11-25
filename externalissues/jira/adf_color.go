package jira

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/flanksource/clicky/markdown"
)

var adfHexColor = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

func coloredADFNode(node markdown.Node, raw json.RawMessage) (markdown.Node, error) {
	var attrs struct {
		Color string `json:"color"`
	}
	if err := json.Unmarshal(raw, &attrs); err != nil {
		return markdown.Node{}, fmt.Errorf("decode color attributes: %w", err)
	}
	if !adfHexColor.MatchString(attrs.Color) {
		return markdown.Node{}, fmt.Errorf("expected a six-digit hexadecimal color, got %q", attrs.Color)
	}
	start := `<span style="color:` + attrs.Color + `">`
	return markdown.Node{
		Kind: "raw-html", Text: node.String(),
		RawHTML: start + node.HTML() + "</span>",
		Source:  start + node.Markdown() + "</span>",
	}, nil
}
