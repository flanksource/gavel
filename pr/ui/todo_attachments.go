package ui

import (
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/flanksource/commons/logger"
	"github.com/google/uuid"
)

// attachmentsDir is where uploaded todo attachments (e.g. React Grab screenshots)
// are persisted. It's centralized under the gavel config dir rather than a
// workspace so it works for both file-backed (.todos) and grite todos. A package
// var so tests can point it at a temp dir, mirroring settingsPath in settings.go.
var attachmentsDir = filepath.Join(os.Getenv("HOME"), ".config", "gavel", "attachments")

// attachmentURLPrefix is the path the dashboard loads stored attachments from; the
// markdown body embeds <prefix>/<id> and handler.go serves it from attachmentsDir.
const attachmentURLPrefix = "/api/todos/attachments/"

// persistMultipartAttachments writes every uploaded file in the form to
// attachmentsDir under a random id (keeping the original extension) and returns a
// summary per file carrying the served URL. Unlike a metadata-only summary this
// actually stores the bytes so the todo can reference and render them.
func persistMultipartAttachments(form *multipart.Form) ([]todoAttachmentSummary, error) {
	if form == nil || len(form.File) == 0 {
		return nil, nil
	}
	if err := os.MkdirAll(attachmentsDir, 0o755); err != nil {
		return nil, fmt.Errorf("create attachments dir: %w", err)
	}
	var attachments []todoAttachmentSummary
	for field, headers := range form.File {
		for _, header := range headers {
			if header == nil || strings.TrimSpace(header.Filename) == "" {
				continue
			}
			summary, err := storeAttachment(field, header)
			if err != nil {
				return nil, err
			}
			attachments = append(attachments, summary)
		}
	}
	sort.Slice(attachments, func(i, j int) bool {
		if attachments[i].Field != attachments[j].Field {
			return attachments[i].Field < attachments[j].Field
		}
		return attachments[i].Filename < attachments[j].Filename
	})
	return attachments, nil
}

func storeAttachment(field string, header *multipart.FileHeader) (todoAttachmentSummary, error) {
	src, err := header.Open()
	if err != nil {
		return todoAttachmentSummary{}, fmt.Errorf("open attachment %q: %w", header.Filename, err)
	}
	defer src.Close() //nolint:errcheck

	contentType := header.Header.Get("Content-Type")
	id := uuid.NewString() + attachmentExt(header.Filename, contentType)
	dst, err := os.Create(filepath.Join(attachmentsDir, id))
	if err != nil {
		return todoAttachmentSummary{}, fmt.Errorf("create attachment file: %w", err)
	}
	written, err := io.Copy(dst, src)
	if closeErr := dst.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return todoAttachmentSummary{}, fmt.Errorf("write attachment %q: %w", header.Filename, err)
	}
	return todoAttachmentSummary{
		Field:       field,
		Filename:    header.Filename,
		ContentType: contentType,
		Size:        written,
		ID:          id,
		URL:         attachmentURLPrefix + id,
		IsImage:     strings.HasPrefix(strings.ToLower(contentType), "image/"),
	}, nil
}

// attachmentExt picks a file extension for the stored id, preferring the original
// filename's extension and falling back to one derived from the content type.
func attachmentExt(filename, contentType string) string {
	if ext := strings.ToLower(filepath.Ext(filename)); ext != "" {
		return ext
	}
	if exts, _ := mime.ExtensionsByType(contentType); len(exts) > 0 {
		return exts[0]
	}
	return ""
}

// handleTodoAttachment serves a previously stored attachment by id. The id must be
// a bare filename — any path separator or traversal segment is rejected so a
// request can't escape attachmentsDir.
func (s *Server) handleTodoAttachment(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" || id != path.Base(id) || strings.ContainsAny(id, `/\`) || strings.Contains(id, "..") {
		http.Error(w, "invalid attachment id", http.StatusBadRequest)
		return
	}
	full := filepath.Join(attachmentsDir, id)
	data, err := os.ReadFile(full)
	if err != nil {
		http.Error(w, "attachment not found", http.StatusNotFound)
		return
	}
	if ct := mime.TypeByExtension(filepath.Ext(id)); ct != "" {
		w.Header().Set("Content-Type", ct)
	}
	w.Header().Set("Cache-Control", "public, max-age=86400")
	if _, err := w.Write(data); err != nil {
		logger.Debugf("write attachment %s: %v", id, err)
	}
}

// todoBodyWithAttachments appends an "## Attachments" section to the todo body.
// Image attachments embed inline as markdown images (the dashboard's Markdown
// renderer renders them); everything else is a plain link.
func todoBodyWithAttachments(body string, attachments []todoAttachmentSummary) string {
	body = strings.TrimSpace(body)
	if len(attachments) == 0 {
		return body
	}
	var sb strings.Builder
	if body != "" {
		sb.WriteString(body)
		sb.WriteString("\n\n")
	}
	sb.WriteString("## Attachments\n\n")
	for _, a := range attachments {
		if a.IsImage {
			fmt.Fprintf(&sb, "![%s](%s)\n", a.Filename, a.URL)
		} else {
			fmt.Fprintf(&sb, "- [%s](%s)\n", a.Filename, a.URL)
		}
	}
	return strings.TrimSpace(sb.String())
}
