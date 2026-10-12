package runtime

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"gorm.io/gorm"
)

// ErrSchemaBehind is the one message a todo command gives when the database it
// opened predates this binary.
//
// It exists because todo commands open the database WITHOUT migrating it — only
// `gavel serve` applies migrations — so a binary built after a schema change
// will happily connect to a database that has never seen it. Every read of the
// missing column then fails somewhere deep, one obscure error per call site,
// instead of once, here, with the command that fixes it.
const ErrSchemaBehind = "database schema is behind this binary; run `gavel serve` once to migrate"

// requiredColumns are the newest columns the todo runtime reads, one per
// migration bundle: Captain's and Gavel's are applied separately, so a database
// can have either without the other. A bundle that has its newest column has
// everything before it.
var requiredColumns = []string{
	"captain_prompt_run_iterations.verification_result",
	"todo_issues.parent_issue_id",
}

// requireCurrentSchema is the whole schema-drift guard: it fails with
// ErrSchemaBehind, naming what is missing, when the database lacks a column
// every issue read now selects.
func requireCurrentSchema(ctx context.Context, db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("native TODO storage: database is nil")
	}
	var present []string
	err := db.WithContext(ctx).Raw(`
		SELECT table_name || '.' || column_name
		FROM information_schema.columns
		WHERE table_schema = 'public'
		  AND table_name || '.' || column_name IN ?`, requiredColumns).Scan(&present).Error
	if err != nil {
		return fmt.Errorf("check native TODO schema: %w", err)
	}
	var missing []string
	for _, column := range requiredColumns {
		if !slices.Contains(present, column) {
			missing = append(missing, column)
		}
	}
	if len(missing) > 0 {
		verb := "is"
		if len(missing) > 1 {
			verb = "are"
		}
		return fmt.Errorf("%s (%s %s missing)", ErrSchemaBehind, strings.Join(missing, ", "), verb)
	}
	return nil
}
