package main

import (
	"fmt"
	"strconv"
	"time"

	ui "github.com/flanksource/gavel/pr/ui"
	"github.com/flanksource/gavel/schemas"
)

const projectActionSchemaVersion = 2

type projectActionCommandProvider struct{}

func (projectActionCommandProvider) Schema(action string) (ui.ProjectActionSchema, error) {
	document, err := schemas.Load(action)
	if err != nil {
		return ui.ProjectActionSchema{}, err
	}
	defaults := make(map[string]any)
	for name, raw := range document.Properties {
		property := raw.(map[string]any)
		if value, present := property["default"]; present {
			defaults[name] = value
		}
	}
	return ui.ProjectActionSchema{
		SchemaVersion: projectActionSchemaVersion,
		Action:        action,
		Schema:        document.Schema,
		Defaults:      defaults,
	}, nil
}

func (projectActionCommandProvider) Args(action string, options map[string]any) ([]string, error) {
	document, err := schemas.Load(action)
	if err != nil {
		return nil, err
	}
	for name := range options {
		if _, ok := document.Properties[name]; !ok {
			return nil, fmt.Errorf("unsupported %s option %q", action, name)
		}
	}
	if err := document.Validate(options); err != nil {
		return nil, fmt.Errorf("invalid %s options: %w", action, err)
	}

	positional := projectActionPositional(action)
	args := make([]string, 0, len(options))
	for _, name := range document.Order {
		if name == positional {
			continue
		}
		value, present := options[name]
		if !present {
			continue
		}
		property := document.Properties[name].(map[string]any)
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

func projectActionPositional(action string) string {
	if action != "test" {
		return "files"
	}
	return "paths"
}

func serializeProjectActionOption(name string, property map[string]any, value any) ([]string, error) {
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
		values, err := projectActionArrayValues(property, value)
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
		return []string{"--" + name + "=" + text}, nil
	default:
		return nil, fmt.Errorf("unsupported schema type %q", property["type"])
	}
}

func projectActionArrayValues(property map[string]any, value any) ([]string, error) {
	values, err := anyValues(value)
	if err != nil {
		return nil, err
	}
	items := property["items"].(map[string]any)
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
