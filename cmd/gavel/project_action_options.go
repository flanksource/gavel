package main

import (
	"fmt"
	"strconv"
	"time"

	ui "github.com/flanksource/gavel/pr/ui"
	"github.com/flanksource/gavel/testrunner/parsers"
	"github.com/spf13/cobra"
)

const projectActionSchemaVersion = 1

type projectActionCommandProvider struct{}

func (projectActionCommandProvider) Schema(action string) (ui.ProjectActionSchema, error) {
	command, positional, exclusions, err := projectActionSchemaCommand(action)
	if err != nil {
		return ui.ProjectActionSchema{}, err
	}
	schema := commandSchema(command, positional, commandSchemaOptions{Exclude: exclusions})
	defaults := projectActionDefaults(schema)
	if action == "commit" {
		defaults["precommit"] = "fail"
	}
	return ui.ProjectActionSchema{
		SchemaVersion: projectActionSchemaVersion,
		Action:        action,
		Schema:        map[string]any(schema),
		Defaults:      defaults,
	}, nil
}

func (provider projectActionCommandProvider) Args(action string, options map[string]any) ([]string, error) {
	definition, err := provider.Schema(action)
	if err != nil {
		return nil, err
	}
	properties, ok := definition.Schema["properties"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%s action schema has invalid properties", action)
	}
	for name := range options {
		if _, ok := properties[name]; !ok {
			return nil, fmt.Errorf("unsupported %s option %q", action, name)
		}
	}

	positional := projectActionPositional(action)
	order, _ := definition.Schema["x-order"].([]string)
	args := make([]string, 0, len(options))
	for _, name := range order {
		if name == positional {
			continue
		}
		value, present := options[name]
		if !present {
			continue
		}
		property, ok := properties[name].(fixtureSchemaProperty)
		if !ok {
			return nil, fmt.Errorf("%s option %q has invalid schema", action, name)
		}
		serialized, err := serializeProjectActionOption(name, property, value)
		if err != nil {
			return nil, fmt.Errorf("%s option %q: %w", action, name, err)
		}
		args = append(args, serialized...)
	}
	if value, present := options[positional]; present {
		paths, err := stringValues(value)
		if err != nil {
			return nil, fmt.Errorf("%s option %q: %w", action, positional, err)
		}
		args = append(args, paths...)
	}
	return args, nil
}

func projectActionSchemaCommand(action string) (*cobra.Command, string, map[string]bool, error) {
	command, err := schemaCommand(action)
	if err != nil {
		return nil, "", nil, err
	}
	switch action {
	case "commit":
		return command, "files", projectActionCommitExclusions(), nil
	case "lint":
		return command, "files", projectActionLintExclusions(), nil
	case "test":
		return command, "paths", projectActionTestExclusions(), nil
	default:
		return nil, "", nil, fmt.Errorf("unsupported project action %q", action)
	}
}

func projectActionPositional(action string) string {
	if action != "test" {
		return "files"
	}
	return "paths"
}

func projectActionDefaults(schema fixtureJSONSchema) map[string]any {
	properties, _ := schema["properties"].(map[string]any)
	defaults := make(map[string]any)
	for name, raw := range properties {
		property, ok := raw.(fixtureSchemaProperty)
		if !ok {
			continue
		}
		if value, present := property["default"]; present {
			defaults[name] = value
		}
	}
	return defaults
}

func serializeProjectActionOption(name string, property fixtureSchemaProperty, value any) ([]string, error) {
	switch property["type"] {
	case "boolean":
		boolean, ok := value.(bool)
		if !ok {
			return nil, fmt.Errorf("expected boolean, got %T", value)
		}
		return []string{"--" + name + "=" + strconv.FormatBool(boolean)}, nil
	case "integer":
		number, err := integerValue(value)
		if err != nil {
			return nil, err
		}
		return []string{"--" + name + "=" + strconv.FormatInt(number, 10)}, nil
	case "number":
		number, ok := value.(float64)
		if !ok {
			return nil, fmt.Errorf("expected number, got %T", value)
		}
		return []string{"--" + name + "=" + strconv.FormatFloat(number, 'g', -1, 64)}, nil
	case "array":
		values, err := projectActionArrayValues(name, property, value)
		if err != nil {
			return nil, err
		}
		args := make([]string, 0, len(values))
		for _, item := range values {
			args = append(args, "--"+name+"="+item)
		}
		return args, nil
	case "string":
		text, ok := value.(string)
		if !ok {
			return nil, fmt.Errorf("expected string, got %T", value)
		}
		if property["format"] == "duration" {
			if _, err := time.ParseDuration(text); err != nil {
				return nil, fmt.Errorf("invalid duration %q", text)
			}
		}
		if err := validateProjectActionEnum(property, text); err != nil {
			return nil, err
		}
		return []string{"--" + name + "=" + text}, nil
	default:
		return nil, fmt.Errorf("unsupported schema type %q", property["type"])
	}
}

