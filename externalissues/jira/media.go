package jira

import (
	"encoding/json"
	"fmt"
	"net/url"
	"slices"
	"strings"

	"github.com/flanksource/clicky/markdown"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/externalissues"
)

type mediaReference struct {
	ID   string
	Name string
}

func findMediaAttachment(reference mediaReference, attachments []externalissues.Attachment) (*externalissues.Attachment, error) {
	var match *externalissues.Attachment
	for i := range attachments {
		if attachments[i].ID == reference.ID && reference.ID != "" {
			match = &attachments[i]
			break
		}
	}
	if match == nil {
		for i := range attachments {
			if reference.Name == "" || attachments[i].Name != reference.Name {
				continue
			}
			if match != nil {
				return nil, fmt.Errorf("ambiguous Jira media %s: multiple attachments named %q", reference.ID, reference.Name)
			}
			match = &attachments[i]
		}
	}
	if match == nil {
		logger.Warnf("Jira media %s (%s): no matching attachment; retaining media reference", reference.ID, reference.Name)
		return nil, nil
	}
	if match.URL == "" {
		return nil, fmt.Errorf("Jira media %s attachment %s has no content URL", reference.ID, match.ID)
	}
	return match, nil
}

func attachmentMediaNode(node markdown.Node, attachment externalissues.Attachment) markdown.Node {
	node.Href = attachment.URL
	if node.String() == attachment.ID {
		node.Children = []markdown.Node{{Kind: "text", Text: attachment.Name}}
	}
	if strings.HasPrefix(strings.ToLower(attachment.MIMEType), "image/") {
		node.Kind = "image"
	}
	return node
}

func resolveADFMedia(node *ADFNode, attachments []externalissues.Attachment) error {
	if len(attachments) == 0 {
		return nil
	}
	if node.Type == "media" || node.Type == "mediaInline" {
		var attrs adfAttrs
		if err := json.Unmarshal(node.Attrs, &attrs); err != nil {
			return fmt.Errorf("decode media attributes: %w", err)
		}
		if attrs.URL == "" && attrs.ID != "" {
			attachment, err := findMediaAttachment(mediaReference{ID: attrs.ID, Name: attrs.Alt}, attachments)
			if err != nil {
				return err
			}
			node.media = attachment
		}
	}
	for i := range node.Content {
		if err := resolveADFMedia(&node.Content[i], attachments); err != nil {
			return err
		}
	}
	return nil
}

func resolveDocumentMedia(node markdown.Node, attachments []externalissues.Attachment) (markdown.Node, error) {
	if len(attachments) == 0 {
		return node, nil
	}
	if node.Kind == "link" && strings.HasPrefix(node.Href, "atlassian-media:") {
		address, err := url.Parse(node.Href)
		if err != nil {
			return markdown.Node{}, fmt.Errorf("parse media reference %q: %w", node.Href, err)
		}
		attachment, err := findMediaAttachment(mediaReference{ID: strings.TrimPrefix(address.Path, "/"), Name: node.String()}, attachments)
		if err != nil {
			return markdown.Node{}, err
		}
		if attachment != nil {
			node = attachmentMediaNode(node, *attachment)
		}
	}
	node.Children, node.Items = slices.Clone(node.Children), slices.Clone(node.Items)
	for _, children := range [][]markdown.Node{node.Children, node.Items} {
		for i := range children {
			resolved, err := resolveDocumentMedia(children[i], attachments)
			if err != nil {
				return markdown.Node{}, err
			}
			children[i] = resolved
		}
	}
	return node, nil
}
