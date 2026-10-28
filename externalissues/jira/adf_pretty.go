package jira

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/clicky/markdown"
)

func (n ADFNode) Render() (api.Textable, error) {
	return n.render("$")
}

func (n ADFNode) render(path string) (api.Textable, error) {
	if n.Type == "table" {
		return n.renderTable(path)
	}
	var attrs adfAttrs
	if len(n.Attrs) > 0 && string(n.Attrs) != "null" {
		if err := json.Unmarshal(n.Attrs, &attrs); err != nil {
			return nil, fmt.Errorf("ADF %s attrs: %w", path, err)
		}
	}
	children := make([]api.Textable, 0, len(n.Content))
	for i, child := range n.Content {
		value, err := child.render(fmt.Sprintf("%s.content[%d]", path, i))
		if err != nil {
			return nil, err
		}
		children = append(children, value)
	}
	inline := api.TextList(children).Join()
	blocks := api.TextList(children).Join(clicky.Text("\n\n"))
	switch n.Type {
	case "doc":
		if n.Version != 1 {
			return nil, fmt.Errorf("ADF %s: unsupported document version %d", path, n.Version)
		}
		return blocks, nil
	case "paragraph":
		return inline, nil
	case "text":
		var text api.Textable = clicky.Text(n.Text)
		for _, mark := range n.Marks {
			var markAttrs adfAttrs
			if len(mark.Attrs) > 0 {
				if err := json.Unmarshal(mark.Attrs, &markAttrs); err != nil {
					return nil, fmt.Errorf("ADF %s mark %s: %w", path, mark.Type, err)
				}
			}
			switch mark.Type {
			case "strong":
				text = api.Text{}.Add(text).Styles("font-bold")
			case "em":
				text = api.Text{}.Add(text).Styles("italic")
			case "strike":
				text = api.Text{}.Add(text).Styles("line-through")
			case "code":
				text = api.Text{}.Add(text).Styles("font-mono")
			case "link":
				text = clicky.Link(markAttrs.Href).Add(text)
			default:
				return nil, fmt.Errorf("ADF %s: unsupported mark %q", path, mark.Type)
			}
		}
		return text, nil
	case "hardBreak":
		return clicky.Text("\n"), nil
	case "heading":
		return clicky.Heading(attrs.Level, inline), nil
	case "blockquote":
		return clicky.Blockquote(blocks), nil
	case "listItem":
		return blocks, nil
	case "bulletList", "orderedList":
		list := clicky.List(children...)
		list.Numbered = n.Type == "orderedList"
		list.Bullet = clicky.Text("- ")
		return list, nil
	case "mention":
		return clicky.Text(attrs.Text), nil
	case "emoji":
		if attrs.ShortName == "" {
			return nil, fmt.Errorf("ADF %s: emoji missing shortName", path)
		}
		return clicky.Text(attrs.ShortName), nil
	case "mediaSingle", "mediaGroup":
		return api.TextList(children).Join(clicky.Text("\n")), nil
	case "media":
		target := attrs.URL
		if target == "" && attrs.ID != "" {
			target = "atlassian-media://" + url.PathEscape(attrs.Collection) + "/" + url.PathEscape(attrs.ID)
		}
		if target == "" {
			return nil, fmt.Errorf("ADF %s: media missing id or URL", path)
		}
		label := attrs.Alt
		if label == "" {
			label = attrs.ID
		}
		return clicky.Link(target).AddText(label), nil
	case "codeBlock":
		var code string
		for _, child := range n.Content {
			switch child.Type {
			case "text":
				code += child.Text
			case "hardBreak":
				code += "\n"
			default:
				return nil, fmt.Errorf("ADF %s: unsupported code block content %q", path, child.Type)
			}
		}
		return clicky.CodeBlock(attrs.Language, code), nil
	default:
		return nil, fmt.Errorf("ADF %s: unsupported node %q", path, n.Type)
	}
}

