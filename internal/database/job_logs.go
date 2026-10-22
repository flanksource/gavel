// Job logs in the blob store: where an organization's GitHub Actions job logs live, and the
// migration step that moves logs still inline in job_log_caches.logs_gz there.
package database

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	commonsdb "github.com/flanksource/commons-db/db"
	"github.com/flanksource/gavel/blob"
	"github.com/flanksource/gavel/tenant"
	"github.com/flanksource/gavel/verify"
	"github.com/google/uuid"
)

// OpenBlobs is the blob store ~/.gavel.yaml configures, or <user cache dir>/gavel/blobs.
func OpenBlobs(ctx context.Context) (blob.Store, error) {
	store, err := verify.OpenGlobalStorage(ctx)
	if err != nil || store != nil {
		return store, err
	}
	dir, err := os.UserCacheDir()
	if err != nil {
		return nil, fmt.Errorf("find the cache directory for the blob store: %w", err)
	}
	return blob.Local(filepath.Join(dir, "gavel", "blobs")), nil
}

// OrgBlobs is org's part of store: orgs/<org>/.
func OrgBlobs(store blob.Store, org uuid.UUID) blob.Store {
	return blob.WithPrefix(store, "orgs/"+org.String()+"/")
}

// JobLogKey is the key of a job's gzipped logs in its organization's blobs.
func JobLogKey(jobID int64) string {
	return fmt.Sprintf("joblogs/%d.log.gz", jobID)
}

// moveInlineJobLogs writes the logs of every job_log_caches row that still holds them inline
// to the blob store and records their key, so the schema can drop logs_gz without losing
// any. A failed write fails the migration, leaving the column and its logs in place.
func moveInlineJobLogs(ctx context.Context, dsn string) (returnErr error) {
	pool, err := commonsdb.NewDB(dsn)
	if err != nil {
		return fmt.Errorf("open Gavel database to move inline job logs: %w", err)
	}
	defer func() {
		if err := pool.Close(); err != nil {
			returnErr = errors.Join(returnErr, err)
		}
	}()
	hasColumn := func(column string) (bool, error) {
		var exists bool
		err := pool.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM information_schema.columns
			WHERE table_schema = 'public' AND table_name = 'job_log_caches' AND column_name = $1)`, column).Scan(&exists)
		if err != nil {
			return false, fmt.Errorf("look for job_log_caches.%s: %w", column, err)
		}
		return exists, nil
	}
	inline, err := hasColumn("logs_gz")
	if err != nil || !inline {
		return err
	}
	// A database from before organizations has no org_id yet: its rows are the local org's,
	// as the schema then records. log_key is added as the schema declares it.
	orgColumn := "org_id"
	if scoped, err := hasColumn("org_id"); err != nil {
		return err
	} else if !scoped {
		orgColumn = "'" + tenant.LocalOrgID.String() + "'::uuid"
	}
	if _, err := pool.ExecContext(ctx, `ALTER TABLE public.job_log_caches ADD COLUMN IF NOT EXISTS log_key text NOT NULL DEFAULT ''`); err != nil {
		return fmt.Errorf("add job_log_caches.log_key: %w", err)
	}
	type job struct {
		org uuid.UUID
		id  int64
	}
	rows, err := pool.QueryContext(ctx, `SELECT `+orgColumn+`, job_id FROM public.job_log_caches WHERE logs_gz IS NOT NULL`)
	if err != nil {
		return fmt.Errorf("list inline job logs: %w", err)
	}
	var jobs []job
	for rows.Next() {
		var j job
		if err := rows.Scan(&j.org, &j.id); err != nil {
			_ = rows.Close()
			return fmt.Errorf("list inline job logs: %w", err)
		}
		jobs = append(jobs, j)
	}
	if err := errors.Join(rows.Err(), rows.Close()); err != nil {
		return fmt.Errorf("list inline job logs: %w", err)
	}
	if len(jobs) == 0 {
		return nil
	}
	store, err := OpenBlobs(ctx)
	if err != nil {
		return fmt.Errorf("move inline job logs: %w", err)
	}
	for _, j := range jobs {
		// One row at a time: a job's logs can be tens of megabytes.
		var gz []byte
		if err := pool.QueryRowContext(ctx, `SELECT logs_gz FROM public.job_log_caches WHERE job_id = $1 AND `+orgColumn+` = $2`, j.id, j.org).Scan(&gz); err != nil {
			return fmt.Errorf("read inline logs of job %d: %w", j.id, err)
		}
		key := JobLogKey(j.id)
		if err := OrgBlobs(store, j.org).Write(ctx, key, bytes.NewReader(gz)); err != nil {
			return fmt.Errorf("move logs of job %d to the blob store: %w", j.id, err)
		}
		if _, err := pool.ExecContext(ctx, `UPDATE public.job_log_caches SET log_key = $1, logs_gz = NULL WHERE job_id = $2 AND `+orgColumn+` = $3`, key, j.id, j.org); err != nil {
			return fmt.Errorf("record moved logs of job %d: %w", j.id, err)
		}
	}
	return nil
}
