// Package fileread reads text files confined to a root directory, for surfaces
// that hand file contents to a browser or an agent.
package fileread

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const (
	MaxBytes   = 2 << 20
	sniffBytes = 8 << 10
)

var (
	ErrNotLocal = errors.New("path must be relative and stay inside the root")
	ErrTooLarge = errors.New("file exceeds 2 MiB")
	ErrBinary   = errors.New("file is binary")
	ErrNotFile  = errors.New("path is a directory")
)

// Read reads path through an os.Root at root, so neither ".." nor a symlink can
// escape it, and rejects oversize or binary content.
func Read(root, path string) ([]byte, error) {
	if !filepath.IsLocal(path) {
		return nil, fmt.Errorf("%w: %q", ErrNotLocal, path)
	}
	dir, err := os.OpenRoot(root)
	if err != nil {
		return nil, err
	}
	defer dir.Close()
	file, err := dir.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		return nil, ErrNotFile
	}
	data, err := io.ReadAll(io.LimitReader(file, MaxBytes+1))
	if err != nil {
		return nil, err
	}
	if len(data) > MaxBytes {
		return nil, ErrTooLarge
	}
	if bytes.IndexByte(data[:min(len(data), sniffBytes)], 0) >= 0 {
		return nil, ErrBinary
	}
	return data, nil
}
