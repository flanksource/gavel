// Where a workspace's snapshots live: its .gavel directory by default, or the blob store the
// serve host's ~/.gavel.yaml configures, under a prefix derived from the workspace's path.
package snapshots

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/flanksource/gavel/blob"
	"github.com/flanksource/gavel/verify"
)

// ErrInvalidRunID reports a run ID that names neither the last pointer nor a run-* snapshot.
var ErrInvalidRunID = errors.New("invalid run id")

// OpenStore returns the store holding workDir's snapshots. It is a variable so tests can point
// it elsewhere.
var OpenStore = func(workDir string) (blob.Store, error) {
	store, err := verify.OpenWorkspaceStorage(context.Background(), workDir)
	if err != nil {
		return nil, fmt.Errorf("open the snapshot store: %w", err)
	}
	if store == nil {
		return blob.Local(filepath.Join(workDir, Dir)), nil
	}
	return store, nil
}

// Location describes where key, one of workDir's snapshots, is stored: a path under workDir's
// .gavel directory by default.
func Location(workDir, key string) string {
	local := filepath.Join(workDir, Dir, key)
	config, err := storageConfig()
	if err != nil || config.IsZero() {
		return local
	}
	prefix, err := blob.WorkspacePrefix(workDir)
	if err != nil {
		return local
	}
	return config.Location(prefix + key)
}

func storageConfig() (blob.Config, error) {
	config, err := verify.LoadGlobalGavelConfig()
	if err != nil {
		return blob.Config{}, fmt.Errorf("read the snapshot storage config: %w", err)
	}
	return config.Storage, nil
}

// ReadRun returns the raw snapshot runID names: PointerLast for the run last.json points at,
// or a run-* stem.
func ReadRun(workDir, runID string) ([]byte, error) {
	store, err := OpenStore(workDir)
	if err != nil {
		return nil, err
	}
	key := runID + ".json"
	if runID == PointerLast {
		pointer, err := loadPointer(store, PointerLast)
		if err != nil {
			return nil, err
		}
		if pointer == nil {
			return nil, fmt.Errorf("%w: no %s pointer", blob.ErrNotFound, PointerLast)
		}
		if key, err = pointerKey(pointer); err != nil {
			return nil, err
		}
	} else if !strings.HasPrefix(runID, PerRunPrefix) || strings.ContainsAny(runID, `/\`) || strings.Contains(runID, "..") {
		return nil, fmt.Errorf("%w %q: want %s or a %s* snapshot name", ErrInvalidRunID, runID, PointerLast, PerRunPrefix)
	}
	return readKey(store, key)
}

// pointerKey is the store key of the snapshot p names. Save records it as .gavel/<key>.
func pointerKey(p *Pointer) (string, error) {
	path := filepath.ToSlash(p.Path)
	key, ok := strings.CutPrefix(path, Dir+"/")
	if !ok || key == "" || strings.Contains(key, "/") {
		return "", fmt.Errorf("pointer names %q, not a snapshot under %s/", p.Path, Dir)
	}
	return key, nil
}

func readKey(store blob.Store, key string) ([]byte, error) {
	reader, err := store.Read(context.Background(), key)
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("read snapshot %s: %w", key, err)
	}
	return data, nil
}

func putJSON(store blob.Store, key string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal %s: %w", key, err)
	}
	return store.Write(context.Background(), key, bytes.NewReader(data))
}
