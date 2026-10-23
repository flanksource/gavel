package snapshots

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// IndexName is the file in .gavel that caches each run's RunInfo, so ListRuns
// decodes a run-*.json once rather than on every listing. Run files are
// immutable and can hold gigabytes of captured output.
const IndexName = "runs-index.json"

// indexVersion is bumped whenever runInfo derives a RunInfo differently, so
// summaries cached by an older gavel are recomputed rather than served.
const indexVersion = 1

type indexFile struct {
	Version int                   `json:"version"`
	Runs    map[string]indexEntry `json:"runs"`
}

// indexEntry is one run's cached summary, valid while the file keeps the size
// and mtime it was summarized at.
type indexEntry struct {
	Size    int64     `json:"size"`
	ModTime time.Time `json:"modTime"`
	Info    RunInfo   `json:"info"`
}

// runIndex maps a run file's name to its cached summary.
type runIndex struct {
	entries map[string]indexEntry
	dirty   bool
}

// loadIndex reads dir's index. A missing index, or one written for another
// indexVersion, starts empty; an undecodable one is an error, since it is only
// ever replaced whole by save.
func loadIndex(dir string) (*runIndex, error) {
	index := &runIndex{entries: map[string]indexEntry{}}
	data, err := os.ReadFile(filepath.Join(dir, IndexName))
	if errors.Is(err, os.ErrNotExist) {
		return index, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read run index: %w", err)
	}
	var file indexFile
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, fmt.Errorf("decode run index %s: %w", filepath.Join(dir, IndexName), err)
	}
	if file.Version == indexVersion && file.Runs != nil {
		index.entries = file.Runs
	}
	return index, nil
}

// lookup returns the cached summary of name when the file is unchanged.
func (ix *runIndex) lookup(name, path string, stat os.FileInfo) (RunInfo, bool) {
	entry, ok := ix.entries[name]
	if !ok || entry.Size != stat.Size() || !entry.ModTime.Equal(stat.ModTime()) {
		return RunInfo{}, false
	}
	entry.Info.Path = path
	return entry.Info, true
}

func (ix *runIndex) store(name string, stat os.FileInfo, info RunInfo) {
	ix.entries[name] = indexEntry{Size: stat.Size(), ModTime: stat.ModTime(), Info: info}
	ix.dirty = true
}

// save replaces dir's index atomically, so a concurrent reader sees the old
// index or the new one, never a partial write.
func (ix *runIndex) save(dir string) error {
	if !ix.dirty {
		return nil
	}
	data, err := json.Marshal(indexFile{Version: indexVersion, Runs: ix.entries})
	if err != nil {
		return fmt.Errorf("encode run index: %w", err)
	}
	tmp, err := os.CreateTemp(dir, "."+IndexName+"-")
	if err != nil {
		return fmt.Errorf("write run index: %w", err)
	}
	defer os.Remove(tmp.Name()) //nolint:errcheck
	if _, err := tmp.Write(data); err != nil {
		tmp.Close() //nolint:errcheck
		return fmt.Errorf("write run index: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("write run index: %w", err)
	}
	if err := os.Rename(tmp.Name(), filepath.Join(dir, IndexName)); err != nil {
		return fmt.Errorf("write run index: %w", err)
	}
	return nil
}
