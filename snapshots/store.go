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
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/flanksource/gavel/blob"
	testui "github.com/flanksource/gavel/testrunner/ui"
	"github.com/flanksource/gavel/utils"
	"github.com/flanksource/gavel/verify"
	"github.com/gofrs/flock"
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
	} else if !validRunID(runID) {
		return nil, fmt.Errorf("%w %q: want %s or a %s* snapshot name", ErrInvalidRunID, runID, PointerLast, PerRunPrefix)
	}
	_, data, err := readSnapshot(store, key)
	return data, err
}

func validRunID(runID string) bool {
	return strings.HasPrefix(runID, PerRunPrefix) && !strings.ContainsAny(runID, `/\`) && !strings.Contains(runID, "..")
}

// pointerKey is the store key of the snapshot p names. Save records it as .gavel/<key>.
func pointerKey(p *Pointer) (string, error) {
	path := filepath.ToSlash(p.Path)
	key, ok := strings.CutPrefix(path, Dir+"/")
	if !ok || SnapshotStem(key) == "" || strings.ContainsAny(key, `/\`) {
		return "", fmt.Errorf("pointer names %q, not a snapshot under %s/", p.Path, Dir)
	}
	return key, nil
}

func openSnapshotKey(store blob.Store, key string) (string, io.ReadCloser, error) {
	return openSnapshot(key, func(candidate string) (io.ReadCloser, error) {
		return store.Read(context.Background(), candidate)
	}, blob.ErrNotFound)
}

func readSnapshot(store blob.Store, key string) (string, []byte, error) {
	name, source, err := openSnapshotKey(store, key)
	if err != nil {
		return name, nil, err
	}
	reader, err := decodeReader(name, source)
	if err != nil {
		return name, nil, err
	}
	defer reader.Close()
	data, err := io.ReadAll(reader)
	if err != nil {
		return name, nil, fmt.Errorf("read snapshot %s: %w", name, err)
	}
	return name, data, nil
}

func putSnapshot(store blob.Store, key string, snap *testui.Snapshot) (string, error) {
	config, err := compressionConfig()
	if err != nil {
		return "", err
	}
	data, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal %s: %w", key, err)
	}
	if snap.Status.Running || config.suffix == "" || int64(len(data)) < config.threshold {
		return key, store.Write(context.Background(), key, bytes.NewReader(data))
	}
	var compressed bytes.Buffer
	if err := compressSnapshot(&compressed, bytes.NewReader(data), config.suffix); err != nil {
		return "", fmt.Errorf("compress %s: %w", key, err)
	}
	name := key + config.suffix
	return name, store.Write(context.Background(), name, &compressed)
}

// replaceSnapshot publishes the new version before removing older encodings.
func replaceSnapshot(store blob.Store, key string, snap *testui.Snapshot) (string, error) {
	name, err := putSnapshot(store, key, snap)
	if err != nil {
		return "", err
	}
	return name, deleteSnapshotKeys(context.Background(), store, key, name)
}

func deleteSnapshotKeys(ctx context.Context, store blob.Store, key, keep string) error {
	for _, candidate := range snapshotKeys(key) {
		if candidate != keep {
			if err := store.Delete(ctx, candidate); err != nil && !errors.Is(err, blob.ErrNotFound) {
				return err
			}
		}
	}
	return nil
}

// DeleteRun removes every stored encoding of a per-run snapshot.
func DeleteRun(ctx context.Context, store blob.Store, runID string) error {
	if !validRunID(runID) {
		return fmt.Errorf("%w %q", ErrInvalidRunID, runID)
	}
	return withSnapshotLock(ctx, store, func() error {
		return deleteSnapshotKeys(ctx, store, runID+".json", "")
	})
}

func withSnapshotLock(ctx context.Context, store blob.Store, action func() error) (err error) {
	dir, local := blob.LocalDirectory(store)
	if !local {
		return action()
	}
	if err := utils.MkdirOutput(dir, 0o755); err != nil {
		return err
	}
	lock := flock.New(filepath.Join(dir, "snapshots.lock"), flock.SetFlag(os.O_CREATE|os.O_RDWR))
	locked, err := lock.TryLockContext(ctx, 50*time.Millisecond)
	if err != nil {
		return fmt.Errorf("lock snapshots in %s: %w", dir, err)
	}
	if !locked {
		return fmt.Errorf("lock snapshots in %s: lock was not acquired", dir)
	}
	defer func() { err = errors.Join(err, lock.Close()) }()
	return action()
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
