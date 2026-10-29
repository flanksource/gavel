package jira

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/gavel/externalissues"
)

type issueReference struct {
	Key    string `json:"key"`
	Fields struct {
		Summary string         `json:"summary"`
		Status  *statusPayload `json:"status"`
	} `json:"fields"`
}

type issueRelationship struct {
	Type struct {
		Inward  string `json:"inward"`
		Outward string `json:"outward"`
	} `json:"type"`
	Inward  *issueReference `json:"inwardIssue"`
	Outward *issueReference `json:"outwardIssue"`
}

func (i Issue) semanticField(field externalissues.Field, id string, options RenderOptions) (fieldView, error) {
	if string(field.Value) == "null" {
		return fieldView{}, nil
	}
	switch id {
	case "resolution":
		var resolution struct{ Name, Description string }
		if err := json.Unmarshal(field.Value, &resolution); err != nil {
			return fieldView{}, err
		}
		if resolution.Name == "" {
			return fieldView{}, fmt.Errorf("resolution requires a name")
		}
		short := clicky.Text(resolution.Name)
		full := short
		if resolution.Description != "" {
			full = full.AddText(" — " + resolution.Description)
		}
		return fieldView{short: short, full: full}, nil
	case "parent":
		var reference issueReference
		if err := json.Unmarshal(field.Value, &reference); err != nil {
			return fieldView{}, err
		}
		text, err := i.referenceText(reference, options)
		return fieldView{short: text, full: text}, err
	case "issuelinks":
		return i.relationshipsField(field.Value, options)
	}
	if len(field.Schema) > 0 {
		var schema struct {
			Custom string `json:"custom"`
		}
		if err := json.Unmarshal(field.Schema, &schema); err != nil {
			return fieldView{}, err
		}
		if schema.Custom == "com.pyxis.greenhopper.jira:gh-epic-link" {
			var key string
			if err := json.Unmarshal(field.Value, &key); err != nil {
				return fieldView{}, err
			}
			text, err := i.referenceText(issueReference{Key: key}, options)
			return fieldView{short: text, full: text}, err
		}
	}
	return renderFieldValue(field.Value)
}

func (i Issue) referenceText(reference issueReference, options RenderOptions) (api.Textable, error) {
	if !syncComponent(reference.Key) || strings.TrimSpace(reference.Key) == "" {
		return nil, fmt.Errorf("issue reference requires a safe issue key")
	}
	if i.Site == "" {
		return nil, fmt.Errorf("issue reference %s requires a Jira site", reference.Key)
	}
	address := strings.TrimRight(i.Site, "/") + "/browse/" + url.PathEscape(reference.Key)
	if options.ResolveIssueURL != nil {
		var err error
		address, err = options.ResolveIssueURL(address)
		if err != nil {
			return nil, err
		}
	}
	label := reference.Key
	if reference.Fields.Summary != "" {
		label += " — " + reference.Fields.Summary
	}
	text := api.Text{}.Add(clicky.Link(address).AddText(label))
	status, err := reference.Fields.Status.status()
	if err != nil {
		return nil, err
	}
	if status != nil && status.Name != "" {
		text = text.AddText(" · ").Add(status.PrettyShort())
	}
	return text, nil
}

func (i Issue) relationshipsField(raw json.RawMessage, options RenderOptions) (fieldView, error) {
	var relationships []issueRelationship
	if err := json.Unmarshal(raw, &relationships); err != nil {
		return fieldView{}, err
	}
	var items api.TextList
	for _, relationship := range relationships {
		if (relationship.Inward == nil) == (relationship.Outward == nil) {
			return fieldView{}, fmt.Errorf("issue link requires exactly one inwardIssue or outwardIssue")
		}
		reference, label := relationship.Inward, relationship.Type.Inward
		if relationship.Outward != nil {
			reference, label = relationship.Outward, relationship.Type.Outward
		}
		if label == "" {
			return fieldView{}, fmt.Errorf("issue link requires a relationship label")
		}
		text, err := i.referenceText(*reference, options)
		if err != nil {
			return fieldView{}, err
		}
		items = append(items, api.Text{}.AddText(label+" ").Add(text))
	}
	if len(items) == 0 {
		return fieldView{}, nil
	}
	text := clicky.List(items...)
	return fieldView{short: text, full: text}, nil
}
