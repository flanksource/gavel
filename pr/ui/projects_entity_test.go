package ui

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func stubProjectTodoCounts(t *testing.T) {
	t.Helper()
	original := projectTodoCounts
	projectTodoCounts = func(context.Context, Project) (todoCounts, error) {
		return todoCounts{}, nil
	}
	t.Cleanup(func() { projectTodoCounts = original })
}

// projectByNameReq builds a request for the per-entity handler with the {name}
// path value populated the way the ServeMux pattern would at runtime.
func projectByNameReq(method, name, body string) *http.Request {
	var r *http.Request
	if body == "" {
		r = httptest.NewRequest(method, "/api/projects/"+name, nil)
	} else {
		r = httptest.NewRequest(method, "/api/projects/"+name, strings.NewReader(body))
	}
	r.SetPathValue("name", name)
	return r
}

func TestHandleProjectsIgnoresLegacyTodoProvider(t *testing.T) {
	originalPath := projectsPath
	projectsPath = filepath.Join(t.TempDir(), "projects.json")
	t.Cleanup(func() { projectsPath = originalPath })

	legacy := []byte(`[
  {
    "name": "clicky",
    "dir": "/srv/clicky",
    "repos": ["flanksource/clicky"],
    "todoProvider": "grite"
  }
]`)
	if err := os.WriteFile(projectsPath, legacy, 0o600); err != nil {
		t.Fatalf("write legacy projects config: %v", err)
	}

	originalCounts := projectTodoCounts
	var countedDir string
	projectTodoCounts = func(_ context.Context, project Project) (todoCounts, error) {
		countedDir = project.ResolvedDir()
		return todoCounts{Total: 14, Open: 12}, nil
	}
	t.Cleanup(func() { projectTodoCounts = originalCounts })

	rec := httptest.NewRecorder()
	(&Server{}).handleProjects(rec, httptest.NewRequest(http.MethodGet, "/api/projects", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}
	if countedDir != "/srv/clicky" {
		t.Fatalf("TODO counts dir = %q, want /srv/clicky", countedDir)
	}

	var got []projectInfo
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal projects response: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("projects response length = %d, want 1", len(got))
	}
	if got[0].Name != "clicky" || got[0].TodoBackend != "db" || got[0].TodoCounts.Total != 14 || got[0].TodoCounts.Open != 12 {
		t.Fatalf("project response = %+v, want Clicky on db with 14 total and 12 open TODOs", got[0])
	}
	if body := rec.Body.String(); strings.Contains(body, "grite") || strings.Contains(body, "todoProvider") {
		t.Fatalf("projects response leaked retired provider configuration: %s", body)
	}
}

func TestHandleProjectGet(t *testing.T) {
	stubProjectTodoCounts(t)
	withProject(t, "gavel", "flanksource/gavel", "web: echo hi\n")

	rec := httptest.NewRecorder()
	(&Server{}).handleProjectByName(rec, projectByNameReq("GET", "gavel", ""))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}
	var got projectInfo
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Name != "gavel" || !got.HasProcfile {
		t.Errorf("GET /api/projects/gavel = %+v, want gavel with hasProcfile=true", got)
	}

	rec = httptest.NewRecorder()
	(&Server{}).handleProjectByName(rec, projectByNameReq("GET", "nope", ""))
	if rec.Code != http.StatusNotFound {
		t.Errorf("GET unknown status = %d, want 404", rec.Code)
	}
}

