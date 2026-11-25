package jira

import (
	"encoding/json"
	"fmt"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/gavel/externalissues"
)

type Attachment struct {
	externalissues.Attachment
}

func (a Attachment) PrettyShort() api.Textable {
	return clicky.Link(a.URL).AddText(a.Name)
}

func (a Attachment) Pretty() api.Text {
	text := api.Text{}.Add(a.PrettyShort()).Append(" (").Add(api.HumanizeBytes(a.Size)).Append(")")
	for _, value := range []string{a.ID, a.MIMEType, a.Author.Name} {
		if value != "" {
			text = text.Append(" · " + value)
		}
	}
	if a.Created != nil {
		text = text.Append(" · ").Add(clicky.Human(*a.Created))
	}
	return text
}

func (a Attachment) Columns() []api.ColumnDef {
	return []api.ColumnDef{clicky.Column("id").Label("ID").Build(), clicky.Column("file").Label("File").Build(), clicky.Column("type").Label("Type").Build(), clicky.Column("size").Label("Size").Build(), clicky.Column("author").Label("Author").Build(), clicky.Column("created").Label("Created").Build()}
}

func (a Attachment) Row() map[string]any {
	row := map[string]any{"id": a.ID, "file": a.PrettyShort(), "type": a.MIMEType, "size": api.HumanizeBytes(a.Size), "author": a.Author.PrettyShort(), "created": ""}
	if a.Created != nil {
		row["created"] = clicky.Human(*a.Created)
	}
	return row
}

func ParseAttachments(raw json.RawMessage) ([]Attachment, error) {
	var payloads []json.RawMessage
	if err := json.Unmarshal(raw, &payloads); err != nil {
		return nil, fmt.Errorf("decode Jira attachments: %w", err)
	}
	items := make([]Attachment, 0, len(payloads))
	for _, payload := range payloads {
		item, err := parseAttachment(payload)
		if err != nil {
			return nil, err
		}
		items = append(items, Attachment{Attachment: item})
	}
	return items, nil
}
