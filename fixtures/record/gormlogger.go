package record

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// inProcess is where WrapGorm sends what it sees. It is a process-global
// because gavel's gorm handle is built once, long before any fixture declares
// `record: sql`, and threading a sink down to it would mean re-opening the
// database per fixture.
//
// Nil until a fixture asks: an unset pointer is one atomic load per query, and
// the delegate logger runs exactly as it did before.
var inProcess atomic.Pointer[StatementLog]

// StatementLog collects statements observed inside this process.
type StatementLog struct {
	mu         sync.Mutex
	statements []Statement
	truncated  bool
}

// Add records one statement, dropping anything past the cap and marking the log
// truncated so the artifact says so rather than quietly ending early.
func (l *StatementLog) Add(statement Statement) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if len(l.statements) >= maxStatements {
		l.truncated = true
		return
	}
	l.statements = append(l.statements, statement)
}

// Statements returns everything recorded so far, oldest first.
func (l *StatementLog) Statements() []Statement {
	l.mu.Lock()
	defer l.mu.Unlock()
	return append([]Statement(nil), l.statements...)
}

// Between narrows to one fixture's window.
func (l *StatementLog) Between(from, to time.Time) []Statement {
	return StatementsBetween(l.Statements(), from, to)
}

// Truncated reports whether the cap was reached.
func (l *StatementLog) Truncated() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.truncated
}

// Save writes the log as a JSONL artifact.
func (l *StatementLog) Save(store *Store, label string, statements []Statement) (Result, error) {
	result, err := SaveStatements(store, label, statements)
	result.Truncated = l.Truncated()
	return result, err
}

// StartInProcess begins recording gavel's own database traffic and returns the
// log it accumulates in. Only one is active at a time — the sink is global, and
// two concurrent in-process recorders would each see the other's queries — so a
// second call replaces the first.
func StartInProcess() *StatementLog {
	log := &StatementLog{}
	inProcess.Store(log)
	return log
}

// StopInProcess detaches the sink. The log keeps what it collected.
func StopInProcess() { inProcess.Store(nil) }

// WrapGorm returns a logger that records every statement gorm traces, on top of
// whatever the delegate already does with it. It is always safe to call — the
// recording is decided per query by whether a fixture has started one — so the
// database can be wired once at startup.
//
// gorm's logger.Interface has exactly one method through which Create, Query,
// Raw and Row all funnel, which is why this is a logger rather than the twelve
// callback registrations the gorm.Plugin route would need for the same reach.
func WrapGorm(delegate logger.Interface) logger.Interface {
	return &gormLogger{Interface: delegate}
}

type gormLogger struct {
	logger.Interface
}

// LogMode is overridden because the embedded delegate's own LogMode returns a
// bare delegate, which would silently drop the recording the first time gorm
// changed level.
func (l *gormLogger) LogMode(level logger.LogLevel) logger.Interface {
	return &gormLogger{Interface: l.Interface.LogMode(level)}
}

// ParamsFilter forwards to the delegate. gorm looks this interface up on the
// concrete logger, so an embedded implementation is invisible to it — without
// this method the delegate's decision to strip bind parameters is silently
// reversed and every value gets inlined into the logged SQL.
func (l *gormLogger) ParamsFilter(ctx context.Context, sql string, params ...any) (string, []any) {
	if filter, ok := l.Interface.(gorm.ParamsFilter); ok {
		return filter.ParamsFilter(ctx, sql, params...)
	}
	return sql, params
}

func (l *gormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	l.Interface.Trace(ctx, begin, fc, err)

	log := inProcess.Load()
	if log == nil {
		return
	}

	sql, rows := fc()
	op, tables := classify(sql)
	statement := Statement{
		started:    begin,
		SQL:        sql,
		Op:         op,
		Tables:     tables,
		Rows:       int(rows),
		DurationMs: time.Since(begin).Milliseconds(),
	}
	if err != nil {
		statement.Error = err.Error()
	}
	log.Add(statement)
}
