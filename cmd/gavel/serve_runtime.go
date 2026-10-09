package main

import (
	"context"
	"fmt"
	"os"
	"time"

	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/captain/pkg/monitor"
	"github.com/flanksource/commons/logger"
	shared "github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/internal/database/todoprojection"
	"gorm.io/gorm"
)

type serveDatabase interface {
	Disabled() bool
	Gorm() *gorm.DB
	DSN() string
	DSNSource() string
}

type serveSessionMonitor interface {
	Run(context.Context) error
	Ready() <-chan struct{}
	IngestStats() monitor.IngestStats
}

// serveProjection is a supervised LISTEN: todoprojection.Projection, which
// keeps TODO activity watermarks current from Captain's row-change feed, and
// ui.GitChangeListener, which pushes git state changes to the dashboard.
type serveProjection interface {
	Run(context.Context) error
	Ready() <-chan struct{}
}

type serveDatabaseMode uint8

const (
	serveDatabaseNoMigrations serveDatabaseMode = iota
	serveDatabaseWithMigrations
)

// projectionRestartDelay spaces restarts of a projection that failed after
// startup. Each restart re-LISTENs and resyncs, so no change is lost for good.
const projectionRestartDelay = 5 * time.Second

type serveRuntimeDependencies struct {
	openDatabase           func(context.Context, serveDatabaseMode) (serveDatabase, error)
	newMonitor             func(*gorm.DB) (serveSessionMonitor, error)
	newProjection          func(*gorm.DB) (serveProjection, error)
	projectionRestartDelay time.Duration
	countLiveSessions      func(context.Context, *gorm.DB) (int64, error)
	logInfo                func(string)
}

var defaultServeRuntimeDependencies = serveRuntimeDependencies{
	openDatabase: func(ctx context.Context, mode serveDatabaseMode) (serveDatabase, error) {
		if mode == serveDatabaseWithMigrations {
			return shared.Shared(ctx, shared.WithMigrations())
		}
		return shared.Shared(ctx)
	},
	newMonitor: func(gormDB *gorm.DB) (serveSessionMonitor, error) {
		db, err := captaindb.Use(gormDB)
		if err != nil {
			return nil, err
		}
		hostID, _ := os.Hostname()
		return monitor.New(monitor.Config{DB: db, HostID: hostID})
	},
	newProjection: func(gormDB *gorm.DB) (serveProjection, error) {
		return todoprojection.New(gormDB)
	},
	projectionRestartDelay: projectionRestartDelay,
	countLiveSessions: func(ctx context.Context, gormDB *gorm.DB) (int64, error) {
		db, err := captaindb.Use(gormDB)
		if err != nil {
			return 0, err
		}
		return db.CountLiveRootSessions(ctx)
	},
	logInfo: func(message string) { logger.Infof("%s", message) },
}

func serveDatabaseStartupMessage(db serveDatabase, liveSessions int64) string {
	if db.Disabled() {
		return "Database Info: source=\"disabled\" dsn=\"\" live_sessions=0"
	}
	return fmt.Sprintf("Database Info: source=%q dsn=%q live_sessions=%d",
		db.DSNSource(), captaindb.MaskDSN(db.DSN()), liveSessions)
}

// serveRuntime is what startServeRuntime started: the monitor's counter reader
// and the shared pool, both nil when persistence is off.
type serveRuntime struct {
	IngestStats func() monitor.IngestStats
	DB          *gorm.DB
}

// startServeRuntime synchronously opens Gavel's process database before any
// serve goroutine or HTTP listener can initialize Captain independently. When
// persistence is enabled, it then runs Captain's continuous session monitor and
// Gavel's row-change projection on that same pool until the serve context is
// cancelled.
func startServeRuntime(ctx context.Context, deps serveRuntimeDependencies, mode serveDatabaseMode) (serveRuntime, error) {
	db, err := deps.openDatabase(ctx, mode)
	if err != nil {
		return serveRuntime{}, fmt.Errorf("initialize Gavel shared database: %w", err)
	}
	if db.Disabled() {
		deps.logInfo(serveDatabaseStartupMessage(db, 0))
		return serveRuntime{}, nil
	}

	mon, err := deps.newMonitor(db.Gorm())
	if err != nil {
		return serveRuntime{}, fmt.Errorf("initialize Captain session monitor: %w", err)
	}
	go func() {
		if err := mon.Run(ctx); err != nil && ctx.Err() == nil {
			logger.Errorf("Captain session monitor stopped: %v", err)
		}
	}()
	select {
	case <-mon.Ready():
	case <-ctx.Done():
		return serveRuntime{}, ctx.Err()
	}
	if err := startProjection(ctx, deps, db.Gorm()); err != nil {
		return serveRuntime{}, err
	}
	liveSessions, err := deps.countLiveSessions(ctx, db.Gorm())
	if err != nil {
		return serveRuntime{}, fmt.Errorf("count live Captain sessions: %w", err)
	}
	deps.logInfo(serveDatabaseStartupMessage(db, liveSessions))
	return serveRuntime{IngestStats: mon.IngestStats, DB: db.Gorm()}, nil
}

// startProjection runs the row-change projection under startSupervised.
func startProjection(ctx context.Context, deps serveRuntimeDependencies, gormDB *gorm.DB) error {
	projection, err := deps.newProjection(gormDB)
	if err != nil {
		return fmt.Errorf("initialize Captain row-change projection: %w", err)
	}
	return startSupervised(ctx, supervisedListener{
		name: "Captain row-change projection", stale: "TODO activity is stale", restartDelay: deps.projectionRestartDelay,
	}, projection)
}

// supervisedListener names a long-lived LISTEN for its logs and errors: stale
// says what goes out of date while it is down.
type supervisedListener struct {
	name, stale  string
	restartDelay time.Duration
}

// startSupervised runs a LISTEN on a plain goroutine, like the session
// monitor, never as a clicky task: a long-lived task holds every task drain
// open. Startup waits until it is ready, so one that cannot run at all (for
// example against a database Gavel's schema was never applied to) fails serve
// loudly. One that fails once running is logged and restarted, and every
// restart catches up on the changes it missed.
func startSupervised(ctx context.Context, listener supervisedListener, run serveProjection) error {
	startFailed := make(chan error, 1)
	go superviseProjection(ctx, listener, run, startFailed)
	select {
	case <-run.Ready():
		return nil
	case err := <-startFailed:
		return fmt.Errorf("start %s: %w", listener.name, err)
	case <-ctx.Done():
		return ctx.Err()
	}
}

func superviseProjection(ctx context.Context, listener supervisedListener, run serveProjection, startFailed chan<- error) {
	for {
		err := run.Run(ctx)
		if ctx.Err() != nil {
			return
		}
		select {
		case <-run.Ready():
		default:
			startFailed <- err
			return
		}
		logger.Errorf("%s stopped, %s until it restarts in %s: %v", listener.name, listener.stale, listener.restartDelay, err)
		select {
		case <-ctx.Done():
			return
		case <-time.After(listener.restartDelay):
		}
	}
}
