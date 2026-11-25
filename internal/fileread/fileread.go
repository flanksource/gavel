// Package fileread reads files confined to a root directory, for surfaces
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
	MaxBytes      = 2 << 20
	MaxImageBytes = 10 << 20
	sniffBytes    = 8 << 10
)

var (
	ErrNotLocal = errors.New("path must be relative and stay inside the root")
	ErrTooLarge = errors.New("file exceeds the size limit")
	ErrBinary   = errors.New("file is binary")
	ErrNotFile  = errors.New("path is a directory")
	ErrNotImage = errors.New("path does not have an image extension")
)

// Options selects how Read validates content. Image mode requires an image
// extension, allows binary content and raises the size cap to MaxImageBytes.
type Options struct {
	Image bool
}

// Read reads path through an os.Root at root, so neither ".." nor a symlink can
// escape it, and rejects oversize content and, outside image mode, binary content.
func Read(root, path string, opts Options) ([]byte, error) {
	if !filepath.IsLocal(path) {
		return nil, fmt.Errorf("%w: %q", ErrNotLocal, path)
	}
	limit := MaxBytes
	if opts.Image {
		if _, ok := ContentType(path); !ok {
			return nil, fmt.Errorf("%w: %q", ErrNotImage, path)
		}
		limit = MaxImageBytes
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
	data, err := io.ReadAll(io.LimitReader(file, int64(limit)+1))
	if err != nil {
		return nil, err
	}
	if len(data) > limit {
		return nil, fmt.Errorf("%w of %d MiB", ErrTooLarge, limit>>20)
	}
	if !opts.Image && bytes.IndexByte(data[:min(len(data), sniffBytes)], 0) >= 0 {
		return nil, ErrBinary
	}
	return data, nil
}
