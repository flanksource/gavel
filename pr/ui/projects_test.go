package ui

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func projectNames(ps []Project) []string {
	names := make([]string, 0, len(ps))
	for _, p := range ps {
		names = append(names, p.Name)
	}
	return names
}

// The catalog is sorted case-insensitively at the store so every surface (the
// dashboard, `gavel projects`, the new-todo form) lists projects the same way.
// Names that differ only by case fall back to a byte comparison so the order
// is deterministic.
func TestProjectsStoreSortsCaseInsensitively(t *testing.T) {
	orig := projectsPath
	projectsPath = filepath.Join(t.TempDir(), "projects.json")
	defer func() { projectsPath = orig }()

	unsorted := []string{"infra", "gavel", "alpha", "Gavel", "Beta"}
	want := []string{"alpha", "Beta", "Gavel", "gavel", "infra"}

	// A catalog written by an older gavel is unsorted on disk.
	raw := `[` +
		`{"name":"infra","dir":"/i"},{"name":"gavel","dir":"/g"},{"name":"alpha","dir":"/a"},` +
		`{"name":"Gavel","dir":"/G"},{"name":"Beta","dir":"/b"}]`
	if err := os.WriteFile(projectsPath, []byte(raw), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := LoadProjects()
	if err != nil {
		t.Fatalf("LoadProjects() = %v", err)
	}
	if !slices.Equal(projectNames(got), want) {
		t.Fatalf("LoadProjects() order = %v, want %v", projectNames(got), want)
	}

	projects := make([]Project, 0, len(unsorted))
	for _, name := range unsorted {
		projects = append(projects, Project{Name: name, Dir: "/" + name})
	}
	if err := SaveProjects(projects); err != nil {
		t.Fatalf("SaveProjects() = %v", err)
	}
	data, err := os.ReadFile(projectsPath)
	if err != nil {
		t.Fatal(err)
	}
	var onDisk []Project
	if err := json.Unmarshal(data, &onDisk); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(projectNames(onDisk), want) {
		t.Fatalf("projects.json order = %v, want %v", projectNames(onDisk), want)
	}
}

func TestProjectsRoundTrip(t *testing.T) {
	orig := projectsPath
	projectsPath = filepath.Join(t.TempDir(), "projects.json")
	defer func() { projectsPath = orig }()

	// A missing file is the empty state, not an error.
	got, err := LoadProjects()
	if err != nil {
		t.Fatalf("LoadProjects() on missing file = %v", err)
	}
	if got != nil {
		t.Errorf("LoadProjects() on missing file = %v, want nil", got)
	}

	want := []Project{
		{Name: "gavel", Dir: "~/go/src/gavel", Repos: []string{"flanksource/gavel"}},
		{Name: "infra", Dir: "/srv/infra", Repos: []string{"acme/infra", "acme/charts"}},
	}
	if err := SaveProjects(want); err != nil {
		t.Fatalf("SaveProjects() = %v", err)
	}

	got, err = LoadProjects()
	if err != nil {
		t.Fatalf("LoadProjects() = %v", err)
	}
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
	got, err := GetProject("alpha")
	if err != nil || got.Dir != "/srv/new" || got.Name != "alpha" {
		t.Errorf("after update GetProject = %+v, %v; want alpha dir=/srv/new", got, err)
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
	if _, err := GetProject("alpha"); !errors.Is(err, ErrProjectNotFound) {
		t.Errorf("GetProject after delete err = %v, want ErrProjectNotFound", err)
	}
	if err := DeleteProject("alpha"); !errors.Is(err, ErrProjectNotFound) {
		t.Errorf("DeleteProject missing err = %v, want ErrProjectNotFound", err)
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
