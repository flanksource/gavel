// Package projects is the project catalog: the GitHub repos each local
// workspace directory backs. It is file-only (~/.config/gavel/projects.json)
// when gavel runs without PostgreSQL and the projects table otherwise; Open
// picks the mode once and callers only see a Service.
package projects

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/todos/query"
	todoruntime "github.com/flanksource/gavel/todos/runtime"
	"github.com/google/uuid"
)

// Catalog errors. Callers map these to a transport status (HTTP code / CLI
// exit) with errors.Is, so the create/update/delete semantics live in one place
// for every surface.
var (
	ErrNotFound = errors.New("unknown project")
	ErrExists   = errors.New("project already exists")
	ErrInvalid  = errors.New("invalid project")
)

// Service is the project catalog. Project names are the public key (comment
// scopes are gavel://project/<name>); ID is stable across the file and DB modes.
type Service interface {
	List(ctx context.Context) ([]Project, error)
	Get(ctx context.Context, name string) (Project, error)
	// ForDir is the most specific project whose directory contains dir.
	ForDir(ctx context.Context, dir string) (Project, error)
	Create(ctx context.Context, p Project) (Project, error)
	// EnsureWeb is the web project named name, created when absent.
	EnsureWeb(ctx context.Context, name string) (Project, error)
	// Update replaces the named project; the name is authoritative, so the body
	// cannot rename it.
	Update(ctx context.Context, name string, p Project) (Project, error)
	Delete(ctx context.Context, name string) error
}

// Project associates one or more GitHub repos with a local workspace directory
// on disk. The directory is where Gavel discovers a Procfile so the PR UI can
// supervise the project's processes alongside its pull requests.
type Project struct {
	ID       uuid.UUID `json:"id"`
	TenantID string    `json:"tenantId,omitempty"`
	Name     string    `json:"name"`
	// Kind is KindDir (a local checkout) or KindWeb (no directory, no repos).
	Kind  string   `json:"kind"`
	Dir   string   `json:"dir"`
	Repos []string `json:"repos"`
	// ReposInferred marks Repos as derived from the directory's GitHub remotes
	// because none were registered. Inferred repos are never written back, so
	// they keep following the checkout's remotes.
	ReposInferred bool `json:"reposInferred,omitempty"`
}

// idNamespace seeds the file-mode project id, so a project keeps one id from
// projects.json into the projects table it is imported into.
// Project kinds. A stored empty kind reads as KindDir.
const (
	KindDir = "dir"
	KindWeb = "web"
)

var idNamespace = uuid.MustParse("6f1d2c1e-4b8a-5d3e-9c2f-7a0b1e3d5c47")

// FileID is the stable id of a project registered in projects.json.
func FileID(name string) uuid.UUID { return uuid.NewSHA1(idNamespace, []byte(name)) }

// ResolvedDir expands a leading "~" in Dir to the user's home directory so the
// stored path can be portable across machines. Other paths are returned as-is.
func (p Project) ResolvedDir() string {
	dir := strings.TrimSpace(p.Dir)
	if dir == "~" || strings.HasPrefix(dir, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, strings.TrimPrefix(dir[1:], "/"))
		}
	}
	return dir
}

// WorkspaceOptions projects the catalog record into the native TODO runtime
// without giving the runtime a second configuration source.
func (p Project) WorkspaceOptions() todoruntime.WorkspaceOptions {
	return todoruntime.WorkspaceOptions{
		Name:         p.Name,
		RootPath:     p.ResolvedDir(),
		Repositories: append([]string(nil), p.Repos...),
	}
}

// TodoWorkspace projects the catalog record into the directory a TODO list
// reads, named the way the projects list displays it.
func (p Project) TodoWorkspace() query.Workspace {
	return query.Workspace{Name: p.Name, Dir: p.ResolvedDir()}
}

// TodoWorkspaces projects a whole catalog, in catalog order.
func TodoWorkspaces(ps []Project) []query.Workspace {
	workspaces := make([]query.Workspace, len(ps))
	for i, p := range ps {
		workspaces[i] = p.TodoWorkspace()
	}
	return workspaces
}

// ForRepo returns the first project whose Repos list contains repo.
func ForRepo(ps []Project, repo string) (Project, bool) {
	for _, p := range ps {
		if slices.Contains(p.Repos, repo) {
			return p, true
		}
	}
	return Project{}, false
}

// ByName returns the project with the given name.
func ByName(ps []Project, name string) (Project, bool) {
	i := slices.IndexFunc(ps, func(p Project) bool { return p.Name == name })
	if i < 0 {
		return Project{}, false
	}
	return ps[i], true
}

// CanonicalPath is path made absolute with its symlinks resolved, so a checkout
// reached through a link and one registered at its real location compare equal.
// A path that does not exist (a moved or deleted project) is only made absolute.
func CanonicalPath(path string) (string, error) {
	absolute, err := filepath.Abs(strings.TrimSpace(path))
	if err != nil {
		return "", err
	}
	resolved, err := filepath.EvalSymlinks(absolute)
	if errors.Is(err, os.ErrNotExist) {
		return absolute, nil
	}
	return resolved, err
}

// withRepos returns the project with a non-nil Repos list. Go marshals a nil
// slice as JSON null, and every consumer of the catalog treats repos as a list,
// so a project registered without repos stays an empty list on the wire.
func (p Project) withRepos() Project {
	if p.Repos == nil {
		p.Repos = []string{}
	}
	return p
}

