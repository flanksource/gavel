// Tests that todo attachments go to the configured blob store, and that a Server scoped to an
// organization reads and writes only that organization's attachments.
package ui

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/flanksource/gavel/blob"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// uploadAttachment stores one PNG through the upload endpoint and returns its id.
func uploadAttachment(t *testing.T, s *Server) string {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreatePart(map[string][]string{
		"Content-Disposition": {`form-data; name="attachment"; filename="screen.png"`},
		"Content-Type":        {"image/png"},
	})
	require.NoError(t, err)
	_, err = part.Write(pngBytes)
	require.NoError(t, err)
	require.NoError(t, writer.Close())
	req := httptest.NewRequest(http.MethodPost, "/api/todos/attachments", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()
	s.handleTodoAttachmentUpload(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	var resp todoAttachmentUploadResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Len(t, resp.Attachments, 1)
	return resp.Attachments[0].ID
}

func getAttachment(s *Server, id string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, attachmentURLPrefix+id, nil)
	req.SetPathValue("id", id)
	rec := httptest.NewRecorder()
	s.handleTodoAttachment(rec, req)
	return rec
}

func TestAttachmentsGoToTheConfiguredStore(t *testing.T) {
	attachmentsDir = t.TempDir()
	bucket := t.TempDir()
	s := &Server{}
	s.SetBlobStore(blob.Local(bucket))

	id := uploadAttachment(t, s)
	assert.FileExists(t, filepath.Join(bucket, "attachments", id))
	entries, err := os.ReadDir(attachmentsDir)
	require.NoError(t, err)
	assert.Empty(t, entries, "nothing is written to the local attachments directory")

	rec := getAttachment(s, id)
	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, pngBytes, rec.Body.Bytes())
}

func TestAttachmentsAreScopedToTheOrganization(t *testing.T) {
	attachmentsDir = t.TempDir()
	bucket := blob.Local(t.TempDir())
	orgA, orgB := uuid.New(), uuid.New()
	a, b := &Server{}, &Server{}
	a.SetOrg(orgA, nil)
	b.SetOrg(orgB, nil)
	a.SetBlobStore(bucket)
	b.SetBlobStore(bucket)

	id := uploadAttachment(t, a)
	assert.Equal(t, http.StatusOK, getAttachment(a, id).Code)
	assert.Equal(t, http.StatusNotFound, getAttachment(b, id).Code, "another org's attachment id is a 404")
	exists, err := bucket.Exists(t.Context(), "orgs/"+orgA.String()+"/attachments/"+id)
	require.NoError(t, err)
	assert.True(t, exists)
}
