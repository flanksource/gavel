// Package comments provides embeddable, outcome-oriented comments independent of tasks and conversations.
package comments

import (
	"context"
	"encoding/json"
	"errors"
	"time"
)

var (
	ErrNotFound  = errors.New("comment not found")
	ErrConflict  = errors.New("comment version or request conflict")
	ErrForbidden = errors.New("comment access denied")
)

type Principal struct {
	TenantID string
	Actor    Actor
}
type Target struct {
	URI      string          `json:"uri"`
	ScopeURI string          `json:"scopeUri"`
	Kind     string          `json:"kind"`
	Revision string          `json:"revision,omitempty"`
	Label    string          `json:"label,omitempty"`
	Selector json.RawMessage `json:"selector,omitempty"`
}
type Outcome struct {
	Kind     string    `json:"kind"`
	Summary  string    `json:"summary,omitempty"`
	Evidence []string  `json:"evidence,omitempty"`
	Actor    Actor     `json:"actor"`
	At       time.Time `json:"at"`
}
type Message struct {
	ID        string          `json:"id"`
	ParentID  string          `json:"parentId,omitempty"`
	Body      string          `json:"bodyMarkdown"`
	Author    Actor           `json:"author"`
	Content   json.RawMessage `json:"content,omitempty"`
	CreatedAt time.Time       `json:"createdAt"`
}
type Link struct {
	Relation string `json:"relation"`
	URI      string `json:"uri"`
}
type Comment struct {
	ID          string          `json:"id"`
	TenantID    string          `json:"tenantId"`
	Target      Target          `json:"target"`
	Body        string          `json:"bodyMarkdown"`
	Author      Actor           `json:"author"`
	WorkflowID  string          `json:"workflowId"`
	Status      string          `json:"status"`
	Phase       Phase           `json:"phase"`
	Assignee    *Actor          `json:"assignee,omitempty"`
	Resolution  *Outcome        `json:"resolution,omitempty"`
	Outcome     *Outcome        `json:"outcome,omitempty"`
	Content     json.RawMessage `json:"content,omitempty"`
	Links       []Link          `json:"links,omitempty"`
	Version     int64           `json:"version"`
	CreatedAt   time.Time       `json:"createdAt"`
	UpdatedAt   time.Time       `json:"updatedAt"`
	DeletedAt   *time.Time      `json:"deletedAt,omitempty"`
	Messages    []Message       `json:"messages,omitempty"`
	Attachments []Attachment    `json:"attachments,omitempty"`
}

func (c Comment) GetID() string   { return c.ID }
func (c Comment) GetName() string { return c.Target.Label }

type Event struct {
	Sequence  int64           `json:"sequence"`
	CommentID string          `json:"commentId"`
	Kind      string          `json:"kind"`
	Actor     Actor           `json:"actor"`
	At        time.Time       `json:"at"`
	Data      json.RawMessage `json:"data,omitempty"`
}
type Filter struct {
	ScopeURI  string `json:"scopeUri,omitempty" flag:"scope"`
	TargetURI string `json:"targetUri,omitempty" flag:"target"`
	Phase     Phase  `json:"phase,omitempty" flag:"phase"`
	Limit     int    `json:"limit" flag:"limit" default:"50"`
	Offset    int    `json:"offset" flag:"offset"`
	Unread    bool   `json:"unread,omitempty" flag:"unread"`
}
type Change struct {
	Comment         Comment
	Message         *Message
	Attachment      *Attachment
	EditMessage     *Message
	DeleteMessageID string
	Event           Event
}
type TransactionStore interface {
	Transaction(context.Context, func(Store) error) error
}
type Store interface {
	Create(context.Context, Comment, string, string, Event) (Comment, error)
	Get(context.Context, string, string) (Comment, error)
	List(context.Context, string, Filter) ([]Comment, error)
	Mutate(context.Context, string, string, int64, func(Comment) (Change, error)) (Comment, error)
	Events(context.Context, string, string, int64, int) ([]Event, error)
}
type Access interface {
	Principal(context.Context) (Principal, error)
	Authorize(context.Context, Principal, string, Target) error
}
type Targets interface {
	Resolve(context.Context, Principal, Target) (Target, error)
}
type Options struct {
	Store           Store
	Access          Access
	Targets         Targets
	Workflows       []Workflow
	Providers       []Provider
	Blobs           Blobs
	AttachmentLimit int64
}
