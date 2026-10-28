package jira

import (
	"encoding/json"
	"fmt"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
)

type Comments struct {
	Raw  json.RawMessage `json:"-"`
	text api.Text
}

func (c Comments) MarshalJSON() ([]byte, error) { return c.Raw, nil }
func (c Comments) Pretty() api.Text             { return c.text }

func ParseComments(issue json.RawMessage) (Comments, error) {
	var response struct {
		Comments json.RawMessage `json:"comments"`
	}
	if err := json.Unmarshal(issue, &response); err != nil {
		return Comments{}, fmt.Errorf("decode issue comments: %w", err)
	}
	if len(response.Comments) == 0 || string(response.Comments) == "null" {
		return Comments{}, fmt.Errorf("full Jira issue response is missing comments")
	}
	var comments []issueComment
	if err := json.Unmarshal(response.Comments, &comments); err != nil {
		return Comments{}, fmt.Errorf("decode comments: %w", err)
	}
	var text api.Text
	for _, comment := range comments {
		body, err := comment.Body.Render()
		if err != nil {
			return Comments{}, fmt.Errorf("comment by %s: %w", comment.Author.DisplayName, err)
		}
		heading := comment.Author.DisplayName + " — " + comment.Created
		if comment.Updated != "" && comment.Updated != comment.Created {
			heading += " (updated " + comment.Updated + ")"
		}
		text = text.Add(clicky.Heading(3, clicky.Text(heading))).NewLine().Add(body).NewLine().NewLine()
	}
	return Comments{Raw: response.Comments, text: text}, nil
}
