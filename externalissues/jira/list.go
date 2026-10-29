package jira

import (
	"encoding/json"
	"fmt"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
)

func (i Issue) Columns() []api.ColumnDef {
	return []api.ColumnDef{clicky.Column("key").Label("Key").Build(), clicky.Column("title").Label("Title").Build(), clicky.Column("status").Label("Status").Build(), clicky.Column("assignee").Label("Assignee").Build()}
}

func (i Issue) Row() map[string]any {
	row := map[string]any{"key": clicky.Link(i.URL).AddText(i.Key), "title": i.Title, "status": "", "assignee": ""}
	if i.Status != nil {
		row["status"] = i.Status.Name
	}
	if i.Assignee != nil {
		row["assignee"] = i.Assignee.PrettyShort()
	}
	return row
}

type ListOptions struct {
	Site         string
	Catalogs     map[string]map[string]FieldDefinition
	IgnoreFields []string
}

func ParseList(raw json.RawMessage, options ListOptions) ([]Issue, error) {
	var response struct {
		Issues *[]json.RawMessage `json:"issues"`
	}
	if err := json.Unmarshal(raw, &response); err != nil {
		return nil, fmt.Errorf("decode Jira issue list: %w", err)
	}
	if response.Issues == nil {
		return nil, fmt.Errorf("Jira issue list has no issues array")
	}
	issues := make([]Issue, 0, len(*response.Issues))
	for _, rawIssue := range *response.Issues {
		var ref struct {
			Key string `json:"key"`
		}
		if err := json.Unmarshal(rawIssue, &ref); err != nil {
			return nil, fmt.Errorf("decode Jira issue row: %w", err)
		}
		issue, err := ParseIssue(rawIssue, ParseOptions{Site: options.Site, Catalog: options.Catalogs[ref.Key], IgnoreFields: options.IgnoreFields})
		if err != nil {
			return nil, err
		}
		issues = append(issues, issue)
	}
	return issues, nil
}
