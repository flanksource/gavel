package fixtures

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/goccy/go-yaml"
)

// A step's config keeps the key order it was authored in: a kind may give
// mapping order meaning (services start in the order they are declared), and
// Go maps would sort it away. orderedObject is a JSON object as an ordered
// list of members; the tree it heads holds orderedObject, []any, string,
// json.Number, bool and nil.
type orderedObject []orderedMember

type orderedMember struct {
	Key   string
	Value any
}

// decodeOrderedJSON parses raw JSON into an ordered tree.
func decodeOrderedJSON(raw []byte) (any, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	value, err := decodeOrderedValue(decoder)
	if err != nil {
		return nil, err
	}
	if _, err := decoder.Token(); err != io.EOF {
		return nil, fmt.Errorf("unexpected data after the JSON value")
	}
	return value, nil
}

func decodeOrderedValue(decoder *json.Decoder) (any, error) {
	token, err := decoder.Token()
	if err != nil {
		return nil, err
	}
	switch delim := token.(type) {
	case json.Delim:
		switch delim {
		case '{':
			object := orderedObject{}
			for decoder.More() {
				key, err := decoder.Token()
				if err != nil {
					return nil, err
				}
				value, err := decodeOrderedValue(decoder)
				if err != nil {
					return nil, err
				}
				object = append(object, orderedMember{Key: key.(string), Value: value})
			}
			_, err := decoder.Token()
			return object, err
		case '[':
			list := []any{}
			for decoder.More() {
				value, err := decodeOrderedValue(decoder)
				if err != nil {
					return nil, err
				}
				list = append(list, value)
			}
			_, err := decoder.Token()
			return list, err
		}
		return nil, fmt.Errorf("unexpected JSON delimiter %v", delim)
	default:
		return token, nil
	}
}

// encodeOrderedJSON writes an ordered tree (or a goccy yaml.MapSlice tree)
// back to JSON, members in order.
func encodeOrderedJSON(value any) ([]byte, error) {
	var buf bytes.Buffer
	if err := writeOrderedJSON(&buf, value); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func writeOrderedJSON(buf *bytes.Buffer, value any) error {
	switch v := value.(type) {
	case orderedObject:
		buf.WriteByte('{')
		for i, member := range v {
			if err := writeMember(buf, i, member.Key, member.Value); err != nil {
				return err
			}
		}
		buf.WriteByte('}')
	case yaml.MapSlice:
		buf.WriteByte('{')
		for i, item := range v {
			key, ok := item.Key.(string)
			if !ok {
				return fmt.Errorf("mapping key %v is a %T, not a string", item.Key, item.Key)
			}
			if err := writeMember(buf, i, key, item.Value); err != nil {
				return err
			}
		}
		buf.WriteByte('}')
	case []any:
		buf.WriteByte('[')
		for i, item := range v {
			if i > 0 {
				buf.WriteByte(',')
			}
			if err := writeOrderedJSON(buf, item); err != nil {
				return err
			}
		}
		buf.WriteByte(']')
	default:
		encoded, err := json.Marshal(v)
		if err != nil {
			return err
		}
		buf.Write(encoded)
	}
	return nil
}

func writeMember(buf *bytes.Buffer, index int, key string, value any) error {
	if index > 0 {
		buf.WriteByte(',')
	}
	encodedKey, err := json.Marshal(key)
	if err != nil {
		return err
	}
	buf.Write(encodedKey)
	buf.WriteByte(':')
	return writeOrderedJSON(buf, value)
}
