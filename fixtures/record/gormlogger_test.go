package record

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// fakeGormLogger is a delegate that records what it was asked to do, so the
// wrapper can be shown to pass everything through rather than replace it.
type fakeGormLogger struct {
	traces  int
	level   logger.LogLevel
	filters int
}

func (f *fakeGormLogger) LogMode(level logger.LogLevel) logger.Interface {
	f.level = level
	return f
}
func (f *fakeGormLogger) Info(context.Context, string, ...any)  {}
func (f *fakeGormLogger) Warn(context.Context, string, ...any)  {}
func (f *fakeGormLogger) Error(context.Context, string, ...any) {}
func (f *fakeGormLogger) Trace(context.Context, time.Time, func() (string, int64), error) {
	f.traces++
}

// ParamsFilter is how commons-db's logger keeps bind values out of logged SQL.
func (f *fakeGormLogger) ParamsFilter(_ context.Context, sql string, _ ...any) (string, []any) {
	f.filters++
	return sql, nil
}

func trace(wrapped logger.Interface, sql string, rows int64, err error) {
	wrapped.Trace(context.Background(), time.Now().Add(-5*time.Millisecond),
		func() (string, int64) { return sql, rows }, err)
}

func TestWrapGormRecordsNothingUntilAFixtureAsks(t *testing.T) {
	delegate := &fakeGormLogger{}
	wrapped := WrapGorm(delegate)

	trace(wrapped, "SELECT * FROM users", 1, nil)

	assert.Equal(t, 1, delegate.traces, "the delegate still logs")
	assert.Nil(t, inProcess.Load(), "no sink is attached until StartInProcess")
}

func TestWrapGormRecordsWhileTheLogIsAttached(t *testing.T) {
	delegate := &fakeGormLogger{}
	wrapped := WrapGorm(delegate)

	log := StartInProcess()
	t.Cleanup(StopInProcess)

	trace(wrapped, "INSERT INTO users (id) VALUES (1)", 1, nil)
	trace(wrapped, "SELECT * FROM missing", 0, errors.New(`relation "missing" does not exist`))

	statements := log.Statements()
	require.Len(t, statements, 2)

	assert.Equal(t, "INSERT", statements[0].Op)
	assert.Equal(t, []string{"users"}, statements[0].Tables)
	assert.Equal(t, 1, statements[0].Rows)
	assert.GreaterOrEqual(t, statements[0].DurationMs, int64(5), "duration is measured from gorm's begin")

	assert.Contains(t, statements[1].Error, `relation "missing" does not exist`)
}

func TestStopInProcessDetachesButKeepsWhatItCollected(t *testing.T) {
	delegate := &fakeGormLogger{}
	wrapped := WrapGorm(delegate)

	log := StartInProcess()
	trace(wrapped, "SELECT 1", 1, nil)
	StopInProcess()
	trace(wrapped, "SELECT 2", 1, nil)

	statements := log.Statements()
	require.Len(t, statements, 1, "only what was traced while attached")
	assert.Equal(t, "SELECT 1", statements[0].SQL)
	assert.Equal(t, 2, delegate.traces, "the delegate keeps logging either way")
}

// gorm resolves ParamsFilter on the concrete logger, so an embedded delegate is
// invisible to it — without an explicit forward, every bind value silently gets
// inlined into the logged SQL.
func TestWrapGormForwardsParamsFilter(t *testing.T) {
	delegate := &fakeGormLogger{}
	wrapped := WrapGorm(delegate)

	filter, ok := wrapped.(gorm.ParamsFilter)
	require.True(t, ok, "the wrapper must satisfy gorm.ParamsFilter")

	sql, params := filter.ParamsFilter(context.Background(), "SELECT $1", "hunter2")
	assert.Equal(t, "SELECT $1", sql)
	assert.Nil(t, params, "the delegate's decision to strip bind values survives wrapping")
	assert.Equal(t, 1, delegate.filters)
}

// LogMode returns a new logger, and gorm calls it during Open. A bare delegate
// coming back would drop the recording the first time the level changed.
func TestLogModeKeepsRecording(t *testing.T) {
	delegate := &fakeGormLogger{}
	wrapped := WrapGorm(delegate).LogMode(logger.Info)

	log := StartInProcess()
	t.Cleanup(StopInProcess)
	trace(wrapped, "SELECT 1", 1, nil)

	assert.Equal(t, logger.Info, delegate.level)
	assert.Len(t, log.Statements(), 1)
}

func TestStatementLogWindowsToOneFixture(t *testing.T) {
	log := &StatementLog{}
	base := time.Date(2026, 8, 2, 12, 0, 0, 0, time.UTC)
	log.Add(Statement{started: base, SQL: "before"})
	log.Add(Statement{started: base.Add(2 * time.Second), SQL: "inside"})

	window := log.Between(base.Add(time.Second), base.Add(5*time.Second))

	require.Len(t, window, 1)
	assert.Equal(t, "inside", window[0].SQL)
}
