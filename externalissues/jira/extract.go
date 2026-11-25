package jira

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/flanksource/clicky/exec"
	"github.com/flanksource/clicky/task"
	"github.com/flanksource/gavel/externalissues"
	"github.com/flanksource/gavel/internal/safefs"
)

//go:embed extract_pdf.py
var extractPDFScript []byte

//go:embed extract_docx.py
var extractDOCXScript []byte

func documentKind(attachment externalissues.Attachment) string {
	switch strings.ToLower(filepath.Ext(attachment.Name)) {
	case ".pdf":
		return "pdf"
	case ".docx":
		return "docx"
	}
	switch strings.ToLower(strings.TrimSpace(strings.Split(attachment.MIMEType, ";")[0])) {
	case "application/pdf":
		return "pdf"
	case "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
		return "docx"
	}
	return ""
}

type documentExtractOptions struct {
	Attachment externalissues.Attachment
	Path       string
	OutputDir  string
}

func extractDocument(ctx context.Context, options documentExtractOptions) (output extractedDocument, returnErr error) {
	return runJiraTask(ctx, "Extract "+options.Attachment.Name, func(ctx context.Context, progress *task.Task) (output extractedDocument, returnErr error) {
		progress.SetProgress(0, 2)
		progress.SetDescription("Validating document")
		attachment := options.Attachment
		source, err := filepath.Abs(options.Path)
		if err != nil {
			return output, err
		}
		if documentKind(attachment) == "docx" {
			if err := validateDOCXMedia(source); err != nil {
				return output, err
			}
		}
		dir, err := os.MkdirTemp(filepath.Dir(source), ".extract-")
		if err != nil {
			return output, err
		}
		defer func() {
			returnErr = errors.Join(returnErr, safefs.SafeRemove(dir, safefs.RemoveContext{AllowedRoot: filepath.Dir(source)}))
		}()
		destination := filepath.Join(dir, "document.md")
		script := filepath.Join(dir, "extract_docx.py")
		if err := os.WriteFile(script, extractDOCXScript, 0o600); err != nil {
			return output, err
		}
		args := []string{"--from", "docx2md==" + extractionTools("docx")["docx2md"], "python", script, source, destination}
		if documentKind(attachment) == "pdf" {
			script := filepath.Join(dir, "extract_pdf.py")
			if err := os.WriteFile(script, extractPDFScript, 0o600); err != nil {
				return output, err
			}
			args = []string{"--from", "pymupdf==" + extractionTools("pdf")["pymupdf"], "python", script, source, destination}
		}
		progress.SetDescription("Converting " + attachment.Name + " to Markdown")
		result := exec.NewExec("uvx", args...).WithoutShell().WithCwd(dir).WithContext(ctx).WithLogger(progress).WithTimeout(5 * time.Minute).WithCaptureLimit(16 * 1024).Run().Result()
		if result.Error != nil {
			return output, fmt.Errorf("extract %s with uvx: %w: %s", attachment.Name, result.Error, strings.TrimSpace(result.Stderr))
		}
		data, err := os.ReadFile(destination)
		if err != nil {
			return output, fmt.Errorf("read extracted Markdown: %w", err)
		}
		if strings.TrimSpace(string(data)) == "" {
			return output, fmt.Errorf("converter produced empty Markdown for %s", attachment.Name)
		}
		progress.SetProgress(1, 2)
		progress.SetDescription("Saving extracted images")
		if err := convertDocumentMetafiles(ctx, dir); err != nil {
			return output, err
		}
		output, err = localizeDocumentImages(documentImageOptions{Markdown: string(data), SourceDir: dir, OutputDir: options.OutputDir})
		if err != nil {
			return output, err
		}
		progress.SetProgress(2, 2)
		progress.SetDescription(fmt.Sprintf("Extracted Markdown and %d images", len(output.Images)))
		return output, nil
	})
}
