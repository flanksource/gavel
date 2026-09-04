package database

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	commonsdb "github.com/flanksource/commons-db/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConcurrentOpenSerializesMigrationLifecycle(t *testing.T) {
	if os.Getenv("GAVEL_DB_EMBEDDED_TEST") == "" {
		t.Skip("set GAVEL_DB_EMBEDDED_TEST=1 to run embedded-postgres migration-lock tests")
	}
	dsn, stop, err := commonsdb.StartEmbedded(commonsdb.EmbeddedConfig{
		DataDir:  filepath.Join(t.TempDir(), "postgres"),
		Database: "gavel_migration_lock",
	})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, stop()) })

	clearDatabaseEnvironment(t)
	t.Setenv(EnvDSN, dsn)

	lockPool, err := commonsdb.NewDB(dsn)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, lockPool.Close()) })
	lockConn, err := lockPool.Conn(t.Context())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, lockConn.Close()) })

	_, err = lockConn.ExecContext(t.Context(),
		`SELECT pg_advisory_lock($1)`, migrationAdvisoryLockID,
	)
	require.NoError(t, err)
	lockHeld := true
	t.Cleanup(func() {
		if !lockHeld {
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), migrationUnlockTimeout)
		defer cancel()
		var unlocked bool
		require.NoError(t, lockConn.QueryRowContext(ctx,
			`SELECT pg_advisory_unlock($1)`, migrationAdvisoryLockID,
		).Scan(&unlocked))
		require.True(t, unlocked)
	})

	const callers = 4
	type openResult struct {
		db  *DB
		err error
	}
	results := make(chan openResult, callers)
	start := make(chan struct{})
	var ready sync.WaitGroup
	ready.Add(callers)
	for range callers {
		go func() {
			ready.Done()
			<-start
			db, err := Open(t.Context(), WithMigrations())
			results <- openResult{db: db, err: err}
		}()
	}
	ready.Wait()
	close(start)

	// The externally held session lock forces every Open onto PostgreSQL's
	// advisory-lock wait queue before any process can begin Captain preflight.
	deadline := time.Now().Add(10 * time.Second)
	for {
		var waiting int
		require.NoError(t, lockConn.QueryRowContext(t.Context(), `
			SELECT count(*)
			FROM pg_locks
			WHERE locktype = 'advisory' AND NOT granted`).Scan(&waiting))
		if waiting >= callers {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("only %d/%d concurrent Open calls waited for the migration advisory lock", waiting, callers)
		}
		time.Sleep(25 * time.Millisecond)
	}

	var captainSchemaAbsent bool
	require.NoError(t, lockConn.QueryRowContext(t.Context(),
		`SELECT to_regclass('public.captain_sessions') IS NULL`,
	).Scan(&captainSchemaAbsent))
	assert.True(t, captainSchemaAbsent, "Captain preflight/migration must not begin before the lifecycle lock is acquired")

	var unlocked bool
	require.NoError(t, lockConn.QueryRowContext(t.Context(),
		`SELECT pg_advisory_unlock($1)`, migrationAdvisoryLockID,
	).Scan(&unlocked))
	require.True(t, unlocked)
	lockHeld = false

	for range callers {
		select {
		case result := <-results:
			require.NoError(t, result.err)
			require.NotNil(t, result.db)
			require.NoError(t, result.db.Close())
		case <-time.After(45 * time.Second):
			t.Fatal("timed out waiting for serialized concurrent database Open calls")
		}
	}

	// Callback failures must still release and close the dedicated lock
	// session. A subsequent session can acquire the same key immediately.
	sentinel := errors.New("migration callback failed")
	err = withMigrationAdvisoryLock(t.Context(), dsn, func() error { return sentinel })
	require.ErrorIs(t, err, sentinel)
	var reacquired bool
	require.NoError(t, lockConn.QueryRowContext(t.Context(),
		`SELECT pg_try_advisory_lock($1)`, migrationAdvisoryLockID,
	).Scan(&reacquired))
	require.True(t, reacquired, "failed migration callback leaked the session advisory lock")
	require.NoError(t, lockConn.QueryRowContext(t.Context(),
		`SELECT pg_advisory_unlock($1)`, migrationAdvisoryLockID,
	).Scan(&unlocked))
	require.True(t, unlocked)
}
