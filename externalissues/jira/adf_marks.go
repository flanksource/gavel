package jira

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/flanksource/clicky/markdown"
)

var adfBorderColor = regexp.MustCompile(`^#[0-9a-fA-F]{6}([0-9a-fA-F]{2})?$`)

func styledADFNode(node markdown.Node, tag, style string) markdown.Node {
	start := "<" + tag
	if style != "" {
		start += ` style="` + style + `"`
	}
	start += ">"
	end := "</" + tag + ">"
	return markdown.Node{Kind: "raw-html", Text: node.String(), RawHTML: start + node.HTML() + end, Source: start + node.Markdown() + end}
}

func subSupADFNode(node markdown.Node, raw json.RawMessage) (markdown.Node, error) {
	var attrs struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(raw, &attrs); err != nil {
		return markdown.Node{}, fmt.Errorf("decode subsup: %w", err)
	}
	if attrs.Type != "sub" && attrs.Type != "sup" {
		return markdown.Node{}, fmt.Errorf("subsup requires sub or sup, received %q", attrs.Type)
	}
	return styledADFNode(node, attrs.Type, ""), nil
}

func borderedADFNode(node markdown.Node, raw json.RawMessage) (markdown.Node, error) {
	var attrs struct {
		Size  float64 `json:"size"`
		Color string  `json:"color"`
	}
	if err := json.Unmarshal(raw, &attrs); err != nil {
		return markdown.Node{}, fmt.Errorf("decode border: %w", err)
	}
	if attrs.Size < 1 || attrs.Size > 3 || !adfBorderColor.MatchString(attrs.Color) {
		return markdown.Node{}, fmt.Errorf("border requires size 1..3 and a hexadecimal color")
	}
	return styledADFNode(node, "span", fmt.Sprintf("border:%gpx solid %s", attrs.Size, attrs.Color)), nil
}
