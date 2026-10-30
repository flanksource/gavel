package jira

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
)

// fieldLabelKeys are the properties that name a Jira user, option, version or component, in precedence order.
var fieldLabelKeys = []string{"displayName", "name", "value"}

type fieldView struct {
	short api.Textable
	full  api.Textable
}

func (v fieldView) PrettyShort() api.Textable { return v.short }
func (v fieldView) PrettyFull() api.Textable  { return v.full }

func renderFieldValue(raw json.RawMessage) (fieldView, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return fieldView{}, err
	}
	return fieldValueText(value)
}

func fieldValueText(value any) (fieldView, error) {
	var text api.Textable
	switch value := value.(type) {
	case nil:
		return fieldView{}, nil
	case string:
		if value == "" {
			value = `""`
		}
		text = clicky.Text(value)
	case bool, json.Number:
		text = clicky.Text(fmt.Sprint(value))
	case []any:
		return fieldList(value)
	case map[string]any:
		return fieldObject(value)
	default:
		return fieldView{}, fmt.Errorf("unsupported Jira field value %T", value)
	}
	return fieldView{short: text, full: text}, nil
}

func fieldList(values []any) (fieldView, error) {
	var short, full api.TextList
	structured := false
	for _, value := range values {
		item, err := fieldValueText(value)
		if err != nil {
			return fieldView{}, err
		}
		if item.short == nil {
			continue
		}
		short = append(short, item.PrettyShort())
		full = append(full, item.PrettyFull())
		switch value.(type) {
		case map[string]any, []any:
			structured = true
		}
	}
	if len(short) == 0 {
		return fieldView{}, nil
	}
	var summary api.Textable = clicky.CompactList(short)
	if structured {
		summary = clicky.List(short...)
	}
	return fieldView{short: summary, full: clicky.List(full...)}, nil
}

func fieldObject(values map[string]any) (fieldView, error) {
	if len(values) == 0 {
		return fieldValueText("{}")
	}
	if label, ok := values["value"].(string); ok {
		return fieldOption(values, label)
	}
	if view, ok, err := semanticFieldObject(values); ok || err != nil {
		return view, err
	}
	if address, ok := values["self"].(string); ok && address != "" {
		if count, ok := fieldCount(values); ok {
			link := clicky.Link(address).AddText(string(count))
			return fieldView{short: link, full: link}, nil
		}
	}
	short := make(map[string]api.Textable, len(values))
	full := make(map[string]api.Textable, len(values))
	for key, value := range values {
		item, err := fieldValueText(value)
		if err != nil {
			return fieldView{}, err
		}
		if item.short == nil {
			continue
		}
		short[key], full[key] = item.PrettyShort(), item.PrettyFull()
	}
	if len(short) == 0 {
		return fieldView{}, nil
	}
	var summary api.Textable = clicky.KeyMap(short)
	for _, key := range fieldLabelKeys {
		label, ok := values[key].(string)
		if !ok || label == "" {
			continue
		}
		summary = clicky.Text(label)
		if link, ok := values["url"].(string); ok && link != "" {
			summary = clicky.Link(link).AddText(label)
			full[key] = summary
		}
		break
	}
	return fieldView{short: summary, full: clicky.KeyMap(full)}, nil
}

func fieldOption(values map[string]any, label string) (fieldView, error) {
	label, err := optionLabel(values, label)
	if err != nil {
		return fieldView{}, err
	}
	return fieldValueText(label)
}

// optionLabel joins a cascading Jira option's values, parent first: "Region / East".
func optionLabel(values map[string]any, label string) (string, error) {
	if values["child"] == nil {
		return label, nil
	}
	child, ok := values["child"].(map[string]any)
	if !ok {
		return "", fmt.Errorf("Jira option child must be an object, received %T", values["child"])
	}
	childLabel, ok := child["value"].(string)
	if !ok {
		return "", fmt.Errorf("Jira option child value must be a string, received %T", child["value"])
	}
	descendant, err := optionLabel(child, childLabel)
	return label + " / " + descendant, err
}

// fieldCount is the count a Jira votes or watchers object stands for.
func fieldCount(values map[string]any) (json.Number, bool) {
	for _, key := range []string{"votes", "watchCount"} {
		if count, ok := values[key].(json.Number); ok {
			return count, true
		}
	}
	return "", false
}
