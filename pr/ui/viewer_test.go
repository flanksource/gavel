// Tests that the PR snapshot's viewer (the @me filter) is the requesting member's GitHub
// login in an organization's dashboard, and the token owner's login locally.
package ui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/tenant"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSnapshotViewerIsTheRequestingMember(t *testing.T) {
	srv := &Server{auth: github.AuthProbeResult{Login: "host-user"}}
	viewer := func(req *http.Request) string {
		t.Helper()
		rec := httptest.NewRecorder()
		srv.handleJSON(rec, req)
		var body struct{ Viewer string }
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
		return body.Viewer
	}

	assert.Equal(t, "host-user", viewer(httptest.NewRequest(http.MethodGet, "/api/prs", nil)), "locally, the token owner")

	user := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/api/prs", nil)
	req = req.WithContext(tenant.WithScope(req.Context(), tenant.Scope{OrgID: uuid.New(), UserID: &user, GitHubLogin: "ada"}))
	assert.Equal(t, "ada", viewer(req), "in an org's dashboard, the member making the request")
}