// withInferredRepos fills an unregistered repo list from the project
// directory's GitHub remotes. A directory that does not exist or is not a git
// checkout has no repos to infer.
func (p Project) withInferredRepos(source string) (Project, error) {
	if p.Kind == KindWeb || len(p.Repos) > 0 {
		return p, nil
	}
	dir := p.ResolvedDir()
	if _, err := os.Stat(dir); errors.Is(err, os.ErrNotExist) {
		return p, nil
	}
	repos, err := github.ReposFromRemotes(dir)
	if err != nil {
		return p, fmt.Errorf("infer repos of project %q from %s: %w", p.Name, source, err)
	}
	p.Repos, p.ReposInferred = repos, len(repos) > 0
	return p, nil
}

// registeredRepos is the repo list a project is stored with: inferred repos are
// never written back.
func (p Project) registeredRepos() []string {
	if p.ReposInferred || p.Repos == nil {
		return []string{}
	}
	return p.Repos
}

// prepared validates a project and applies the stored defaults before it is
// read back: the non-nil repo list and its inferred repos.
func prepared(ps []Project, source string) ([]Project, error) {
	for i := range ps {
		var err error
		ps[i] = ps[i].withKind()
		if ps[i], err = ps[i].withInferredRepos(source); err != nil {
			return nil, err
		}
		ps[i] = ps[i].withRepos()
	}
	return sortProjects(ps), nil
}

// sortProjects orders the catalog case-insensitively by name, so every surface
// that reads the store lists projects the same way. Names that differ only by
// case fall back to a byte comparison to keep the order deterministic.
func sortProjects(ps []Project) []Project {
	slices.SortStableFunc(ps, func(a, b Project) int {
		if c := strings.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name)); c != 0 {
			return c
		}
		return strings.Compare(a.Name, b.Name)
	})
	return ps
}

// mostSpecific resolves dir to the most-specific project root in ps,
// comparing both through their symlinks. source names the catalog in errors.
func mostSpecific(ps []Project, dir, source string) (Project, error) {
	absolute, err := CanonicalPath(dir)
	if err != nil {
		return Project{}, fmt.Errorf("resolve project directory %q using %s: %w", dir, source, err)
	}
	var match Project
	matchLength := -1
	for _, project := range ps {
		if project.Kind != KindDir {
			continue
		}
		root, err := CanonicalPath(project.ResolvedDir())
		if err != nil {
			return Project{}, fmt.Errorf("resolve project %q directory %q from %s: %w", project.Name, project.Dir, source, err)
		}
		relative, err := filepath.Rel(root, absolute)
		if err != nil {
			return Project{}, fmt.Errorf("compare project %q directory %q using %s: %w", project.Name, root, source, err)
		}
		if relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			continue
		}
		if len(root) > matchLength {
			match, matchLength = project, len(root)
		}
	}
	if matchLength < 0 {
		return Project{}, fmt.Errorf("%w: project for directory %q was not found in %s", ErrNotFound, absolute, source)
	}
	return match, nil
}

// validUpdate applies Update's rules to p against the stored existing project:
// neither the name nor the kind can change, the result must validate, and unchanged inferred repos
// stay inferred instead of being registered (a form or CLI update echoes the
// repos it was shown).
func validUpdate(existing, p Project) (Project, error) {
	p.ID, p.TenantID, p.Name = existing.ID, existing.TenantID, existing.Name
	p.ReposInferred = existing.ReposInferred && slices.Equal(p.Repos, existing.Repos)
	if p.Kind == "" {
		p.Kind = existing.Kind
	}
	if p.Kind != existing.Kind {
		return Project{}, fmt.Errorf("%w: project %q is a %s project and cannot become %s", ErrInvalid, existing.Name, existing.Kind, p.Kind)
	}
	return p, p.validate()
}

// withKind defaults an empty kind to KindDir.
func (p Project) withKind() Project {
	if p.Kind == "" {
		p.Kind = KindDir
	}
	return p
}

// validate checks a project is storable: named, a dir project with a dir, a
// web project with neither dir nor repos.
func (p Project) validate() error {
	switch {
	case strings.TrimSpace(p.Name) == "":
		return fmt.Errorf("%w: a name is required", ErrInvalid)
	case p.Kind == KindDir && strings.TrimSpace(p.Dir) == "":
		return fmt.Errorf("%w: dir project %q needs a dir", ErrInvalid, p.Name)
	case p.Kind == KindWeb && (p.Dir != "" || len(p.registeredRepos()) > 0):
		return fmt.Errorf("%w: web project %q has no dir or repos", ErrInvalid, p.Name)
	case p.Kind != KindDir && p.Kind != KindWeb:
		return fmt.Errorf("%w: project %q has unknown kind %q", ErrInvalid, p.Name, p.Kind)
	}
	return nil
}

// ensureWeb is EnsureWeb over any catalog: the existing web project, or a new
// one. A dir project of that name is a conflict, not a web project.
func ensureWeb(ctx context.Context, catalog Service, name string) (Project, error) {
	existing, err := catalog.Get(ctx, name)
	if errors.Is(err, ErrNotFound) {
		existing, err = catalog.Create(ctx, Project{Name: name, Kind: KindWeb})
		if errors.Is(err, ErrExists) {
			existing, err = catalog.Get(ctx, name)
		}
	}
	if err != nil {
		return Project{}, err
	}
	if existing.Kind != KindWeb {
		return Project{}, fmt.Errorf("%w: %q is a %s project, not a web project", ErrExists, name, existing.Kind)
	}
	return existing, nil
}
