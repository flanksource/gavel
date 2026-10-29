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
	Preview         bool
	Context         string
	ResolveIssueURL func(string) (string, error)
}

func renderDocument(document *externalissues.Document, options documentRenderOptions) (api.Textable, error) {
	if document == nil {
		return clicky.Text(""), nil
	}
	var node markdown.Node
	switch document.Format {
	case "text":
		var text string
		if err := json.Unmarshal(document.Content, &text); err != nil {
			return nil, err
		}
		node = markdown.Node{Kind: "text", Text: text}
	case "adf":
		node = renderADF(document.Content, strings.TrimSpace(options.Context+" $"))
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
	var text api.Text
	for _, comment := range comments {
		bodyOptions := options
		bodyOptions.Context = fmt.Sprintf("%s comment %s by %s", options.Context, comment.ID, comment.Author.Name)
		body, err := renderDocument(comment.Body, bodyOptions)
		if err != nil {
			return nil, fmt.Errorf("comment %s by %s: %w", comment.ID, comment.Author.Name, err)
		}
		heading := comment.Author.PrettyShort()
		if comment.Author.Name == "" && comment.Author.ID == "" {
			heading = clicky.Text("Comment " + comment.ID)
		}
		metadata := map[string]any{}
		if comment.ID != "" {
			metadata["ID"] = comment.ID
		}
		if comment.Created != nil {
			metadata["Created"] = clicky.Human(*comment.Created)
		}
		if comment.Updated != nil && (comment.Created == nil || !comment.Updated.Equal(*comment.Created)) {
			metadata["Updated"] = clicky.Human(*comment.Updated)
		}
		text = text.Add(clicky.Heading(3, heading)).NewLine().NewLine()
		if len(metadata) > 0 {
			text = text.Add(clicky.KeyMap(metadata)).NewLine().NewLine()
		}
		text = text.Add(body).NewLine().NewLine()
	}
	return text, nil
}
