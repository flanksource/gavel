package ui

import (
	"bytes"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/flanksource/gavel/procfile"
)

// withProject points projectsPath at a temp file holding a single project whose
// dir contains the given Procfile body (empty body = no Procfile written). It
// returns the project's directory.
func withProject(t *testing.T, name, repo, procfileBody string) string {
	t.Helper()
	dir := t.TempDir()
	if procfileBody != "" {
		if err := os.WriteFile(filepath.Join(dir, "Procfile"), []byte(procfileBody), 0o644); err != nil {
			t.Fatalf("write Procfile: %v", err)
		}
	}
	orig := projectsPath
	projectsPath = filepath.Join(t.TempDir(), "projects.json")
	t.Cleanup(func() { projectsPath = orig })
	SaveProjects([]Project{{Name: name, Dir: dir, Repos: []string{repo}}})
	return dir
}

func TestHandleProjectsListsHasProcfile(t *testing.T) {
	withProject(t, "gavel", "flanksource/gavel", "web: echo hi\n")

	rec := httptest.NewRecorder()
	(&Server{}).handleProjects(rec, httptest.NewRequest("GET", "/api/projects", nil))
	if rec.Code != 200 {
		t.Fatalf("status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}
	var got []projectInfo
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(got) != 1 || got[0].Name != "gavel" || !got[0].HasProcfile {
		t.Errorf("GET /api/projects = %+v, want one gavel project with hasProcfile=true", got)
	}
}

func TestHandleProjectsCreate(t *testing.T) {
	withProject(t, "gavel", "flanksource/gavel", "")

	body := `{"name":"infra","dir":"/srv/infra","repos":["acme/infra"]}`
	rec := httptest.NewRecorder()
	(&Server{}).handleProjects(rec, httptest.NewRequest("POST", "/api/projects", strings.NewReader(body)))
	if rec.Code != http.StatusCreated {
		t.Fatalf("POST status = %d, want 201; body = %q", rec.Code, rec.Body.String())
	}
	if _, ok := projectByName(LoadProjects(), "infra"); !ok {
		t.Error("POST /api/projects did not persist the new project")
	}
}

func TestHandleProjectsCreateConflict(t *testing.T) {
	withProject(t, "gavel", "flanksource/gavel", "")

	body := `{"name":"gavel","dir":"/elsewhere"}`
	rec := httptest.NewRecorder()
	(&Server{}).handleProjects(rec, httptest.NewRequest("POST", "/api/projects", strings.NewReader(body)))
	if rec.Code != http.StatusConflict {
		t.Fatalf("POST duplicate name status = %d, want 409; body = %q", rec.Code, rec.Body.String())
	}
	if p, _ := projectByName(LoadProjects(), "gavel"); p.Dir == "/elsewhere" {
		t.Error("conflicting create overwrote the existing project")
	}
}

func TestHandleProjectsPostValidation(t *testing.T) {
	withProject(t, "gavel", "flanksource/gavel", "")
	tests := []struct {
		name    string
		body    string
		wantSts int
	}{
		{"malformed json", `{not json`, 400},
		{"missing name", `{"dir":"/srv"}`, 400},
		{"missing dir", `{"name":"x"}`, 400},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			(&Server{}).handleProjects(rec, httptest.NewRequest("POST", "/api/projects", strings.NewReader(tc.body)))
			if rec.Code != tc.wantSts {
				t.Errorf("status = %d, want %d; body = %q", rec.Code, tc.wantSts, rec.Body.String())
			}
		})
	}
}

