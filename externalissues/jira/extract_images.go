package jira

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/flanksource/clicky/markdown"
	"golang.org/x/net/html"
)

type extractedDocument struct {
	Markdown string
	Images   []string
}

type documentImageOptions struct {
	Markdown  string
	SourceDir string
	OutputDir string
}

func localizeDocumentImages(options documentImageOptions) (result extractedDocument, returnErr error) {
	content, err := documentHTMLImages(options.Markdown)
	if err != nil {
		return result, err
	}
	document, err := markdown.ParseString(content)
	if err != nil {
		return result, err
	}
	source, err := os.OpenRoot(options.SourceDir)
	if err != nil {
		return result, err
	}
	defer func() { returnErr = errors.Join(returnErr, source.Close()) }()
	output, err := os.OpenRoot(options.OutputDir)
	if err != nil {
		return result, err
	}
	defer func() { returnErr = errors.Join(returnErr, output.Close()) }()
	var visit func(*markdown.Node) error
	visit = func(node *markdown.Node) error {
		if node.Kind == "image" {
			path, err := localizeDocumentImage(source, output, node.Href)
			if err != nil {
				return err
			}
			node.Href = filepath.ToSlash(path)
			if !slices.Contains(result.Images, node.Href) {
				result.Images = append(result.Images, node.Href)
			}
		}
		for _, nodes := range [][]markdown.Node{node.Children, node.Items} {
			for i := range nodes {
				if err := visit(&nodes[i]); err != nil {
					return err
				}
			}
		}
		return nil
	}
	if err := visit(&document.Root); err != nil {
		return result, err
	}
	result.Markdown = document.Markdown()
	return result, nil
}

func localizeDocumentImage(source, output *os.Root, address string) (string, error) {
	u, err := url.Parse(address)
	if err != nil || u.IsAbs() || u.Host != "" || !filepath.IsLocal(u.Path) || strings.Contains(u.Path, "\\") {
		return "", fmt.Errorf("unsafe extracted image path %q", address)
	}
	data, err := source.ReadFile(u.Path)
	if err != nil {
		return "", fmt.Errorf("read extracted image %q: %w", address, err)
	}
	return storeDocumentImage(output, data)
}

func storeDocumentImage(root *os.Root, data []byte) (string, error) {
	extensions := map[string]string{"image/png": ".png", "image/jpeg": ".jpg", "image/gif": ".gif", "image/webp": ".webp", "image/bmp": ".bmp", "image/x-icon": ".ico"}
	mime := http.DetectContentType(data)
	extension, ok := extensions[mime]
	if !ok && validateDocumentSVG(data) == nil {
		extension, ok = ".svg", true
	}
	if !ok {
		return "", fmt.Errorf("extracted image has unsupported type %s", mime)
	}
	path := filepath.Join("attachments", "cas", fmt.Sprintf("%x%s", sha256.Sum256(data), extension))
	if err := root.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return "", err
	}
	if existing, err := root.ReadFile(path); err == nil {
		if !bytes.Equal(data, existing) {
			return "", fmt.Errorf("image content does not match CAS filename %s", path)
		}
		return path, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", err
	}
	file, err := root.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return "", err
	}
	_, writeErr := file.Write(data)
	if err := errors.Join(writeErr, file.Close()); err != nil {
		return "", errors.Join(err, root.Remove(path))
	}
	return path, nil
}

func documentHTMLImages(content string) (string, error) {
	var output strings.Builder
	tokens := html.NewTokenizer(strings.NewReader(content))
	for {
		kind := tokens.Next()
		if kind == html.ErrorToken {
			if errors.Is(tokens.Err(), io.EOF) {
				return output.String(), nil
			}
			return "", tokens.Err()
		}
		raw := string(tokens.Raw())
		if kind != html.StartTagToken && kind != html.SelfClosingTagToken {
			output.WriteString(raw)
			continue
		}
		token := tokens.Token()
		if token.Data != "img" {
			output.WriteString(raw)
			continue
		}
		image := markdown.Node{Kind: "image"}
		var alt string
		for _, attr := range token.Attr {
			switch attr.Key {
			case "src":
				image.Href = attr.Val
			case "alt", "id":
				if alt == "" || attr.Key == "alt" {
					alt = attr.Val
				}
			}
		}
		image.Children = []markdown.Node{{Kind: "text", Text: alt}}
		output.WriteString(image.Markdown())
	}
}
