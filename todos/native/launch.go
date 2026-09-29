package native

import (
	"context"
	"errors"
	"fmt"

	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ErrDatabasePoolMismatch = errors.New("captain and native TODO repositories use different database pools")

// PromptRunLaunch combines Captain's authoritative records with the native
// issue after all three were committed by one database transaction.
type PromptRunLaunch struct {
	Session       *captaindb.Session
	PromptRun     *captaindb.PromptRun
	Issue         *Issue
	DispatchOwned bool
}

// LaunchCoordinator is the native-only execution boundary. Only the
// PostgreSQL runtime constructs one.
//
// Captain and Gavel must share the exact underlying *sql.DB pool: Captain owns
// every write to its own tables and runs Gavel's link hooks inside its
// transactions, where Gavel writes only todo_* rows through the same handle.
type LaunchCoordinator struct {
	captain    *captaindb.DB
	repository *Repository
}

func NewLaunchCoordinator(captain *captaindb.DB, repository *Repository) (*LaunchCoordinator, error) {
	if captain == nil || captain.Gorm() == nil {
		return nil, fmt.Errorf("%w: Captain database is nil", ErrInvalidInput)
	}
	if repository == nil || repository.db == nil {
		return nil, fmt.Errorf("%w: native repository is nil", ErrInvalidInput)
	}
	captainSQL, captainErr := captain.Gorm().DB()
	repositorySQL, repositoryErr := repository.db.DB()
	if captainErr != nil || repositoryErr != nil || captainSQL != repositorySQL {
		return nil, ErrDatabasePoolMismatch
	}
	return &LaunchCoordinator{captain: captain, repository: repository}, nil
}

// AttachPromptRun is the host link of Captain's prompt-run admission: it runs
// inside the admission transaction tx, after Captain created (or idempotently
// resolved) the run and its session tree, and attaches and activates that run
// on exactly one native issue. It writes only todo_* rows; returning an error
// rolls Captain's admission back and nothing is dispatched.
//
// attachment.PromptRunID may be left unset: the admitted run is authoritative,
// and a different ID is rejected rather than silently replaced.
func (c *LaunchCoordinator) AttachPromptRun(
	ctx context.Context,
	tx *captaindb.DB,
	run *captaindb.PromptRun,
	attachment PromptRunAttachment,
) (*PromptRunLaunch, error) {
	if run == nil || run.ID == uuid.Nil {
		return nil, fmt.Errorf("%w: an admitted Captain prompt run is required", ErrInvalidInput)
	}
	if !inCaptainTransaction(tx) {
		return nil, fmt.Errorf("%w: prompt run %s must be attached inside Captain's admission transaction", ErrInvalidInput, run.ID)
	}
	if attachment.PromptRunID != uuid.Nil && attachment.PromptRunID != run.ID {
		return nil, fmt.Errorf("%w: attachment names prompt run %s but Captain admitted %s", ErrInvalidInput, attachment.PromptRunID, run.ID)
	}
	attachment.PromptRunID = run.ID
	repositoryTx, err := NewRepository(tx.Gorm())
	if err != nil {
		return nil, err
	}
	integration, err := NewExecutionIntegration(repositoryTx)
	if err != nil {
		return nil, err
	}
	issue, dispatchOwned, err := integration.activatePromptRun(ctx, attachment)
	if err != nil {
		return nil, err
	}
	session, err := tx.GetSession(ctx, run.SessionID)
	if err != nil {
		return nil, fmt.Errorf("read admitted session %s of prompt run %s: %w", run.SessionID, run.ID, err)
	}
	return &PromptRunLaunch{Session: session, PromptRun: run, Issue: issue, DispatchOwned: dispatchOwned}, nil
}

// inCaptainTransaction mirrors GORM's own nested-transaction check: a handle
// scoped to a transaction runs on a *sql.Tx, which commits.
func inCaptainTransaction(tx *captaindb.DB) bool {
	if tx == nil || tx.Gorm() == nil || tx.Gorm().Statement == nil {
		return false
	}
	committer, ok := tx.Gorm().Statement.ConnPool.(gorm.TxCommitter)
	return ok && committer != nil
}
