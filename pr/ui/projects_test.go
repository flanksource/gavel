package ui

import (
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
