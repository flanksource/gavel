package ui

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestProjectsRoundTrip(t *testing.T) {
	orig := projectsPath
	projectsPath = filepath.Join(t.TempDir(), "projects.json")
	defer func() { projectsPath = orig }()

	// A missing file is the empty state, not an error.
	if got := LoadProjects(); got != nil {
		t.Errorf("LoadProjects() on missing file = %v, want nil", got)
	}

	want := []Project{
		{Name: "gavel", Dir: "~/go/src/gavel", Repos: []string{"flanksource/gavel"}},
		{Name: "infra", Dir: "/srv/infra", Repos: []string{"acme/infra", "acme/charts"}},
	}
	SaveProjects(want)

	got := LoadProjects()
	if len(got) != len(want) {
		t.Fatalf("LoadProjects() returned %d projects, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i].Name != want[i].Name || got[i].Dir != want[i].Dir || len(got[i].Repos) != len(want[i].Repos) {
			t.Errorf("project[%d] = %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestResolvedDirExpandsHome(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	tests := []struct {
		dir  string
		want string
	}{
		{"~/go/src/gavel", filepath.Join(home, "go/src/gavel")},
		{"~", home},
		{"/abs/path", "/abs/path"},
		{"relative/path", "relative/path"},
	}
	for _, tc := range tests {
		if got := (Project{Dir: tc.dir}).ResolvedDir(); got != tc.want {
			t.Errorf("ResolvedDir(%q) = %q, want %q", tc.dir, got, tc.want)
		}
	}
}

func TestProjectForRepo(t *testing.T) {
	ps := []Project{
		{Name: "gavel", Repos: []string{"flanksource/gavel"}},
		{Name: "infra", Repos: []string{"acme/infra", "acme/charts"}},
	}

	if p, ok := ProjectForRepo(ps, "acme/charts"); !ok || p.Name != "infra" {
		t.Errorf("ProjectForRepo(acme/charts) = %+v, %v; want infra project", p, ok)
	}
	if _, ok := ProjectForRepo(ps, "unknown/repo"); ok {
		t.Error("ProjectForRepo(unknown/repo) returned a match, want miss")
	}
}

func TestUpsertProject(t *testing.T) {
	ps := []Project{{Name: "gavel", Dir: "/old", Repos: []string{"flanksource/gavel"}}}

	ps = upsertProject(ps, Project{Name: "gavel", Dir: "/new", Repos: []string{"flanksource/gavel"}})
	if len(ps) != 1 || ps[0].Dir != "/new" {
		t.Errorf("upsert of existing name should replace in place, got %+v", ps)
	}

	ps = upsertProject(ps, Project{Name: "infra", Dir: "/srv", Repos: []string{"acme/infra"}})
	if len(ps) != 2 || ps[1].Name != "infra" {
		t.Errorf("upsert of new name should append, got %+v", ps)
	}
}

// TestProjectServiceCRUD exercises the shared create/update/delete/get service
// that both the HTTP entity and the `gavel projects` CLI build on, including the
// sentinel errors callers map to a status code.
func TestProjectServiceCRUD(t *testing.T) {
	orig := projectsPath
	projectsPath = filepath.Join(t.TempDir(), "projects.json")
	defer func() { projectsPath = orig }()

	if err := CreateProject(Project{Name: "alpha", Dir: "/srv/alpha"}); err != nil {
		t.Fatalf("CreateProject = %v, want nil", err)
	}
	if err := CreateProject(Project{Name: "alpha", Dir: "/other"}); !errors.Is(err, ErrProjectExists) {
		t.Errorf("duplicate CreateProject err = %v, want ErrProjectExists", err)
	}
	if err := CreateProject(Project{Name: "beta"}); !errors.Is(err, ErrProjectInvalid) {
		t.Errorf("CreateProject without dir err = %v, want ErrProjectInvalid", err)
	}

	// Update keeps the name from the id even if the body carries a different one.
	if err := UpdateProject("alpha", Project{Name: "renamed", Dir: "/srv/new"}); err != nil {
		t.Fatalf("UpdateProject = %v, want nil", err)
	}
	got, ok := GetProject("alpha")
	if !ok || got.Dir != "/srv/new" || got.Name != "alpha" {
		t.Errorf("after update GetProject = %+v, %v; want alpha dir=/srv/new", got, ok)
	}
	if err := UpdateProject("ghost", Project{Dir: "/x"}); !errors.Is(err, ErrProjectNotFound) {
		t.Errorf("UpdateProject missing err = %v, want ErrProjectNotFound", err)
	}
	if err := UpdateProject("alpha", Project{Dir: ""}); !errors.Is(err, ErrProjectInvalid) {
		t.Errorf("UpdateProject empty dir err = %v, want ErrProjectInvalid", err)
	}

	if err := DeleteProject("alpha"); err != nil {
		t.Fatalf("DeleteProject = %v, want nil", err)
	}
	if _, ok := GetProject("alpha"); ok {
		t.Error("GetProject after delete returned a project, want miss")
	}
	if err := DeleteProject("alpha"); !errors.Is(err, ErrProjectNotFound) {
		t.Errorf("DeleteProject missing err = %v, want ErrProjectNotFound", err)
	}
}

func TestResolveTodoBackend(t *testing.T) {
	withTodos := t.TempDir()
	if err := os.Mkdir(filepath.Join(withTodos, ".todos"), 0o755); err != nil {
		t.Fatalf("mkdir .todos: %v", err)
	}
	noTodos := t.TempDir()

	tests := []struct {
		dir        string
		configured string
		wantName   string
		wantAuto   bool
	}{
		{noTodos, "grite", "grite", false},
		{noTodos, "todos", "todos", false},
		{withTodos, "", "todos", true},   // auto-detected from .todos dir
		{noTodos, "", "grite", true},     // auto, no .todos → grite
		{noTodos, "auto", "grite", true}, // explicit "auto" behaves like empty
	}
	for _, tc := range tests {
		name, auto := resolveTodoBackend(tc.dir, tc.configured)
		if name != tc.wantName || auto != tc.wantAuto {
			t.Errorf("resolveTodoBackend(%q, %q) = (%q, %v), want (%q, %v)",
				tc.dir, tc.configured, name, auto, tc.wantName, tc.wantAuto)
		}
	}

	if got := TodoBackendLabel(withTodos, ""); got != "todos (auto)" {
		t.Errorf("TodoBackendLabel(auto) = %q, want %q", got, "todos (auto)")
	}
	if got := TodoBackendLabel(noTodos, "grite"); got != "grite" {
		t.Errorf("TodoBackendLabel(explicit) = %q, want %q", got, "grite")
	}
}

// Guard against accidental writes to the real config during tests.
func TestProjectsPathUnderConfigDir(t *testing.T) {
	if filepath.Base(projectsPath) != "projects.json" {
		t.Errorf("projectsPath basename = %q, want projects.json", filepath.Base(projectsPath))
	}
	if dir := filepath.Base(filepath.Dir(projectsPath)); dir != "gavel" {
		t.Errorf("projectsPath parent dir = %q, want gavel", dir)
	}
	_ = os.Getenv("HOME")
}
