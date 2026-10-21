// Deletes run data older than its window across every organization: job logs, test and lint
// runs, archived trace stores, linter executions and sandbox runs, each blob before the row naming it.
package retention

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"time"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/blob"
	"github.com/flanksource/gavel/internal/safefs"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// batchSize is how many rows each query selects or deletes at once.
const batchSize = 500

const defaultWindow = 60 * 24 * time.Hour

// Stores are where the deleted rows' blobs live.
type Stores struct {
	// Org is an organization's part of the blob store, holding job and sandbox logs; it may
	// return nil when there is none.
	Org func(org uuid.UUID) blob.Store
	// Workspace is the store holding the run snapshots of org's workspace workDir, as named in
	// test_run_caches.workspace_dir.
	Workspace func(org uuid.UUID, workDir string) (blob.Store, error)
	// Traces is the configured blob store, holding the trace stores runs archived under
	// workspaces/*/traces/<run>/; nil when trace stores are only on local disk.
	Traces blob.Store
	// Transcripts is the directory holding the transcripts agent runs uploaded, under
	// <org>/<run>/; empty when there is none.
	Transcripts string
}

// Report counts what one Run deleted, per category.
type Report struct {
	JobLogs          int
	TestRuns         int
	LintRuns         int
	Traces           int
	LinterExecutions int
	SandboxRuns      int
}

// Run deletes everything older than its retention.<category> window, measured back from now.
func Run(ctx context.Context, db *gorm.DB, stores Stores, now time.Time) (Report, error) {
	var report Report
	windows := map[string]time.Duration{}
	for _, key := range []string{"job_logs", "test_runs", "lint_runs", "traces", "linter_executions", "sandbox_runs"} {
		window, err := Window("retention."+key, defaultWindow)
		if err != nil {
			return report, err
		}
		windows[key] = window
	}
	cutoff := func(key string) time.Time { return now.Add(-windows[key]) }
	db = db.WithContext(ctx)

	var err error
	if report.JobLogs, err = deleteJobLogs(ctx, db, stores, cutoff("job_logs")); err != nil {
		return report, err
	}
	if report.TestRuns, err = deleteRuns(ctx, db, stores, []string{"test", "test+lint"}, cutoff("test_runs")); err != nil {
		return report, err
	}
	if report.LintRuns, err = deleteRuns(ctx, db, stores, []string{"lint"}, cutoff("lint_runs")); err != nil {
		return report, err
	}
	if report.Traces, err = deleteTraces(ctx, stores.Traces, cutoff("traces")); err != nil {
		return report, err
	}
	if report.LinterExecutions, err = deleteLinterExecutions(db, cutoff("linter_executions")); err != nil {
		return report, err
	}
	if report.SandboxRuns, err = deleteSandboxRuns(ctx, db, stores, cutoff("sandbox_runs")); err != nil {
		return report, err
	}
	logger.Infof("retention: deleted %d job logs, %d test runs, %d lint runs, %d trace stores, %d linter executions, %d sandbox runs",
		report.JobLogs, report.TestRuns, report.LintRuns, report.Traces, report.LinterExecutions, report.SandboxRuns)
	return report, nil
}

// deleteOrgBlob deletes key from the organization's store; a row with no key has no blob.
func deleteOrgBlob(ctx context.Context, stores Stores, org uuid.UUID, key, what string) error {
	if key == "" {
		return nil
	}
	var store blob.Store
	if stores.Org != nil {
		store = stores.Org(org)
	}
	if store == nil {
		return fmt.Errorf("delete %s %s: no blob store for organization %s", what, key, org)
	}
	if err := store.Delete(ctx, key); err != nil && !errors.Is(err, blob.ErrNotFound) {
		return fmt.Errorf("delete %s %s of organization %s: %w", what, key, org, err)
	}
	return nil
}

func deleteJobLogs(ctx context.Context, db *gorm.DB, stores Stores, cutoff time.Time) (int, error) {
	deleted := 0
	for {
		var rows []struct {
			OrgID  uuid.UUID
			JobID  int64
			LogKey string
		}
		if err := db.Raw(`SELECT org_id, job_id, log_key FROM job_log_caches WHERE fetched_at < ? ORDER BY org_id, job_id LIMIT ?`,
			cutoff, batchSize).Scan(&rows).Error; err != nil {
			return deleted, fmt.Errorf("list job logs older than %s: %w", cutoff.Format(time.RFC3339), err)
		}
		for _, row := range rows {
			if err := deleteOrgBlob(ctx, stores, row.OrgID, row.LogKey, "job log"); err != nil {
				return deleted, err
			}
			if err := db.Exec(`DELETE FROM job_log_caches WHERE org_id = ? AND job_id = ?`, row.OrgID, row.JobID).Error; err != nil {
				return deleted, fmt.Errorf("delete job log %d of organization %s: %w", row.JobID, row.OrgID, err)
			}
			deleted++
		}
		if len(rows) < batchSize {
			return deleted, nil
		}
	}
}

