package runconfig

import (
	"encoding/json"
	"fmt"
	"net/url"
	"path/filepath"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"time"
)

func Apply(target, options any, explicit map[string]bool) error {
	destination := reflect.ValueOf(target).Elem()
	value := reflect.ValueOf(options)
	for i := range value.NumField() {
		field, configured := value.Type().Field(i), value.Field(i)
		if configured.IsNil() || explicit[key(field)] {
			continue
		}
		actual := destination.FieldByName(field.Name)
		if !actual.IsValid() || !actual.CanSet() {
			return fmt.Errorf("configured option %s has no destination in %T", key(field), target)
		}
		if configured.Kind() == reflect.Pointer && actual.Kind() != reflect.Pointer {
			configured = configured.Elem()
		}
		if actual.Type() == reflect.TypeFor[time.Duration]() {
			duration, err := time.ParseDuration(configured.String())
			if err != nil {
				return fmt.Errorf("option %s: %w", key(field), err)
			}
			actual.SetInt(int64(duration))
		} else if configured.Type().ConvertibleTo(actual.Type()) {
			actual.Set(configured.Convert(actual.Type()))
		} else {
			return fmt.Errorf("configured option %s cannot project %s into %s", key(field), configured.Type(), actual.Type())
		}
	}
	configured := value.FieldByName("Ignore")
	workDir := destination.FieldByName("WorkDir")
	if configured.IsValid() && !configured.IsNil() && !explicit["ignore"] && workDir.IsValid() && workDir.String() != "" {
		patterns := make([]string, configured.Len())
		for j := range configured.Len() {
			patterns[j] = configured.Index(j).String()
			if filepath.IsAbs(patterns[j]) {
				rel, err := filepath.Rel(workDir.String(), patterns[j])
				if err != nil {
					return err
				}
				patterns[j] = filepath.ToSlash(rel)
			}
		}
		destination.FieldByName("Ignore").Set(reflect.ValueOf(patterns))
	}
	return nil
}

func Validate(options any, tool bool) error {
	value := reflect.ValueOf(options)
	for i := range value.NumField() {
		field, configured := value.Type().Field(i), value.Field(i)
		if configured.IsNil() {
			continue
		}
		if tool && field.Tag.Get("scope") == "run" {
			return fmt.Errorf("option %s applies to the entire run and cannot be set per tool", key(field))
		}
		if configured.Kind() == reflect.Pointer {
			configured = configured.Elem()
		}
		if field.Tag.Get("duration") != "" {
			if duration, err := time.ParseDuration(configured.String()); err != nil || duration < 0 {
				return fmt.Errorf("option %s requires a nonnegative duration, got %q", key(field), configured.String())
			}
		}
		if choices := field.Tag.Get("enum"); choices != "" && !slices.Contains(strings.Split(choices, ","), configured.String()) {
			return fmt.Errorf("option %s must be one of %s", key(field), choices)
		}
		if minimum := field.Tag.Get("min"); minimum != "" {
			limit, err := strconv.ParseInt(minimum, 10, 64)
			if err != nil {
				return err
			}
			if configured.Int() < limit {
				return fmt.Errorf("option %s must be at least %d", key(field), limit)
			}
		}
	}
	return nil
}

func ResolvePaths(options any, base string) {
	value := reflect.ValueOf(options).Elem()
	for i := range value.NumField() {
		field := value.Field(i)
		if value.Type().Field(i).Tag.Get("path") == "" || field.IsNil() {
			continue
		}
		resolve := func(path string) string {
			if path == "" || filepath.IsAbs(path) {
				return path
			}
			if value.Type().Field(i).Tag.Get("path") == "uri" {
				if parsed, err := url.Parse(path); err == nil && parsed.IsAbs() {
					return path
				}
			}
			return filepath.Join(base, path)
		}
		if field.Kind() == reflect.Pointer {
			field.Elem().SetString(resolve(field.Elem().String()))
		} else {
			for j := range field.Len() {
				field.Index(j).SetString(resolve(field.Index(j).String()))
			}
		}
	}
}

func Present(options any) map[string]bool {
	fields := map[string]bool{}
	value := reflect.ValueOf(options)
	for i := range value.NumField() {
		field := value.Type().Field(i)
		if !field.IsExported() || value.Field(i).IsZero() {
			continue
		}
		name := field.Tag.Get("flag")
		if name == "" {
			name = strings.Split(field.Tag.Get("yaml"), ",")[0]
		}
		if name != "" && name != "-" {
			fields[name] = true
		}
	}
	return fields
}

func JSONFields(data []byte) (map[string]bool, error) {
	var document map[string]json.RawMessage
	if err := json.Unmarshal(data, &document); err != nil {
		return nil, err
	}
	fields := make(map[string]bool, len(document))
	for name := range document {
		fields[strings.ReplaceAll(name, "_", "-")] = true
	}
	return fields, nil
}

func key(field reflect.StructField) string { return strings.Split(field.Tag.Get("json"), ",")[0] }
