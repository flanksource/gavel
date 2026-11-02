package jira

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/flanksource/clicky/markdown"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/externalissues"
	"github.com/flanksource/gavel/externalissues/syncdoc"
)

type syncDownloadOptions struct {
	Attachment *externalissues.Attachment
	Record     *syncdoc.Attachment
	Root       *os.Root
	Stage      string
	Relative   string
	Force      bool
}

func (c *Client) syncDownload(ctx context.Context, o syncDownloadOptions) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if !o.Force && o.Record.Local != nil {
		cached, err := syncdoc.ReadArtifact(o.Root, *o.Record.Local)
		if err != nil && !errors.Is(err, os.ErrNotExist) && !errors.Is(err, syncdoc.ErrArtifactHashMismatch) {
			return fmt.Errorf("read cached attachment %s: %w", o.Attachment.ID, err)
		}
		if err == nil && o.Attachment.Size > 0 && int64(len(cached)) != o.Attachment.Size {
			err = fmt.Errorf("expected %d bytes, cached %d", o.Attachment.Size, len(cached))
		}
		if err == nil {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := os.WriteFile(filepath.Join(o.Stage, o.Relative), cached, 0600); err != nil {
				return err
			}
			copy := *o.Record.Local
			copy.Path = filepath.ToSlash(o.Relative)
			o.Record.Local = &copy
			logger.Infof("Reusing verified attachment %s (%s)", o.Attachment.ID, o.Attachment.Name)
			return nil
		}
		logger.Warnf("attachment %s cache: %v; refreshing", o.Attachment.ID, err)
	}
	result, err := c.DownloadAttachment(ctx, DownloadOptions{ID: o.Attachment.ID, Output: filepath.Join(o.Stage, o.Relative)})
	if err != nil {
		return err
	}
	previous := o.Record.Local
	o.Record.Local = &syncdoc.Artifact{Path: filepath.ToSlash(o.Relative), SHA256: result.SHA256, ETag: result.ETag, DownloadedAt: time.Now().UTC()}
	if previous != nil {
		o.Record.Local.Extra = previous.Extra
	}
	return nil
}

type syncExtractionOptions struct {
	Attachment externalissues.Attachment
	Record     *syncdoc.Attachment
	Root       *os.Root
	Stage      string
	Path       string
	Force      bool
	Result     *SyncResult
}

func extractionTools(kind string) map[string]string {
	if kind == "pdf" {
		return map[string]string{"pymupdf": "2.0"}
	}
	return map[string]string{"docx2md": "1.0.5", "emf-converter": "4.24.0"}
}

func extractionFingerprint(kind string) string {
	script := extractDOCXScript
	if kind == "pdf" {
		script = extractPDFScript
	}
	return syncdoc.Hash(append([]byte(fmt.Sprintf("pipeline-v1:%s:%v:", kind, extractionTools(kind))), script...))
}

func syncExtraction(ctx context.Context, o syncExtractionOptions) (extractedDocument, error) {
	fingerprint := extractionFingerprint(documentKind(o.Attachment))
	if !o.Force && o.Record.Extraction != nil && o.Record.Extraction.SourceSHA256 == o.Record.Local.SHA256 && o.Record.Extraction.PipelineFingerprint == fingerprint {
		body, err := cachedExtraction(o)
		if err == nil {
			logger.Infof("Reusing extraction for attachment %s", o.Attachment.ID)
			return body, nil
		}
		logger.Warnf("attachment %s extraction cache: %v; regenerating", o.Attachment.ID, err)
	}
	body, err := extractDocument(ctx, documentExtractOptions{Attachment: o.Attachment, Path: o.Path, OutputDir: o.Stage})
	if err != nil {
		return body, err
	}
	record := &syncdoc.Extraction{SourceSHA256: o.Record.Local.SHA256, PipelineFingerprint: fingerprint, Tools: extractionTools(documentKind(o.Attachment)), ExtractedAt: time.Now().UTC()}
	if o.Record.Extraction != nil {
		record.Extra = o.Record.Extraction.Extra
	}
	for _, path := range body.Images {
		data, err := os.ReadFile(filepath.Join(o.Stage, filepath.FromSlash(path)))
		if err != nil {
			return body, err
		}
		mime := http.DetectContentType(data)
		if validateDocumentSVG(data) == nil {
			mime = "image/svg+xml"
		}
		record.Images = append(record.Images, syncdoc.Artifact{Path: path, SHA256: syncdoc.Hash(data), MIMEType: mime})
	}
	cache, err := rebaseExtraction(body.Markdown, true)
	if err != nil {
		return body, err
	}
	record.Markdown = syncdoc.Artifact{Path: "attachments/extracted/" + syncdoc.Hash([]byte(cache)) + ".md", SHA256: syncdoc.Hash([]byte(cache))}
	if err := os.MkdirAll(filepath.Join(o.Stage, "attachments/extracted"), 0700); err != nil {
		return body, err
	}
	if err := os.WriteFile(filepath.Join(o.Stage, record.Markdown.Path), []byte(cache), 0600); err != nil {
		return body, err
	}
	o.Record.Extraction = record
	o.Result.artifacts = append(o.Result.artifacts, filepath.Join(filepath.Dir(o.Result.Path), record.Markdown.Path))
	return body, nil
}

func cachedExtraction(o syncExtractionOptions) (extractedDocument, error) {
	record := o.Record.Extraction
	data, err := syncdoc.ReadArtifact(o.Root, record.Markdown)
	if err != nil {
		return extractedDocument{}, err
	}
	body, err := rebaseExtraction(string(data), false)
	if err != nil {
		return extractedDocument{}, err
	}
	result := extractedDocument{Markdown: body}
	for _, artifact := range append(append([]syncdoc.Artifact{}, record.Images...), record.Markdown) {
		data, err := syncdoc.ReadArtifact(o.Root, artifact)
		if err != nil {
			return result, err
		}
		target := filepath.Join(o.Stage, artifact.Path)
		if err := os.MkdirAll(filepath.Dir(target), 0700); err != nil {
			return result, err
		}
		if err := os.WriteFile(target, data, 0600); err != nil {
			return result, err
		}
	}
	for _, image := range record.Images {
		result.Images = append(result.Images, image.Path)
	}
	o.Result.artifacts = append(o.Result.artifacts, filepath.Join(filepath.Dir(o.Result.Path), record.Markdown.Path))
	return result, nil
}

func rebaseExtraction(body string, toCache bool) (string, error) {
	doc, err := markdown.ParseString(body)
	if err != nil {
		return "", err
	}
	var visit func(*markdown.Node)
	visit = func(n *markdown.Node) {
		if n.Kind == "image" {
			if toCache {
				n.Href = strings.Replace(n.Href, "attachments/cas/", "../cas/", 1)
			} else {
				n.Href = strings.Replace(n.Href, "../cas/", "attachments/cas/", 1)
			}
		}
		for _, nodes := range [][]markdown.Node{n.Children, n.Items} {
			for i := range nodes {
				visit(&nodes[i])
			}
		}
	}
	visit(&doc.Root)
	return doc.Markdown(), nil
}
