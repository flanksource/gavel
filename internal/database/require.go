package database

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// ErrUnavailable identifies commands and APIs that require durable PostgreSQL
// state while the process database is disabled or unconfigured.
var ErrUnavailable = errors.New("gavel database unavailable")

// Require returns the process-owned GORM pool or an actionable unavailable
// error. Cache-only consumers should use Shared directly because they may
// legitimately degrade to an in-memory/no-cache path.
func Require(ctx context.Context, feature string) (*gorm.DB, error) {
	db, err := Shared(ctx)
	if err != nil {
		return nil, err
	}
	if !db.Disabled() {
		return db.Gorm(), nil
	}
	feature = strings.TrimSpace(feature)
	if feature == "" {
		feature = "this operation"
	}
	return nil, fmt.Errorf("%s requires PostgreSQL: %w; set %s or configure embedded PostgreSQL with gavel system install --embedded", feature, ErrUnavailable, EnvDSN)
}