// deleteRuns deletes test_run_caches rows of kinds started before cutoff, and the run-*
// snapshots they name. A sha-* row names a snapshot last.json may point at, so only its row goes.
func deleteRuns(ctx context.Context, db *gorm.DB, stores Stores, kinds []string, cutoff time.Time) (int, error) {
	workspaces := map[string]blob.Store{}
	deleted := 0
	for {
		var rows []struct {
			OrgID        uuid.UUID
			WorkspaceDir string
			RunID        string
		}
		if err := db.Raw(`SELECT org_id, workspace_dir, run_id FROM test_run_caches WHERE kind IN ? AND started_ts < ? ORDER BY org_id, workspace_dir, run_id LIMIT ?`,
			kinds, cutoff.UnixNano(), batchSize).Scan(&rows).Error; err != nil {
			return deleted, fmt.Errorf("list %s runs older than %s: %w", strings.Join(kinds, "/"), cutoff.Format(time.RFC3339), err)
		}
		for _, row := range rows {
			if strings.HasPrefix(row.RunID, "run-") {
				store, err := workspaceStore(stores, workspaces, row.OrgID, row.WorkspaceDir)
				if err != nil {
					return deleted, err
				}
				if err := store.Delete(ctx, row.RunID+".json"); err != nil && !errors.Is(err, blob.ErrNotFound) {
					return deleted, fmt.Errorf("delete run snapshot %s of %s: %w", row.RunID, row.WorkspaceDir, err)
				}
			}
			if err := db.Exec(`DELETE FROM test_run_caches WHERE org_id = ? AND workspace_dir = ? AND run_id = ?`,
				row.OrgID, row.WorkspaceDir, row.RunID).Error; err != nil {
				return deleted, fmt.Errorf("delete run %s of %s: %w", row.RunID, row.WorkspaceDir, err)
			}
			deleted++
		}
		if len(rows) < batchSize {
			return deleted, nil
		}
	}
}

func workspaceStore(stores Stores, opened map[string]blob.Store, org uuid.UUID, workDir string) (blob.Store, error) {
	cacheKey := org.String() + "\x00" + workDir
	if store, ok := opened[cacheKey]; ok {
		return store, nil
	}
	if stores.Workspace == nil {
		return nil, fmt.Errorf("delete runs of %s: no snapshot store for workspaces", workDir)
	}
	store, err := stores.Workspace(org, workDir)
	if err != nil {
		return nil, fmt.Errorf("open the snapshot store of %s: %w", workDir, err)
	}
	opened[cacheKey] = store
	return store, nil
}

// deleteTraces deletes the archived trace stores whose files were all last modified before
// cutoff. A store is a run's directory of files, so it goes whole or not at all.
func deleteTraces(ctx context.Context, store blob.Store, cutoff time.Time) (int, error) {
	if store == nil {
		return 0, nil
	}
	objects, err := store.List(ctx, "workspaces/")
	if err != nil {
		return 0, fmt.Errorf("list trace stores: %w", err)
	}
	runs := map[string][]string{}
	var fresh []string
	for _, object := range objects {
		run, ok := traceRun(object.Key)
		if !ok {
			continue
		}
		runs[run] = append(runs[run], object.Key)
		if !object.ModTime.Before(cutoff) {
			fresh = append(fresh, run)
		}
	}
	for _, run := range fresh {
		delete(runs, run)
	}
	for run, keys := range runs {
		for _, key := range keys {
			if err := store.Delete(ctx, key); err != nil && !errors.Is(err, blob.ErrNotFound) {
				return 0, fmt.Errorf("delete trace store %s: %w", run, err)
			}
		}
	}
	return len(runs), nil
}

// traceRun is the run directory, workspaces/<workspace>/traces/<run>/, that key is a file of.
func traceRun(key string) (string, bool) {
	parts := strings.SplitN(key, "/", 5)
	if len(parts) < 5 || parts[0] != "workspaces" || parts[2] != "traces" {
		return "", false
	}
	return strings.Join(parts[:4], "/") + "/", true
}

func deleteLinterExecutions(db *gorm.DB, cutoff time.Time) (int, error) {
	deleted := 0
	for {
		result := db.Exec(`DELETE FROM linter_executions WHERE id IN (SELECT id FROM linter_executions WHERE executed_at < ? LIMIT ?)`, cutoff, batchSize)
		if result.Error != nil {
			return deleted, fmt.Errorf("delete linter executions older than %s: %w", cutoff.Format(time.RFC3339), result.Error)
		}
		deleted += int(result.RowsAffected)
		if result.RowsAffected < batchSize {
			return deleted, nil
		}
	}
}

func deleteSandboxRuns(ctx context.Context, db *gorm.DB, stores Stores, cutoff time.Time) (int, error) {
	deleted := 0
	for {
		var rows []struct {
			ID     uuid.UUID
			OrgID  uuid.UUID
			LogKey string
		}
		if err := db.Raw(`SELECT id, org_id, log_key FROM sandbox_runs WHERE finished_at < ? ORDER BY id LIMIT ?`,
			cutoff, batchSize).Scan(&rows).Error; err != nil {
			return deleted, fmt.Errorf("list sandbox runs finished before %s: %w", cutoff.Format(time.RFC3339), err)
		}
		for _, row := range rows {
			if err := deleteOrgBlob(ctx, stores, row.OrgID, row.LogKey, "sandbox log"); err != nil {
				return deleted, err
			}
			if stores.Transcripts != "" {
				dir := filepath.Join(stores.Transcripts, row.OrgID.String(), row.ID.String())
				// A run that uploaded no transcripts has no directory, nor maybe its org.
				err := safefs.SafeRemove(dir, safefs.RemoveContext{AllowedRoot: stores.Transcripts})
				if err != nil && !errors.Is(err, fs.ErrNotExist) {
					return deleted, fmt.Errorf("delete the transcripts of sandbox run %s: %w", row.ID, err)
				}
			}
			if err := db.Exec(`DELETE FROM sandbox_runs WHERE id = ?`, row.ID).Error; err != nil {
				return deleted, fmt.Errorf("delete sandbox run %s: %w", row.ID, err)
			}
			deleted++
		}
		if len(rows) < batchSize {
			return deleted, nil
		}
	}
}
