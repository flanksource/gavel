package externalissues

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"unicode"
)

type Fields map[string]Field

func (f Fields) Keys() map[string]string {
	ids := make([]string, 0, len(f))
	keys, counts := make(map[string]string, len(f)), make(map[string]int)
	for id, field := range f {
		ids = append(ids, id)
		key := strings.Join(strings.FieldsFunc(strings.ToLower(field.Name), func(r rune) bool {
			return !unicode.IsLetter(r) && !unicode.IsDigit(r)
		}), "_")
		if key == "" {
			key = id
		}
		keys[id] = key
		counts[key]++
	}
	slices.Sort(ids)
	used, next := make(map[string]bool, len(f)), make(map[string]int)
	for _, id := range ids {
		base := keys[id]
		if counts[base] == 1 {
			used[base] = true
			continue
		}
		for {
			next[base]++
			key := fmt.Sprintf("%s_%d", base, next[base])
			if counts[key] == 0 && !used[key] {
				keys[id], used[key] = key, true
				break
			}
		}
	}
	return keys
}

func (f Fields) MarshalJSON() ([]byte, error) {
	if f == nil {
		return []byte("null"), nil
	}
	values := make(map[string]json.RawMessage, len(f))
	for id, key := range f.Keys() {
		values[key] = f[id].Value
	}
	return json.Marshal(values)
}

func (f *Fields) UnmarshalJSON(data []byte) error {
	var values map[string]json.RawMessage
	if err := json.Unmarshal(data, &values); err != nil {
		return err
	}
	if values == nil {
		*f = nil
		return nil
	}
	fields := make(Fields, len(values))
	for key, value := range values {
		fields[key] = Field{ID: key, Name: key, Value: value}
	}
	*f = fields
	return nil
}

func (Fields) JSONSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": true}
}
