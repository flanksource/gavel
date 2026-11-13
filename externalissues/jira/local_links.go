package jira

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/flanksource/clicky/markdown"
)

func indexIssueFiles(root string) (map[string]string, error) {
	files := make(map[string]string)
	projects, err := os.ReadDir(root)
	if os.IsNotExist(err) {
		return files, nil
	}
	if err != nil {
		return nil, fmt.Errorf("index local Jira issues: %w", err)
	}
	for _, project := range projects {
		if !project.IsDir() || strings.HasPrefix(project.Name(), ".") {
			continue
		}
		entries, err := os.ReadDir(filepath.Join(root, project.Name()))
		if err != nil {
			return nil, fmt.Errorf("index Jira project %s: %w", project.Name(), err)
		}
		for _, entry := range entries {
			if !entry.Type().IsRegular() || !strings.HasSuffix(entry.Name(), ".md") || strings.HasPrefix(entry.Name(), ".") {
				continue
			}
			key := strings.TrimSuffix(entry.Name(), ".md")
			path := filepath.Join(root, project.Name(), entry.Name())
			if previous, ok := files[key]; ok {
				return nil, fmt.Errorf("ambiguous local issue %s: %s and %s", key, previous, path)
			}
			files[key] = path
		}
	}
	return files, nil
}

type localLinkOptions struct {
	Site  string
	Path  string
	Files map[string]string
}

func (options localLinkOptions) resolve(address string) (string, error) {
	parsed, err := url.Parse(address)
	if err != nil {
		return "", fmt.Errorf("parse issue link: %w", err)
	}
	site, err := url.Parse(options.Site)
	if err != nil {
		return "", fmt.Errorf("parse Jira site: %w", err)
	}
	key, jira := strings.CutPrefix(parsed.Path, "/browse/")
	if !jira || !strings.EqualFold(site.Host, parsed.Host) || !strings.EqualFold(site.Scheme, parsed.Scheme) {
		return address, nil
	}
	path, exists := options.Files[key]
	if !exists {
		return address, nil
	}
	relative, err := filepath.Rel(filepath.Dir(options.Path), path)
	if err != nil {
		return "", fmt.Errorf("relative issue link %s: %w", key, err)
	}
	return (&url.URL{Path: filepath.ToSlash(relative), Fragment: parsed.Fragment}).String(), nil
}

func resolveDocumentLinks(node markdown.Node, resolve func(string) (string, error)) (markdown.Node, error) {
	if resolve == nil {
		return node, nil
	}
	if node.Kind == "link" {
		address, err := resolve(node.Href)
		if err != nil {
			return markdown.Node{}, err
		}
		node.Href = address
	}
	node.Children, node.Items = slices.Clone(node.Children), slices.Clone(node.Items)
	for _, children := range [][]markdown.Node{node.Children, node.Items} {
		for index, child := range children {
			value, err := resolveDocumentLinks(child, resolve)
			if err != nil {
				return markdown.Node{}, err
			}
			children[index] = value
		}
	}
	return node, nil
}
