package postgres

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"github.com/flanksource/commons-db/migrate"
	_ "github.com/lib/pq"
)

//go:embed schema/*.hcl
var schema embed.FS

// Provision explicitly provisions a fresh comments schema; constructing a store never performs DDL.
func Provision(ctx context.Context, dsn string) error {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return err
	}
	defer db.Close()
	var exists bool
	if err = db.QueryRowContext(ctx, "SELECT to_regclass('public.comments') IS NOT NULL").Scan(&exists); err != nil {
		return err
	}
	if exists {
		var count int
		err = db.QueryRowContext(ctx, `SELECT count(*) FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'comments' AND column_name IN ('id','tenant_id','scope_uri','target_uri','phase','version','request_id','request_hash','payload')`).Scan(&count)
		if err != nil {
			return err
		}
		if count != 9 {
			return fmt.Errorf("existing comments table is not owned by this service; conversion is outside fresh provisioning")
		}
	}
	return migrate.Apply(ctx, dsn, schema, migrate.WithDir("schema"), migrate.WithName("gavel-comments"))
}
