package jira

import (
	"encoding/json"
	"fmt"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
)

type Attachment struct {
	ID       string     `json:"id"`
	Filename string     `json:"filename"`
	Content  string     `json:"content"`
	MIMEType string     `json:"mimeType"`
	Size     int64      `json:"size"`
	Created  string     `json:"created"`
	Author   NamedValue `json:"author"`
}

func (a Attachment) Columns() []api.ColumnDef {
	return []api.ColumnDef{clicky.Column("id").Label("ID").Build(), clicky.Column("file").Label("File").Build(), clicky.Column("type").Label("Type").Build(), clicky.Column("size").Label("Size").Build(), clicky.Column("author").Label("Author").Build(), clicky.Column("created").Label("Created").Build()}
}

func (a Attachment) Row() map[string]any {
	return map[string]any{"id": a.ID, "file": clicky.Link(a.Content).AddText(a.Filename), "type": a.MIMEType, "size": api.HumanizeBytes(a.Size), "author": a.Author.DisplayName, "created": a.Created}
}

func ParseAttachments(raw json.RawMessage) ([]Attachment, error) {
	var items []Attachment
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil, fmt.Errorf("decode Jira attachments: %w", err)
	}
	return items, nil
}
