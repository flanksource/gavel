package snapshots

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/flanksource/commons/duration"
	"github.com/flanksource/commons/properties"
	"github.com/flanksource/gavel/blob"
)

type CacheStats struct {
	Path  string
	Bytes int64
}

// CacheInfo measures the regular files in the workspace's local .gavel directory.
func CacheInfo(workDir string) (CacheStats, error) {
	if workDir == "" {
		return CacheStats{}, errors.New("cache location requires a workspace")
	}
	stats := CacheStats{Path: filepath.Join(workDir, Dir)}
	err := filepath.WalkDir(stats.Path, func(path string, entry fs.DirEntry, err error) error {
		if path == stats.Path && errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		if err != nil || !entry.Type().IsRegular() {
			return err
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		stats.Bytes += info.Size()
		return nil
	})
	return stats, err
}

type CacheAge struct {
	minBytes int64
	age      time.Duration
}

// ParseCacheAges accepts repeated duration or size:duration cleanup rules.
func ParseCacheAges(values []string) ([]CacheAge, error) {
	if len(values) == 0 {
		return nil, errors.New("cache cleanup requires at least one --age rule")
	}
	rules := make([]CacheAge, 0, len(values))
	for _, value := range values {
		rule := CacheAge{}
		age := value
		if size, rest, ok := strings.Cut(value, ":"); ok {
			bytes, err := properties.ParseBytes(size)
			if err != nil || bytes < 0 {
				return nil, fmt.Errorf("invalid --age %q: want a non-negative byte size before ':'", value)
			}
			rule.minBytes, age = bytes, rest
		}
		parsed, err := duration.ParseDuration(strings.TrimSpace(age))
		if err != nil || parsed <= 0 {
			return nil, fmt.Errorf("invalid --age %q: want a positive duration such as 30d or 10MB:10d", value)
		}
		rule.age = time.Duration(parsed)
		rules = append(rules, rule)
	}
	return rules, nil
}

// CleanupCache deletes settled snapshots that match any rule and have no pointer to them.
func CleanupCache(ctx context.Context, workDir string, rules []CacheAge, now time.Time) (int, error) {
	if len(rules) == 0 {
		return 0, errors.New("cache cleanup requires at least one age rule")
	}
	return editCache(ctx, workDir, func(root *os.Root, entries []fs.DirEntry) (int, error) {
		protected, err := cachePointers(root, entries)
		if err != nil {
			return 0, err
		}
		files, err := selectCacheFiles(root, entries, func(file fs.FileInfo) bool {
			return len(protected[snapshotJSONName(file.Name())]) == 0 && slices.ContainsFunc(rules, func(rule CacheAge) bool {
				return file.Size() >= rule.minBytes && file.ModTime().Before(now.Add(-rule.age))
			})
		})
		if err != nil {
			return 0, err
		}
		for i, file := range files {
			if err := ctx.Err(); err != nil {
				return i, err
			}
			if err := root.Remove(file.Name()); err != nil {
				return i, err
			}
		}
		return len(files), nil
	})
}

func editCache(ctx context.Context, workDir string, edit func(*os.Root, []fs.DirEntry) (int, error)) (int, error) {
	if workDir == "" {
		return 0, errors.New("editing the cache requires a workspace")
	}
	dir := filepath.Join(workDir, Dir)
	info, err := os.Lstat(dir)
	if errors.Is(err, fs.ErrNotExist) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	if !info.IsDir() {
		return 0, fmt.Errorf("cache %s must be a directory", dir)
	}
	var count int
	err = withSnapshotLock(ctx, blob.Local(dir), func() error {
		root, err := os.OpenRoot(dir)
		if err != nil {
			return err
		}
		defer root.Close()
		entries, err := fs.ReadDir(root.FS(), ".")
		if err != nil {
			return err
		}
		count, err = edit(root, entries)
		return err
	})
	return count, err
}

func selectCacheFiles(root *os.Root, entries []fs.DirEntry, match func(fs.FileInfo) bool) ([]fs.FileInfo, error) {
	var files []fs.FileInfo
	for _, entry := range entries {
		if !entry.Type().IsRegular() || !isCacheSnapshot(entry.Name()) {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			return nil, err
		}
		if !match(info) {
			continue
		}
		settled, err := settledCacheSnapshot(root, entry.Name())
		if err != nil {
			return nil, fmt.Errorf("inspect cached snapshot %s: %w", entry.Name(), err)
		}
		if settled {
			files = append(files, info)
		}
	}
	return files, nil
}

func isCacheSnapshot(name string) bool {
	stem := SnapshotStem(name)
	if validRunID(stem) || strings.HasPrefix(stem, "sha-") {
		return true
	}
	if number, ok := strings.CutPrefix(stem, "pr-"); ok {
		id, err := strconv.Atoi(number)
		return err == nil && id > 0
	}
	return false
}

type cachePointer struct {
	name string
	data Pointer
}

func cachePointers(root *os.Root, entries []fs.DirEntry) (map[string][]cachePointer, error) {
	pointers := make(map[string][]cachePointer)
	for _, entry := range entries {
		if !entry.Type().IsRegular() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			return nil, err
		}
		// A pointer contains a relative snapshot filename, a git SHA, and an optional dirty hash.
		if info.Size() > 4096 && entry.Name() != PointerLast+".json" {
			continue
		}
		data, err := fs.ReadFile(root.FS(), entry.Name())
		if err != nil {
			return nil, err
		}
		var pointer Pointer
		if err := json.Unmarshal(data, &pointer); err != nil {
			if entry.Name() == PointerLast+".json" {
				return nil, fmt.Errorf("decode %s: %w", entry.Name(), err)
			}
			continue
		}
		if pointer.SHA == "" && entry.Name() != PointerLast+".json" {
			continue
		}
		key, err := pointerKey(&pointer)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", entry.Name(), err)
		}
		base := snapshotJSONName(key)
		pointers[base] = append(pointers[base], cachePointer{name: entry.Name(), data: pointer})
	}
	return pointers, nil
}
