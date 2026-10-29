package jira

import (
	"encoding/json"
	"fmt"

	"github.com/flanksource/clicky"
	"github.com/flanksource/gavel/externalissues"
)

type Comments []externalissues.Comment

func (c Comments) FormatView(options clicky.FormatOptions) (any, error) {
	switch options.ResolveFormat() {
	case "json", "ndjson", "yaml", "csv", "toon":
		return []externalissues.Comment(c), nil
	default:
		return renderComments(c, documentRenderOptions{})
	}
}

func ParseComments(issue json.RawMessage) (Comments, error) {
	var response struct {
		Comments *[]json.RawMessage `json:"comments"`
	}
	if err := json.Unmarshal(issue, &response); err != nil {
		return nil, fmt.Errorf("decode issue comments: %w", err)
	}
	if response.Comments == nil {
		return nil, fmt.Errorf("full Jira issue response is missing comments")
	}
	comments := make(Comments, 0, len(*response.Comments))
	for _, raw := range *response.Comments {
		comment, err := parseComment(raw)
		if err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}
	return comments, nil
}
