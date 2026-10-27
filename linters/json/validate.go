package jsonlint

import (
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/flanksource/gavel/models"
	"github.com/santhosh-tekuri/jsonschema/v6"
	"gopkg.in/yaml.v3"
)

func (v *validator) validate(input fileInput, doc document) ([]models.Violation, error) {
	reference := input.schema
	if reference == "" {
		var err error
		reference, err = schemaDeclaration(doc.value)
		if err != nil {
			return nil, err
		}
	}
	if reference == "" {
		return nil, nil
	}
	schema, err := v.compile(reference, filepath.Dir(input.path))
	if err != nil {
		return nil, err
	}
	err = schema.Validate(doc.value)
	if err == nil {
		return nil, v.ctx.Err()
	}
	var invalid *jsonschema.ValidationError
	if !errors.As(err, &invalid) {
		return nil, err
	}
	return validationViolations(input.path, doc, invalid), nil
}

func validationViolations(path string, doc document, invalid *jsonschema.ValidationError) []models.Violation {
	var violations []models.Violation
	if len(invalid.Causes) > 0 {
		for _, cause := range invalid.Causes {
			violations = append(violations, validationViolations(path, doc, cause)...)
		}
		return violations
	}
	node := sourceNode(&doc.node, invalid.InstanceLocation)
	line, column := max(1, node.Line)+doc.offset, max(1, node.Column)
	location := ""
	for _, part := range invalid.InstanceLocation {
		location += "/" + strings.ReplaceAll(strings.ReplaceAll(part, "~", "~0"), "/", "~1")
	}
	return []models.Violation{violation(path, line, column, "schema", fmt.Sprintf("%s: %s", location, invalid.Error()))}
}

func sourceNode(node *yaml.Node, location []string) *yaml.Node {
	if node.Kind == yaml.DocumentNode && len(node.Content) > 0 {
		node = node.Content[0]
	}
	for _, part := range location {
		if node.Kind == yaml.AliasNode {
			node = node.Alias
		}
		switch node.Kind {
		case yaml.MappingNode:
			for i := 0; i < len(node.Content); i += 2 {
				if node.Content[i].Value == part {
					node = node.Content[i+1]
					break
				}
			}
		case yaml.SequenceNode:
			index, err := strconv.Atoi(part)
			if err == nil && index >= 0 && index < len(node.Content) {
				node = node.Content[index]
			}
		}
	}
	return node
}
