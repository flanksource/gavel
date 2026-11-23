package snapshots

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/flanksource/gavel/blob"
)

// CompressCache converts settled JSON snapshots at or above the configured threshold.
func CompressCache(ctx context.Context, workDir string) (int, error) {
	config, err := compressionConfig()
	if err != nil {
		return 0, err
	}
	if config.suffix == "" {
		return 0, fmt.Errorf("set %s to zstd or gzip to compress the cache", compressionDefaultProperty)
	}
	return editCache(ctx, workDir, func(root *os.Root, entries []fs.DirEntry) (int, error) {
		files, err := selectCacheFiles(root, entries, func(file fs.FileInfo) bool {
			return strings.HasSuffix(file.Name(), ".json") && file.Size() >= config.threshold
		})
		if err != nil {
			return 0, err
		}
		pointers, err := cachePointers(root, entries)
		if err != nil {
			return 0, err
		}
		for i, file := range files {
			if err := ctx.Err(); err != nil {
				return i, err
			}
			if err := compressCacheFile(root, file, config.suffix, pointers[file.Name()]); err != nil {
				return i, fmt.Errorf("compress %s: %w", file.Name(), err)
			}
		}
		return len(files), nil
	})
}

func compressCacheFile(root *os.Root, info fs.FileInfo, suffix string, pointers []cachePointer) error {
	source, err := root.Open(info.Name())
	if err != nil {
		return err
	}
	defer source.Close()
	target := info.Name() + suffix
	targetInfo, err := root.Lstat(target)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		if err := putCompressedCacheFile(root, target, source, info, suffix); err != nil {
			return err
		}
	case err != nil:
		return err
	case !targetInfo.Mode().IsRegular():
		return fmt.Errorf("%s is not a regular file", target)
	default:
		existing, err := root.Open(target)
		if err != nil {
			return err
		}
		if err := checkCompressedCopy(target, source, existing); err != nil {
			return err
		}
	}
	store := blob.Local(root.Name())
	for _, pointer := range pointers {
		if filepath.Base(pointer.data.Path) != info.Name() {
			continue
		}
		pointer.data.Path = filepath.Join(Dir, target)
		if err := putJSON(store, pointer.name, pointer.data); err != nil {
			return err
		}
	}
	return root.Remove(info.Name())
}

func putCompressedCacheFile(root *os.Root, target string, source io.Reader, info fs.FileInfo, suffix string) error {
	temp := ".snapshot-" + rand.Text()
	file, err := root.OpenFile(temp, os.O_WRONLY|os.O_CREATE|os.O_EXCL, info.Mode().Perm())
	if err != nil {
		return err
	}
	defer root.Remove(temp)
	err = compressSnapshot(file, source, suffix)
	if err == nil {
		err = file.Sync()
	}
	if err := errors.Join(err, file.Close()); err != nil {
		return err
	}
	if err := root.Chmod(temp, info.Mode().Perm()); err != nil {
		return err
	}
	if err := root.Chtimes(temp, info.ModTime(), info.ModTime()); err != nil {
		return err
	}
	return root.Link(temp, target)
}

func checkCompressedCopy(name string, source io.Reader, compressed io.ReadCloser) error {
	decoded, err := decodeReader(name, compressed)
	if err != nil {
		return err
	}
	defer decoded.Close()
	want, got := sha256.New(), sha256.New()
	if _, err := io.Copy(want, source); err != nil {
		return err
	}
	if _, err := io.Copy(got, decoded); err != nil {
		return err
	}
	if !bytes.Equal(want.Sum(nil), got.Sum(nil)) {
		return fmt.Errorf("existing %s has different contents", name)
	}
	return nil
}

func settledCacheSnapshot(root *os.Root, name string) (bool, error) {
	file, err := root.Open(name)
	if err != nil {
		return false, err
	}
	reader, err := decodeReader(name, file)
	if err != nil {
		return false, err
	}
	defer reader.Close()
	decoder := json.NewDecoder(reader)
	start, err := decoder.Token()
	if err != nil || start != json.Delim('{') {
		return false, err
	}
	// Snapshots put status before the potentially large tests and lint results.
	for decoder.More() {
		key, err := decoder.Token()
		if err != nil {
			return false, err
		}
		if key == "status" {
			var status *struct{ Running bool }
			err := decoder.Decode(&status)
			return err == nil && status != nil && !status.Running, err
		}
		if key == "path" {
			return false, nil
		}
		if err := decoder.Decode(new(json.RawMessage)); err != nil {
			return false, err
		}
	}
	_, err = decoder.Token()
	return false, err
}
