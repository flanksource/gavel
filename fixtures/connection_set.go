package fixtures

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/flanksource/commons/logger"
	"github.com/goccy/go-yaml"
)

// ConnectionSet is a fixture's `connections:` declaration. The list form names
// connections (catalog names or connection:// references) that each become the
// default for their resolved type (Defaults); the map form declares named
// connections, each a string reference or an inline commons-db connection
// object (Named). A declaration is one form or the other.
type ConnectionSet struct {
	Defaults []string
	Named    map[string]any
}

// NamedKeys are the map form's names, sorted.
func (s *ConnectionSet) NamedKeys() []string {
	keys := make([]string, 0, len(s.Named))
	for key := range s.Named {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func (s *ConnectionSet) UnmarshalJSON(b []byte) error {
	var raw any
	if err := json.Unmarshal(b, &raw); err != nil {
		return fmt.Errorf("connections: %w", err)
	}
	return s.parse(raw)
}

func (s *ConnectionSet) UnmarshalYAML(unmarshal func(any) error) error {
	var raw any
	if err := unmarshal(&raw); err != nil {
		return fmt.Errorf("connections: %w", err)
	}
	return s.parse(raw)
}

func (s ConnectionSet) MarshalJSON() ([]byte, error) { return json.Marshal(s.declared()) }

func (s ConnectionSet) MarshalYAML() (any, error) { return s.declared(), nil }

// declared is the set in the form it was declared in.
func (s ConnectionSet) declared() any {
	if s.Named != nil {
		return s.Named
	}
	return s.Defaults
}

func (s *ConnectionSet) parse(raw any) error {
	switch value := raw.(type) {
	case []any:
		defaults := make([]string, 0, len(value))
		for i, entry := range value {
			ref, ok := entry.(string)
			if !ok || ref == "" {
				return fmt.Errorf("connections[%d] must be a connection name or connection:// reference, got %T %v", i, entry, entry)
			}
			defaults = append(defaults, ref)
		}
		*s = ConnectionSet{Defaults: defaults}
	case map[string]any:
		named := make(map[string]any, len(value))
		for name, entry := range value {
			if err := checkNamedConnection(name, entry); err != nil {
				return err
			}
			named[name] = entry
		}
		*s = ConnectionSet{Named: named}
	default:
		return fmt.Errorf("connections must be a list of connection references or a map of name to reference or inline connection, got %T %v", raw, raw)
	}
	return nil
}

func checkNamedConnection(name string, entry any) error {
	switch value := entry.(type) {
	case string:
		if value != "" {
			return nil
		}
	case map[string]any:
		return nil
	}
	return fmt.Errorf("connections.%s must be a connection name, connection:// reference or inline connection object, got %T %v", name, entry, entry)
}

// legacyConnectionKey is the retired single-connection front-matter key.
const legacyConnectionKey = "connection"

// unmarshalFrontMatter decodes a markdown document's front matter, reading the
// retired `connection: <ref>` key as `connections: [<ref>]`.
func unmarshalFrontMatter(data []byte) (*FrontMatter, error) {
	var frontMatter FrontMatter
	if err := yaml.Unmarshal(data, &frontMatter); err != nil {
		return nil, fmt.Errorf("failed to parse YAML front-matter: %w", err)
	}
	legacy, ok := frontMatter.Metadata[legacyConnectionKey]
	if !ok {
		return &frontMatter, nil
	}
	ref, isString := legacy.(string)
	switch {
	case !isString || ref == "":
		return nil, fmt.Errorf("front-matter connection: must be a connection name or connection:// reference, got %T %v; declare an inline connection under connections: {<name>: {...}}", legacy, legacy)
	case frontMatter.Connections != nil:
		return nil, fmt.Errorf("front matter declares both connection: and connections: (move %q into connections)", ref)
	}
	logger.Warnf("front-matter connection: %s is deprecated, use connections: [%s]", ref, ref)
	frontMatter.Connections = &ConnectionSet{Defaults: []string{ref}}
	delete(frontMatter.Metadata, legacyConnectionKey)
	return &frontMatter, nil
}
