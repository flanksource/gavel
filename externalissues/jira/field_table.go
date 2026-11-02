package jira

import (
	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
)

type fieldRow struct {
	Name  string
	Value api.Textable
}

func (fieldRow) Columns() []api.ColumnDef {
	return []api.ColumnDef{clicky.Column("field").Label("Field").Build(), clicky.Column("value").Label("Value").Build()}
}

func (f fieldRow) Row() map[string]any {
	return map[string]any{"field": f.Name, "value": f.Value}
}
