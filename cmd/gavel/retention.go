// Runs retention from gavel serve: loads the properties file that sets the retention windows, and
// deletes expired run data at startup and every retention.interval as a "Retention cleanup" task.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	clickytask "github.com/flanksource/clicky/task"
	commonscontext "github.com/flanksource/commons/context"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/commons/properties"
	"github.com/flanksource/gavel/blob"
	ghcache "github.com/flanksource/gavel/github/cache"
	"github.com/flanksource/gavel/internal/retention"
	"github.com/flanksource/gavel/pr/ui"
	"github.com/flanksource/gavel/snapshots"
	"github.com/flanksource/gavel/verify"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// defaultPropertiesFile is loaded when --properties-file is unset and it exists.
var defaultPropertiesFile = "/etc/gavel/gavel.properties"

// loadServeProperties loads path, or defaultPropertiesFile when path is empty and that file exists.
func loadServeProperties(path string) error {
	if path == "" {
		if _, err := os.Stat(defaultPropertiesFile); errors.Is(err, os.ErrNotExist) {
			return nil
		}
		path = defaultPropertiesFile
	} else if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("--properties-file %s: %w", path, err)
	}
	if err := properties.Global.LoadFile(path); err != nil {
		return fmt.Errorf("load properties from %s: %w", path, err)
	}
	return nil
}

// startRetention runs a retention pass now and then every retention.interval (1h by default)
// until ctx ends. cache supplies each organization's job logs; nil leaves them alone.
func startRetention(ctx context.Context, db *gorm.DB, cache *ghcache.Store) error {
	interval, err := retention.Window("retention.interval", time.Hour)
	if err != nil {
		return err
	}
	recordings, err := verify.OpenGlobalStorage(ctx)
	if err != nil {
		return err
	}
	stores := retention.Stores{Recordings: recordings}
	if cache != nil {
		stores.Org = cache.OrgBlobs
	}
	stores.Workspace = retentionWorkspaces(stores.Org)
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			runRetentionPass(db, stores)
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}()
	return nil
}

func runRetentionPass(db *gorm.DB, stores retention.Stores) {
	task := clickytask.StartTask("Retention cleanup", func(ctx commonscontext.Context, _ *clickytask.Task) (retention.Report, error) {
		return retention.Run(ctx, db, stores, time.Now())
	})
	if _, err := task.GetResult(); err != nil {
		logger.Warnf("retention cleanup: %v", err)
	}
}

// retentionWorkspaces maps a test_run_caches workspace to the store of its run snapshots: a
// repository's uploaded runs (repo:<id>) are in its organization's store, a local workspace's in
// its snapshot store.
func retentionWorkspaces(orgBlobs func(uuid.UUID) blob.Store) func(uuid.UUID, string) (blob.Store, error) {
	return func(org uuid.UUID, workDir string) (blob.Store, error) {
		id, ok := strings.CutPrefix(workDir, "repo:")
		if !ok {
			return snapshots.OpenStore(workDir)
		}
		repo, err := uuid.Parse(id)
		if err != nil {
			return nil, fmt.Errorf("workspace %s does not name a repository: %w", workDir, err)
		}
		var store blob.Store
		if orgBlobs != nil {
			store = orgBlobs(org)
		}
		if store == nil {
			return nil, fmt.Errorf("workspace %s: no blob store for organization %s", workDir, org)
		}
		return blob.WithPrefix(store, ui.RepoRunsPrefix(repo)), nil
	}
}
