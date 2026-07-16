package database

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/clicky/shutdown"
	commonsdb "github.com/flanksource/commons-db/db"
	"github.com/flanksource/commons-db/migrate"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/service"
	"github.com/spf13/pflag"
	"gorm.io/gorm"
)

const (
	// EnvDSN is the preferred PostgreSQL connection string for all Gavel data.
	EnvDSN = "GAVEL_DB_DSN"
	// EnvDisable disables all database-backed caches when set to "off".
	EnvDisable = "GAVEL_DB"

	// LegacyEnvDSN and LegacyEnvDisable remain supported for compatibility.
	LegacyEnvDSN     = "GAVEL_GITHUB_CACHE_DSN"
	LegacyEnvDisable = "GAVEL_GITHUB_CACHE"
	databaseURLFlag  = "db-url"

	// migrationAdvisoryLockID is the ASCII encoding of "GavelDBM". A stable
	// PostgreSQL session lock serializes independently started Gavel processes
	// across the complete Captain-then-Gavel migration lifecycle.
	migrationAdvisoryLockID int64 = 0x476176656c44424d
	migrationUnlockTimeout        = 5 * time.Second
)

var databaseURL string

// BindDatabaseURLFlag exposes the shared process database as a root persistent
// flag. The explicit CLI value wins over environment variables and db.json.
func BindDatabaseURLFlag(flags *pflag.FlagSet) {
	flags.StringVar(&databaseURL, databaseURLFlag, "", "PostgreSQL database URL (overrides environment and db.json)")
}

//go:embed schema/*.hcl schema/*.sql
var schemaFS embed.FS

// DB owns one application database pool and the metadata used by status views.
// A disabled DB is a valid handle with no underlying connection.
type DB struct {
	gorm          *gorm.DB
	disabled      bool
	shared        bool
	migrated      bool
	dsn           string
	dsnSource     string
	disableSource string
}

// Option configures a Gavel database open.
type Option func(*openOptions)

type openOptions struct {
	migrate bool
}

// WithMigrations applies Captain's and Gavel's schemas before opening.
func WithMigrations() Option {
	return func(options *openOptions) { options.migrate = true }
}

func resolveOpenOptions(optionFns ...Option) openOptions {
	var options openOptions
	for _, option := range optionFns {
		if option != nil {
			option(&options)
		}
	}
	return options
}

type openDependencies struct {
	disabled func() (string, bool)
	resolve  func() (string, string, error)
	migrate  func(context.Context, string) error
	open     func(string, *gorm.Config) (*gorm.DB, error)
}

var defaultOpenDependencies = openDependencies{
	disabled: disabledByEnvironment,
	resolve:  resolveDSN,
	migrate:  migrateDatabase,
	open:     commonsdb.NewGorm,
}

// Open resolves and opens the configured PostgreSQL backend. It does not run
// migrations unless WithMigrations is supplied. With no configuration it
// returns a disabled handle so commands can continue without caching.
func Open(ctx context.Context, options ...Option) (*DB, error) {
	return open(ctx, defaultOpenDependencies, options...)
}

func open(ctx context.Context, deps openDependencies, optionFns ...Option) (*DB, error) {
	options := resolveOpenOptions(optionFns...)
	if source, disabled := deps.disabled(); disabled {
		logger.Debugf("gavel database disabled via %s=off", source)
		return &DB{disabled: true, disableSource: source}, nil
	}

	dsn, source, err := deps.resolve()
	if err != nil {
		return nil, err
	}
	if dsn == "" {
		logger.Debugf("gavel database disabled: %s not set and no db config", EnvDSN)
		return &DB{disabled: true}, nil
	}

	if options.migrate {
		if err := deps.migrate(ctx, dsn); err != nil {
			return nil, err
		}
	}

	gormDB, err := deps.open(dsn, commonsdb.DefaultGormConfig())
	if err != nil {
		return nil, fmt.Errorf("open gavel database: %w", err)
	}
	return &DB{gorm: gormDB, migrated: options.migrate, dsn: dsn, dsnSource: source}, nil
}

func migrateDatabase(ctx context.Context, dsn string) error {
	return withMigrationAdvisoryLock(ctx, dsn, func() error {
		// Captain owns its schema and must migrate before Gavel installs any
		// cross-owner constraints or projections. Both are covered by the same
		// cross-process lifecycle lock.
		if err := captaindb.Migrate(ctx, dsn); err != nil {
			return err
		}
		return applyGavelSchema(ctx, dsn)
	})
}

func applyGavelSchema(ctx context.Context, dsn string) error {
	if err := migrate.Apply(ctx, dsn, schemaFS,
		migrate.WithDir("schema"),
		migrate.WithName("gavel"),
		migrate.WithExclude(
			"captain_*",
			"todo_issue_prompt_runs.todo_issue_prompt_runs_captain_prompt_run_fkey",
			"todo_issue_plans.todo_issue_plans_captain_plan_fkey",
			"todo_issue_plan_revision_details",
			"todo_issue_runtime",
		),
	); err != nil {
		return fmt.Errorf("migrate gavel database: %w", err)
	}
	return nil
}

