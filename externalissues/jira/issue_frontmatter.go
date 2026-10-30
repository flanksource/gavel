package jira

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/flanksource/gavel/externalissues/syncdoc"
)

// maxFrontmatterText bounds a rich-text field copied into frontmatter: a single
// line of at most this many characters. Longer text stays in the Markdown body only.
const maxFrontmatterText = 200

// promotedFields are written as their own gavel.issue keys, never again under fields.
var promotedFields = map[string]bool{"resolution": true, "resolutiondate": true, "duedate": true, "components": true, "statuscategory": true}

// syncMetadata is the issue as the gavel.issue frontmatter block.
func (i Issue) syncMetadata() (*syncdoc.Issue, error) {
	block := &syncdoc.Issue{Labels: i.Labels, Created: wholeSeconds(i.Created), Updated: wholeSeconds(i.Updated)}
	if i.Type != nil {
		block.Type = i.Type.Name
	}
	if i.Status != nil {
		block.Status, block.StatusCategory = i.Status.Name, i.Status.Category
	}
	if i.Priority != nil {
		block.Priority = i.Priority.Name
	}
	if i.Assignee != nil {
		block.Assignee = i.Assignee.Name
	}
	if i.Reporter != nil {
		block.Reporter = i.Reporter.Name
	}
	if i.Project != nil {
		block.Project = i.Project.Name
	}
	if err := i.promoteFields(block); err != nil {
		return nil, fmt.Errorf("issue %s: %w", i.Key, err)
	}
	fields, err := i.frontmatterFields()
	if err != nil {
		return nil, err
	}
	block.Fields = fields
	return block, nil
}

func (i Issue) promoteFields(block *syncdoc.Issue) error {
	var resolution struct{ Name string }
	var resolved *string
	var components []struct{ Name string }
	for id, target := range map[string]any{"resolution": &resolution, "resolutiondate": &resolved, "duedate": &block.Due, "components": &components} {
		if field, ok := i.Fields[id]; ok {
			if err := json.Unmarshal(field.Value, target); err != nil {
				return fmt.Errorf("field %s: %w", id, err)
			}
		}
	}
	at, err := parseTime(resolved, "resolutiondate")
	if err != nil {
		return err
	}
	block.Resolution, block.Resolved = resolution.Name, wholeSeconds(at)
	for _, component := range components {
		block.Components = append(block.Components, component.Name)
	}
	return nil
}

func (i Issue) frontmatterFields() (map[string]any, error) {
	labels := i.fieldLabels()
	fields := make(map[string]any)
	for id, field := range i.Fields {
		if promotedFields[strings.ToLower(id)] {
			continue
		}
		value, err := frontmatterField(id, field.Format, field.Schema, field.Value)
		if err != nil {
			return nil, fmt.Errorf("issue %s field %s (%s): %w", i.Key, labels[id], id, err)
		}
		if value != nil {
			fields[labels[id]] = value
		}
	}
	if len(fields) == 0 {
		return nil, nil
	}
	return fields, nil
}

func frontmatterField(id, format string, schema, raw json.RawMessage) (any, error) {
	switch {
	case format == "markdown":
		var text string
		if err := json.Unmarshal(raw, &text); err != nil {
			return nil, err
		}
		if text = strings.TrimSpace(text); strings.Contains(text, "\n") || utf8.RuneCountInString(text) > maxFrontmatterText {
			return nil, nil
		}
		return frontmatterValue(text, false)
	case id == "parent" && string(raw) != "null":
		var reference issueReference
		if err := json.Unmarshal(raw, &reference); err != nil {
			return nil, err
		}
		return frontmatterValue(reference.Key, false)
	case id == "issuelinks":
		return relationshipLabels(raw)
	}
	var kind struct{ Type string }
	if len(schema) > 0 {
		if err := json.Unmarshal(schema, &kind); err != nil {
			return nil, err
		}
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return nil, err
	}
	return frontmatterValue(value, kind.Type == "datetime")
}

// frontmatterValue reduces a decoded Jira value to typed YAML data: objects that
// name a user, option or version become that name, empty values become nil.
func frontmatterValue(value any, datetime bool) (any, error) {
	switch value := value.(type) {
	case nil:
		return nil, nil
	case string:
		if value == "" {
			return nil, nil
		}
		if !datetime {
			return value, nil
		}
		at, err := parseTime(&value, "datetime")
		if err != nil {
			return nil, err
		}
		return at.Format(time.RFC3339Nano), nil
	case bool:
		return value, nil
	case json.Number:
		if number, err := value.Int64(); err == nil {
			return number, nil
		}
		return value.Float64()
	case []any:
		var items []any
		for _, item := range value {
			reduced, err := frontmatterValue(item, datetime)
			if err != nil {
				return nil, err
			}
			if reduced != nil {
				items = append(items, reduced)
			}
		}
		if len(items) == 0 {
			return nil, nil
		}
		return items, nil
	case map[string]any:
		return frontmatterObject(value)
	default:
		return nil, fmt.Errorf("unsupported Jira field value %T", value)
	}
}

func frontmatterObject(values map[string]any) (any, error) {
	if count, ok := fieldCount(values); ok {
		return frontmatterValue(count, false)
	}
	if label, ok := values["value"].(string); ok {
		label, err := optionLabel(values, label)
		if err != nil {
			return nil, err
		}
		return frontmatterValue(label, false)
	}
	for _, key := range fieldLabelKeys {
		if label, ok := values[key].(string); ok && label != "" {
			return label, nil
		}
	}
	object := make(map[string]any)
	for key, value := range values {
		reduced, err := frontmatterValue(value, false)
		if err != nil {
			return nil, fmt.Errorf("property %s: %w", key, err)
		}
		if reduced != nil {
			object[key] = reduced
		}
	}
	if len(object) == 0 {
		return nil, nil
	}
	return object, nil
}

func relationshipLabels(raw json.RawMessage) (any, error) {
	var relationships []issueRelationship
	if err := json.Unmarshal(raw, &relationships); err != nil {
		return nil, err
	}
	var items []any
	for _, relationship := range relationships {
		reference, label, err := relationship.target()
		if err != nil {
			return nil, err
		}
		items = append(items, label+" "+reference.Key)
	}
	if len(items) == 0 {
		return nil, nil
	}
	return items, nil
}

func wholeSeconds(at *time.Time) *time.Time {
	if at == nil {
		return nil
	}
	truncated := at.Truncate(time.Second)
	return &truncated
}
