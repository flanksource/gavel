package jira

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/flanksource/clicky/api"
	"github.com/flanksource/clicky/markdown"
	"github.com/flanksource/commons/logger"
)

func (n ADFNode) Render() (api.Textable, error) {
	return n.toNode("$"), nil
}

func (n ADFNode) renderNode(path string) (markdown.Node, error) {
	var attrs adfAttrs
	if len(n.Attrs) > 0 && string(n.Attrs) != "null" {
		if err := json.Unmarshal(n.Attrs, &attrs); err != nil {
			return markdown.Node{}, fmt.Errorf("ADF %s attrs: %w", path, err)
		}
	}
	if n.Type == "table" {
		return n.tableNode(path)
	}
	node, err := n.nodeValue(path, attrs)
	if err != nil {
		return markdown.Node{}, fmt.Errorf("ADF %s: %w", path, err)
	}
	if n.Type == "codeBlock" {
		return node, nil
	}
	for i, child := range n.Content {
		value := child.toNode(fmt.Sprintf("%s.content[%d]", path, i))
		if n.Type == "bulletList" || n.Type == "orderedList" || n.Type == "taskList" {
			if n.Type == "taskList" && child.Type == "taskList" {
				value = markdown.Node{Kind: "list_item", Children: []markdown.Node{value}}
			} else if (n.Type == "taskList" && child.Type != "taskItem") || (n.Type != "taskList" && child.Type != "listItem") {
				return markdown.Node{}, fmt.Errorf("ADF %s: invalid list child %q", path, child.Type)
			}
			node.Items = append(node.Items, value)
		} else {
			node.Children = append(node.Children, value)
		}
	}
	if n.Type == "taskItem" {
		node.Children = []markdown.Node{{Kind: "paragraph", Children: node.Children}}
	}
	return n.markedNode(node, path), nil
}

func (n ADFNode) nodeValue(path string, attrs adfAttrs) (markdown.Node, error) {
	node := markdown.Node{}
	switch n.Type {
	case "doc":
		if n.Version != 1 {
			return node, fmt.Errorf("ADF %s: unsupported document version %d", path, n.Version)
		}
		node.Kind = "document"
	case "paragraph", "blockquote":
		node.Kind = n.Type
	case "text":
		node.Kind, node.Text = "text", n.Text
	case "hardBreak":
		node.Kind = "linebreak"
	case "inlineCard":
		return cardNode(attrs)
	case "date":
		return dateNode(attrs.Timestamp)
	case "rule":
		node.Kind = "thematic_break"
	case "panel":
		return panelNode(attrs.PanelType)
	case "heading":
		if attrs.Level < 1 || attrs.Level > 6 {
			return node, fmt.Errorf("ADF %s: invalid heading level %d", path, attrs.Level)
		}
		node.Kind, node.Level = "heading", attrs.Level
	case "listItem":
		node.Kind = "list_item"
	case "taskItem":
		if attrs.State != "TODO" && attrs.State != "DONE" {
			return node, fmt.Errorf("ADF %s: invalid task state %q", path, attrs.State)
		}
		checked := attrs.State == "DONE"
		node.Kind, node.Checked = "list_item", &checked
	case "bulletList", "orderedList", "taskList":
		node.Kind, node.Ordered = "list", n.Type == "orderedList"
	case "mention":
		node.Kind, node.Text = "text", attrs.Text
	case "emoji":
		if attrs.ShortName == "" {
			return node, fmt.Errorf("ADF %s: emoji missing shortName", path)
		}
		node.Kind, node.Text = "text", attrs.ShortName
		if attrs.Text != "" {
			node.Text = attrs.Text
		}
	case "mediaSingle", "mediaGroup":
		node.Kind = "document"
	case "media", "mediaInline":
		return mediaNode(attrs, path)
	case "codeBlock":
		return n.codeNode(attrs, path)
	case "tableRow":
		node.Kind = "table_row"
	case "tableHeader", "tableCell":
		if attrs.Colspan > 1 || attrs.Rowspan > 1 {
			return node, fmt.Errorf("ADF %s: merged table cells are unsupported", path)
		}
		node.Kind = "table_cell"
	default:
		return node, fmt.Errorf("ADF %s: unsupported node %q", path, n.Type)
	}
	return node, nil
}

