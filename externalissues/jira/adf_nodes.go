package jira

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/flanksource/clicky/markdown"
)

func cardNode(attrs adfAttrs) (markdown.Node, error) {
	address, label := attrs.URL, ""
	if len(attrs.Data) > 0 {
		if address != "" {
			return markdown.Node{}, fmt.Errorf("inlineCard requires either url or data, not both")
		}
		var data struct {
			ID   string `json:"@id"`
			URL  string `json:"url"`
			Name string `json:"name"`
		}
		if err := json.Unmarshal(attrs.Data, &data); err != nil {
			return markdown.Node{}, fmt.Errorf("inlineCard data: %w", err)
		}
		address, label = data.URL, data.Name
		if address == "" {
			address = data.ID
		}
	}
	parsed, err := url.Parse(address)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return markdown.Node{}, fmt.Errorf("inlineCard requires an HTTP URL")
	}
	if label == "" {
		label = address
		if key, ok := strings.CutPrefix(parsed.Path, "/browse/"); ok && key != "" && !strings.Contains(key, "/") {
			label = key
		}
	}
	return markdown.Node{Kind: "link", Href: address, Children: []markdown.Node{{Kind: "text", Text: label}}}, nil
}

func dateNode(timestamp string) (markdown.Node, error) {
	millis, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return markdown.Node{}, fmt.Errorf("date requires an epoch-millisecond timestamp: %w", err)
	}
	return markdown.Node{Kind: "text", Text: time.UnixMilli(millis).UTC().Format(time.DateOnly)}, nil
}

func panelNode(kind string) (markdown.Node, error) {
	titles := map[string]string{"info": "Info", "note": "Note", "warning": "Warning", "success": "Success", "error": "Error"}
	title, ok := titles[kind]
	if !ok {
		return markdown.Node{}, fmt.Errorf("unsupported panel type %q", kind)
	}
	return markdown.Node{Kind: "blockquote", Children: []markdown.Node{{Kind: "paragraph", Children: []markdown.Node{{Kind: "strong", Children: []markdown.Node{{Kind: "text", Text: title}}}}}}}, nil
}