func TestHandleProcStatusForProject(t *testing.T) {
	withProject(t, "gavel", "flanksource/gavel", "web: echo hi\nworker: echo bye\n")

	rec := httptest.NewRecorder()
	(&Server{}).handleProcStatus(rec, httptest.NewRequest("GET", "/api/proc/status?project=gavel", nil))
	if rec.Code != 200 {
		t.Fatalf("status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}
	var got procStatus
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !got.HasProcfile {
		t.Fatalf("hasProcfile = false, want true; body = %q", rec.Body.String())
	}
	if got.Running {
		t.Error("running = true with no supervisor started, want false")
	}
	if len(got.Processes) != 2 {
		t.Fatalf("processes = %d, want 2 (web, worker)", len(got.Processes))
	}
	for _, p := range got.Processes {
		if p.Status != "stopped" {
			t.Errorf("process %q status = %q, want stopped", p.Name, p.Status)
		}
	}
}

func TestHandleProcStatusNoProcfileIsNotError(t *testing.T) {
	withProject(t, "gavel", "flanksource/gavel", "") // dir exists, no Procfile

	rec := httptest.NewRecorder()
	(&Server{}).handleProcStatus(rec, httptest.NewRequest("GET", "/api/proc/status?project=gavel", nil))
	if rec.Code != 200 {
		t.Fatalf("status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}
	var got procStatus
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.HasProcfile {
		t.Error("hasProcfile = true for a dir without a Procfile, want false")
	}
}

func TestHandleProcStatusAllProjectsKeyedByRepo(t *testing.T) {
	withProject(t, "gavel", "flanksource/gavel", "web: echo hi\n")

	rec := httptest.NewRecorder()
	(&Server{}).handleProcStatus(rec, httptest.NewRequest("GET", "/api/proc/status", nil))
	if rec.Code != 200 {
		t.Fatalf("status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}
	var got map[string]procStatus
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	st, ok := got["flanksource/gavel"]
	if !ok || !st.HasProcfile {
		t.Errorf("status map = %+v, want flanksource/gavel with hasProcfile=true", got)
	}
}

func TestHandleProcStatusUnknownProject(t *testing.T) {
	withProject(t, "gavel", "flanksource/gavel", "web: echo hi\n")

	rec := httptest.NewRecorder()
	(&Server{}).handleProcStatus(rec, httptest.NewRequest("GET", "/api/proc/status?project=nope", nil))
	if rec.Code != 404 {
		t.Errorf("status = %d, want 404; body = %q", rec.Code, rec.Body.String())
	}
}

func TestHandleProcFaviconValidation(t *testing.T) {
	withProject(t, "gavel", "flanksource/gavel", "web: echo hi\n")
	s := &Server{}
	tests := []struct {
		name    string
		method  string
		target  string
		wantSts int
	}{
		{"wrong method", "POST", "/api/proc/favicon?project=gavel&port=3000", 405},
		{"missing project", "GET", "/api/proc/favicon?port=3000", 400},
		{"missing port", "GET", "/api/proc/favicon?project=gavel", 400},
		{"invalid port", "GET", "/api/proc/favicon?project=gavel&port=nope", 400},
		{"unknown project", "GET", "/api/proc/favicon?project=nope&port=3000", 404},
		{"undiscovered port", "GET", "/api/proc/favicon?project=gavel&port=3000", 404},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			s.handleProcFavicon(rec, httptest.NewRequest(tc.method, tc.target, nil))
			if rec.Code != tc.wantSts {
				t.Errorf("status = %d, want %d; body = %q", rec.Code, tc.wantSts, rec.Body.String())
			}
		})
	}
}

func TestHandleProcFaviconAllowedDiscoveredPort(t *testing.T) {
	dir := withProject(t, "gavel", "flanksource/gavel", "web: echo hi\n")

	icon := []byte{0x00, 0x00, 0x01, 0x00}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte(`<!doctype html><html><head></head><body>ok</body></html>`))
		case "/favicon.ico":
			w.Header().Set("Content-Type", "image/x-icon")
			_, _ = w.Write(icon)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	_, portStr, err := net.SplitHostPort(strings.TrimPrefix(srv.URL, "http://"))
	if err != nil {
		t.Fatalf("split test server port: %v", err)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		t.Fatalf("parse test server port: %v", err)
	}
	stateDir, err := procfile.StateDir(dir)
	if err != nil {
		t.Fatalf("state dir: %v", err)
	}
	if err := procfile.WriteState(stateDir, procfile.State{
		Processes: []procfile.ProcState{{
			Name:    "web",
			Command: "echo hi",
			Status:  procfile.StatusRunning,
			LogFile: procfile.LogPath(stateDir, "web"),
			Ports:   []int{port},
		}},
	}); err != nil {
		t.Fatalf("write state: %v", err)
	}

	rec := httptest.NewRecorder()
	(&Server{}).handleProcFavicon(rec, httptest.NewRequest("GET", "/api/proc/favicon?project=gavel&port="+portStr, nil))
	if rec.Code != 200 {
		t.Fatalf("status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}
	if got := rec.Header().Get("Content-Type"); !strings.Contains(got, "image/x-icon") {
		t.Errorf("content-type = %q, want image/x-icon", got)
	}
	if !bytes.Equal(rec.Body.Bytes(), icon) {
		t.Errorf("body = %v, want icon bytes %v", rec.Body.Bytes(), icon)
	}
}

func TestHandleProcControlValidation(t *testing.T) {
	withProject(t, "gavel", "flanksource/gavel", "web: echo hi\n")
	s := &Server{}
	tests := []struct {
		name    string
		method  string
		path    string
		body    string
		wantSts int
	}{
		{"wrong method", "GET", "/api/proc/start", `{}`, 405},
		{"malformed json", "POST", "/api/proc/start", `{not json`, 400},
		{"unknown project", "POST", "/api/proc/start", `{"project":"nope"}`, 404},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			s.handleProcControl(rec, httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body)))
			if rec.Code != tc.wantSts {
				t.Errorf("status = %d, want %d; body = %q", rec.Code, tc.wantSts, rec.Body.String())
			}
		})
	}
}

func TestGitChangeCount(t *testing.T) {
	// A directory that is not a git work tree is reported as an error, so the
	// caller can omit the field rather than claiming zero changes.
	if _, err := gitChangeCount(t.TempDir()); err == nil {
		t.Error("gitChangeCount on a non-git dir = nil error, want error")
	}

	dir := t.TempDir()
	if out, err := runGit(dir, "init"); err != nil {
		t.Fatalf("git init: %v\n%s", err, out)
	}

	if n, err := gitChangeCount(dir); err != nil || n != 0 {
		t.Fatalf("gitChangeCount(clean repo) = (%d, %v), want (0, nil)", n, err)
	}

	for _, name := range []string{"a.txt", "b.txt"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(name), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	if n, err := gitChangeCount(dir); err != nil || n != 2 {
		t.Fatalf("gitChangeCount(2 untracked) = (%d, %v), want (2, nil)", n, err)
	}
}

func runGit(dir string, args ...string) ([]byte, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	return cmd.CombinedOutput()
}

func TestHandleProcLogsUnknownProject(t *testing.T) {
	withProject(t, "gavel", "flanksource/gavel", "web: echo hi\n")

	rec := httptest.NewRecorder()
	(&Server{}).handleProcLogs(rec, httptest.NewRequest("GET", "/api/proc/logs?project=nope", nil))
	if rec.Code != 404 {
		t.Errorf("status = %d, want 404; body = %q", rec.Code, rec.Body.String())
	}
}
