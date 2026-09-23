package entity

import (
	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/gavel/todos/labels"
	"github.com/flanksource/gavel/todos/query"
	"github.com/flanksource/gavel/todos/types"
)

// Summary is a TODO as a list returns it: what a caller scanning a backlog
// needs to pick one, and nothing else. A list across every project that carried
// each TODO's body, plan and verification ran to half a megabyte a page — far
// more than an agent can be handed. `todo get` returns the whole record.
type Summary struct {
	// ID is the short id, which resolves back to the TODO wherever a full id is
	// accepted.
	ID       string         `json:"id"`
	Title    string         `json:"title"`
	Status   types.Status   `json:"status"`
	Priority types.Priority `json:"priority"`
	Labels   []string       `json:"labels,omitempty"`
	// Project is the short name of the project the TODO belongs to — the name a
	// list is filtered by.
	Project string `json:"project,omitempty"`

	// labelDefinitions keep a rendered row's label colours the ones the
	// provider resolved; they are presentation, so they stay off the wire.
	labelDefinitions labels.Definitions
}

func summarize(todo *types.TODO) Summary {
	return Summary{
		ID:               todo.DisplayID(),
		Title:            todo.Title,
		Status:           todo.Status,
		Priority:         todo.Priority,
		Labels:           todo.Labels,
		Project:          query.ShortProjectName(todo.Workspace),
		labelDefinitions: todo.LabelDefinitions,
	}
}

func summarizeAll(todos types.TODOS) []Summary {
	summaries := make([]Summary, len(todos))
	for i, todo := range todos {
		summaries[i] = summarize(todo)
	}
	return summaries
}

func (s Summary) GetID() string   { return s.ID }
func (s Summary) GetName() string { return s.Title }

// summaryColumns are the TODO row cells a summary keeps; the rest describe the
// run history a summary leaves out.
var summaryColumns = []string{"ID", "Title", "Status", "Priority", "Labels"}

// PrettyRow renders the cells a full TODO row renders for the fields a summary
// keeps, so a list reads the same in a table whichever shape it came from, and
// names the project by its short name.
func (s Summary) PrettyRow(opts interface{}) map[string]api.Text {
	todo := types.TODO{ShortID: s.ID, Labels: s.Labels, LabelDefinitions: s.labelDefinitions}
	todo.Title, todo.Status, todo.Priority = s.Title, s.Status, s.Priority
	full := todo.PrettyRow(opts)
	row := make(map[string]api.Text, len(summaryColumns)+1)
	for _, column := range summaryColumns {
		if cell, ok := full[column]; ok {
			row[column] = cell
		}
	}
	row["Project"] = clicky.Text(s.Project, "order-1 text-muted")
	return row
}