// withMigrationAdvisoryLock holds one PostgreSQL session-level advisory lock
// on a dedicated sql.Conn for the entire callback. The connection never
// returns to the pool while the lock is held, and closing it is a final lock
// release safeguard if the explicit unlock fails.
func withMigrationAdvisoryLock(ctx context.Context, dsn string, callback func() error) (returnErr error) {
	pool, err := commonsdb.NewDB(dsn)
	if err != nil {
		return fmt.Errorf("open Gavel migration lock database: %w", err)
	}
	defer func() {
		if err := pool.Close(); err != nil {
			returnErr = errors.Join(returnErr, fmt.Errorf("close Gavel migration lock database: %w", err))
		}
	}()

	conn, err := pool.Conn(ctx)
	if err != nil {
		return fmt.Errorf("open Gavel migration lock connection: %w", err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			returnErr = errors.Join(returnErr, fmt.Errorf("close Gavel migration lock connection: %w", err))
		}
	}()

	if _, err := conn.ExecContext(ctx, `SELECT pg_advisory_lock($1)`, migrationAdvisoryLockID); err != nil {
		return fmt.Errorf("acquire Gavel migration lock: %w", err)
	}
	defer func() {
		unlockCtx, cancel := context.WithTimeout(context.Background(), migrationUnlockTimeout)
		defer cancel()

		var unlocked bool
		if err := conn.QueryRowContext(unlockCtx,
			`SELECT pg_advisory_unlock($1)`, migrationAdvisoryLockID,
		).Scan(&unlocked); err != nil {
			returnErr = errors.Join(returnErr, fmt.Errorf("release Gavel migration lock: %w", err))
		} else if !unlocked {
			returnErr = errors.Join(returnErr, errors.New("release Gavel migration lock: lock was not held by the dedicated connection"))
		}
	}()

	return callback()
}

func disabledByEnvironment() (string, bool) {
	if value, ok := os.LookupEnv(EnvDisable); ok && strings.TrimSpace(value) != "" {
		return EnvDisable, strings.EqualFold(strings.TrimSpace(value), "off")
	}
	return LegacyEnvDisable, strings.EqualFold(strings.TrimSpace(os.Getenv(LegacyEnvDisable)), "off")
}

func resolveDSN() (string, string, error) {
	if dsn := strings.TrimSpace(databaseURL); dsn != "" {
		return dsn, "--" + databaseURLFlag, nil
	}
	if dsn := strings.TrimSpace(os.Getenv(EnvDSN)); dsn != "" {
		return dsn, EnvDSN, nil
	}
	if dsn := strings.TrimSpace(os.Getenv(LegacyEnvDSN)); dsn != "" {
		return dsn, LegacyEnvDSN, nil
	}

	cfg, err := service.LoadDBConfig()
	if err != nil {
		return "", "", fmt.Errorf("load db config: %w", err)
	}
	switch cfg.Mode {
	case "":
		return "", "", nil
	case service.DBModeDSN:
		if cfg.DSN == "" {
			return "", "", fmt.Errorf("db config: mode=dsn but DSN is empty")
		}
		return cfg.DSN, "db.json (dsn)", nil
	case service.DBModeEmbedded:
		return resolveEmbeddedDSN()
	default:
		return "", "", fmt.Errorf("unknown db mode %q (expected %q or %q)", cfg.Mode, service.DBModeDSN, service.DBModeEmbedded)
	}
}

func resolveEmbeddedDSN() (string, string, error) {
	if running, err := service.FindRunningEmbeddedPostgres(); err != nil {
		logger.V(1).Infof("probe running embedded postgres: %v", err)
	} else if running != nil {
		logger.Debugf("reusing embedded postgres (pid=%d, port=%d)", running.PID, running.Port)
		dsn := service.EmbeddedDSN(running.Port)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := commonsdb.EnsurePerformanceDiagnostics(ctx, dsn); err != nil {
			return "", "", fmt.Errorf("validate reused Gavel embedded PostgreSQL: %w", err)
		}
		return dsn, "db.json (embedded, reused)", nil
	}

	embeddedConfig, err := service.EmbeddedPostgresConfig()
	if err != nil {
		return "", "", err
	}
	dsn, stop, err := commonsdb.StartEmbedded(embeddedConfig)
	if err != nil {
		return "", "", fmt.Errorf("start embedded postgres: %w", err)
	}
	shutdown.AddHookWithPriority("embedded-postgres", shutdown.PriorityCritical, func() {
		if err := stop(); err != nil {
			logger.Warnf("stop embedded postgres: %v", err)
		}
	})
	return dsn, "db.json (embedded)", nil
}

// Disabled reports whether persistence is unavailable by configuration.
func (db *DB) Disabled() bool { return db == nil || db.disabled || db.gorm == nil }

// Gorm returns the application connection, or nil for a disabled database.
func (db *DB) Gorm() *gorm.DB {
	if db == nil {
		return nil
	}
	return db.gorm
}

// DSN returns the resolved connection string. Callers must redact it before display.
func (db *DB) DSN() string {
	if db == nil {
		return ""
	}
	return db.dsn
}

// DSNSource names the environment variable or config source selected by Open.
func (db *DB) DSNSource() string {
	if db == nil {
		return ""
	}
	return db.dsnSource
}

// DisableSource names the environment variable that disabled the database.
func (db *DB) DisableSource() string {
	if db == nil {
		return ""
	}
	return db.disableSource
}

// Close releases the underlying SQL pool.
func (db *DB) Close() error {
	if db.Disabled() {
		return nil
	}
	if db.shared {
		return nil
	}
	return db.closeUnderlying()
}

func (db *DB) closeUnderlying() error {
	if db == nil || db.gorm == nil {
		return nil
	}
	sqlDB, err := db.gorm.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
