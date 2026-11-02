package jira

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/flanksource/gavel/externalissues"
)

type personPayload struct {
	AccountID   string            `json:"accountId"`
	DisplayName string            `json:"displayName"`
	URL         string            `json:"url"`
	Self        string            `json:"self"`
	AvatarURLs  map[string]string `json:"avatarUrls"`
}

func (p *personPayload) person() *externalissues.Person {
	if p == nil {
		return nil
	}
	person := &externalissues.Person{ID: p.AccountID, Name: p.DisplayName, URL: p.URL}
	if person.URL == "" {
		person.URL = p.Self
	}
	for _, size := range []string{"24x24", "32x32", "48x48", "16x16"} {
		if p.AvatarURLs[size] != "" {
			person.AvatarURL = p.AvatarURLs[size]
			break
		}
	}
	return person
}

type commentPayload struct {
	ID      string          `json:"id"`
	Author  personPayload   `json:"author"`
	Created *string         `json:"created"`
	Updated *string         `json:"updated"`
	Body    json.RawMessage `json:"body"`
}

func parseComment(raw json.RawMessage) (externalissues.Comment, error) {
	var p commentPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return externalissues.Comment{}, fmt.Errorf("decode Jira comment: %w", err)
	}
	comment := externalissues.Comment{ID: p.ID, Author: *p.Author.person()}
	var err error
	if comment.Created, err = parseTime(p.Created, "comment "+p.ID+" created"); err != nil {
		return comment, err
	}
	if comment.Updated, err = parseTime(p.Updated, "comment "+p.ID+" updated"); err != nil {
		return comment, err
	}
	comment.Body = parseDocument(p.Body)
	return comment, nil
}

type attachmentPayload struct {
	ID       string        `json:"id"`
	Filename string        `json:"filename"`
	Content  string        `json:"content"`
	MIMEType string        `json:"mimeType"`
	Size     int64         `json:"size"`
	Created  *string       `json:"created"`
	Author   personPayload `json:"author"`
}

func parseAttachment(raw json.RawMessage) (externalissues.Attachment, error) {
	var p attachmentPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return externalissues.Attachment{}, fmt.Errorf("decode Jira attachment: %w", err)
	}
	created, err := parseTime(p.Created, "attachment "+p.ID+" created")
	return externalissues.Attachment{ID: p.ID, Name: p.Filename, URL: p.Content, MIMEType: p.MIMEType, Size: p.Size, Created: created, Author: *p.Author.person()}, err
}

func parseTime(value *string, field string) (*time.Time, error) {
	if value == nil {
		return nil, nil
	}
	for _, layout := range []string{time.RFC3339Nano, "2006-01-02T15:04:05.999999999-0700", time.DateOnly} {
		if parsed, err := time.Parse(layout, *value); err == nil {
			return &parsed, nil
		}
	}
	return nil, fmt.Errorf("%s: invalid timestamp %q", field, *value)
}

func parseDocument(raw json.RawMessage) *externalissues.Document {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	var value string
	if err := json.Unmarshal(raw, &value); err == nil {
		return &externalissues.Document{Format: "text", Content: raw}
	}
	return &externalissues.Document{Format: "adf", Content: raw}
}
