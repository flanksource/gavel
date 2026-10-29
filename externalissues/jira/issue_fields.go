package jira

import (
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/clicky/markdown"
	"github.com/flanksource/gavel/externalissues"
)

func (i Issue) renderFields(options RenderOptions) (api.Textable, error) {
	labels := i.fieldLabels()
	rows := make([]fieldRow, 0, len(labels))
	var sections api.Text
	for _, id := range slices.SortedFunc(maps.Keys(labels), func(a, b string) int { return strings.Compare(labels[a], labels[b]) }) {
		value, checklist, err := i.renderField(i.Fields[id], id, options)
		if err != nil {
			return nil, fmt.Errorf("issue %s field %s: %w", i.Key, id, err)
		}
		if value == nil {
			continue
		}
		if options.Full && checklist {
			sections = sections.NewLine().NewLine().Add(clicky.Heading(2, clicky.Text(labels[id]))).NewLine().NewLine().Add(value)
		} else {
			rows = append(rows, fieldRow{Name: labels[id], Value: value})
		}
	}
	if len(sections.Children) == 0 {
		if len(rows) == 0 {
			return api.Text{}, nil
		}
		return api.NewTableFrom(rows), nil
	}
	if len(rows) == 0 {
		return sections, nil
	}
	return api.Text{}.Add(api.NewTableFrom(rows)).Add(sections), nil
}

func (i Issue) fieldLabels() map[string]string {
	ids := make([]string, 0, len(i.Fields))
	counts := make(map[string]int)
	for id, field := range i.Fields {
		ids = append(ids, id)
		label := field.Name
		if label == "" {
			label = id
		}
		counts[label]++
	}
	slices.Sort(ids)
	labels, used := make(map[string]string), make(map[string]bool)
	for _, id := range ids {
		label := i.Fields[id].Name
		if label == "" {
			label = id
		}
		if counts[label] > 1 {
			label += " [" + id + "]"
		}
		for used[label] {
			label += " [" + id + "]"
		}
		labels[id], used[label] = label, true
	}
	return labels
}

func (i Issue) renderField(field externalissues.Field, id string, options RenderOptions) (api.Textable, bool, error) {
	view, err := i.semanticField(field, id, options)
	if err != nil {
		return nil, false, err
	}
	value := view.PrettyShort()
	if options.Full {
		value = view.PrettyFull()
	}
	if value == nil {
		return nil, false, nil
	}
	checklist := strings.Contains(strings.ToLower(field.Name), "checklist")
	if field.Format != "markdown" && !checklist {
		return value, false, nil
	}
	var source string
	if err := json.Unmarshal(field.Value, &source); err != nil {
		if field.Format == "markdown" {
			return nil, false, err
		}
		return value, false, nil
	}
	if source == "" {
		return clicky.Text(`""`), false, nil
	}
	doc, err := markdown.ParseString(source, markdown.WithFrontmatter(false))
	if err != nil {
		return nil, false, err
	}
	if checklist {
		doc.Root = checklistDocument(doc.Root)
	}
	doc.Root, err = resolveDocumentLinks(doc.Root, options.ResolveIssueURL)
	return doc, hasChecklist(doc.Root), err
}
