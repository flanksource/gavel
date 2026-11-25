package projects

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/flanksource/gavel/internal/safefs"
	"github.com/flanksource/gavel/tenant"
)

// File is the file-only catalog: a JSON list at Path. A project's ID is
// FileID(name), so it matches the id it gets when the file is imported into
// the projects table.
type File struct {
	Path string
}

// NewFile is the catalog stored in the projects.json at path.
func NewFile(path string) *File { return &File{Path: path} }

// fileRecord is one project as projects.json stores it: ids and inferred repos
// are derived on read, never written.
type fileRecord struct {
	Name  string   `json:"name"`
	Kind  string   `json:"kind,omitempty"`
	Dir   string   `json:"dir"`
	Repos []string `json:"repos"`
}

// List reads the catalog, sorted by name. A missing file is the normal "no
// projects configured yet" state; read and parse failures name the file.
func (f *File) List(context.Context) ([]Project, error) {
	records, err := f.records()
	if err != nil {
		return nil, err
	}
	ps := make([]Project, len(records))
	for i, r := range records {
		ps[i] = Project{ID: FileID(r.Name), TenantID: tenant.LocalOrgID.String(), Name: r.Name, Kind: r.Kind, Dir: r.Dir, Repos: r.Repos}
	}
	return prepared(ps, f.Path)
}

func (f *File) records() ([]fileRecord, error) {
	data, err := os.ReadFile(f.Path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read gavel projects file %s: %w", f.Path, err)
	}
	var records []fileRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return nil, fmt.Errorf("parse gavel projects file %s: %w", f.Path, err)
	}
	return records, nil
}

func (f *File) Get(ctx context.Context, name string) (Project, error) {
	ps, err := f.List(ctx)
	if err != nil {
		return Project{}, err
	}
	if p, ok := ByName(ps, name); ok {
		return p, nil
	}
	return Project{}, fmt.Errorf("%w: %q was not found in %s", ErrNotFound, name, f.Path)
}

func (f *File) ForDir(ctx context.Context, dir string) (Project, error) {
	ps, err := f.List(ctx)
	if err != nil {
		return Project{}, err
	}
	return mostSpecific(ps, dir, f.Path)
}

// Create persists a new project, failing if required fields are missing
// (ErrInvalid) or the name is already taken (ErrExists).
func (f *File) Create(ctx context.Context, p Project) (Project, error) {
	p = p.withKind()
	if err := p.validate(); err != nil {
		return Project{}, err
	}
	ps, err := f.List(ctx)
	if err != nil {
		return Project{}, err
	}
	if _, ok := ByName(ps, p.Name); ok {
		return Project{}, fmt.Errorf("%w: %q in %s", ErrExists, p.Name, f.Path)
	}
	if err := f.save(append(ps, p)); err != nil {
		return Project{}, err
	}
	return f.Get(ctx, p.Name)
}

func (f *File) EnsureWeb(ctx context.Context, name string) (Project, error) {
	return ensureWeb(ctx, f, name)
}

func (f *File) Update(ctx context.Context, name string, p Project) (Project, error) {
	ps, err := f.List(ctx)
	if err != nil {
		return Project{}, err
	}
	i := slices.IndexFunc(ps, func(existing Project) bool { return existing.Name == name })
	if i < 0 {
		return Project{}, fmt.Errorf("%w: %q was not found in %s", ErrNotFound, name, f.Path)
	}
	if ps[i], err = validUpdate(ps[i], p); err != nil {
		return Project{}, err
	}
	if err := f.save(ps); err != nil {
		return Project{}, err
	}
	return f.Get(ctx, name)
}

func (f *File) Delete(ctx context.Context, name string) error {
	ps, err := f.List(ctx)
	if err != nil {
		return err
	}
	i := slices.IndexFunc(ps, func(p Project) bool { return p.Name == name })
	if i < 0 {
		return fmt.Errorf("%w: %q was not found in %s", ErrNotFound, name, f.Path)
	}
	return f.save(slices.Delete(ps, i, i+1))
}

// save writes the catalog sorted by name, dropping inferred repos.
func (f *File) save(ps []Project) error {
	if err := os.MkdirAll(filepath.Dir(f.Path), 0o755); err != nil {
		return fmt.Errorf("create gavel projects directory for %s: %w", f.Path, err)
	}
	sorted := sortProjects(slices.Clone(ps))
	records := make([]fileRecord, len(sorted))
	for i, p := range sorted {
		records[i] = fileRecord{Name: p.Name, Kind: p.withKind().Kind, Dir: p.Dir, Repos: p.registeredRepos()}
	}
	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal gavel projects file %s: %w", f.Path, err)
	}
	if err := f.write(data); err != nil {
		return fmt.Errorf("write gavel projects file %s: %w", f.Path, err)
	}
	return nil
}

func (f *File) write(data []byte) (err error) {
	dir := filepath.Dir(f.Path)
	file, err := os.CreateTemp(dir, ".projects-*.json")
	if err != nil {
		return err
	}
	path := file.Name()
	defer func() {
		if file != nil {
			err = errors.Join(err, file.Close())
		}
		err = errors.Join(err, safefs.SafeRemove(path, safefs.RemoveContext{AllowedRoot: dir}))
	}()
	if err := file.Chmod(0o644); err != nil {
		return err
	}
	if _, err := file.Write(data); err != nil {
		return err
	}
	if err := file.Sync(); err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		file = nil
		return err
	}
	file = nil
	return os.Rename(path, f.Path)
}
