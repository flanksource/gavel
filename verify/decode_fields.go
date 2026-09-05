package verify

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"gopkg.in/yaml.v3"
)

type fieldInspector struct {
	format  string
	unknown map[string]bool
}

func (f fieldInspector) walk(node *yaml.Node, target reflect.Type, path string) {
	if node == nil || target == nil {
		return
	}
	if node.Kind == yaml.DocumentNode {
		for _, child := range node.Content {
			f.walk(child, target, path)
		}
		return
	}
	if node.Kind == yaml.AliasNode {
		f.walk(node.Alias, target, path)
		return
	}
	target = f.wireType(target)
	if target == nil {
		return
	}
	switch target.Kind() {
	case reflect.Struct:
		f.walkStruct(node, target, path)
	case reflect.Map:
		if node.Kind == yaml.MappingNode {
			for i := 0; i+1 < len(node.Content); i += 2 {
				f.walk(node.Content[i+1], target.Elem(), fieldPath(path, node.Content[i].Value))
			}
		}
	case reflect.Slice, reflect.Array:
		if node.Kind == yaml.SequenceNode {
			for i, child := range node.Content {
				f.walk(child, target.Elem(), fmt.Sprintf("%s[%d]", path, i))
			}
		}
	}
}

func (f fieldInspector) wireType(target reflect.Type) reflect.Type {
	for target.Kind() == reflect.Pointer {
		target = target.Elem()
	}
	value := reflect.New(target)
	if shape, ok := value.Interface().(DecodeFields); ok {
		return reflect.TypeOf(shape.DecodeFields())
	}
	if f.format == "json" && value.Type().Implements(reflect.TypeFor[json.Unmarshaler]()) {
		return nil
	}
	if f.format == "yaml" && value.Type().Implements(reflect.TypeFor[yaml.Unmarshaler]()) {
		return nil
	}
	return target
}

func (f fieldInspector) walkStruct(node *yaml.Node, target reflect.Type, path string) {
	if node.Kind != yaml.MappingNode {
		return
	}
	fields, additional := f.structFields(target, map[reflect.Type]bool{})
	for i := 0; i+1 < len(node.Content); i += 2 {
		key, value := node.Content[i], node.Content[i+1]
		if key.Tag == "!!merge" && f.format == "yaml" {
			f.walkMerge(value, target, path)
			continue
		}
		field := fields[key.Value]
		if field == nil && f.format == "json" {
			for name, candidate := range fields {
				if strings.EqualFold(name, key.Value) {
					field = candidate
					break
				}
			}
		}
		if field == nil {
			field = additional
		}
		if field == nil {
			f.unknown[fieldPath(path, key.Value)] = true
			continue
		}
		f.walk(value, field, fieldPath(path, key.Value))
	}
}

func (f fieldInspector) walkMerge(node *yaml.Node, target reflect.Type, path string) {
	if node.Kind == yaml.SequenceNode {
		for _, child := range node.Content {
			f.walk(child, target, path)
		}
		return
	}
	f.walk(node, target, path)
}

func (f fieldInspector) structFields(target reflect.Type, seen map[reflect.Type]bool) (map[string]reflect.Type, reflect.Type) {
	fields := map[string]reflect.Type{}
	var additional reflect.Type
	if seen[target] {
		return fields, nil
	}
	seen[target] = true
	defer delete(seen, target)
	for i := 0; i < target.NumField(); i++ {
		field := target.Field(i)
		if !field.IsExported() {
			continue
		}
		tag := strings.Split(field.Tag.Get(f.format), ",")
		if tag[0] == "-" {
			continue
		}
		inline := strings.Contains(","+field.Tag.Get(f.format)+",", ",inline,") || (f.format == "json" && field.Anonymous && tag[0] == "")
		if inline {
			fieldType := field.Type
			for fieldType.Kind() == reflect.Pointer {
				fieldType = fieldType.Elem()
			}
			if fieldType.Kind() == reflect.Map {
				additional = fieldType.Elem()
			} else if fieldType.Kind() == reflect.Struct {
				nested, extra := f.structFields(fieldType, seen)
				for name, typ := range nested {
					if fields[name] == nil {
						fields[name] = typ
					}
				}
				if extra != nil {
					additional = extra
				}
			}
			continue
		}
		name := tag[0]
		if name == "" {
			name = field.Name
			if f.format == "yaml" {
				name = strings.ToLower(name)
			}
		}
		fields[name] = field.Type
	}
	return fields, additional
}

func fieldPath(parent, key string) string {
	if parent == "" {
		return key
	}
	return parent + "." + key
}
