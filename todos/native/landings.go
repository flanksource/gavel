package native

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

// ErrAlreadyLanded reports a second landing of a run that already has one.
var ErrAlreadyLanded = errors.New("native todo run already landed")

// LandingVia is how a run's commits reached a branch.
type LandingVia string

const (
	// LandingMerge cherry-picked the run's commits onto the checkout's branch.
	LandingMerge LandingVia = "merge"
	// LandingPR opened a pull request carrying the run's commits.
	LandingPR LandingVia = "pr"
)

// RunLanding records how one run step's worktree commits were landed.
type RunLanding struct {
	PromptRunID  uuid.UUID  `gorm:"column:prompt_run_id" json:"promptRunId"`
	Via          LandingVia `gorm:"column:via" json:"via"`
	TargetBranch string     `gorm:"column:target_branch" json:"targetBranch"`
	// LandedSHA is the tip that carries the run's commits: the checkout's HEAD
	// after a merge, the pull request's topic head after a PR.
	LandedSHA string `gorm:"column:landed_sha" json:"landedSha"`
	PRNumber  *int   `gorm:"column:pr_number" json:"prNumber,omitempty"`
	PRURL     string `gorm:"column:pr_url" json:"prUrl,omitempty"`
	// CommitCount is how many commits the landing carried (Setup..Head at land
	// time), so LandedSHA~CommitCount..LandedSHA is the landed work even after
	// the run's branch is deleted.
	CommitCount int `gorm:"column:commit_count" json:"commitCount"`
	// BranchDeletedAt is when the run's worktree branch was deleted, once every
	// commit on it was confirmed landed; nil while the branch is kept.
	BranchDeletedAt *time.Time `gorm:"column:branch_deleted_at" json:"branchDeletedAt,omitempty"`
	LandedAt        time.Time  `gorm:"column:created_at" json:"landedAt"`
}

const runLandingColumns = `prompt_run_id, via, target_branch, landed_sha, pr_number,
	COALESCE(pr_url, '') AS pr_url, commit_count, branch_deleted_at, created_at`

func (l RunLanding) validate() error {
	if l.PromptRunID == uuid.Nil {
		return fmt.Errorf("%w: landing prompt run ID is required", ErrInvalidInput)
	}
	if strings.TrimSpace(l.TargetBranch) == "" || strings.TrimSpace(l.LandedSHA) == "" {
		return fmt.Errorf("%w: landing of run %s needs a target branch and a landed sha", ErrInvalidInput, l.PromptRunID)
	}
	if l.CommitCount < 1 {
		return fmt.Errorf("%w: landing of run %s carried %d commits; it must carry at least one", ErrInvalidInput, l.PromptRunID, l.CommitCount)
	}
	switch l.Via {
	case LandingMerge:
		if l.PRNumber != nil || l.PRURL != "" {
			return fmt.Errorf("%w: a merge landing of run %s carries no pull request", ErrInvalidInput, l.PromptRunID)
		}
	case LandingPR:
		if l.PRNumber == nil || *l.PRNumber <= 0 || strings.TrimSpace(l.PRURL) == "" {
			return fmt.Errorf("%w: a pr landing of run %s needs the pull request number and URL", ErrInvalidInput, l.PromptRunID)
		}
	default:
		return fmt.Errorf("%w: landing via %q is neither %q nor %q", ErrInvalidInput, l.Via, LandingMerge, LandingPR)
	}
	return nil
}

// RecordLanding stores a run's landing and a run_landed event in one
// transaction. The git work it describes has already happened, so it takes the
// issue lock without an expected version: a concurrent edit to the issue must
// not lose the record of commits that already reached a branch.
func (r *Repository) RecordLanding(ctx context.Context, issueID uuid.UUID, landing RunLanding, actor string) (*RunLanding, error) {
	if issueID == uuid.Nil {
		return nil, fmt.Errorf("%w: landing issue ID is required", ErrInvalidInput)
	}
	if err := landing.validate(); err != nil {
		return nil, err
	}
	var recorded RunLanding
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		issue, err := lockExecutionIssue(tx, issueID)
		if err != nil {
			return err
		}
		result := tx.Raw(`
			INSERT INTO todo_run_landings
				(prompt_run_id, issue_id, via, target_branch, landed_sha, pr_number, pr_url, commit_count, branch_deleted_at, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, now())
			RETURNING `+runLandingColumns,
			landing.PromptRunID, issueID, landing.Via, landing.TargetBranch, landing.LandedSHA,
			landing.PRNumber, nullableString(landing.PRURL), landing.CommitCount, landing.BranchDeletedAt,
		).Scan(&recorded)
		if result.Error != nil {
			return mapLandingError(result.Error, issueID, landing.PromptRunID)
		}
		_, err = recordMutation(tx, issue.lockedIssue(), EventInput{
			Kind: "run_landed", Actor: actor, Payload: landingPayload(landing),
		})
		return err
	})
	if err != nil {
		return nil, err
	}
	return &recorded, nil
}

// ListLandings returns the issue's run landings, oldest first.
func (r *Repository) ListLandings(ctx context.Context, issueID uuid.UUID) ([]RunLanding, error) {
	var landings []RunLanding
	result := r.db.WithContext(ctx).Raw(`
		SELECT `+runLandingColumns+`
		FROM todo_run_landings WHERE issue_id = ? ORDER BY created_at, prompt_run_id`, issueID,
	).Scan(&landings)
	return landings, result.Error
}

// ListLandingsForRuns reads the landings of the named prompt runs in one query,
// keyed by run; a run that was never landed has no entry. The todo list diffs
// every issue's latest run, so the per-issue ListLandings would be an N+1.
func (r *Repository) ListLandingsForRuns(ctx context.Context, promptRunIDs []uuid.UUID) (map[uuid.UUID]RunLanding, error) {
	byRun := make(map[uuid.UUID]RunLanding, len(promptRunIDs))
	if len(promptRunIDs) == 0 {
		return byRun, nil
	}
	var landings []RunLanding
	result := r.db.WithContext(ctx).Raw(`
		SELECT `+runLandingColumns+`
		FROM todo_run_landings WHERE prompt_run_id IN ?`, promptRunIDs,
	).Scan(&landings)
	if result.Error != nil {
		return nil, result.Error
	}
	for _, landing := range landings {
		byRun[landing.PromptRunID] = landing
	}
	return byRun, nil
}

func landingPayload(landing RunLanding) map[string]any {
	payload := map[string]any{
		"promptRunId":  landing.PromptRunID,
		"via":          landing.Via,
		"targetBranch": landing.TargetBranch,
		"landedSha":    landing.LandedSHA,
		"commitCount":  landing.CommitCount,
	}
	if landing.PRNumber != nil {
		payload["prNumber"] = *landing.PRNumber
	}
	if landing.PRURL != "" {
		payload["prUrl"] = landing.PRURL
	}
	if landing.BranchDeletedAt != nil {
		payload["branchDeletedAt"] = *landing.BranchDeletedAt
	}
	return payload
}

func mapLandingError(err error, issueID, promptRunID uuid.UUID) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return err
	}
	switch pgErr.Code {
	case "23505":
		return fmt.Errorf("%w: run %s", ErrAlreadyLanded, promptRunID)
	case "23503":
		return fmt.Errorf("%w: prompt run %s is not linked to issue %s", ErrLinkConflict, promptRunID, issueID)
	default:
		return err
	}
}
