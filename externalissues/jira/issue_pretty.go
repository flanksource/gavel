package jira

import (
	"fmt"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
)

type RenderOptions struct {
	Full            bool
	ResolveIssueURL func(string) (string, error)
}

func (i Issue) FormatView(options clicky.FormatOptions) (any, error) {
	switch options.ResolveFormat() {
	case "json", "ndjson", "yaml", "csv", "toon":
		return i.Issue, nil
	default:
		return i.Render(RenderOptions{Full: i.Full})
	}
}

func (i Issue) PrettyShort() api.Textable {
	return clicky.Link(i.URL).AddText(i.Key + " — " + i.Title)
}

func (i Issue) Pretty() api.Text {
	text := api.Text{}.Add(clicky.Heading(3, i.PrettyShort()))
	var metadata api.TextList
	if i.Status != nil && i.Status.Name != "" {
		metadata = append(metadata, i.Status.PrettyShort())
	}
	if i.Priority != nil && i.Priority.Name != "" {
		metadata = append(metadata, i.Priority.PrettyShort())
	}
	if i.Assignee != nil && i.Assignee.Name != "" {
		metadata = append(metadata, i.Assignee.PrettyShort())
	}
	if i.Project != nil && i.Project.Name != "" {
		metadata = append(metadata, i.Project.PrettyShort())
	}
	if len(metadata) > 0 {
		text = text.NewLine().NewLine().Add(metadata.Join(clicky.Text(" · ")))
	}
	return text
}

func (i Issue) Render(options RenderOptions) (api.Textable, error) {
	full := options.Full
	text := i.Pretty()
	if full {
		text = text.NewLine().NewLine().Add(i.PrettyMetadata())
	}
	if i.Description != nil {
		body, err := renderDocument(i.Description, documentRenderOptions{Preview: !full, Context: "issue " + i.Key + " description", ResolveIssueURL: options.ResolveIssueURL, Attachments: i.Attachments})
		if err != nil {
			return nil, fmt.Errorf("issue %s description: %w", i.Key, err)
		}
		if full {
			text = text.NewLine().NewLine().Add(clicky.Heading(2, clicky.Text("Description")))
		}
		text = text.NewLine().NewLine().Add(body)
	}
	if len(i.Labels) > 0 {
		text = text.NewLine().NewLine().Add(clicky.KeyValue("Labels", clicky.CompactList(i.Labels)))
	}
	fields, err := i.renderFields(options)
	if err != nil {
		return nil, err
	}
	if len(i.Fields) > 0 {
		text = text.NewLine().NewLine().Add(fields)
	}
	if !full {
		if len(i.Attachments) > 0 {
			var links api.TextList
			for _, attachment := range i.Attachments {
				links = append(links, (Attachment{Attachment: attachment}).PrettyShort())
			}
			text = text.NewLine().NewLine().Add(clicky.KeyValue("Attachments", clicky.List(links...)))
		}
		return text.Add(i.PrettyCounts()), nil
	}
	if len(i.Attachments) > 0 {
		text = text.NewLine().NewLine().Add(clicky.Heading(2, clicky.Text("Attachments"))).NewLine().NewLine()
		items := make([]Attachment, 0, len(i.Attachments))
		for _, attachment := range i.Attachments {
			items = append(items, Attachment{Attachment: attachment})
		}
		text = text.Add(api.NewTableFrom(items))
	}
	if len(i.Comments) > 0 {
		comments, err := renderComments(i.Comments, documentRenderOptions{Context: "issue " + i.Key, ResolveIssueURL: options.ResolveIssueURL, Attachments: i.Attachments})
		if err != nil {
			return nil, fmt.Errorf("issue %s: %w", i.Key, err)
		}
		text = text.NewLine().NewLine().Add(comments)
	}
	return text, nil
}

func (i Issue) PrettyMetadata() api.Textable {
	metadata := map[string]any{"Provider": i.Provider, "Site": i.Site, "ID": i.ID}
	if i.Type != nil {
		metadata["Type"] = i.Type.PrettyShort()
	}
	if i.Reporter != nil {
		metadata["Reporter"] = i.Reporter.PrettyShort()
	}
	if i.Created != nil {
		metadata["Created"] = clicky.Human(*i.Created)
	}
	if i.Updated != nil {
		metadata["Updated"] = clicky.Human(*i.Updated)
	}
	return clicky.KeyMap(metadata)
}

func (i Issue) PrettyCounts() api.Textable {
	var counts api.TextList
	for _, item := range []struct {
		count *int
		label string
	}{{i.CommentCount, "comment"}, {i.AttachmentCount, "attachment"}} {
		if item.count == nil {
			continue
		}
		label := item.label
		if *item.count != 1 {
			label += "s"
		}
		counts = append(counts, clicky.Text(fmt.Sprintf("%d %s", *item.count, label)))
	}
	if len(counts) == 0 {
		return api.Text{}
	}
	return api.Text{}.NewLine().NewLine().Add(counts.Join(clicky.Text(" · ")))
}
