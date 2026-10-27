package runconfig

import (
	"reflect"
	"strconv"
	"strings"
)

func Schema(options any, tool bool) map[string]any {
	properties := map[string]any{}
	typeOf := reflect.TypeOf(options)
	for i := range typeOf.NumField() {
		field := typeOf.Field(i)
		t := field.Type
		if t.Kind() == reflect.Pointer {
			t = t.Elem()
		}
		node := map[string]any{"description": field.Tag.Get("help")}
		switch t.Kind() {
		case reflect.Bool:
			node["type"] = "boolean"
		case reflect.Int:
			node["type"] = "integer"
		case reflect.String:
			node["type"] = "string"
		case reflect.Slice:
			node["type"], node["items"] = "array", map[string]any{"type": "string"}
		}
		if choices := field.Tag.Get("enum"); choices != "" {
			node["enum"] = strings.Split(choices, ",")
		}
		if minimum := field.Tag.Get("min"); minimum != "" {
			limit, err := strconv.Atoi(minimum)
			if err != nil {
				panic(err)
			}
			node["minimum"] = limit
		}
		if tool && field.Tag.Get("scope") == "run" {
			node["not"] = map[string]any{}
		}
		properties[key(field)] = node
	}
	return map[string]any{"type": "object", "additionalProperties": false, "properties": properties}
}
