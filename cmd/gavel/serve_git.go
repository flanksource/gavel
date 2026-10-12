package main

import (
	"context"
	"fmt"
	"time"

	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/commons/logger"
	gavelctx "github.com/flanksource/gavel/context"
	"github.com/flanksource/gavel/git/gitstate"
	"github.com/flanksource/gavel/pr/ui"
	"gorm.io/gorm"
)

const (
	// gitFeedRestartDelay spaces restarts of a git change feed that failed
	// after startup; each restart wakes the dashboard to catch up.
	gitFeedRestartDelay = 5 * time.Second
	// gitActiveSessionsInterval is how often the working directories of live
	// agent sessions are re-read into the tracker's hot rescan set.
	gitActiveSessionsInterval = 10 * time.Second
	// liveSessionPageSize is the page size of that read.
	liveSessionPageSize = 500
)

// startGitTracker starts the process's git state tracker over the shared
// pool; nil, without error, when the process runs without a database.
func startGitTracker(ctx context.Context, db *gorm.DB) (*gitstate.Tracker, error) {
	if db == nil {
		return nil, nil
	}
	tracker := gitstate.NewTracker(gitstate.Options{Store: gitstate.NewStore(db), Context: ctx})
	if err := tracker.Start(); err != nil {
		return nil, fmt.Errorf("start git state tracker: %w", err)
	}
	return tracker, nil
}

// serveContext is the dashboard's root context, carrying the git state
// tracker when there is one; without it the git endpoints answer 503.
func serveContext(ctx context.Context, tracker *gitstate.Tracker) gavelctx.Context {
	if tracker == nil {
		return gavelctx.New(ctx)
	}
	return gavelctx.New(ctx, gavelctx.WithGitTracker(tracker))
}

// startGitChangeFeed pushes git state changes to srv's /api/git/stream through
// a supervised LISTEN on the shared pool, and keeps the tracker's hot set on
// the directories live agent sessions work in. It does nothing without a
// tracker.
func startGitChangeFeed(ctx context.Context, srv *ui.Server, tracker *gitstate.Tracker, db *gorm.DB) error {
	if tracker == nil {
		return nil
	}
	pool, err := db.DB()
	if err != nil {
		return fmt.Errorf("git change feed: %w", err)
	}
	err = startSupervised(ctx, supervisedListener{
		name: "git change feed", stale: "the dashboard's git views refresh only every 30s", restartDelay: gitFeedRestartDelay,
	}, srv.GitChangeListener(pool))
	if err != nil {
		return err
	}
	captain, err := captaindb.Use(db)
	if err != nil {
		return fmt.Errorf("git tracker active sessions: %w", err)
	}
	go refreshActiveSessions(ctx, tracker, captain, gitActiveSessionsInterval)
	return nil
}

// refreshActiveSessions hands the tracker the working directories of live
// agent sessions every interval, so the worktrees agents are editing are
// rescanned at the hot cadence.
func refreshActiveSessions(ctx context.Context, tracker *gitstate.Tracker, captain *captaindb.DB, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		dirs, err := liveSessionDirs(ctx, captain)
		switch {
		case ctx.Err() != nil:
			return
		case err != nil:
			logger.Warnf("git tracker: keeping the previous agent worktrees hot: %v", err)
		default:
			tracker.SetActive(dirs)
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

// liveSessionDirs lists the recorded and process working directories of every
// session with a live process.
func liveSessionDirs(ctx context.Context, captain *captaindb.DB) ([]string, error) {
	var dirs []string
	cursor := ""
	for {
		page, err := captain.ListSessionSummaries(ctx, captaindb.SessionListFilter{LiveOnly: true, Limit: liveSessionPageSize, Cursor: cursor})
		if err != nil {
			return nil, fmt.Errorf("list live agent sessions: %w", err)
		}
		for _, row := range page.Rows {
			for _, dir := range []*string{row.CWD, row.ProcessCWD} {
				if dir != nil && *dir != "" {
					dirs = append(dirs, *dir)
				}
			}
		}
		if page.NextCursor == "" {
			return dirs, nil
		}
		cursor = page.NextCursor
	}
}