func TestHandleProjectUpdate(t *testing.T) {
	withProject(t, "gavel", "flanksource/gavel", "")

	rec := httptest.NewRecorder()
	(&Server{}).handleProjectByName(rec, projectByNameReq("PUT", "gavel", `{"dir":"/new","repos":["flanksource/gavel"]}`))
	if rec.Code != http.StatusOK {
		t.Fatalf("PUT status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}
	projects, err := LoadProjects()
	if err != nil {
		t.Fatalf("LoadProjects = %v", err)
	}
	p, ok := projectByName(projects, "gavel")
	if !ok || p.Dir != "/new" {
		t.Errorf("after PUT, project = %+v, want dir=/new", p)
	}

	rec = httptest.NewRecorder()
	(&Server{}).handleProjectByName(rec, projectByNameReq("PUT", "nope", `{"dir":"/x"}`))
	if rec.Code != http.StatusNotFound {
		t.Errorf("PUT unknown status = %d, want 404", rec.Code)
	}

	rec = httptest.NewRecorder()
	(&Server{}).handleProjectByName(rec, projectByNameReq("PUT", "gavel", `{"dir":""}`))
	if rec.Code != http.StatusBadRequest {
		t.Errorf("PUT empty dir status = %d, want 400", rec.Code)
	}
}

func TestHandleProjectDelete(t *testing.T) {
	withProject(t, "gavel", "flanksource/gavel", "")

	rec := httptest.NewRecorder()
	(&Server{}).handleProjectByName(rec, projectByNameReq("DELETE", "gavel", ""))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("DELETE status = %d, want 204; body = %q", rec.Code, rec.Body.String())
	}
	projects, err := LoadProjects()
	if err != nil {
		t.Fatalf("LoadProjects = %v", err)
	}
	if _, ok := projectByName(projects, "gavel"); ok {
		t.Error("DELETE did not remove the project")
	}

	rec = httptest.NewRecorder()
	(&Server{}).handleProjectByName(rec, projectByNameReq("DELETE", "gavel", ""))
	if rec.Code != http.StatusNotFound {
		t.Errorf("DELETE missing status = %d, want 404", rec.Code)
	}
}

func TestHandleProjectsClickyTable(t *testing.T) {
	stubProjectTodoCounts(t)
	withProject(t, "gavel", "flanksource/gavel", "")

	req := httptest.NewRequest("GET", "/api/projects", nil)
	req.Header.Set("Accept", "application/json+clicky")
	rec := httptest.NewRecorder()
	(&Server{}).handleProjects(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}

	var doc struct {
		Version int `json:"version"`
		Node    struct {
			Kind    string `json:"kind"`
			Columns []struct {
				Name string `json:"name"`
			} `json:"columns"`
			Rows []struct {
				Cells map[string]json.RawMessage `json:"cells"`
			} `json:"rows"`
		} `json:"node"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &doc); err != nil {
		t.Fatalf("unmarshal clicky doc: %v\nbody=%s", err, rec.Body.String())
	}
	if doc.Version != 1 || doc.Node.Kind != "table" {
		t.Fatalf("clicky doc = version %d kind %q, want version 1 kind table", doc.Version, doc.Node.Kind)
	}
	cols := make(map[string]bool)
	for _, c := range doc.Node.Columns {
		cols[c.Name] = true
	}
	for _, want := range []string{"name", "dir", "provider"} {
		if !cols[want] {
			t.Errorf("clicky table missing column %q; got %v", want, cols)
		}
	}
	if len(doc.Node.Rows) != 1 {
		t.Fatalf("clicky table rows = %d, want 1", len(doc.Node.Rows))
	}
	if _, ok := doc.Node.Rows[0].Cells["name"]; !ok {
		t.Errorf("clicky row missing 'name' cell; got %v", doc.Node.Rows[0].Cells)
	}
}

func TestHandleOpenAPIProjectsSurface(t *testing.T) {
	rec := httptest.NewRecorder()
	(&Server{}).handleOpenAPI(rec, httptest.NewRequest("GET", "/api/openapi.json", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}

	var spec struct {
		XClicky struct {
			Surfaces []struct {
				Key    string `json:"key"`
				Entity string `json:"entity"`
			} `json:"surfaces"`
		} `json:"x-clicky"`
		Paths map[string]map[string]struct {
			OperationID string `json:"operationId"`
			XClicky     struct {
				Surface string `json:"surface"`
				Verb    string `json:"verb"`
				Scope   string `json:"scope"`
				IDParam string `json:"idParam"`
			} `json:"x-clicky"`
		} `json:"paths"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &spec); err != nil {
		t.Fatalf("unmarshal openapi: %v", err)
	}

	if len(spec.XClicky.Surfaces) != 1 || spec.XClicky.Surfaces[0].Key != "projects" || spec.XClicky.Surfaces[0].Entity != "project" {
		t.Fatalf("x-clicky surfaces = %+v, want one projects/project surface", spec.XClicky.Surfaces)
	}

	// Every CRUD verb is present and scoped correctly; entity-scoped ops carry idParam=name.
	wantVerbs := map[string]struct {
		path, scope, idParam string
	}{
		"list":   {"/api/projects", "collection", ""},
		"create": {"/api/projects", "collection", ""},
		"get":    {"/api/projects/{name}", "entity", "name"},
		"update": {"/api/projects/{name}", "entity", "name"},
		"delete": {"/api/projects/{name}", "entity", "name"},
	}
	gotVerbs := map[string]bool{}
	for path, methods := range spec.Paths {
		for _, op := range methods {
			v := op.XClicky.Verb
			if v == "" {
				continue
			}
			gotVerbs[v] = true
			want, ok := wantVerbs[v]
			if !ok {
				t.Errorf("unexpected verb %q", v)
				continue
			}
			if path != want.path || op.XClicky.Scope != want.scope || op.XClicky.IDParam != want.idParam {
				t.Errorf("verb %q at %s scope=%q idParam=%q, want %s scope=%q idParam=%q",
					v, path, op.XClicky.Scope, op.XClicky.IDParam, want.path, want.scope, want.idParam)
			}
		}
	}
	for v := range wantVerbs {
		if !gotVerbs[v] {
			t.Errorf("openapi missing verb %q", v)
		}
	}
}
