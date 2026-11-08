package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/flanksource/deps/pkg/config"
	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/fixtures/kinds/dbconn"
)

type ConnectionLookup struct {
	Ref string `json:"ref"`
	App string `json:"app,omitempty"`
}

// existingConnection borrows the saved connection an entry names. A
// connection no layer defines is absent (nil, nil): the entry is then
// provisioned. Any other resolution failure is returned.
func existingConnection(ctx context.Context, step *fixtures.StepContext, entry Entry) (*depsService, map[string]any, error) {
	if entry.Connection == nil {
		return nil, nil, nil
	}
	resolved, err := dbconn.Resolve(ctx, step, entry.Connection.Ref, dbconn.Options{App: entry.Connection.App})
	if errors.Is(err, dbconn.ErrNotFound) {
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, err
	}
	connection := resolved.Connection
	_, spec, ok := config.GetService(entry.Deps)
	if !ok || connection.Type != spec.Type {
		return nil, nil, fmt.Errorf("entry %q: saved connection type %q does not match service %q", entry.Name, connection.Type, entry.Deps)
	}
	driver, endpoint, err := dbconn.Endpoint(connection)
	if err != nil {
		return nil, nil, err
	}
	db, err := sql.Open(driver, endpoint.String())
	if err != nil {
		return nil, nil, err
	}
	probeCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := db.PingContext(probeCtx); err != nil {
		return nil, nil, errors.Join(fmt.Errorf("entry %q: existing connection probe: %w", entry.Name, err), db.Close())
	}
	port, err := strconv.Atoi(endpoint.Port())
	if err != nil {
		return nil, nil, errors.Join(err, db.Close())
	}
	return &depsService{borrowed: db}, map[string]any{
		"type": "deps", "source": "existing", "ready": true,
		"url": endpoint.String(), "host": endpoint.Hostname(), "port": port,
		"connection": map[string]any{"type": connection.Type, "url": endpoint.String()},
	}, nil
}
