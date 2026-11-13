// Package blob adapts Gavel's configured blob storage to tenant-scoped comment attachments.
package blob

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"io"

	"github.com/flanksource/gavel/blob"
	"github.com/flanksource/gavel/comments"
)

type Store struct{ Store blob.Store }

func key(tenant, id string) string {
	return "comments/" + base64.RawURLEncoding.EncodeToString([]byte(tenant)) + "/" + base64.RawURLEncoding.EncodeToString([]byte(id))
}
func (s Store) Put(ctx context.Context, tenant, id string, data []byte) error {
	return s.Store.Write(ctx, key(tenant, id), bytes.NewReader(data))
}
func (s Store) Open(ctx context.Context, tenant, id string) (io.ReadCloser, error) {
	r, err := s.Store.Read(ctx, key(tenant, id))
	if errors.Is(err, blob.ErrNotFound) {
		return nil, comments.ErrNotFound
	}
	return r, err
}
func (s Store) Delete(ctx context.Context, tenant, id string) error {
	return s.Store.Delete(ctx, key(tenant, id))
}
