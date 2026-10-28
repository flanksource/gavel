package jira

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
)

type Issue struct {
	Raw  json.RawMessage `json:"-"`
	text api.Text
}

type issueComment struct {
	Author struct {
		DisplayName string `json:"displayName"`
	} `json:"author"`
	Created string  `json:"created"`
	Updated string  `json:"updated"`
	Body    ADFNode `json:"body"`
}

type issueAttachment struct {
	Filename string     `json:"filename"`
	Content  string     `json:"content"`
	ID       string     `json:"id"`
	Size     int64      `json:"size"`
	MIMEType string     `json:"mimeType"`
	Created  string     `json:"created"`
	Author   NamedValue `json:"author"`
}

func (i Issue) MarshalJSON() ([]byte, error) { return i.Raw, nil }

func ParseIssue(raw json.RawMessage) (Issue, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return Issue{}, fmt.Errorf("decode Jira issue: %w", err)
	}
	var key, summary, link string
	if err := json.Unmarshal(fields["key"], &key); err != nil || key == "" {
		return Issue{}, fmt.Errorf("Jira issue has no valid key")
	}
	if err := json.Unmarshal(fields["summary"], &summary); err != nil {
		return Issue{}, fmt.Errorf("issue %s summary: %w", key, err)
	}
	if rawURL := fields["url"]; len(rawURL) > 0 && string(rawURL) != "null" {
		if err := json.Unmarshal(rawURL, &link); err != nil {
			return Issue{}, fmt.Errorf("issue %s URL: %w", key, err)
		}
	}
	text := api.Text{}.Add(clicky.Heading(1, clicky.Link(link).AddText(key+" — "+summary))).NewLine().NewLine()
	keys := make([]string, 0, len(fields))
	for name := range fields {
		keys = append(keys, name)
	}
	sort.Strings(keys)
	for _, name := range keys {
		if name == "key" || name == "summary" || name == "url" || name == "description" || name == "comment" || name == "comments" || name == "attachment" || bytes.Equal(fields[name], []byte("null")) {
			continue
		}
		var scalar string
		if json.Unmarshal(fields[name], &scalar) == nil {
			if scalar != "" {
				text = text.Add(clicky.Text(name + ": " + scalar)).NewLine()
			}
			continue
		}
		var doc ADFNode
		if json.Unmarshal(fields[name], &doc) == nil && doc.Type == "doc" {
			value, err := doc.Render()
			if err != nil {
				return Issue{}, fmt.Errorf("issue %s field %s: %w", key, name, err)
			}
			text = text.Add(clicky.Heading(2, clicky.Text(name))).NewLine().Add(value).NewLine()
			continue
		}
		text = text.Add(clicky.Text(name + ": ")).Add(clicky.CodeBlock("json", string(fields[name]))).NewLine()
	}
	if rawDescription := fields["description"]; len(rawDescription) > 0 && string(rawDescription) != "null" {
		var description ADFNode
		if err := json.Unmarshal(rawDescription, &description); err != nil {
			return Issue{}, fmt.Errorf("issue %s description: %w", key, err)
		}
		rendered, err := description.Render()
		if err != nil {
			return Issue{}, fmt.Errorf("issue %s description: %w", key, err)
		}
		text = text.NewLine().Add(clicky.Heading(2, clicky.Text("Description"))).NewLine().Add(rendered).NewLine()
	}
	var attachments []issueAttachment
	if rawAttachments := fields["attachment"]; len(rawAttachments) > 0 && string(rawAttachments) != "null" {
		if err := json.Unmarshal(rawAttachments, &attachments); err != nil {
			return Issue{}, fmt.Errorf("issue %s attachments: %w", key, err)
		}
	}
	if len(attachments) > 0 {
		text = text.NewLine().Add(clicky.Heading(2, clicky.Text("Attachments"))).NewLine()
		for _, attachment := range attachments {
			text = text.Add(clicky.Link(attachment.Content).AddText(attachment.Filename)).Append(fmt.Sprintf(" (%s, %s, %d bytes, %s, %s)", attachment.ID, attachment.MIMEType, attachment.Size, attachment.Author.DisplayName, attachment.Created)).NewLine()
		}
	}
	var comments []issueComment
	if rawComments := fields["comments"]; len(rawComments) > 0 && string(rawComments) != "null" {
		if err := json.Unmarshal(rawComments, &comments); err != nil {
			return Issue{}, fmt.Errorf("issue %s comments: %w", key, err)
		}
	}
	if len(comments) > 0 {
		text = text.NewLine().Add(clicky.Heading(2, clicky.Text("Comments"))).NewLine()
		for _, comment := range comments {
			rendered, err := comment.Body.Render()
			if err != nil {
				return Issue{}, fmt.Errorf("issue %s comment by %s: %w", key, comment.Author.DisplayName, err)
			}
			heading := comment.Author.DisplayName + " — " + comment.Created
			if comment.Updated != "" && comment.Updated != comment.Created {
				heading += " (updated " + comment.Updated + ")"
			}
			text = text.NewLine().Add(clicky.Heading(3, clicky.Text(heading))).NewLine().Add(rendered).NewLine()
		}
	}
	return Issue{Raw: raw, text: text}, nil
}

func (i Issue) Pretty() api.Text { return i.text }
