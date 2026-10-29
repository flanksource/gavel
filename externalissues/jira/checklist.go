package jira

import (
	"strings"

	"github.com/flanksource/clicky/markdown"
)

func checklistDocument(node markdown.Node) markdown.Node {
	for index, child := range node.Children {
		node.Children[index] = checklistDocument(child)
	}
	for index, child := range node.Items {
		node.Items[index] = checklistDocument(child)
	}
	if node.Kind != "paragraph" {
		return node
	}
	lines := [][]markdown.Node{{}}
	for _, child := range node.Children {
		if child.Kind == "linebreak" {
			lines = append(lines, nil)
		} else if child.Kind == "text" {
			for index, part := range strings.Split(child.Text, "\n") {
				if index > 0 {
					lines = append(lines, nil)
				}
				if part != "" {
					item := child
					item.Text = part
					lines[len(lines)-1] = append(lines[len(lines)-1], item)
				}
			}
		} else {
			lines[len(lines)-1] = append(lines[len(lines)-1], child)
		}
	}
	var blocks []markdown.Node
	found := false
	for _, line := range lines {
		item, checked := checklistItem(line)
		if checked {
			found = true
			if len(blocks) == 0 || blocks[len(blocks)-1].Kind != "list" {
				blocks = append(blocks, markdown.Node{Kind: "list"})
			}
			blocks[len(blocks)-1].Items = append(blocks[len(blocks)-1].Items, item)
		} else {
			blocks = append(blocks, markdown.Node{Kind: "paragraph", Children: line})
		}
	}
	if !found {
		return node
	}
	return markdown.Node{Kind: "document", Children: blocks}
}

func checklistItem(line []markdown.Node) (markdown.Node, bool) {
	if len(line) == 0 || line[0].Kind != "text" {
		return markdown.Node{}, false
	}
	text := strings.TrimLeft(line[0].Text, " \t")
	for _, marker := range []struct {
		Text    string
		Checked bool
	}{
		{":cross_mark:", false}, {":x:", false}, {"❌", false}, {"✗", false},
		{":white_check_mark:", true}, {":heavy_check_mark:", true}, {":check_mark:", true}, {"✅", true}, {"✔", true}, {"☑", true},
	} {
		if rest, ok := strings.CutPrefix(text, marker.Text); ok {
			line[0].Text = strings.TrimLeft(rest, " \t")
			checked := marker.Checked
			return markdown.Node{Kind: "list_item", Checked: &checked, Children: []markdown.Node{{Kind: "paragraph", Children: line}}}, true
		}
	}
	return markdown.Node{}, false
}

func hasChecklist(node markdown.Node) bool {
	if node.Checked != nil {
		return true
	}
	for _, child := range node.Children {
		if hasChecklist(child) {
			return true
		}
	}
	for _, child := range node.Items {
		if hasChecklist(child) {
			return true
		}
	}
	return false
}
