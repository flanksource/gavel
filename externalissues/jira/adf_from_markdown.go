package jira

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/flanksource/clicky/markdown"
)

func MarkdownToADF(source string) (ADFNode, error) {
	if strings.TrimSpace(source) == "" {
		return ADFNode{}, fmt.Errorf("Jira comment body must not be empty")
	}
	document, err := markdown.ParseString(source)
	if err != nil {
		return ADFNode{}, err
	}
	return markdownToADF(document.Root, nil)
}

func markdownToADF(node markdown.Node, marks []ADFMark) (ADFNode, error) {
	value := ADFNode{Type: node.Kind}
	switch node.Kind {
	case "document":
		value.Type, value.Version = "doc", 1
	case "paragraph", "blockquote":
	case "text":
		return ADFNode{Type: "text", Text: node.Text, Marks: marks}, nil
	case "strong", "emphasis", "strike", "link":
		mark := ADFMark{Type: node.Kind}
		if node.Kind == "emphasis" {
			mark.Type = "em"
		}
		if node.Kind == "link" {
			attrs, err := json.Marshal(map[string]string{"href": node.Href})
			if err != nil {
				return ADFNode{}, err
			}
			mark.Attrs = attrs
		}
		value.Type = "inlineContainer"
		marks = append(append([]ADFMark{}, marks...), mark)
	case "code":
		return ADFNode{Type: "text", Text: node.Text, Marks: append(append([]ADFMark{}, marks...), ADFMark{Type: "code"})}, nil
	case "linebreak":
		value.Type = "hardBreak"
	case "heading":
		attrs, err := json.Marshal(map[string]int{"level": node.Level})
		if err != nil {
			return ADFNode{}, err
		}
		value.Attrs = attrs
	case "code_block":
		attrs, err := json.Marshal(map[string]string{"language": node.Language})
		if err != nil {
			return ADFNode{}, err
		}
		return ADFNode{Type: "codeBlock", Attrs: attrs, Content: []ADFNode{{Type: "text", Text: node.Source}}}, nil
	case "list":
		value.Type = "bulletList"
		if node.Ordered {
			value.Type = "orderedList"
		}
		node.Children = node.Items
	case "list_item":
		if node.Checked != nil {
			return ADFNode{}, fmt.Errorf("Jira Markdown task lists are not supported")
		}
		value.Type = "listItem"
	case "table":
	case "table_header", "table_row":
		value.Type = "tableRow"
	case "table_cell":
		value.Type = "tableCell"
		node.Children = []markdown.Node{{Kind: "paragraph", Children: node.Children}}
	default:
		return ADFNode{}, fmt.Errorf("unsupported Jira Markdown node %q", node.Kind)
	}
	for _, child := range node.Children {
		converted, err := markdownToADF(child, marks)
		if err != nil {
			return ADFNode{}, err
		}
		if converted.Type == "inlineContainer" {
			value.Content = append(value.Content, converted.Content...)
		} else {
			value.Content = append(value.Content, converted)
		}
	}
	return value, nil
}
