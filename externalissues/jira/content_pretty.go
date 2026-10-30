package jira

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/clicky/markdown"
	"github.com/flanksource/gavel/externalissues"
)

const previewRunes = 240

type documentRenderOptions struct {
	Attachments     []externalissues.Attachment
	Preview         bool
	Context         string
	ResolveIssueURL func(string) (string, error)
}

func renderDocument(document *externalissues.Document, options documentRenderOptions) (api.Textable, error) {
	if document == nil {
		return clicky.Text(""), nil
	}
	var node markdown.Node
	var err error
	switch document.Format {
	case "text":
		var text string
		if err := json.Unmarshal(document.Content, &text); err != nil {
			return nil, err
		}
		node = markdown.Node{Kind: "text", Text: text}
	case "adf":
		node, err = renderADF(document.Content, adfRenderOptions{Path: strings.TrimSpace(options.Context + " $"), Attachments: options.Attachments})
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported document format %q", document.Format)
	}
	if options.Preview {
		return clicky.Text(documentPreview(node)), nil
	}
	return resolveDocumentLinks(node, options.ResolveIssueURL)
}

func documentPreview(node markdown.Node) string {
	var content strings.Builder
	var visit func(markdown.Node)
	visit = func(node markdown.Node) {
		content.WriteString(node.Text)
		if node.Text == "" {
			content.WriteString(node.Source)
		}
		for _, child := range node.Children {
			visit(child)
		}
		for _, item := range node.Items {
			visit(item)
		}
		switch node.Kind {
		case "paragraph", "heading", "linebreak", "list_item", "table_cell", "code_block":
			content.WriteByte(' ')
		}
	}
	visit(node)
	text := []rune(strings.Join(strings.Fields(content.String()), " "))
	if len(text) > previewRunes {
		return string(text[:previewRunes-1]) + "…"
	}
	return string(text)
}

func renderComments(comments []externalissues.Comment, options documentRenderOptions) (api.Textable, error) {
	if len(comments) == 0 {
		return api.Text{}, nil
	}
	text := api.Text{}.Add(clicky.Heading(2, clicky.Text("Comments"))).NewLine().NewLine()
	for _, comment := range comments {
		bodyOptions := options
		bodyOptions.Context = fmt.Sprintf("%s comment %s by %s", options.Context, comment.ID, comment.Author.Name)
		body, err := renderDocument(comment.Body, bodyOptions)
		if err != nil {
			return nil, fmt.Errorf("comment %s by %s: %w", comment.ID, comment.Author.Name, err)
		}
		author := comment.Author.Name
		if author == "" {
			author = comment.Author.ID
		}
		if author == "" {
			author = "Unknown author"
		}
		heading := author + " commented"
		if comment.Created != nil {
			heading += " " + comment.Created.Format("2006-01-02 15:04 -07:00")
		}
		text = text.Add(clicky.Text(heading, "font-bold")).NewLine().NewLine().Add(body).NewLine().NewLine()
		text = text.Add(markdown.Node{Kind: "raw-html", Source: "<hr/>", RawHTML: "<hr/>"}).NewLine().NewLine()
	}
	return text, nil
}