func projectActionArrayValues(name string, property fixtureSchemaProperty, value any) ([]string, error) {
	values, err := anyValues(value)
	if err != nil {
		return nil, err
	}
	items, _ := property["items"].(fixtureSchemaProperty)
	serialized := make([]string, 0, len(values))
	for _, value := range values {
		switch items["type"] {
		case "integer":
			number, err := integerValue(value)
			if err != nil {
				return nil, err
			}
			serialized = append(serialized, strconv.FormatInt(number, 10))
		default:
			text, ok := value.(string)
			if !ok {
				return nil, fmt.Errorf("expected string array item, got %T", value)
			}
			if name == "framework" {
				if _, err := parsers.ParseFramework(text); err != nil {
					return nil, err
				}
			} else if err := validateProjectActionEnum(items, text); err != nil {
				return nil, err
			}
			serialized = append(serialized, text)
		}
	}
	return serialized, nil
}

func anyValues(value any) ([]any, error) {
	switch values := value.(type) {
	case []any:
		return values, nil
	case []string:
		result := make([]any, len(values))
		for index := range values {
			result[index] = values[index]
		}
		return result, nil
	default:
		return nil, fmt.Errorf("expected array, got %T", value)
	}
}

func stringValues(value any) ([]string, error) {
	values, err := anyValues(value)
	if err != nil {
		return nil, err
	}
	result := make([]string, 0, len(values))
	for _, value := range values {
		text, ok := value.(string)
		if !ok {
			return nil, fmt.Errorf("expected string array item, got %T", value)
		}
		result = append(result, text)
	}
	return result, nil
}

func integerValue(value any) (int64, error) {
	switch number := value.(type) {
	case int:
		return int64(number), nil
	case float64:
		if number != float64(int64(number)) {
			return 0, fmt.Errorf("expected integer, got %v", number)
		}
		return int64(number), nil
	default:
		return 0, fmt.Errorf("expected integer, got %T", value)
	}
}

func validateProjectActionEnum(property fixtureSchemaProperty, value string) error {
	values, ok := property["enum"].([]string)
	if !ok {
		return nil
	}
	for _, allowed := range values {
		if allowed == value {
			return nil
		}
	}
	return fmt.Errorf("unsupported value %q", value)
}

func projectActionTestExclusions() map[string]bool {
	return map[string]bool{
		"addr": true, "auto-stop": true, "detach": true, "idle-timeout": true,
		"ui": true, "work-dir": true,
	}
}

func projectActionCommitExclusions() map[string]bool {
	return map[string]bool{
		"auto-merge": true, "batch": true, "fixup": true, "interactive": true, "merge-type": true,
		"no-autosquash": true, "push": true, "since": true, "stage": true,
		"summary": true, "tree": true, "work-dir": true, "yes": true,
	}
}

func projectActionLintExclusions() map[string]bool {
	return map[string]bool{
		"addr": true, "ai-fix": true, "ai-fix-max-iterations": true, "allowed-tools": true,
		"api-key": true, "backend": true, "bare": true, "budget": true, "debug": true,
		"disallowed-tools": true, "edit": true, "effort": true, "fix": true, "hooks": true,
		"fallback": true, "max-tokens": true, "max-turns": true, "mcp": true, "memory": true,
		"mode": true, "model": true,
		"no-cache": true, "no-hooks": true, "no-mcp": true, "no-memory": true,
		"no-project": true, "no-skills": true, "no-user": true, "permission-mode": true,
		"profile": true, "project": true, "resume": true, "skill-dir": true, "skills": true,
		"temperature": true, "triage": true, "ui": true, "user": true, "work-dir": true,
		"yes": true,
	}
}
