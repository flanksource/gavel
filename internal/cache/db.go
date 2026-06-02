package cache

import (
	"database/sql"
	"fmt"
	"sync"

	commonsLogger "github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/models"
	// Pure-Go SQLite gorm dialector backed by modernc.org/sqlite (via the
	// clarkmcc/gorm-sqlite fork, see the go.mod replace). It imports
	// modernc.org/sqlite directly rather than vendoring its own copy, so the
	// "sqlite" database/sql driver is registered exactly once even when another
	// modernc consumer (e.g. clicky) is linked into the same binary. Keeps the
	// build CGO-free (no mattn/go-sqlite3).
	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	// Registers the "sqlite" database/sql driver that the dialector above opens
	// against. The clarkmcc fork does not register it itself, and registration
	// is shared (idempotent within a single linked copy) with other modernc
	// consumers such as clicky.
	_ "modernc.org/sqlite"
)

// DB wraps GORM DB with mutex synchronization for write operations
type DB struct {
	conn    *gorm.DB
	writeMu sync.Mutex // Protects write operations
}

// NewDB creates a new synchronized GORM database wrapper and, for sqlite,
// auto-migrates the lint-violation models. Callers that own a different
// schema (e.g. github/cache) should use NewDBRaw to skip that migration.
func NewDB(driverName, dataSourceName string) (*DB, error) {
	db, err := NewDBRaw(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}
	if driverName == "sqlite" {
		if err := db.Migrate(); err != nil {
			return nil, fmt.Errorf("failed to run migrations: %w", err)
		}
	}
	return db, nil
}

// NewDBRaw opens the underlying database without running any migrations.
// Use this when the caller owns its own schema (the github cache does) and
// does not want lint-violation tables auto-created alongside its data.
func NewDBRaw(driverName, dataSourceName string) (*DB, error) {
	logMode := logger.Silent
	if commonsLogger.IsLevelEnabled(3) {
		logMode = logger.Info // SQL query logging at -vvv
	}
	config := &gorm.Config{Logger: logger.Default.LogMode(logMode)}

	var gormDB *gorm.DB
	var err error
	switch driverName {
	case "sqlite":
		gormDB, err = gorm.Open(sqlite.Open(dataSourceName), config)
		if err != nil {
			return nil, err
		}
		if err := configureSQLitePragmas(gormDB); err != nil {
			return nil, err
		}
	case "postgres":
		gormDB, err = gorm.Open(postgres.Open(dataSourceName), config)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", driverName)
	}

	return &DB{conn: gormDB}, nil
}

func configureSQLitePragmas(gormDB *gorm.DB) error {
	sqlDB, err := gormDB.DB()
	if err != nil {
		return err
	}
	pragmas := []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA busy_timeout=5000",
		"PRAGMA foreign_keys=ON",
		"PRAGMA synchronous=NORMAL",
	}
	for _, p := range pragmas {
		if _, err := sqlDB.Exec(p); err != nil {
			return err
		}
	}
	return nil
}

// GormDB returns the underlying GORM database instance
func (db *DB) GormDB() *gorm.DB {
	return db.conn
}

// Exec executes a query with mutex protection for writes
func (db *DB) Exec(query string, args ...interface{}) (sql.Result, error) {
	db.writeMu.Lock()
	defer db.writeMu.Unlock()
	sqlDB, err := db.conn.DB()
	if err != nil {
		return nil, err
	}
	return sqlDB.Exec(query, args...)
}

// Begin starts a transaction with mutex protection
func (db *DB) Begin() (*Tx, error) {
	db.writeMu.Lock()
	sqlDB, err := db.conn.DB()
	if err != nil {
		db.writeMu.Unlock()
		return nil, err
	}
	tx, err := sqlDB.Begin()
	if err != nil {
		db.writeMu.Unlock()
		return nil, err
	}
	return &Tx{tx: tx, db: db}, nil
}

// Query performs read operations (no mutex needed for reads)
func (db *DB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	sqlDB, err := db.conn.DB()
	if err != nil {
		return nil, err
	}
	return sqlDB.Query(query, args...)
}

// QueryRow performs single row reads (no mutex needed for reads)
func (db *DB) QueryRow(query string, args ...interface{}) *sql.Row {
	sqlDB, _ := db.conn.DB()
	return sqlDB.QueryRow(query, args...)
}

// Prepare prepares a statement
func (db *DB) Prepare(query string) (*sql.Stmt, error) {
	sqlDB, err := db.conn.DB()
	if err != nil {
		return nil, err
	}
	return sqlDB.Prepare(query)
}

// Close closes the database connection
func (db *DB) Close() error {
	db.writeMu.Lock()
	defer db.writeMu.Unlock()
	sqlDB, err := db.conn.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// Migrate runs automatic migrations for all models
func (db *DB) Migrate() error {
	return db.conn.AutoMigrate(
		&models.Violation{},
		&FileScan{},
	)
}

// Tx wraps sql.Tx to ensure mutex is released on commit/rollback
type Tx struct {
	tx       *sql.Tx
	db       *DB
	finished bool // Track if transaction is already finished
}

// Exec executes a query within the transaction
func (t *Tx) Exec(query string, args ...interface{}) (sql.Result, error) {
	return t.tx.Exec(query, args...)
}

// Prepare prepares a statement within the transaction
func (t *Tx) Prepare(query string) (*sql.Stmt, error) {
	return t.tx.Prepare(query)
}

// Query performs a query within the transaction
func (t *Tx) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return t.tx.Query(query, args...)
}

// QueryRow performs a single row query within the transaction
func (t *Tx) QueryRow(query string, args ...interface{}) *sql.Row {
	return t.tx.QueryRow(query, args...)
}

// Commit commits the transaction and releases the write lock
func (t *Tx) Commit() error {
	if t.finished {
		return nil // Already committed or rolled back
	}
	t.finished = true
	defer t.db.writeMu.Unlock()
	return t.tx.Commit()
}

// Rollback rolls back the transaction and releases the write lock
func (t *Tx) Rollback() error {
	if t.finished {
		return nil // Already committed or rolled back
	}
	t.finished = true
	defer t.db.writeMu.Unlock()
	return t.tx.Rollback()
}
