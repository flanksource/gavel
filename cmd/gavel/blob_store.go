// Opens the blob store the serve host's ~/.gavel.yaml configures, which gavel serve hands to its
// dashboards for attachments.
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/flanksource/gavel/blob"
	"github.com/flanksource/gavel/verify"
)

// configuredBlobStore opens the storage section of ~/.gavel.yaml, or returns nil when it has none,
// so callers keep their usual local directories.
func configuredBlobStore(ctx context.Context) (blob.Store, error) {
	config, err := verify.LoadGlobalGavelConfig()
	if err != nil {
		return nil, err
	}
	if config.Storage.IsZero() {
		return nil, nil
	}
	store, err := blob.Open(ctx, config.Storage)
	if err != nil {
		home, _ := os.UserHomeDir()
		return nil, fmt.Errorf("open the storage in %s: %w", filepath.Join(home, verify.GavelConfigFileName), err)
	}
	return store, nil
}
