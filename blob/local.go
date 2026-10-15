// A Store over a local directory. Every operation goes through os.Root, so a symlink planted
// under the directory can't redirect a read, write or delete to a file outside it.
package blob

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"sort"
	"strings"
)

type localStore struct {
	root string
}

// Local returns a Store over the directory root, which is created on the first write.
func Local(root string) Store {
	return &localStore{root: root}
}

// open opens the root directory; a root that doesn't exist yet holds no objects.
func (s *localStore) open() (*os.Root, error) {
	root, err := os.OpenRoot(s.root)
	if err != nil {
		return nil, fmt.Errorf("open blob directory %s: %w", s.root, err)
	}
	return root, nil
}

func (s *localStore) Write(ctx context.Context, key string, r io.Reader) error {
	if err := validateKey(key); err != nil {
		return err
	}
	if err := os.MkdirAll(s.root, 0o755); err != nil {
		return fmt.Errorf("create blob directory %s: %w", s.root, err)
	}
	root, err := s.open()
	if err != nil {
		return err
	}
	defer root.Close()
	if dir := path.Dir(key); dir != "." {
		if err := root.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("write blob %s: %w", key, err)
		}
	}
	file, err := root.Create(key)
	if err != nil {
		return fmt.Errorf("write blob %s: %w", key, err)
	}
	if _, err := io.Copy(file, r); err != nil {
		_ = file.Close()
		return fmt.Errorf("write blob %s: %w", key, err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("write blob %s: %w", key, err)
	}
	return nil
}

func (s *localStore) Read(ctx context.Context, key string) (io.ReadCloser, error) {
	if err := validateKey(key); err != nil {
		return nil, err
	}
	root, err := s.open()
	if errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("%w: %s", ErrNotFound, key)
	}
	if err != nil {
		return nil, err
	}
	defer root.Close()
	file, err := root.Open(key)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("%w: %s", ErrNotFound, key)
	}
	if err != nil {
		return nil, fmt.Errorf("read blob %s: %w", key, err)
	}
	return file, nil
}

func (s *localStore) Exists(ctx context.Context, key string) (bool, error) {
	if err := validateKey(key); err != nil {
		return false, err
	}
	root, err := s.open()
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	defer root.Close()
	_, err = root.Stat(key)
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("check blob %s: %w", key, err)
	}
	return true, nil
}

func (s *localStore) Delete(ctx context.Context, key string) error {
	if err := validateKey(key); err != nil {
		return err
	}
	root, err := s.open()
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	defer root.Close()
	if err := root.Remove(key); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("delete blob %s: %w", key, err)
	}
	return nil
}

// List walks the prefix's directory without following symlinks, and lists regular files only.
func (s *localStore) List(ctx context.Context, prefix string) ([]Object, error) {
	if err := validatePrefix(prefix); err != nil {
		return nil, err
	}
	root, err := s.open()
	if errors.Is(err, fs.ErrNotExist) {
		return []Object{}, nil
	}
	if err != nil {
		return nil, err
	}
	defer root.Close()
	dir := "."
	if i := strings.LastIndex(prefix, "/"); i >= 0 {
		dir = prefix[:i]
	}
	objects := []Object{}
	err = fs.WalkDir(root.FS(), dir, func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) && name == dir {
				return fs.SkipAll
			}
			return err
		}
		if !entry.Type().IsRegular() || !strings.HasPrefix(name, prefix) {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		objects = append(objects, Object{Key: name, Size: info.Size(), ModTime: info.ModTime()})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("list blobs under %q in %s: %w", prefix, s.root, err)
	}
	sort.Slice(objects, func(i, j int) bool { return objects[i].Key < objects[j].Key })
	return objects, nil
}
