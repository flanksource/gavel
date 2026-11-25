package projects

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/flanksource/gavel/tenant"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const dbSource = "the projects table"

// DB is the catalog in the projects table, scoped to the tenant on each call's
// context. The local organization's first use imports the projects.json at
// ImportPath once; the file is never read again.
type DB struct {
	db         *gorm.DB
	importPath string
	imported   sync.Map
}

// NewDB is the catalog in db's projects table, importing importPath once.
func NewDB(db *gorm.DB, importPath string) *DB {
	return &DB{db: db, importPath: importPath}
}

type projectRow struct {
	ID       uuid.UUID
	TenantID string
	Name     string
	Kind     string
	Dir      string
	Repos    string
}

const selectProjects = `SELECT id, tenant_id, name, kind, COALESCE(dir, '') AS dir, repos::text AS repos FROM projects WHERE tenant_id = ?`

func (d *DB) List(ctx context.Context) ([]Project, error) {
	tenantID, err := d.tenant(ctx)
	if err != nil {
		return nil, err
	}
	var rows []projectRow
	if err := d.db.WithContext(ctx).Raw(selectProjects, tenantID).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("list projects of tenant %s: %w", tenantID, err)
	}
	return projectsFromRows(rows)
}

func (d *DB) Get(ctx context.Context, name string) (Project, error) {
	tenantID, err := d.tenant(ctx)
	if err != nil {
		return Project{}, err
	}
	var rows []projectRow
	if err := d.db.WithContext(ctx).Raw(selectProjects+` AND name = ?`, tenantID, name).Scan(&rows).Error; err != nil {
		return Project{}, fmt.Errorf("get project %q of tenant %s: %w", name, tenantID, err)
	}
	if len(rows) == 0 {
		return Project{}, fmt.Errorf("%w: %q was not found in %s", ErrNotFound, name, dbSource)
	}
	ps, err := projectsFromRows(rows)
	if err != nil {
		return Project{}, err
	}
	return ps[0], nil
}

func (d *DB) ForDir(ctx context.Context, dir string) (Project, error) {
	ps, err := d.List(ctx)
	if err != nil {
		return Project{}, err
	}
	return mostSpecific(ps, dir, dbSource)
}

func (d *DB) Create(ctx context.Context, p Project) (Project, error) {
	p = p.withKind()
	if err := p.validate(); err != nil {
		return Project{}, err
	}
	tenantID, err := d.tenant(ctx)
	if err != nil {
		return Project{}, err
	}
	created, err := insertProject(d.db.WithContext(ctx), tenantID, uuid.New(), p)
	if err != nil {
		return Project{}, err
	}
	if !created {
		return Project{}, fmt.Errorf("%w: %q in %s", ErrExists, p.Name, dbSource)
	}
	return d.Get(ctx, p.Name)
}

func (d *DB) EnsureWeb(ctx context.Context, name string) (Project, error) {
	return ensureWeb(ctx, d, name)
}

func (d *DB) Update(ctx context.Context, name string, p Project) (Project, error) {
	existing, err := d.Get(ctx, name)
	if err != nil {
		return Project{}, err
	}
	if p, err = validUpdate(existing, p); err != nil {
		return Project{}, err
	}
	repos, err := json.Marshal(p.registeredRepos())
	if err != nil {
		return Project{}, fmt.Errorf("marshal repos of project %q: %w", name, err)
	}
	if err := d.db.WithContext(ctx).Exec(`UPDATE projects SET dir = NULLIF(?, ''), repos = ?::jsonb, updated_at = now() WHERE id = ?`,
		p.Dir, string(repos), existing.ID).Error; err != nil {
		return Project{}, fmt.Errorf("update project %q: %w", name, err)
	}
	return d.Get(ctx, name)
}

func (d *DB) Delete(ctx context.Context, name string) error {
	tenantID, err := d.tenant(ctx)
	if err != nil {
		return err
	}
	result := d.db.WithContext(ctx).Exec(`DELETE FROM projects WHERE tenant_id = ? AND name = ?`, tenantID, name)
	if result.Error != nil {
		return fmt.Errorf("delete project %q: %w", name, result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("%w: %q was not found in %s", ErrNotFound, name, dbSource)
	}
	return nil
}

// tenant is the calling tenant's id, its projects.json imported first.
func (d *DB) tenant(ctx context.Context) (string, error) {
	tenantID := tenant.FromContext(ctx).OrgID.String()
	if _, done := d.imported.Load(tenantID); done {
		return tenantID, nil
	}
	if tenantID == tenant.LocalOrgID.String() {
		if err := d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error { return d.importFile(tx, tenantID) }); err != nil {
			return "", fmt.Errorf("import gavel projects file %s into %s: %w", d.importPath, dbSource, err)
		}
	}
	d.imported.Store(tenantID, struct{}{})
	return tenantID, nil
}

// importFile copies projects.json into an empty projects table, keeping each
// project's file-mode id, and records the import so it runs once.
func (d *DB) importFile(tx *gorm.DB, tenantID string) error {
	recorded := tx.Exec(`INSERT INTO project_imports (tenant_id, source) VALUES (?, ?) ON CONFLICT (tenant_id) DO NOTHING`, tenantID, d.importPath)
	if recorded.Error != nil || recorded.RowsAffected == 0 {
		return recorded.Error
	}
	var existing int64
	if err := tx.Raw(`SELECT count(*) FROM projects WHERE tenant_id = ?`, tenantID).Scan(&existing).Error; err != nil || existing > 0 {
		return err
	}
	records, err := NewFile(d.importPath).records()
	if err != nil {
		return err
	}
	for _, r := range records {
		created, err := insertProject(tx, tenantID, FileID(r.Name), Project{Name: r.Name, Kind: r.Kind, Dir: r.Dir, Repos: r.Repos})
		if err != nil {
			return err
		}
		if !created {
			return fmt.Errorf("%w: %q is listed twice", ErrExists, r.Name)
		}
	}
	return nil
}

// insertProject adds p under id, reporting false when the tenant already has
// a project of that name.
func insertProject(tx *gorm.DB, tenantID string, id uuid.UUID, p Project) (bool, error) {
	repos, err := json.Marshal(p.registeredRepos())
	if err != nil {
		return false, fmt.Errorf("marshal repos of project %q: %w", p.Name, err)
	}
	result := tx.Exec(`INSERT INTO projects (id, tenant_id, name, kind, dir, repos) VALUES (?, ?, ?, ?, NULLIF(?, ''), ?::jsonb) ON CONFLICT (tenant_id, name) DO NOTHING`,
		id, tenantID, p.Name, p.withKind().Kind, p.Dir, string(repos))
	if result.Error != nil {
		return false, fmt.Errorf("insert project %q: %w", p.Name, result.Error)
	}
	return result.RowsAffected == 1, nil
}

func projectsFromRows(rows []projectRow) ([]Project, error) {
	ps := make([]Project, len(rows))
	for i, row := range rows {
		ps[i] = Project{ID: row.ID, TenantID: row.TenantID, Name: row.Name, Kind: row.Kind, Dir: row.Dir}
		if err := json.Unmarshal([]byte(row.Repos), &ps[i].Repos); err != nil {
			return nil, fmt.Errorf("parse repos of project %q in %s: %w", row.Name, dbSource, err)
		}
	}
	return prepared(ps, dbSource)
}
