package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/flanksource/captain/pkg/monitor"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type fakeServeDatabase struct {
	disabled bool
	gorm     *gorm.DB
	dsn      string
	source   string
}

func (d fakeServeDatabase) Disabled() bool    { return d.disabled }
func (d fakeServeDatabase) Gorm() *gorm.DB    { return d.gorm }
func (d fakeServeDatabase) DSN() string       { return d.dsn }
func (d fakeServeDatabase) DSNSource() string { return d.source }

type fakeServeMonitor struct {
	started chan context.Context
	stopped chan struct{}
	ready   chan struct{}
	stats   monitor.IngestStats
}

func (m *fakeServeMonitor) Run(ctx context.Context) error {
	m.started <- ctx
	close(m.ready)
	<-ctx.Done()
	close(m.stopped)
	return nil
}

func (m *fakeServeMonitor) Ready() <-chan struct{} { return m.ready }

func (m *fakeServeMonitor) IngestStats() monitor.IngestStats { return m.stats }

func TestStartServeRuntimeStartsMonitorOnSharedPool(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	pool := &gorm.DB{}
	mon := &fakeServeMonitor{
		started: make(chan context.Context, 1), stopped: make(chan struct{}), ready: make(chan struct{}),
		stats: monitor.IngestStats{FilesIngested: 3, MessagesParsed: 900, MessagesOffered: 9},
	}
	var monitorPool *gorm.DB
	var countPool *gorm.DB
	var openedMode serveDatabaseMode
	var logs []string

	ingestStats, err := startServeRuntime(ctx, serveRuntimeDependencies{
		openDatabase: func(_ context.Context, mode serveDatabaseMode) (serveDatabase, error) {
			openedMode = mode
			return fakeServeDatabase{gorm: pool, dsn: "postgres://captain:secret@db.internal/gavel", source: "--db-url"}, nil
		},
		newMonitor: func(gormDB *gorm.DB) (serveSessionMonitor, error) {
			monitorPool = gormDB
			return mon, nil
		},
		countLiveSessions: func(_ context.Context, gormDB *gorm.DB) (int64, error) {
			countPool = gormDB
			return 7, nil
		},
		logInfo: func(message string) { logs = append(logs, message) },
	}, serveDatabaseWithMigrations)
	require.NoError(t, err)
	require.Equal(t, serveDatabaseWithMigrations, openedMode)
	require.Same(t, pool, monitorPool)
	require.Same(t, pool, countPool)
	require.Equal(t, []string{`Database Info: source="--db-url" dsn="postgres://captain:REDACTED@db.internal/gavel" live_sessions=7`}, logs)
	// The dashboard's own profiler is useless for the monitor it embeds unless
	// the monitor's counters come back out with it.
	require.NotNil(t, ingestStats)
	require.Equal(t, mon.stats, ingestStats())
	select {
	case startedCtx := <-mon.started:
		require.Same(t, ctx, startedCtx)
	case <-time.After(time.Second):
		t.Fatal("session monitor did not start")
	}

	cancel()
	select {
	case <-mon.stopped:
	case <-time.After(time.Second):
		t.Fatal("session monitor did not stop with the serve context")
	}
}

func TestStartServeRuntimeSkipsMonitorWhenDatabaseDisabled(t *testing.T) {
	monitorCalled := false
	countCalled := false
	var logs []string
	ingestStats, err := startServeRuntime(t.Context(), serveRuntimeDependencies{
		openDatabase: func(context.Context, serveDatabaseMode) (serveDatabase, error) {
			return fakeServeDatabase{disabled: true}, nil
		},
		newMonitor: func(*gorm.DB) (serveSessionMonitor, error) {
			monitorCalled = true
			return nil, nil
		},
		countLiveSessions: func(context.Context, *gorm.DB) (int64, error) {
			countCalled = true
			return 0, nil
		},
		logInfo: func(message string) { logs = append(logs, message) },
	}, serveDatabaseNoMigrations)
	require.NoError(t, err)
	require.Nil(t, ingestStats, "there are no ingest counters without a monitor to keep them")
	require.False(t, monitorCalled)
	require.False(t, countCalled)
	require.Equal(t, []string{`Database Info: source="disabled" dsn="" live_sessions=0`}, logs)
}

func TestStartServeRuntimeSurfacesInitializationErrors(t *testing.T) {
	t.Run("database", func(t *testing.T) {
		_, err := startServeRuntime(t.Context(), serveRuntimeDependencies{
			openDatabase: func(context.Context, serveDatabaseMode) (serveDatabase, error) {
				return nil, errors.New("database unavailable")
			},
		}, serveDatabaseNoMigrations)
		require.EqualError(t, err, "initialize Gavel shared database: database unavailable")
	})

	t.Run("monitor", func(t *testing.T) {
		_, err := startServeRuntime(t.Context(), serveRuntimeDependencies{
			openDatabase: func(context.Context, serveDatabaseMode) (serveDatabase, error) {
				return fakeServeDatabase{gorm: &gorm.DB{}}, nil
			},
			newMonitor: func(*gorm.DB) (serveSessionMonitor, error) {
				return nil, errors.New("monitor unavailable")
			},
			logInfo: func(string) {},
		}, serveDatabaseNoMigrations)
		require.EqualError(t, err, "initialize Captain session monitor: monitor unavailable")
	})
}

func TestStartServeRuntimeSurfacesLiveSessionCountError(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	mon := &fakeServeMonitor{started: make(chan context.Context, 1), stopped: make(chan struct{}), ready: make(chan struct{})}

	_, err := startServeRuntime(ctx, serveRuntimeDependencies{
		openDatabase: func(context.Context, serveDatabaseMode) (serveDatabase, error) {
			return fakeServeDatabase{gorm: &gorm.DB{}, dsn: "postgres://db/gavel", source: "test"}, nil
		},
		newMonitor: func(*gorm.DB) (serveSessionMonitor, error) { return mon, nil },
		countLiveSessions: func(context.Context, *gorm.DB) (int64, error) {
			return 0, errors.New("view unavailable")
		},
		logInfo: func(string) { t.Fatal("startup info must not be logged after a count error") },
	}, serveDatabaseNoMigrations)
	require.EqualError(t, err, "count live Captain sessions: view unavailable")
}