func (n ADFNode) markedNode(node markdown.Node, path string) markdown.Node {
	for _, mark := range n.Marks {
		if mark.Type == "code" && n.Type == "text" {
			node.Kind = "code"
		}
	}
	for index, mark := range n.Marks {
		value, err := n.applyMark(node, mark)
		if err != nil {
			logger.Warnf("ADF %s.marks[%d] %s: %v", path, index, mark.Type, err)
			continue
		}
		node = value
	}
	return node
}

func (n ADFNode) applyMark(node markdown.Node, mark ADFMark) (markdown.Node, error) {
	switch mark.Type {
	case "code":
		if n.Type != "text" {
			return node, fmt.Errorf("code mark requires text")
		}
	case "strong", "strike":
		return markdown.Node{Kind: mark.Type, Children: []markdown.Node{node}}, nil
	case "em":
		return markdown.Node{Kind: "emphasis", Children: []markdown.Node{node}}, nil
	case "underline", "subsup":
		if n.Type != "text" {
			return node, fmt.Errorf("%s mark requires text", mark.Type)
		}
		if mark.Type == "subsup" {
			return subSupADFNode(node, mark.Attrs)
		}
		return styledADFNode(node, "u", ""), nil
	case "border":
		if n.Type != "media" && n.Type != "mediaInline" {
			return node, fmt.Errorf("border mark requires media")
		}
		return borderedADFNode(node, mark.Attrs)
	case "textColor":
		if n.Type != "text" {
			return node, fmt.Errorf("textColor mark requires text")
		}
		return coloredADFNode(node, mark.Attrs)
	case "link":
		var attrs adfAttrs
		if err := json.Unmarshal(mark.Attrs, &attrs); err != nil {
			return node, fmt.Errorf("decode link: %w", err)
		}
		if attrs.Href == "" {
			return node, fmt.Errorf("link missing href")
		}
		return markdown.Node{Kind: "link", Href: attrs.Href, Children: []markdown.Node{node}}, nil
	default:
		return node, fmt.Errorf("unsupported mark %q", mark.Type)
	}
	return node, nil
}

func mediaNode(attrs adfAttrs, path string) (markdown.Node, error) {
	target := attrs.URL
	if target == "" && attrs.ID != "" {
		target = "atlassian-media://" + url.PathEscape(attrs.Collection) + "/" + url.PathEscape(attrs.ID)
	}
	if target == "" {
		return markdown.Node{}, fmt.Errorf("ADF %s: media missing id or URL", path)
	}
	label := attrs.Alt
	if label == "" {
		label = attrs.ID
	}
	return markdown.Node{Kind: "link", Href: target, Children: []markdown.Node{{Kind: "text", Text: label}}}, nil
}

func (n ADFNode) codeNode(attrs adfAttrs, path string) (markdown.Node, error) {
	node := markdown.Node{Kind: "code_block", Language: attrs.Language}
	for _, child := range n.Content {
		switch child.Type {
		case "text":
			node.Source += child.Text
		case "hardBreak":
			node.Source += "\n"
		default:
			return markdown.Node{}, fmt.Errorf("ADF %s: unsupported code block content %q", path, child.Type)
		}
	}
	return node, nil
}

func (n ADFNode) tableNode(path string) (markdown.Node, error) {
	if len(n.Content) == 0 || len(n.Content[0].Content) == 0 {
		return markdown.Node{}, fmt.Errorf("ADF %s: table has no cells", path)
	}
	columns := len(n.Content[0].Content)
	hasHeader := true
	for _, cell := range n.Content[0].Content {
		hasHeader = hasHeader && cell.Type == "tableHeader"
	}
	table := markdown.Node{Kind: "table"}
	if !hasHeader {
		header := markdown.Node{Kind: "table_header"}
		for i := range columns {
			header.Children = append(header.Children, markdown.Node{Kind: "table_cell", Children: []markdown.Node{{Kind: "text", Text: fmt.Sprintf("Column %d", i+1)}}})
		}
		table.Children = append(table.Children, header)
	}
	for rowIndex, row := range n.Content {
		if row.Type != "tableRow" || len(row.Content) != columns {
			return markdown.Node{}, fmt.Errorf("ADF %s: invalid table row %d", path, rowIndex)
		}
		for _, cell := range row.Content {
			if cell.Type != "tableHeader" && cell.Type != "tableCell" {
				return markdown.Node{}, fmt.Errorf("ADF %s: invalid table cell %q", path, cell.Type)
			}
		}
		value := row.toNode(fmt.Sprintf("%s.content[%d]", path, rowIndex))
		if rowIndex == 0 && hasHeader {
			value.Kind = "table_header"
		}
		table.Children = append(table.Children, value)
	}
	return table, nil
}
