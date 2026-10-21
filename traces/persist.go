// Persist copies a closed run's trace store to the workspace's configured blob store, so the
// run's traces outlive the workspace they were captured in.
package traces

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/flanksource/gavel/blob"
)

// Persist copies every file of the run's store to traces/<runID>/ in dst. Call it once Close
// has flushed the store. A nil dst — no storage configured — leaves the store only in
// .gavel/traces.
func (h *Host) Persist(ctx context.Context, dst blob.Store) error {
	if dst == nil {
		return nil
	}
	return filepath.WalkDir(h.dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		rel, err := filepath.Rel(h.dir, path)
		if err != nil {
			return err
		}
		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("persist trace store of run %s: %w", h.runID, err)
		}
		defer file.Close()
		if err := dst.Write(ctx, "traces/"+h.runID+"/"+filepath.ToSlash(rel), file); err != nil {
			return fmt.Errorf("persist trace store of run %s: %w", h.runID, err)
		}
		return nil
	})
}
