package jira

import (
	"encoding/json"
	"fmt"

	"github.com/flanksource/clicky/markdown"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/externalissues"
)

type adfRenderOptions struct {
	Path        string
	Attachments []externalissues.Attachment
}

func renderADF(raw json.RawMessage, options adfRenderOptions) (markdown.Node, error) {
	var doc ADFNode
	if err := json.Unmarshal(raw, &doc); err != nil {
		logger.Warnf("ADF %s: decode ADF: %v; retaining raw JSON", options.Path, err)
		return markdown.Node{Kind: "code_block", Language: "json", Source: string(raw)}, nil
	}
	if err := resolveADFMedia(&doc, options.Attachments); err != nil {
		return markdown.Node{}, fmt.Errorf("ADF %s: %w", options.Path, err)
	}
	return doc.toNode(options.Path), nil
}

func (n ADFNode) toNode(path string) markdown.Node {
	node, err := n.renderNode(path)
	if err == nil {
		return node
	}
	logger.Warnf("%v; retaining ADF content", err)
	node = markdown.Node{Kind: "document"}
	if n.Text != "" {
		node.Children = append(node.Children, markdown.Node{Kind: "text", Text: n.Text})
	}
	for index, child := range n.Content {
		node.Children = append(node.Children, child.toNode(fmt.Sprintf("%s.content[%d]", path, index)))
	}
	if len(node.Children) > 0 {
		return n.markedNode(node, path)
	}
	raw, err := json.Marshal(n)
	if err != nil {
		logger.Warnf("ADF %s: retain node: %v", path, err)
		return markdown.Node{Kind: "text", Text: fmt.Sprintf("[ADF %s: %v]", n.Type, err)}
	}
	return markdown.Node{Kind: "code", Text: string(raw)}
}
