// The connections a local gavel run reads: the file layers every Flanksource tool shares, then
// the local organization's encrypted catalog when it holds any git connections.
package forge

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/flanksource/commons-db/connection/catalog"
	"github.com/flanksource/commons-db/connection/store"
	dbcontext "github.com/flanksource/commons-db/context"
	"github.com/flanksource/commons-db/models"
	"github.com/flanksource/gavel/connections"
	"github.com/flanksource/gavel/tenant"
	"gorm.io/gorm"
)

// GitTypes are the connection types that hold repository credentials.
var GitTypes = []string{models.ConnectionTypeGit, models.ConnectionTypeGithub, models.ConnectionTypeGitlab, models.ConnectionTypeAzureDevops}

// Local are the connections of a local gavel run in dir: the global, app and dir/.connections
// file layers, then, when db holds git connections, the local catalog. A nil db is no catalog.
// closeSources releases the catalog's key.
func Local(ctx context.Context, dir string, db *gorm.DB) (sources []Source, closeSources func() error, err error) {
	layers, err := store.DefaultLayers("gavel")
	if err != nil {
		return nil, nil, err
	}
	layers[2].Dir = filepath.Join(dir, ".connections")
	files, err := store.New(layers...)
	if err != nil {
		return nil, nil, err
	}
	sources, closeSources = []Source{files}, func() error { return nil }
	if db == nil {
		return sources, closeSources, nil
	}
	var held int64
	if err := db.WithContext(ctx).Model(&catalog.Record{}).Where("org_id = ? AND type IN ?", tenant.LocalOrgID, GitTypes).Count(&held).Error; err != nil {
		return nil, nil, fmt.Errorf("count the local catalog's git connections: %w", err)
	}
	if held == 0 {
		return sources, closeSources, nil
	}
	encrypted, closeCatalog, err := connections.OpenStore(ctx, connections.Options{DB: db, Scope: tenant.Local()})
	if err != nil {
		return nil, nil, fmt.Errorf("the local connection catalog holds git connections, but its key is unavailable: %w", err)
	}
	return append(sources, gitCatalog{ctx: ctx, db: db, store: encrypted}), closeCatalog, nil
}

// gitCatalog lists the catalog's git connections, decrypting no others.
type gitCatalog struct {
	ctx   context.Context
	db    *gorm.DB
	store *catalog.Store
}

func (c gitCatalog) ListConnections() ([]*models.Connection, error) {
	var ids []string
	if err := c.db.WithContext(c.ctx).Model(&catalog.Record{}).Where("org_id = ? AND type IN ?", tenant.LocalOrgID, GitTypes).
		Order("namespace, name").Pluck("id", &ids).Error; err != nil {
		return nil, fmt.Errorf("list the local catalog's git connections: %w", err)
	}
	listed := make([]*models.Connection, 0, len(ids))
	for _, id := range ids {
		connection, err := c.store.ResolveConnection(dbcontext.NewContext(c.ctx), id)
		if err != nil {
			return nil, err
		}
		listed = append(listed, connection)
	}
	return listed, nil
}
