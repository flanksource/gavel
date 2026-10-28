package jira

import (
	"encoding/json"
	"fmt"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
)

type NamedValue struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
}

type IssueRow struct {
	Key      string     `json:"key"`
	Summary  string     `json:"summary"`
	URL      string     `json:"url"`
	Status   NamedValue `json:"status"`
	Assignee NamedValue `json:"assignee"`
}

func (r IssueRow) Columns() []api.ColumnDef {
	return []api.ColumnDef{clicky.Column("key").Label("Key").Build(), clicky.Column("summary").Label("Summary").Build(), clicky.Column("status").Label("Status").Build(), clicky.Column("assignee").Label("Assignee").Build()}
}

func (r IssueRow) Row() map[string]any {
	return map[string]any{"key": clicky.Link(r.URL).AddText(r.Key), "summary": r.Summary, "status": r.Status.Name, "assignee": r.Assignee.DisplayName}
}

func ParseList(raw json.RawMessage) ([]IssueRow, error) {
	var response struct {
		Issues []IssueRow `json:"issues"`
	}
	if err := json.Unmarshal(raw, &response); err != nil {
		return nil, fmt.Errorf("decode Jira issue list: %w", err)
	}
	return response.Issues, nil
}
