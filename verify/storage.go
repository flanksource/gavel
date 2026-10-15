// Opens the blob store the serve host's ~/.gavel.yaml configures: for the whole host, or for one
// workspace under its own prefix.
package verify

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/flanksource/gavel/blob"
)

// OpenGlobalStorage opens the storage section of ~/.gavel.yaml, or returns nil when it has none,
// so callers keep their usual local directories.
func OpenGlobalStorage(ctx context.Context) (blob.Store, error) {
	config, err := LoadGlobalGavelConfig()
	if err != nil {
		return nil, err
	}
	if config.Storage.IsZero() {
		return nil, nil
	}
	store, err := blob.Open(ctx, config.Storage)
	if err != nil {
		home, _ := os.UserHomeDir()
		return nil, fmt.Errorf("open the storage in %s: %w", filepath.Join(home, GavelConfigFileName), err)
	}
	return store, nil
}

// OpenWorkspaceStorage is OpenGlobalStorage's store under workDir's prefix (see
// blob.WorkspacePrefix), or nil when ~/.gavel.yaml configures none.
func OpenWorkspaceStorage(ctx context.Context, workDir string) (blob.Store, error) {
	store, err := OpenGlobalStorage(ctx)
	if err != nil || store == nil {
		return nil, err
	}
	prefix, err := blob.WorkspacePrefix(workDir)
	if err != nil {
		return nil, err
	}
	return blob.WithPrefix(store, prefix), nil
}
