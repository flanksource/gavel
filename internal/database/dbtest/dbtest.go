// Package dbtest asserts PostgreSQL failures that integration tests trigger on purpose,
// keeping the expected error out of the SQL log while still checking its constraint.
package dbtest

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// Quiet returns db without SQL error logging, for statements a test expects to fail.
// Pair it with RequireConstraintViolation so the expected failure is still asserted.
func Quiet(db *gorm.DB) *gorm.DB {
	return db.Session(&gorm.Session{Logger: gormlogger.Discard})
}

// RequireConstraintViolation fails the test unless err is a PostgreSQL error raised by constraint.
func RequireConstraintViolation(t testing.TB, err error, constraint string) {
	t.Helper()
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		t.Fatalf("expected a violation of %s, got %v", constraint, err)
	}
	if pgErr.ConstraintName != constraint {
		t.Fatalf("expected a violation of %s, got %s: %s", constraint, pgErr.ConstraintName, pgErr.Message)
	}
}
