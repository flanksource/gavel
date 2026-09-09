package runtime

import (
	"fmt"

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

// requireVerificationColumn asserts the column a verification verdict is read
// from exists. It is the whole schema-drift guard: verification_result is the
// newest column the todo runtime depends on, so a database that has it has
// everything before it, and a database that does not needs the same one fix
// whatever else it is missing.
func requireVerificationColumn(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("native TODO storage: database is nil")
	}
	var present bool
	err := db.Raw(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_schema = 'public'
			  AND table_name = 'captain_prompt_run_iterations'
			  AND column_name = 'verification_result'
		)`).Scan(&present).Error
	if err != nil {
		return fmt.Errorf("check Captain verification schema: %w", err)
	}
	if !present {
		return fmt.Errorf("%s (captain_prompt_run_iterations.verification_result is missing)", ErrSchemaBehind)
	}
	return nil
}
