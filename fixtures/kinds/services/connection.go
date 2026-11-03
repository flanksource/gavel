package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/flanksource/commons-db/connection/store"
	dbcontext "github.com/flanksource/commons-db/context"
	"github.com/flanksource/commons-db/models"
	"github.com/flanksource/deps/pkg/config"
	"github.com/flanksource/gavel/fixtures"
	_ "github.com/microsoft/go-mssqldb"
	"github.com/microsoft/go-mssqldb/msdsn"
)

type ConnectionLookup struct {
	Ref string `json:"ref"`
	App string `json:"app,omitempty"`
}

func lookupConnection(ctx context.Context, step *fixtures.StepContext, lookup *ConnectionLookup) (*models.Connection, error) {
	app := lookup.App
	if app == "" {
		app = "gavel"
	}
	layers, err := store.DefaultLayers(app)
	if err != nil {
		return nil, err
	}
	layers[len(layers)-1].Dir = filepath.Join(step.WorkDir, ".connections")
	catalog, err := store.New(layers...)
	if err != nil {
		return nil, err
	}
	dbctx := catalog.Attach(dbcontext.NewContext(ctx))
	connection, err := catalog.ResolveConnection(dbctx, lookup.Ref)
	if err != nil || connection == nil {
		return nil, err
	}
	return dbcontext.HydrateConnection(dbctx, connection)
}

func existingConnection(ctx context.Context, step *fixtures.StepContext, entry Entry) (*depsService, map[string]any, error) {
	if entry.Connection == nil {
		return nil, nil, nil
	}
	connection, err := lookupConnection(ctx, step, entry.Connection)
	if err != nil || connection == nil {
		return nil, nil, err
	}
	_, spec, ok := config.GetService(entry.Deps)
	if !ok || connection.Type != spec.Type {
		return nil, nil, fmt.Errorf("entry %q: saved connection type %q does not match service %q", entry.Name, connection.Type, entry.Deps)
	}
	driver, endpoint, err := connectionEndpoint(connection)
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

func connectionEndpoint(connection *models.Connection) (string, *url.URL, error) {
	if connection.Type == "sql_server" {
		cfg, err := msdsn.Parse(connection.URL)
		if err != nil {
			return "", nil, errors.New("invalid SQL Server connection DSN")
		}
		endpoint := cfg.URL()
		params := endpoint.Query()
		for key, value := range cfg.Parameters {
			switch key {
			case "server", "port", "user id", "password", "database":
				continue
			}
			params.Set(key, value)
		}
		endpoint.RawQuery = params.Encode()
		return "sqlserver", endpoint, nil
	}
	if connection.Type != "postgres" {
		return "", nil, fmt.Errorf("existing connection type %q is unsupported", connection.Type)
	}
	endpoint, err := url.Parse(connection.URL)
	if err != nil || endpoint.Hostname() == "" || !strings.HasPrefix(endpoint.Scheme, "postgres") {
		return "", nil, errors.New("invalid PostgreSQL connection URL")
	}
	if endpoint.Port() == "" {
		endpoint.Host = net.JoinHostPort(endpoint.Hostname(), "5432")
	}
	return "postgres", endpoint, nil
}
