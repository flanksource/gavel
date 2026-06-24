package ui

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/flanksource/gavel/github"
)

// pngBytes is a 1x1 transparent PNG — enough to assert the bytes round-trip.
var pngBytes = []byte{
	0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00, 0x00, 0x00, 0x0d,
	0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
	0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0x15, 0xc4, 0x89,
}

// uploadScreenshot creates a todo with one PNG attachment and returns the create
// response so a test can inspect the persisted attachment.
func uploadScreenshot(t *testing.T, s *Server, workDir string) todoNewResponse {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if err := writer.WriteField("title", "Screenshot todo"); err != nil {
		t.Fatalf("write field: %v", err)
	}
	part, err := writer.CreatePart(map[string][]string{
		"Content-Disposition": {`form-data; name="attachment"; filename="screen.png"`},
		"Content-Type":        {"image/png"},
	})
	if err != nil {
		t.Fatalf("create file part: %v", err)
	}
	if _, err := part.Write(pngBytes); err != nil {
		t.Fatalf("write file part: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart: %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/todos/new?provider=todos&dir="+url.QueryEscape(workDir), &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	s.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want 201; body = %q", rec.Code, rec.Body.String())
	}
	var resp todoNewResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal create response: %v", err)
	}
	return resp
}

func TestAttachmentPersistAndServe(t *testing.T) {
	attachmentsDir = t.TempDir()
	workDir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: workDir}}

	resp := uploadScreenshot(t, s, workDir)
	if len(resp.Attachments) != 1 {
		t.Fatalf("attachments = %+v, want 1", resp.Attachments)
	}
	att := resp.Attachments[0]
	if !att.IsImage || att.URL != attachmentURLPrefix+att.ID {
		t.Fatalf("unexpected attachment: %+v", att)
	}

	// The bytes are persisted on disk under the stored id.
	stored, err := os.ReadFile(filepath.Join(attachmentsDir, att.ID))
	if err != nil {
		t.Fatalf("read stored attachment: %v", err)
	}
	if !bytes.Equal(stored, pngBytes) {
		t.Fatalf("stored bytes != uploaded bytes")
	}

	// The body embeds the image inline as markdown pointing at the served URL.
	wantImg := "![screen.png](" + att.URL + ")"
	if !strings.Contains(resp.Todo.Body, wantImg) {
		t.Fatalf("body missing image embed %q: %q", wantImg, resp.Todo.Body)
	}

	// GET serves the bytes back with an image content-type.
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, att.URL, nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("serve status = %d, want 200", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "image/png") {
		t.Fatalf("content-type = %q, want image/png", ct)
	}
	if !bytes.Equal(rec.Body.Bytes(), pngBytes) {
		t.Fatalf("served bytes != uploaded bytes")
	}
}

func TestAttachmentRejectsTraversal(t *testing.T) {
	attachmentsDir = t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: t.TempDir()}}
	for _, id := range []string{"..%2f..%2fetc%2fpasswd", "..", "foo/bar"} {
		rec := httptest.NewRecorder()
		s.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, attachmentURLPrefix+id, nil))
		if rec.Code == http.StatusOK {
			t.Fatalf("id %q served 200, want rejection", id)
		}
	}
}
