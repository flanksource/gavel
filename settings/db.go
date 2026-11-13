package settings

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/flanksource/gavel/tenant"
	"github.com/flanksource/gavel/verify"
	"github.com/flanksource/repomap"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DB keeps every layer in gavel_settings. The files of a target's chain are
// synced into it on each Resolve (recording the repository HEAD they were read
// at); a position without a file holds a layer authored in the database.
// Rows belong to the tenant of each call's context, as projects do.
type DB struct {
	db *gorm.DB
}

var _ Service = (*DB)(nil)

// NewDB is the DB-mode service over db.
func NewDB(db *gorm.DB) (*DB, error) {
	if db == nil {
		return nil, errors.New("database settings need a database")
	}
	return &DB{db: db}, nil
}

func tenantOf(ctx context.Context) string { return tenant.FromContext(ctx).OrgID.String() }

// row is one gavel_settings row.
type row struct {
	ProjectID *uuid.UUID
	Path      *string
	File      *string
	GitSHA    *string
	Payload   json.RawMessage
}

func (d *DB) Resolve(ctx context.Context, target Target) (Resolved, error) {
	positions, files, err := diskChain(target)
	if err != nil {
		return Resolved{}, err
	}
	if err := d.sync(ctx, target, positions, files); err != nil {
		return Resolved{}, err
	}
	authored, err := d.authored(ctx, target, positions)
	if err != nil {
		return Resolved{}, err
	}
	var layers []Layer
	for i, position := range positions {
		positions[i].DB = len(position.Files) == 0 && position.Kind != KindParent
		for _, layer := range files {
			if layer.Kind == position.Kind && layer.Path == position.Path {
				layers = append(layers, layer)
			}
		}
		if layer, ok := authored[key(position.Kind, position.Path)]; ok && positions[i].DB {
			layers = append(layers, layer)
		}
	}
	return compose(positions, layers)
}

func (d *DB) Save(ctx context.Context, target Target, ref Ref, edit func(*verify.GavelConfig) error) error {
	positions, _, err := diskChain(target)
	if err != nil {
		return err
	}
	position, err := positionFor(positions, ref)
	if err != nil {
		return err
	}
	if ref.File != "" {
		if err := saveFile(position, ref.File, edit); err != nil {
			return err
		}
		_, err = d.Resolve(ctx, target)
		return err
	}
	switch {
	case position.Kind == KindParent:
		return fmt.Errorf("%s has no identity in the project, so it takes no database layer; save it to a file", position.Dir)
	case len(position.Files) > 0:
		return fmt.Errorf("%s already has %v: save to one of them, or add a .gavel.<name>.yaml snippet", position.Dir, position.Files)
	}
	projectID, path, err := identity(target, position)
	if err != nil {
		return err
	}
	tenantID := tenantOf(ctx)
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing []row
		if err := tx.Raw(`SELECT payload FROM gavel_settings WHERE tenant_id = ? AND project_id IS NOT DISTINCT FROM ? AND path IS NOT DISTINCT FROM ? AND file IS NULL FOR UPDATE`, tenantID, projectID, path).Scan(&existing).Error; err != nil {
			return err
		}
		var cfg verify.GavelConfig
		if len(existing) > 0 {
			if err := json.Unmarshal(existing[0].Payload, &cfg); err != nil {
				return fmt.Errorf("decode the database layer at %s: %w", position.Dir, err)
			}
		}
		if err := edit(&cfg); err != nil {
			return err
		}
		return upsert(tx, tenantID, projectID, path, nil, nil, cfg)
	})
}

