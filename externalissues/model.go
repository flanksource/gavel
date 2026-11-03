package externalissues

import (
	"encoding/json"
	"time"
)

type StatusCategory string

const (
	StatusTodo       StatusCategory = "todo"
	StatusInProgress StatusCategory = "in_progress"
	StatusDone       StatusCategory = "done"
)

type Reference struct {
	ID   string `json:"id,omitempty"`
	Key  string `json:"key,omitempty"`
	Name string `json:"name,omitempty"`
	URL  string `json:"url,omitempty"`
}

type Status struct {
	Reference
	Category StatusCategory `json:"category,omitempty"`
}

type Person struct {
	ID        string `json:"id,omitempty"`
	Name      string `json:"name,omitempty"`
	URL       string `json:"url,omitempty"`
	AvatarURL string `json:"avatar_url,omitempty"`
}

// Document preserves a provider's rich content without requiring a renderer to decode an issue.
type Document struct {
	Format  string          `json:"format"`
	Content json.RawMessage `json:"content"`
}

type Field struct {
	ID     string          `json:"id"`
	Name   string          `json:"name,omitempty"`
	Schema json.RawMessage `json:"-"`
	Format string          `json:"format,omitempty"`
	Value  json.RawMessage `json:"value"`
}

type Comment struct {
	ID      string     `json:"id,omitempty"`
	Author  Person     `json:"author"`
	Created *time.Time `json:"created,omitempty"`
	Updated *time.Time `json:"updated,omitempty"`
	Body    *Document  `json:"body,omitempty"`
}

type Attachment struct {
	ID       string     `json:"id,omitempty"`
	Name     string     `json:"name"`
	URL      string     `json:"url"`
	MIMEType string     `json:"mime_type,omitempty"`
	Size     int64      `json:"size"`
	Created  *time.Time `json:"created,omitempty"`
	Author   Person     `json:"author"`
}

type Issue struct {
	Provider        string       `json:"provider"`
	Site            string       `json:"site,omitempty"`
	ID              string       `json:"id,omitempty"`
	Key             string       `json:"key"`
	Title           string       `json:"title"`
	URL             string       `json:"url,omitempty"`
	Status          *Status      `json:"status,omitempty"`
	Assignee        *Person      `json:"assignee,omitempty"`
	Reporter        *Person      `json:"reporter,omitempty"`
	Project         *Reference   `json:"project,omitempty"`
	Type            *Reference   `json:"type,omitempty"`
	Priority        *Reference   `json:"priority,omitempty"`
	Created         *time.Time   `json:"created,omitempty"`
	Updated         *time.Time   `json:"updated,omitempty"`
	Description     *Document    `json:"description,omitempty"`
	Labels          []string     `json:"labels,omitempty"`
	Comments        []Comment    `json:"comments,omitempty"`
	Attachments     []Attachment `json:"attachments,omitempty"`
	CommentCount    *int         `json:"comment_count,omitempty"`
	AttachmentCount *int         `json:"attachment_count,omitempty"`
	Fields          Fields       `json:"fields"`
}