func (n ADFNode) renderTable(path string) (api.Textable, error) {
	if len(n.Content) == 0 || len(n.Content[0].Content) == 0 {
		return nil, fmt.Errorf("ADF %s: table has no cells", path)
	}
	first := n.Content[0]
	hasHeader := true
	for _, cell := range first.Content {
		hasHeader = hasHeader && cell.Type == "tableHeader"
	}
	table := markdown.Node{Kind: "table"}
	header := markdown.Node{Kind: "table_header"}
	for i, cell := range first.Content {
		if hasHeader {
			value, err := cell.renderTableCell(fmt.Sprintf("%s.content[0].content[%d]", path, i))
			if err != nil {
				return nil, err
			}
			header.Children = append(header.Children, value)
		} else {
			header.Children = append(header.Children, markdown.Node{Kind: "table_cell", Children: []markdown.Node{{Kind: "text", Text: fmt.Sprintf("Column %d", i+1)}}})
		}
	}
	table.Children = append(table.Children, header)
	for rowIndex, row := range n.Content {
		if rowIndex == 0 && hasHeader {
			continue
		}
		if row.Type != "tableRow" || len(row.Content) != len(header.Children) {
			return nil, fmt.Errorf("ADF %s: invalid table row %d", path, rowIndex)
		}
		values := markdown.Node{Kind: "table_row"}
		for i, cell := range row.Content {
			value, err := cell.renderTableCell(fmt.Sprintf("%s.content[%d].content[%d]", path, rowIndex, i))
			if err != nil {
				return nil, err
			}
			values.Children = append(values.Children, value)
		}
		table.Children = append(table.Children, values)
	}
	return table, nil
}

func (n ADFNode) renderTableCell(path string) (markdown.Node, error) {
	if n.Type != "tableHeader" && n.Type != "tableCell" {
		return markdown.Node{}, fmt.Errorf("ADF %s: invalid table cell %q", path, n.Type)
	}
	var attrs struct {
		Colspan int `json:"colspan"`
		Rowspan int `json:"rowspan"`
	}
	if len(n.Attrs) > 0 {
		if err := json.Unmarshal(n.Attrs, &attrs); err != nil {
			return markdown.Node{}, fmt.Errorf("ADF %s: %w", path, err)
		}
	}
	if attrs.Colspan > 1 || attrs.Rowspan > 1 {
		return markdown.Node{}, fmt.Errorf("ADF %s: merged table cells are unsupported", path)
	}
	result := markdown.Node{Kind: "table_cell"}
	for i, child := range n.Content {
		value, err := child.tableNode(fmt.Sprintf("%s.content[%d]", path, i))
		if err != nil {
			return markdown.Node{}, err
		}
		result.Children = append(result.Children, value)
	}
	return result, nil
}

func (n ADFNode) tableNode(path string) (markdown.Node, error) {
	result := markdown.Node{}
	switch n.Type {
	case "paragraph":
		result.Kind = "paragraph"
	case "text":
		result.Kind, result.Text = "text", n.Text
	case "hardBreak":
		result.Kind = "linebreak"
	default:
		return result, fmt.Errorf("ADF %s: unsupported table content %q", path, n.Type)
	}
	for i, child := range n.Content {
		value, err := child.tableNode(fmt.Sprintf("%s.content[%d]", path, i))
		if err != nil {
			return markdown.Node{}, err
		}
		result.Children = append(result.Children, value)
	}
	for _, mark := range n.Marks {
		switch mark.Type {
		case "strong", "em", "strike":
			kind := mark.Type
			if kind == "em" {
				kind = "emphasis"
			}
			result = markdown.Node{Kind: kind, Children: []markdown.Node{result}}
		case "link":
			var attrs adfAttrs
			if err := json.Unmarshal(mark.Attrs, &attrs); err != nil {
				return markdown.Node{}, fmt.Errorf("ADF %s link: %w", path, err)
			}
			result = markdown.Node{Kind: "link", Href: attrs.Href, Children: []markdown.Node{result}}
		default:
			return markdown.Node{}, fmt.Errorf("ADF %s: unsupported table mark %q", path, mark.Type)
		}
	}
	return result, nil
}
