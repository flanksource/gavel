package comments

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"mime"
	"strings"
	"time"

	"github.com/google/uuid"
)

type ReadMarker struct {
	CommentID string    `json:"commentId"`
	Version   int64     `json:"version"`
	At        time.Time `json:"at"`
}
type ReadStore interface {
	MarkRead(context.Context, Principal, ReadMarker) (ReadMarker, error)
	ReadMarkers(context.Context, Principal, []string) ([]ReadMarker, error)
}
type Page struct {
	Items  []Comment    `json:"items"`
	Total  int64        `json:"total"`
	Limit  int          `json:"limit"`
	Offset int          `json:"offset"`
	Read   []ReadMarker `json:"read"`
}
type PageStore interface {
	Page(context.Context, Principal, Filter) (Page, error)
}

func (s *Service) Page(ctx context.Context, f Filter) (Page, error) {
	p, err := s.principal(ctx)
	if err != nil {
		return Page{}, err
	}
	if f.ScopeURI == "" || f.Limit < 1 || f.Limit > 200 || f.Offset < 0 {
		return Page{}, ErrInvalid
	}
	if err = s.access.Authorize(ctx, p, "list", Target{ScopeURI: f.ScopeURI, URI: f.TargetURI}); err != nil {
		return Page{}, err
	}
	store, ok := s.store.(PageStore)
	if !ok {
		return Page{}, fmt.Errorf("comment store does not support pagination")
	}
	page, err := store.Page(ctx, p, f)
	if err != nil {
		return Page{}, err
	}
	for _, c := range page.Items {
		if err = s.access.Authorize(ctx, p, "read", c.Target); err != nil {
			return Page{}, err
		}
	}
	return page, nil
}
func (s *Service) MarkRead(ctx context.Context, id string, version int64) (ReadMarker, error) {
	c, err := s.Get(ctx, id)
	if err != nil {
		return ReadMarker{}, err
	}
	if version < 1 || version > c.Version {
		return ReadMarker{}, ErrInvalid
	}
	p, err := s.principal(ctx)
	if err != nil {
		return ReadMarker{}, err
	}
	store, ok := s.store.(ReadStore)
	if !ok {
		return ReadMarker{}, fmt.Errorf("comment store does not support read markers")
	}
	return store.MarkRead(ctx, p, ReadMarker{CommentID: id, Version: version, At: time.Now().UTC()})
}

type Attachment struct {
	ID        string    `json:"id"`
	MessageID string    `json:"messageId,omitempty"`
	Name      string    `json:"name"`
	MediaType string    `json:"mediaType"`
	Size      int64     `json:"size"`
	SHA256    string    `json:"sha256"`
	Author    Actor     `json:"author"`
	CreatedAt time.Time `json:"createdAt"`
}
type Blobs interface {
	Put(context.Context, string, string, []byte) error
	Open(context.Context, string, string) (io.ReadCloser, error)
	Delete(context.Context, string, string) error
}
type Upload struct {
	CommentID       string
	MessageID       string
	ExpectedVersion int64
	Name            string
	MediaType       string
	Body            io.Reader
}

func (s *Service) Upload(ctx context.Context, req Upload) (Comment, error) {
	if s.blobs == nil {
		return Comment{}, fmt.Errorf("attachments are not configured")
	}
	p, err := s.principal(ctx)
	if err != nil {
		return Comment{}, err
	}
	c, err := s.Get(ctx, req.CommentID)
	if err != nil {
		return Comment{}, err
	}
	if err = s.access.Authorize(ctx, p, "attach", c.Target); err != nil {
		return Comment{}, err
	}
	if req.Body == nil || req.Name == "" || strings.ContainsAny(req.Name, "/\\\r\n") || req.ExpectedVersion < 1 {
		return Comment{}, ErrInvalid
	}
	if _, _, err = mime.ParseMediaType(req.MediaType); err != nil {
		return Comment{}, fmt.Errorf("%w: invalid attachment media type", ErrInvalid)
	}
	data, err := io.ReadAll(io.LimitReader(req.Body, s.attachmentLimit+1))
	if err != nil {
		return Comment{}, err
	}
	if int64(len(data)) > s.attachmentLimit || len(data) == 0 {
		return Comment{}, fmt.Errorf("%w: attachment size must be 1..%d bytes", ErrInvalid, s.attachmentLimit)
	}
	sum := sha256.Sum256(data)
	a := Attachment{ID: uuid.NewString(), MessageID: req.MessageID, Name: req.Name, MediaType: req.MediaType, Size: int64(len(data)), SHA256: hex.EncodeToString(sum[:]), Author: p.Actor, CreatedAt: time.Now().UTC()}
	if err = s.blobs.Put(ctx, p.TenantID, a.ID, data); err != nil {
		return Comment{}, err
	}
	out, err := s.store.Mutate(ctx, p.TenantID, c.ID, req.ExpectedVersion, func(current Comment) (Change, error) {
		if current.Phase == Final {
			return Change{}, ErrConflict
		}
		if a.MessageID != "" {
			found := false
			for _, m := range current.Messages {
				found = found || m.ID == a.MessageID
			}
			if !found {
				return Change{}, ErrInvalid
			}
		}
		current.Version++
		current.UpdatedAt = a.CreatedAt
		return Change{Comment: current, Attachment: &a, Event: Event{CommentID: c.ID, Kind: "attachment", Actor: p.Actor, At: a.CreatedAt}}, nil
	})
	if err != nil {
		if cleanup := s.blobs.Delete(context.WithoutCancel(ctx), p.TenantID, a.ID); cleanup != nil {
			return Comment{}, fmt.Errorf("upload: %w; remove unlinked blob: %v", err, cleanup)
		}
	}
	return out, err
}
func (s *Service) Download(ctx context.Context, id, attachmentID string) (Attachment, io.ReadCloser, error) {
	if s.blobs == nil {
		return Attachment{}, nil, fmt.Errorf("attachments are not configured")
	}
	c, err := s.Get(ctx, id)
	if err != nil {
		return Attachment{}, nil, err
	}
	for _, a := range c.Attachments {
		if a.ID == attachmentID {
			reader, err := s.blobs.Open(ctx, c.TenantID, a.ID)
			return a, reader, err
		}
	}
	return Attachment{}, nil, ErrNotFound
}