// sync makes the file rows of target's positions match its files: each file's
// row is written when its config changed (recording the HEAD it was read at),
// and a row whose file is gone is deleted.
func (d *DB) sync(ctx context.Context, target Target, positions []Position, files []Layer) error {
	tenantID := tenantOf(ctx)
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		heads := map[string]string{}
		for _, position := range positions {
			// A web target's project and path positions hold no files: the
			// file rows there are a disk project's own (its pages are web
			// targets too), so they are never this target's to drop.
			if position.Kind == KindParent || target.web() && position.Kind != KindUser {
				continue
			}
			projectID, path, err := identity(target, position)
			if err != nil {
				return err
			}
			vanished := tx.Where("tenant_id = ? AND project_id IS NOT DISTINCT FROM ? AND path IS NOT DISTINCT FROM ? AND file IS NOT NULL", tenantID, projectID, path)
			if len(position.Files) > 0 {
				vanished = vanished.Where("file NOT IN ?", position.Files)
			}
			if err := vanished.Table("gavel_settings").Delete(nil).Error; err != nil {
				return fmt.Errorf("drop vanished config files of %s: %w", position.Dir, err)
			}
			for _, layer := range files {
				if layer.Kind != position.Kind || layer.Path != position.Path {
					continue
				}
				sha, err := headOf(heads, position.Dir)
				if err != nil {
					return err
				}
				file := layer.File
				if err := upsert(tx, tenantID, projectID, path, &file, &sha, layer.Config); err != nil {
					return fmt.Errorf("sync %s: %w", layer.Location, err)
				}
			}
		}
		return nil
	})
}

// authored is the database layers of target's positions, by position.
func (d *DB) authored(ctx context.Context, target Target, positions []Position) (map[string]Layer, error) {
	var rows []row
	query := `SELECT project_id, path, payload FROM gavel_settings WHERE tenant_id = ? AND file IS NULL AND (project_id IS NULL AND path IS NULL`
	args := []any{tenantOf(ctx)}
	if target.Project != nil {
		query += ` OR project_id = ?`
		args = append(args, target.Project.ID)
	}
	if err := d.db.WithContext(ctx).Raw(query+`)`, args...).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("read the database settings layers: %w", err)
	}
	out := map[string]Layer{}
	for _, r := range rows {
		kind, path := KindUser, ""
		if r.ProjectID != nil {
			kind = KindProject
			if r.Path != nil {
				kind, path = KindPath, *r.Path
			}
		}
		var cfg verify.GavelConfig
		if err := json.Unmarshal(r.Payload, &cfg); err != nil {
			return nil, fmt.Errorf("decode the database settings layer %s %q: %w", kind, path, err)
		}
		out[key(kind, path)] = Layer{Ref: Ref{Kind: kind, Path: path}, Location: "db", Config: cfg}
	}
	return out, nil
}

func key(kind Kind, path string) string { return string(kind) + "\x00" + path }

// identity is position's (project_id, path) columns: both null for the user's layer.
func identity(target Target, position Position) (*uuid.UUID, *string, error) {
	switch position.Kind {
	case KindUser:
		return nil, nil, nil
	case KindProject, KindPath:
		if target.Project == nil {
			return nil, nil, fmt.Errorf("%s position %s needs the target's project", position.Kind, position.Dir)
		}
		id := target.Project.ID
		if position.Kind == KindProject {
			return &id, nil, nil
		}
		path := position.Path
		return &id, &path, nil
	}
	return nil, nil, fmt.Errorf("%s position %s has no database identity", position.Kind, position.Dir)
}

func upsert(tx *gorm.DB, tenantID string, projectID *uuid.UUID, path, file, sha *string, cfg verify.GavelConfig) error {
	payload, err := verify.MarshalGavelConfig(cfg)
	if err != nil {
		return err
	}
	return tx.Exec(`INSERT INTO gavel_settings (tenant_id, project_id, path, file, git_sha, payload) VALUES (?, ?, ?, ?, ?, ?::jsonb)
ON CONFLICT (tenant_id, COALESCE(project_id, '00000000-0000-0000-0000-000000000000'::uuid), COALESCE(path, ''), COALESCE(file, ''))
DO UPDATE SET payload = EXCLUDED.payload, git_sha = EXCLUDED.git_sha, updated_at = now()
WHERE gavel_settings.payload IS DISTINCT FROM EXCLUDED.payload`, tenantID, projectID, path, file, sha, string(payload)).Error
}

// headOf is the HEAD commit of the repository holding dir, or "none" outside git.
func headOf(cache map[string]string, dir string) (string, error) {
	root := repomap.FindGitRoot(dir)
	if root == "" {
		return "none", nil
	}
	if sha, ok := cache[root]; ok {
		return sha, nil
	}
	out, err := exec.Command("git", "-C", root, "rev-parse", "HEAD").Output()
	if err != nil {
		return "", fmt.Errorf("read the HEAD of %s: %w", root, err)
	}
	cache[root] = strings.TrimSpace(string(out))
	return cache[root], nil
}
