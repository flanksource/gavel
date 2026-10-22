// A Store that keeps every key under a fixed prefix of another Store, so callers sharing one
// bucket (e.g. one per organization) can't read or list each other's objects.
package blob

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

type prefixStore struct {
	store  Store
	prefix string
}

// WithPrefix returns a Store whose keys live under prefix in store. The prefix must be a valid
// key ending in "/"; any other prefix is a programming error and panics.
func WithPrefix(store Store, prefix string) Store {
	if !strings.HasSuffix(prefix, "/") || validateKey(strings.TrimSuffix(prefix, "/")) != nil {
		panic(fmt.Sprintf("blob.WithPrefix: prefix %q must be a valid key ending in \"/\"", prefix))
	}
	return &prefixStore{store: store, prefix: prefix}
}

func (s *prefixStore) Write(ctx context.Context, key string, r io.Reader) error {
	if err := validateKey(key); err != nil {
		return err
	}
	return s.store.Write(ctx, s.prefix+key, r)
}

func (s *prefixStore) Read(ctx context.Context, key string) (io.ReadCloser, error) {
	if err := validateKey(key); err != nil {
		return nil, err
	}
	reader, err := s.store.Read(ctx, s.prefix+key)
	if errors.Is(err, ErrNotFound) {
		return nil, fmt.Errorf("%w: %s", ErrNotFound, key)
	}
	return reader, err
}

func (s *prefixStore) Exists(ctx context.Context, key string) (bool, error) {
	if err := validateKey(key); err != nil {
		return false, err
	}
	return s.store.Exists(ctx, s.prefix+key)
}

func (s *prefixStore) Delete(ctx context.Context, key string) error {
	if err := validateKey(key); err != nil {
		return err
	}
	return s.store.Delete(ctx, s.prefix+key)
}

func (s *prefixStore) List(ctx context.Context, prefix string) ([]Object, error) {
	if err := validatePrefix(prefix); err != nil {
		return nil, err
	}
	objects, err := s.store.List(ctx, s.prefix+prefix)
	if err != nil {
		return nil, err
	}
	for i := range objects {
		objects[i].Key = strings.TrimPrefix(objects[i].Key, s.prefix)
	}
	return objects, nil
}

// WorkspacePrefix keeps workspaces sharing one store apart: workspaces/<16 hex chars of the
// SHA-256 of workDir's absolute path>/.
func WorkspacePrefix(workDir string) (string, error) {
	abs, err := filepath.Abs(workDir)
	if err != nil {
		return "", fmt.Errorf("resolve workspace %q: %w", workDir, err)
	}
	sum := sha256.Sum256([]byte(abs))
	return "workspaces/" + hex.EncodeToString(sum[:])[:16] + "/", nil
}
