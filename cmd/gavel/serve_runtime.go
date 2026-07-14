package main

import (
	"context"
	"fmt"
	"os"

	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/captain/pkg/monitor"
	"github.com/flanksource/commons/logger"
	shared "github.com/flanksource/gavel/internal/database"
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
}

type serveRuntimeDependencies struct {
	openDatabase      func(context.Context) (serveDatabase, error)
	newMonitor        func(*gorm.DB) (serveSessionMonitor, error)
	countLiveSessions func(context.Context, *gorm.DB) (int64, error)
	logInfo           func(string)
}

var defaultServeRuntimeDependencies = serveRuntimeDependencies{
	openDatabase: func(ctx context.Context) (serveDatabase, error) {
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

// startServeRuntime synchronously opens Gavel's process database before any
// serve goroutine or HTTP listener can initialize Captain independently. When
// persistence is enabled, it then runs Captain's continuous session monitor on
// that same pool until the serve context is cancelled.
func startServeRuntime(ctx context.Context, deps serveRuntimeDependencies) error {
	db, err := deps.openDatabase(ctx)
	if err != nil {
		return fmt.Errorf("initialize Gavel shared database: %w", err)
	}
	if db.Disabled() {
		deps.logInfo(serveDatabaseStartupMessage(db, 0))
		return nil
	}

	mon, err := deps.newMonitor(db.Gorm())
	if err != nil {
		return fmt.Errorf("initialize Captain session monitor: %w", err)
	}
	go func() {
		if err := mon.Run(ctx); err != nil && ctx.Err() == nil {
			logger.Errorf("Captain session monitor stopped: %v", err)
		}
	}()
	select {
	case <-mon.Ready():
	case <-ctx.Done():
		return ctx.Err()
	}
	liveSessions, err := deps.countLiveSessions(ctx, db.Gorm())
	if err != nil {
		return fmt.Errorf("count live Captain sessions: %w", err)
	}
	deps.logInfo(serveDatabaseStartupMessage(db, liveSessions))
	return nil
}
