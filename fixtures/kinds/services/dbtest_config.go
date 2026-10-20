package services

import (
	"encoding/json"
	"fmt"

	"github.com/flanksource/gavel/fixtures"
)

type DBTestConfig struct {
	Default    *bool            `json:"default,omitempty"`
	Migrations *MigrationConfig `json:"migrations,omitempty"`
}

type MigrationConfig struct {
	Dir       string                     `json:"dir"`
	Name      string                     `json:"name,omitempty"`
	Schema    string                     `json:"schema,omitempty"`
	Variables map[string]json.RawMessage `json:"variables,omitempty"`
}

type databaseSelection struct {
	config   map[string]any
	selected *bool
}

type databasePlan struct {
	databases []databaseSelection
	entries   map[int][]rawEntry
}

func collectDatabases(steps []fixtures.StepSpec) (databasePlan, error) {
	plan := databasePlan{entries: make(map[int][]rawEntry)}
	for i, step := range steps {
		if step.Kind != "services" {
			continue
		}
		raws, err := parseConfig(step.Config)
		if err != nil {
			return plan, err
		}
		for _, raw := range raws {
			config, exists := raw.body["dbtest"]
			if !exists {
				continue
			}
			entry, err := decodeEntry(raw.name, raw.body)
			if err != nil {
				return plan, err
			}
			if step.ID == "" {
				return plan, fmt.Errorf("a services step needs an id")
			}
			object, ok := config.(map[string]any)
			if !ok {
				return plan, fmt.Errorf("entry %q: dbtest must be an object", raw.name)
			}
			plan.databases = append(plan.databases, databaseSelection{config: object, selected: entry.DBTest.Default})
			plan.entries[i] = raws
		}
	}
	return plan, nil
}

func (Kind) PrepareSteps(steps []fixtures.StepSpec) error {
	plan, err := collectDatabases(steps)
	if err != nil {
		return err
	}
	selected := 0
	for _, db := range plan.databases {
		if db.selected != nil && *db.selected {
			selected++
		}
	}
	if len(plan.databases) > 1 && selected != 1 {
		return fmt.Errorf("multiple databases require exactly one dbtest.default: true (found %d)", selected)
	}
	for _, db := range plan.databases {
		value := len(plan.databases) == 1
		if db.selected != nil {
			value = *db.selected
		}
		db.config["default"] = value
	}
	for i, raws := range plan.entries {
		list := make([]map[string]any, 0, len(raws))
		for _, raw := range raws {
			raw.body["name"] = raw.name
			list = append(list, raw.body)
		}
		encoded, err := json.Marshal(map[string]any{"services": list})
		if err != nil {
			return err
		}
		steps[i].Config = encoded
	}
	return nil
}
