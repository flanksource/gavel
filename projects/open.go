package projects

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/flanksource/gavel/internal/database"
)

var process struct {
	sync.Mutex
	service  Service
	testPath string
}

// FilePath is the projects.json the file mode reads and the DB mode imports:
// ~/.config/gavel/projects.json, or the file a test pointed the process at.
func FilePath() (string, error) {
	process.Lock()
	defer process.Unlock()
	if process.testPath != "" {
		return process.testPath, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve the gavel projects file: %w", err)
	}
	return filepath.Join(home, ".config", "gavel", "projects.json"), nil
}

// Open picks the catalog mode: the projects table when PostgreSQL is
// available (importing FilePath once), the file otherwise. Pick it once per
// process; Shared does.
func Open(ctx context.Context) (Service, error) {
	path, err := FilePath()
	if err != nil {
		return nil, err
	}
	db, err := database.Require(ctx, "the projects table")
	if errors.Is(err, database.ErrUnavailable) {
		return NewFile(path), nil
	}
	if err != nil {
		return nil, err
	}
	return NewDB(db, path), nil
}

// Shared is the process's catalog, opened by the first caller. A failed open
// is not cached, so a later caller retries it.
func Shared(ctx context.Context) (Service, error) {
	process.Lock()
	service := process.service
	process.Unlock()
	if service != nil {
		return service, nil
	}
	service, err := Open(ctx)
	if err != nil {
		return nil, err
	}
	process.Lock()
	defer process.Unlock()
	if process.service == nil {
		process.service = service
	}
	return process.service, nil
}

// UseFileForTest makes Shared the file-only catalog at path, for another
// package's tests, and returns the restore.
func UseFileForTest(path string) (restore func()) {
	process.Lock()
	defer process.Unlock()
	previousService, previousPath := process.service, process.testPath
	process.service, process.testPath = NewFile(path), path
	return func() {
		process.Lock()
		defer process.Unlock()
		process.service, process.testPath = previousService, previousPath
	}
}
