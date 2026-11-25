package jsonlint

import (
	"errors"
	"path/filepath"

	"github.com/flanksource/gavel/internal/yamldiag"
	"github.com/flanksource/gavel/models"
	"github.com/santhosh-tekuri/jsonschema/v6"
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
		reference = input.fallback
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
	return yamldiag.Violations(input.path, input.data, &doc.node, doc.offset, invalid), nil
}
