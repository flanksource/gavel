package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/commons-db/migrate"
	"github.com/jackc/pgx/v5"
	"github.com/zclconf/go-cty/cty"
	ctyjson "github.com/zclconf/go-cty/cty/json"

	"github.com/flanksource/gavel/fixtures"
)

type databaseService struct {
	connection *sql.DB
	cleanup    func() error
	once       sync.Once
	err        error
}

func startDBTest(ctx context.Context, step *fixtures.StepContext, entry Entry) (*databaseService, map[string]any, error) {
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}
	opts := dbtest.Options{Name: entry.Name}
	if cfg := entry.DBTest.Migrations; cfg != nil {
		provisioner, err := migrationProvisioner(step.WorkDir, cfg)
		if err != nil {
			return nil, nil, fmt.Errorf("entry %q: %w", entry.Name, err)
		}
		opts.Provisioner = provisioner
	}
	database, cleanup, err := dbtest.Open(opts)
	if err != nil {
		return nil, nil, fmt.Errorf("entry %q: provision database: %w", entry.Name, err)
	}
	svc := &databaseService{cleanup: cleanup}
	outputs, err := databaseOutputs(database.DSN())
	if err == nil {
		svc.connection, err = sql.Open("postgres", database.DSN())
	}
	if err == nil {
		err = svc.connection.PingContext(ctx)
	}
	if err == nil {
		err = ctx.Err()
	}
	if err != nil {
		return nil, nil, errors.Join(fmt.Errorf("entry %q: connect database: %w", entry.Name, err), svc.stop())
	}
	return svc, outputs, nil
}

func migrationProvisioner(workDir string, cfg *MigrationConfig) (dbtest.Provisioner, error) {
	dir := cfg.Dir
	if !filepath.IsAbs(dir) {
		dir = filepath.Join(workDir, dir)
	}
	info, err := os.Stat(dir)
	if err != nil {
		return nil, fmt.Errorf("migration directory %q: %w", dir, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("migration directory %q is not a directory", dir)
	}
	opts := []migrate.Option{}
	if cfg.Name != "" {
		opts = append(opts, migrate.WithName(cfg.Name))
	}
	if cfg.Schema != "" {
		opts = append(opts, migrate.WithSchema(cfg.Schema))
	}
	variables := make(map[string]cty.Value, len(cfg.Variables))
	for name, data := range cfg.Variables {
		typ, err := ctyjson.ImpliedType(data)
		if err != nil {
			return nil, fmt.Errorf("migration variable %q: %w", name, err)
		}
		value, err := ctyjson.Unmarshal(data, typ)
		if err != nil {
			return nil, fmt.Errorf("migration variable %q: %w", name, err)
		}
		variables[name] = value
	}
	if len(variables) > 0 {
		opts = append(opts, migrate.WithVariables(variables))
	}
	return migrate.NewProvisioner(os.DirFS(dir), opts...), nil
}

func databaseOutputs(dsn string) (map[string]any, error) {
	cfg, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse provisioned database connection: %w", err)
	}
	parsed, err := url.Parse(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse provisioned database URL: %w", err)
	}
	sslmode := parsed.Query().Get("sslmode")
	if sslmode == "" {
		sslmode = "prefer"
	}
	env := map[string]any{
		"DATABASE_URL": dsn, "GAVEL_DB_DSN": dsn, "COMMONS_DB_URL": dsn, "COMMONS_DB_CREATE": "false",
		"PGHOST": cfg.Host, "PGPORT": strconv.Itoa(int(cfg.Port)), "PGDATABASE": cfg.Database,
		"PGUSER": cfg.User, "PGPASSWORD": cfg.Password, "PGSSLMODE": sslmode,
	}
	return map[string]any{
		"type": "dbtest", "ready": true, "url": dsn, "host": cfg.Host, "port": int(cfg.Port),
		"database": cfg.Database, "username": cfg.User, "password": cfg.Password, "env": env,
		"connection": map[string]any{"url": dsn, "type": "postgres", "username": cfg.User, "password": cfg.Password,
			"properties": map[string]any{"host": cfg.Host, "port": int(cfg.Port), "database": cfg.Database}},
	}, nil
}

func (s *databaseService) observe(ctx context.Context) map[string]any {
	if err := s.connection.PingContext(ctx); err != nil {
		return map[string]any{"ready": false, "error": err.Error()}
	}
	return map[string]any{"ready": true}
}

func (s *databaseService) stop() error {
	s.once.Do(func() {
		if s.connection != nil {
			s.err = s.connection.Close()
		}
		s.err = errors.Join(s.err, s.cleanup())
	})
	return s.err
}
