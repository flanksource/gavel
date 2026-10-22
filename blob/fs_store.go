// A Store over a commons-db filesystem, S3 or GCS. It validates keys, maps each backend's
// not-found error to ErrNotFound, and turns its listings into sorted keys.
package blob

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"sort"
	"strings"

	gcs "cloud.google.com/go/storage"
	smithyhttp "github.com/aws/smithy-go/transport/http"
	commonsfs "github.com/flanksource/commons-db/fs"
)

// fsStore is a Store over a commons-db filesystem.
type fsStore struct {
	fs commonsfs.FilesystemRW
	// literalListing is set for GCS, whose ReadDir takes a literal prefix rather than a glob.
	literalListing bool
}

func (s *fsStore) Write(ctx context.Context, key string, r io.Reader) error {
	if err := validateKey(key); err != nil {
		return err
	}
	if _, err := s.fs.Write(ctx, key, r); err != nil {
		return fmt.Errorf("write blob %s: %w", key, err)
	}
	return nil
}

func (s *fsStore) Read(ctx context.Context, key string) (io.ReadCloser, error) {
	if err := validateKey(key); err != nil {
		return nil, err
	}
	reader, err := s.fs.Read(ctx, key)
	if isNotFound(err) {
		return nil, fmt.Errorf("%w: %s", ErrNotFound, key)
	}
	if err != nil {
		return nil, fmt.Errorf("read blob %s: %w", key, err)
	}
	return reader, nil
}

func (s *fsStore) Exists(ctx context.Context, key string) (bool, error) {
	if err := validateKey(key); err != nil {
		return false, err
	}
	_, err := s.fs.Stat(key)
	if isNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("check blob %s: %w", key, err)
	}
	return true, nil
}

func (s *fsStore) Delete(ctx context.Context, key string) error {
	if err := validateKey(key); err != nil {
		return err
	}
	if err := s.fs.Delete(ctx, key); err != nil {
		return fmt.Errorf("delete blob %s: %w", key, err)
	}
	return nil
}

func (s *fsStore) List(ctx context.Context, prefix string) ([]Object, error) {
	if err := validatePrefix(prefix); err != nil {
		return nil, err
	}
	// List the directory part of the prefix and keep the keys that match all of it, so a
	// prefix may end mid-name ("runs/run-").
	dir := ""
	if i := strings.LastIndex(prefix, "/"); i >= 0 {
		dir = prefix[:i+1]
	}
	pattern := dir + "**"
	if s.literalListing {
		pattern = dir
	}
	entries, err := s.fs.ReadDir(pattern)
	if isNotFound(err) {
		return []Object{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("list blobs under %q: %w", prefix, err)
	}
	objects := []Object{}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		key := entry.FullPath()
		if strings.HasPrefix(key, prefix) {
			objects = append(objects, Object{Key: key, Size: entry.Size(), ModTime: entry.ModTime()})
		}
	}
	sort.Slice(objects, func(i, j int) bool { return objects[i].Key < objects[j].Key })
	return objects, nil
}

// isNotFound reports each backend's missing-object error: S3's HTTP 404 and GCS's
// ErrObjectNotExist, plus fs.ErrNotExist.
func isNotFound(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, fs.ErrNotExist) || errors.Is(err, gcs.ErrObjectNotExist) {
		return true
	}
	var response *smithyhttp.ResponseError
	return errors.As(err, &response) && response.HTTPStatusCode() == http.StatusNotFound
}
