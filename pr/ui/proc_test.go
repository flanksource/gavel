package ui

import (
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
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

func TestHandleProjectsUpsert(t *testing.T) {
	withProject(t, "gavel", "flanksource/gavel", "")

	body := `{"name":"infra","dir":"/srv/infra","repos":["acme/infra"]}`
	rec := httptest.NewRecorder()
	(&Server{}).handleProjects(rec, httptest.NewRequest("POST", "/api/projects", strings.NewReader(body)))
	if rec.Code != 200 {
		t.Fatalf("POST status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}
	if _, ok := projectByName(LoadProjects(), "infra"); !ok {
		t.Error("POST /api/projects did not persist the new project")
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

func TestHandleProcLogsUnknownProject(t *testing.T) {
	withProject(t, "gavel", "flanksource/gavel", "web: echo hi\n")

	rec := httptest.NewRecorder()
	(&Server{}).handleProcLogs(rec, httptest.NewRequest("GET", "/api/proc/logs?project=nope", nil))
	if rec.Code != 404 {
		t.Errorf("status = %d, want 404; body = %q", rec.Code, rec.Body.String())
	}
}
