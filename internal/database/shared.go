package database

import (
	"context"
	"errors"
	"fmt"
	"sync"

	captaincli "github.com/flanksource/captain/pkg/cli"
	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/clicky/shutdown"
)

var processDB struct {
	sync.Mutex
	db *DB
}

// Shared returns the single database handle owned by the current Gavel
// process. Failed opens are not cached, so a transient connection or migration
// failure can be retried by a later caller. Callers must not own the returned
// pool; Close is intentionally a no-op on shared handles.
func Shared(ctx context.Context, optionFns ...Option) (*DB, error) {
	processDB.Lock()
	defer processDB.Unlock()

	if processDB.db != nil {
		if resolveOpenOptions(optionFns...).migrate && !processDB.db.migrated {
			return nil, errors.New("gavel serve cannot migrate after the process database was opened without migrations")
		}
		return processDB.db, nil
	}

	db, err := Open(ctx, optionFns...)
	if err != nil {
		return nil, err
	}
	if !db.Disabled() {
		captainDB, err := captaindb.Use(db.Gorm())
		if err != nil {
			_ = db.closeUnderlying()
			return nil, fmt.Errorf("reuse Gavel pool for Captain: %w", err)
		}
		if err := captaincli.ConfigureNativeDatabase(captainDB.Gorm()); err != nil {
			_ = db.closeUnderlying()
			return nil, fmt.Errorf("configure Captain database: %w", err)
		}
	}
	db.shared = true
	processDB.db = db
	shutdown.AddHookWithPriority("gavel-database", shutdown.PriorityDatabase, func() {
		_ = closeShared()
	})
	return db, nil
}

// ResetDisabledSharedForTest clears a disabled process handle so tests can
// exercise different environment configurations. An enabled shared handle is
// intentionally single-lifetime: Captain holds the same pool by identity and
// rejects replacement, just as production does.
func ResetDisabledSharedForTest() error {
	processDB.Lock()
	defer processDB.Unlock()
	if processDB.db == nil {
		return nil
	}
	if !processDB.db.Disabled() {
		return errors.New("cannot reset an enabled shared database")
	}
	processDB.db = nil
	return nil
}

func closeShared() error {
	processDB.Lock()
	defer processDB.Unlock()
	if processDB.db == nil {
		return nil
	}
	db := processDB.db
	processDB.db = nil
	return db.closeUnderlying()
}
