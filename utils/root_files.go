package utils

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

func openWithin(base, name string) (*os.Root, string, error) {
	resolved, err := ResolveWithin(base, name)
	if err != nil {
		return nil, "", err
	}
	abs, err := filepath.Abs(base)
	if err != nil {
		return nil, "", fmt.Errorf("resolve base directory %q: %w", base, err)
	}
	rel, err := filepath.Rel(filepath.Clean(abs), resolved)
	if err != nil {
		return nil, "", err
	}
	root, err := os.OpenRoot(abs)
	if err != nil {
		return nil, "", err
	}
	return root, rel, nil
}

// ReadFileWithin reads name under base, refusing any path, including one
// reached through a symlink, that leaves base.
func ReadFileWithin(base, name string) ([]byte, error) {
	root, rel, err := openWithin(base, name)
	if err != nil {
		return nil, err
	}
	defer root.Close()
	return root.ReadFile(rel)
}

// WriteFileWithin atomically replaces name under base, keeping the existing
// file's permissions or using perm for a new file. Symlinks leaving base are
// refused.
func WriteFileWithin(base, name string, data []byte, perm fs.FileMode) error {
	root, rel, err := openWithin(base, name)
	if err != nil {
		return err
	}
	defer root.Close()
	if info, err := root.Stat(rel); err == nil {
		perm = info.Mode().Perm()
	} else if !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	suffix := make([]byte, 8)
	if _, err := rand.Read(suffix); err != nil {
		return err
	}
	tmp := rel + ".tmp-" + hex.EncodeToString(suffix)
	file, err := root.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	if err := writeAndChmod(file, data, perm); err != nil {
		_ = root.Remove(tmp)
		return err
	}
	if err := root.Rename(tmp, rel); err != nil {
		_ = root.Remove(tmp)
		return err
	}
	return nil
}

func writeAndChmod(file *os.File, data []byte, perm fs.FileMode) error {
	if err := file.Chmod(perm); err != nil {
		_ = file.Close()
		return err
	}
	if _, err := file.Write(data); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}
