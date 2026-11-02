package jira

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/flanksource/gavel/externalissues"
)

type Issue struct {
	externalissues.Issue
	Full               bool `json:"-"`
	attachmentsIgnored bool
}

type ParseOptions struct {
	Site         string
	Catalog      map[string]FieldDefinition
	IgnoreFields []string
}

type issuePayload struct {
	ID          string                    `json:"id"`
	Key         string                    `json:"key"`
	Summary     string                    `json:"summary"`
	URL         string                    `json:"url"`
	Status      *statusPayload            `json:"status"`
	Assignee    *personPayload            `json:"assignee"`
	Reporter    *personPayload            `json:"reporter"`
	Project     *externalissues.Reference `json:"project"`
	IssueType   *externalissues.Reference `json:"issuetype"`
	Priority    *externalissues.Reference `json:"priority"`
	Created     *string                   `json:"created"`
	Updated     *string                   `json:"updated"`
	Description json.RawMessage           `json:"description"`
	Labels      []string                  `json:"labels"`
	Comments    []json.RawMessage         `json:"comments"`
	Comment     *struct {
		Total *int `json:"total"`
	} `json:"comment"`
	Attachments []json.RawMessage `json:"attachment"`
}

type statusPayload struct {
	externalissues.Reference
	Category struct {
		Key string `json:"key"`
	} `json:"statusCategory"`
}

func ParseIssue(raw json.RawMessage, options ParseOptions) (Issue, error) {
	raw, err := options.filter(raw)
	if err != nil {
		return Issue{}, err
	}
	var payload issuePayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return Issue{}, fmt.Errorf("decode Jira issue: %w", err)
	}
	if strings.TrimSpace(payload.Key) == "" || strings.TrimSpace(payload.Summary) == "" {
		return Issue{}, fmt.Errorf("jira issue requires a valid key and summary")
	}
	issue := Issue{Issue: externalissues.Issue{
		Provider: "jira", Site: options.Site, ID: payload.ID, Key: payload.Key,
		Title: payload.Summary, URL: payload.URL, Project: payload.Project,
		Type: payload.IssueType, Priority: payload.Priority, Labels: payload.Labels,
		Assignee: payload.Assignee.person(), Reporter: payload.Reporter.person(),
	}, attachmentsIgnored: options.ignores("attachment")}
	if err := issue.decodeContent(payload); err != nil {
		return Issue{}, fmt.Errorf("issue %s: %w", issue.Key, err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return Issue{}, fmt.Errorf("issue %s fields: %w", issue.Key, err)
	}
	for key := range fields {
		if issueCoreField(key) {
			delete(fields, key)
		}
	}
	mapped, err := MapFields(fields, options.Catalog)
	if err != nil {
		return Issue{}, fmt.Errorf("issue %s fields: %w", issue.Key, err)
	}
	issue.Fields = mapped
	return issue, nil
}

func (i *Issue) decodeContent(payload issuePayload) (err error) {
	if payload.Comment != nil {
		i.CommentCount = payload.Comment.Total
	}
	if payload.Comments != nil && i.CommentCount == nil {
		count := len(payload.Comments)
		i.CommentCount = &count
	}
	if payload.Attachments != nil {
		count := len(payload.Attachments)
		i.AttachmentCount = &count
	}
	if i.Status, err = payload.Status.status(); err != nil {
		return err
	}
	if i.Created, err = parseTime(payload.Created, "created"); err != nil {
		return err
	}
	if i.Updated, err = parseTime(payload.Updated, "updated"); err != nil {
		return err
	}
	i.Description = parseDocument(payload.Description)
	for _, raw := range payload.Comments {
		comment, err := parseComment(raw)
		if err != nil {
			return err
		}
		i.Comments = append(i.Comments, comment)
	}
	for _, raw := range payload.Attachments {
		attachment, err := parseAttachment(raw)
		if err != nil {
			return err
		}
		i.Attachments = append(i.Attachments, attachment)
	}
	return nil
}

func issueCoreField(name string) bool {
	switch strings.ToLower(name) {
	case "id", "key", "summary", "url", "description", "comments", "comment", "attachment", "status", "assignee", "reporter", "project", "issuetype", "priority", "created", "updated", "labels":
		return true
	default:
		return false
	}
}

func (p *statusPayload) status() (*externalissues.Status, error) {
	if p == nil {
		return nil, nil
	}
	status := &externalissues.Status{Reference: p.Reference}
	switch p.Category.Key {
	case "":
	case "new":
		status.Category = externalissues.StatusTodo
	case "indeterminate":
		status.Category = externalissues.StatusInProgress
	case "done":
		status.Category = externalissues.StatusDone
	default:
		return nil, fmt.Errorf("status %s: unsupported category %q", p.ID, p.Category.Key)
	}
	return status, nil
}
