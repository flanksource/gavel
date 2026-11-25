package comments

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"mime"
	"slices"
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
	if f, err = s.locatePath(ctx, f); err != nil {
		return Page{}, err
	}
	switch {
	case f.ScopeURI == "":
		return Page{}, fmt.Errorf("%w: scope or path is required", ErrInvalid)
	case f.Limit < 1 || f.Limit > 200:
		return Page{}, fmt.Errorf("%w: limit must be 1..200, got %d", ErrInvalid, f.Limit)
	case f.Offset < 0:
		return Page{}, fmt.Errorf("%w: offset must not be negative, got %d", ErrInvalid, f.Offset)
	case f.After != "" && f.Offset != 0:
		return Page{}, fmt.Errorf("%w: after and offset are exclusive", ErrInvalid)
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

// locatePath narrows f to the scope and targets the host places f.Path at.
func (s *Service) locatePath(ctx context.Context, f Filter) (Filter, error) {
	if f.Path == "" {
		return f, nil
	}
	if s.locate == nil {
		return f, fmt.Errorf("%w: this host cannot list comments by path", ErrInvalid)
	}
	if f.TargetURI != "" || f.TargetPrefix != "" {
		return f, fmt.Errorf("%w: path and target filters are exclusive", ErrInvalid)
	}
	l, err := s.locate(ctx, f.Path)
	if err != nil {
		return f, err
	}
	if f.ScopeURI != "" && f.ScopeURI != l.ScopeURI {
		return f, fmt.Errorf("%w: %s is in %s, not in scope %s", ErrInvalid, f.Path, l.ScopeURI, f.ScopeURI)
	}
	f.ScopeURI, f.TargetURI, f.TargetPrefix = l.ScopeURI, l.TargetURI, l.TargetPrefix
	return f, nil
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

// ReadTarget is a page whose threads are in front of the reader.
type ReadTarget struct {
	ScopeURI string `json:"scopeUri"`
	URI      string `json:"uri"`
}

// targetPage is how many threads TargetThreads reads per page.
const targetPage = 200

// TargetThreads pages every thread on t, by id, with the caller's read
// version of each. It pages by keyset (Filter.After), so a thread deleted or
// created while it reads neither skips nor repeats another.
func (s *Service) TargetThreads(ctx context.Context, t ReadTarget) ([]Comment, map[string]int64, error) {
	if t.ScopeURI == "" || t.URI == "" {
		return nil, nil, fmt.Errorf("%w: a target needs scopeUri and uri, got %+v", ErrInvalid, t)
	}
	threads, reads := []Comment{}, map[string]int64{}
	for after := ""; ; {
		page, err := s.Page(ctx, Filter{ScopeURI: t.ScopeURI, TargetURI: t.URI, Limit: targetPage, After: after})
		if err != nil {
			return nil, nil, err
		}
		threads = append(threads, page.Items...)
		for _, r := range page.Read {
			reads[r.CommentID] = r.Version
		}
		if len(page.Items) < targetPage {
			return threads, reads, nil
		}
		after = page.Items[len(page.Items)-1].ID
	}
}

// MarkTargetRead marks every thread on t read at its current version, for the
// caller, and returns how many it marked: threads already read are skipped.
func (s *Service) MarkTargetRead(ctx context.Context, t ReadTarget) (int, error) {
	threads, reads, err := s.TargetThreads(ctx, t)
	if err != nil {
		return 0, err
	}
	marked := 0
	for _, c := range threads {
		if reads[c.ID] >= c.Version {
			continue
		}
		if _, err := s.MarkRead(ctx, c.ID, c.Version); err != nil {
			return marked, err
		}
		marked++
	}
	return marked, nil
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

// CreateAttachment is a file a create stores on its new root before the root
// is routed or delivered, so the agent sees it with the thread. Data is
// base64 in JSON.
type CreateAttachment struct {
	Filename    string `json:"filename"`
	ContentType string `json:"contentType"`
	Data        []byte `json:"data"`
}

// MaxCreateAttachments bounds the attachments one create carries.
const MaxCreateAttachments = 10

// AttachmentLimit is the largest attachment, in bytes, this service stores; 0 without blobs.
func (s *Service) AttachmentLimit() int64 { return s.attachmentLimit }

// newAttachment validates file and describes it as p's attachment, to
// messageID when set.
func (s *Service) newAttachment(p Principal, messageID string, file CreateAttachment) (Attachment, error) {
	if s.blobs == nil {
		return Attachment{}, fmt.Errorf("%w: attachments are not configured on this comments host", ErrUnsupported)
	}
	if file.Filename == "" || strings.ContainsAny(file.Filename, "/\\\r\n") {
		return Attachment{}, fmt.Errorf("%w: attachment name %q must be a plain file name", ErrInvalid, file.Filename)
	}
	if _, _, err := mime.ParseMediaType(file.ContentType); err != nil {
		return Attachment{}, fmt.Errorf("%w: invalid attachment media type %q", ErrInvalid, file.ContentType)
	}
	if size := int64(len(file.Data)); size > s.attachmentLimit || size == 0 {
		return Attachment{}, fmt.Errorf("%w: attachment %s is %d bytes; it must be 1..%d", ErrInvalid, file.Filename, size, s.attachmentLimit)
	}
	sum := sha256.Sum256(file.Data)
	return Attachment{ID: uuid.NewString(), MessageID: messageID, Name: file.Filename, MediaType: file.ContentType, Size: int64(len(file.Data)), SHA256: hex.EncodeToString(sum[:]), Author: p.Actor, CreatedAt: time.Now().UTC()}, nil
}

// attachTo is the change appending a, already stored as a blob, to a thread.
func attachTo(p Principal, a Attachment) func(Comment) (Change, error) {
	return func(current Comment) (Change, error) {
		if current.Phase == Final {
			return Change{}, ErrConflict
		}
		if a.MessageID != "" && !slices.ContainsFunc(current.Messages, func(m Message) bool { return m.ID == a.MessageID }) {
			return Change{}, ErrInvalid
		}
		current.Version++
		current.UpdatedAt = a.CreatedAt
		return Change{Comment: current, Attachment: &a, Event: Event{CommentID: current.ID, Kind: "attachment", Actor: p.Actor, At: a.CreatedAt}}, nil
	}
}

func (s *Service) Upload(ctx context.Context, req Upload) (Comment, error) {
	if s.blobs == nil {
		return Comment{}, fmt.Errorf("%w: attachments are not configured on this comments host", ErrUnsupported)
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
	if req.Body == nil || req.ExpectedVersion < 1 {
		return Comment{}, fmt.Errorf("%w: an upload needs a body and expectedVersion", ErrInvalid)
	}
	data, err := io.ReadAll(io.LimitReader(req.Body, s.attachmentLimit+1))
	if err != nil {
		return Comment{}, err
	}
	a, err := s.newAttachment(p, req.MessageID, CreateAttachment{Filename: req.Name, ContentType: req.MediaType, Data: data})
	if err != nil {
		return Comment{}, err
	}
	if err = s.blobs.Put(ctx, p.TenantID, a.ID, data); err != nil {
		return Comment{}, err
	}
	// An attachment only appends, and its preconditions are checked against the
	// latest thread, so changes made since the uploader's version (an agent
	// delivery claiming the new thread) do not conflict; a version the thread
	// has not reached does.
	out, err := s.WithLatest(ctx, c.ID, func(latest Comment) (Comment, error) {
		if req.ExpectedVersion > latest.Version {
			return Comment{}, fmt.Errorf("%w: expectedVersion %d is newer than the thread's %d", ErrConflict, req.ExpectedVersion, latest.Version)
		}
		return s.store.Mutate(ctx, p.TenantID, c.ID, latest.Version, attachTo(p, a))
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
		return Attachment{}, nil, fmt.Errorf("%w: attachments are not configured on this comments host", ErrUnsupported)
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
