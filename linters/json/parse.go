package jsonlint

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/flanksource/gavel/models"
	"github.com/santhosh-tekuri/jsonschema/v6"
	"gopkg.in/yaml.v3"
)

type fileInput struct {
	path    string
	data    []byte
	schema  string
	workDir string
}

type document struct {
	value  any
	node   yaml.Node
	offset int
}

var yamlLine = regexp.MustCompile(`line (\d+)`)
var modelineSchema = regexp.MustCompile(`\$schema\s*=\s*("[^"]+"|'[^']+'|\S+)`)

func (v *validator) lintFile(input fileInput) ([]models.Violation, error) {
	if input.schema != "" {
		location, err := ResolveSchema(input.schema, input.workDir)
		if err != nil {
			return nil, err
		}
		input.schema = location
	}
	switch strings.ToLower(filepath.Ext(input.path)) {
	case ".ndjson":
		return v.lintRecords(input)
	case ".yaml", ".yml":
		return v.lintYAML(input)
	default:
		doc, err := parseJSON(input.data, 0)
		if err != nil {
			return []models.Violation{syntaxViolation(input.path, input.data, err, 0)}, nil
		}
		return v.validate(input, doc)
	}
}

func (v *validator) lintRecords(input fileInput) ([]models.Violation, error) {
	reader := bufio.NewReader(bytes.NewReader(input.data))
	var violations []models.Violation
	for line := 0; ; line++ {
		if err := v.ctx.Err(); err != nil {
			return violations, err
		}
		record, err := reader.ReadBytes('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			return violations, err
		}
		if len(bytes.TrimSpace(record)) > 0 {
			doc, parseErr := parseJSON(record, line)
			if parseErr != nil {
				violations = append(violations, syntaxViolation(input.path, record, parseErr, line))
			} else {
				found, validationErr := v.validate(input, doc)
				violations = append(violations, found...)
				if validationErr != nil {
					return violations, validationErr
				}
			}
		}
		if errors.Is(err, io.EOF) {
			return violations, nil
		}
	}
}

func (v *validator) lintYAML(input fileInput) ([]models.Violation, error) {
	if input.schema == "" {
		var err error
		input.schema, err = yamlHeader(input.data)
		if err != nil {
			return nil, err
		}
		if input.schema != "" {
			input.schema, err = ResolveSchema(input.schema, filepath.Dir(input.path))
			if err != nil {
				return nil, err
			}
		}
	}
	if input.schema != "" {
		if _, err := v.compile(input.schema, filepath.Dir(input.path)); err != nil {
			return nil, err
		}
	}
	decoder := yaml.NewDecoder(bytes.NewReader(input.data))
	var violations []models.Violation
	for {
		if err := v.ctx.Err(); err != nil {
			return violations, err
		}
		doc := document{}
		if err := decoder.Decode(&doc.node); errors.Is(err, io.EOF) {
			return violations, nil
		} else if err != nil {
			return append(violations, syntaxViolation(input.path, input.data, err, 0)), nil
		}
		found, err := v.validateYAML(input, doc)
		violations = append(violations, found...)
		if err != nil {
			return violations, err
		}
	}
}

func (v *validator) validateYAML(input fileInput, doc document) ([]models.Violation, error) {
	var err error
	if input.schema == "" {
		input.schema, err = yamlSchema(&doc.node)
		if err != nil {
			return nil, err
		}
	}
	if input.schema == "" {
		if err := doc.node.Decode(&doc.value); err != nil {
			return []models.Violation{syntaxViolation(input.path, input.data, err, 0)}, nil
		}
		return nil, nil
	}
	doc.value, err = yamlValue(&doc.node, map[*yaml.Node]bool{})
	if err != nil {
		return []models.Violation{syntaxViolation(input.path, input.data, err, 0)}, nil
	}
	return v.validate(input, doc)
}

func parseJSON(data []byte, offset int) (document, error) {
	value, err := jsonschema.UnmarshalJSON(bytes.NewReader(data))
	if err != nil {
		return document{}, err
	}
	doc := document{value: value, offset: offset}
	if err := yaml.Unmarshal(data, &doc.node); err != nil {
		return document{}, err
	}
	return doc, nil
}

func yamlHeader(data []byte) (string, error) {
	for _, line := range bytes.Split(data, []byte("\n")) {
		text := strings.TrimSpace(strings.TrimPrefix(string(line), "\ufeff"))
		if text == "" {
			continue
		}
		if !strings.HasPrefix(text, "#") {
			break
		}
		if !strings.HasPrefix(strings.TrimSpace(strings.TrimPrefix(text, "#")), "yaml-language-server:") {
			continue
		}
		match := modelineSchema.FindStringSubmatch(text)
		if len(match) > 0 {
			reference := strings.Trim(match[1], "\"'")
			if strings.TrimSpace(reference) == "" {
				return "", fmt.Errorf("YAML header requires a nonempty $schema path or URL")
			}
			return reference, nil
		}
		if strings.Contains(text, "$schema") {
			return "", fmt.Errorf("YAML header requires a nonempty $schema path or URL")
		}
	}
	return "", nil
}

func syntaxViolation(path string, data []byte, err error, offset int) models.Violation {
	line, column := 1, 1
	var syntax *json.SyntaxError
	if errors.As(err, &syntax) {
		position := min(int(syntax.Offset)-1, len(data))
		if position >= 0 {
			line += bytes.Count(data[:position], []byte("\n"))
			column = position - bytes.LastIndexByte(data[:position], '\n')
		}
	} else if match := yamlLine.FindStringSubmatch(err.Error()); len(match) > 0 {
		line, _ = strconv.Atoi(match[1])
	}
	return violation(path, line+offset, column, "syntax", err.Error())
}

func violation(path string, line, column int, rule, message string) models.Violation {
	result := models.NewViolationBuilder().WithFile(path).WithLocation(line, column).
		WithSource("json").WithRuleFromLinter("json", rule).WithMessage(message).Build()
	result.Severity = models.SeverityError
	return result
}
