// Package blob stores run results, attachments, recordings and logs on local disk or object storage.
// Keys are slash-separated and never absolute.
package blob

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
)

// ErrNotFound reports a key with no object. Read wraps it; Delete treats a missing key as done.
var ErrNotFound = errors.New("blob not found")

// ErrInvalidKey reports a key that is empty, absolute, has an empty, "." or ".." segment, or
// contains a backslash.
var ErrInvalidKey = errors.New("invalid blob key")

// Store reads and writes objects by key.
type Store interface {
	// Write stores r under key, replacing any object already there.
	Write(ctx context.Context, key string, r io.Reader) error
	// Read opens the object at key; a missing key is ErrNotFound.
	Read(ctx context.Context, key string) (io.ReadCloser, error)
	Exists(ctx context.Context, key string) (bool, error)
	// Delete removes the object at key; a missing key is not an error.
	Delete(ctx context.Context, key string) error
	// List returns the objects whose keys start with prefix, sorted by key. An empty prefix
	// lists every object.
	List(ctx context.Context, prefix string) ([]Object, error)
}

// Object is one stored object.
type Object struct {
	Key     string
	Size    int64
	ModTime time.Time
}

// validateKey rejects keys that could escape a store's root or name a directory.
func validateKey(key string) error {
	switch {
	case key == "":
		return fmt.Errorf("%w: the key is empty", ErrInvalidKey)
	case strings.HasPrefix(key, "/"):
		return fmt.Errorf("%w: key %q is absolute", ErrInvalidKey, key)
	case strings.Contains(key, `\`):
		return fmt.Errorf("%w: key %q has a backslash", ErrInvalidKey, key)
	}
	for _, segment := range strings.Split(key, "/") {
		switch segment {
		case "":
			return fmt.Errorf("%w: key %q has an empty segment", ErrInvalidKey, key)
		case ".":
			return fmt.Errorf("%w: key %q has a \".\" segment", ErrInvalidKey, key)
		case "..":
			return fmt.Errorf("%w: key %q has a \"..\" segment", ErrInvalidKey, key)
		}
	}
	return nil
}

// validatePrefix accepts an empty prefix, or a key optionally ending in a slash.
func validatePrefix(prefix string) error {
	if prefix == "" {
		return nil
	}
	return validateKey(strings.TrimSuffix(prefix, "/"))
}
