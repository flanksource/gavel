package verify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strings"

	"github.com/flanksource/commons/logger"
	"gopkg.in/yaml.v3"
)

// UnknownFieldPolicy controls schema drift, never type or semantic validation.
type UnknownFieldPolicy uint8

const (
	UnknownFieldsDefault UnknownFieldPolicy = iota
	UnknownFieldsWarn
	UnknownFieldsError
)

const DefaultUnknownFieldPolicy = UnknownFieldsWarn

type DecodeOptions struct {
	Source string
	Policy UnknownFieldPolicy
	// Warn defaults to the application logger.
	Warn func(string)
}

// DecodeFields lets custom decoders expose their object wire shape for field
// inspection. Return a struct or alias without custom unmarshalling methods.
// Custom decoders without this contract remain responsible for their own keys.
type DecodeFields interface {
	DecodeFields() any
}

func DecodeJSON(data []byte, target any, options DecodeOptions) error {
	if err := options.validate(target); err != nil {
		return err
	}
	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("%s: %w", options.source(), err)
	}
	var document yaml.Node
	if err := yaml.Unmarshal(data, &document); err != nil {
		return fmt.Errorf("%s: %w", options.source(), err)
	}
	return options.check(&document, reflect.TypeOf(target), "json")
}

func DecodeYAML(data []byte, target any, options DecodeOptions) error {
	if err := options.validate(target); err != nil {
		return err
	}
	document, err := decodeYAMLDocument(data)
	if err != nil {
		return fmt.Errorf("%s: %w", options.source(), err)
	}
	if err := document.Decode(target); err != nil {
		return fmt.Errorf("%s: %w", options.source(), err)
	}
	return options.check(document, reflect.TypeOf(target), "yaml")
}

func decodeYAMLDocument(data []byte) (*yaml.Node, error) {
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	var document yaml.Node
	if err := decoder.Decode(&document); err != nil && err != io.EOF {
		return nil, err
	}
	var extra yaml.Node
	if err := decoder.Decode(&extra); err != io.EOF {
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("expected one YAML document")
	}
	return &document, nil
}

func (o DecodeOptions) validate(target any) error {
	if o.Policy != UnknownFieldsDefault && o.Policy != UnknownFieldsWarn && o.Policy != UnknownFieldsError {
		return fmt.Errorf("%s: invalid unknown field policy %d", o.source(), o.Policy)
	}
	value := reflect.ValueOf(target)
	if !value.IsValid() || value.Kind() != reflect.Pointer || value.IsNil() {
		return fmt.Errorf("%s: decode target must be a non-nil pointer", o.source())
	}
	return nil
}

func (o DecodeOptions) source() string {
	if o.Source == "" {
		return "input"
	}
	return o.Source
}

func (o DecodeOptions) check(document *yaml.Node, target reflect.Type, format string) error {
	if o.Policy == UnknownFieldsDefault {
		o.Policy = DefaultUnknownFieldPolicy
	}
	fields := fieldInspector{format: format, unknown: map[string]bool{}}
	fields.walk(document, target, "")
	var warnings []string
	for path := range fields.unknown {
		warnings = append(warnings, fmt.Sprintf("%s: unknown field %s", o.source(), path))
	}
	sort.Strings(warnings)
	if len(warnings) > 0 && o.Policy == UnknownFieldsError {
		return fmt.Errorf("%s", strings.Join(warnings, "\n"))
	}
	for _, warning := range warnings {
		if o.Warn != nil {
			o.Warn(warning)
		} else {
			logger.Warnf("%s", warning)
		}
	}
	return nil
}
