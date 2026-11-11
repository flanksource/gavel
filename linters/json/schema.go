package jsonlint

import (
	"context"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"gopkg.in/yaml.v3"
)

type validator struct {
	ctx      context.Context
	compiler *jsonschema.Compiler
	schemas  map[string]*jsonschema.Schema
}

func newValidator(ctx context.Context) *validator {
	compiler := jsonschema.NewCompiler()
	compiler.UseLoader(schemaLoader{ctx: ctx})
	return &validator{ctx: ctx, compiler: compiler, schemas: map[string]*jsonschema.Schema{}}
}

// ResolveSchema anchors a schema path before lint invocations change working directories.
func ResolveSchema(reference, directory string) (string, error) {
	parsed, err := url.Parse(reference)
	if err != nil {
		return "", fmt.Errorf("schema %q: %w", reference, err)
	}
	if parsed.IsAbs() {
		if parsed.Scheme != "file" && parsed.Scheme != "http" && parsed.Scheme != "https" && parsed.Scheme != BuiltinScheme {
			return "", fmt.Errorf("unsupported schema URL scheme %q", parsed.Scheme)
		}
		return parsed.String(), nil
	}
	absolute, err := filepath.Abs(directory)
	if err != nil {
		return "", err
	}
	base := &url.URL{Scheme: "file", Path: filepath.ToSlash(absolute) + "/"}
	return base.ResolveReference(parsed).String(), nil
}

func (v *validator) compile(reference, directory string) (*jsonschema.Schema, error) {
	if err := v.ctx.Err(); err != nil {
		return nil, err
	}
	location, err := ResolveSchema(reference, directory)
	if err != nil {
		return nil, err
	}
	if schema, ok := v.schemas[location]; ok {
		return schema, nil
	}
	schema, err := v.compiler.Compile(location)
	if err != nil {
		return nil, fmt.Errorf("load schema %s: %w", location, err)
	}
	v.schemas[location] = schema
	return schema, nil
}

type schemaLoader struct{ ctx context.Context }

func (l schemaLoader) Load(location string) (any, error) {
	if err := l.ctx.Err(); err != nil {
		return nil, err
	}
	parsed, err := url.Parse(location)
	if err != nil {
		return nil, err
	}
	if parsed.Scheme == "file" {
		path, err := (jsonschema.FileLoader{}).ToFile(location)
		if err != nil {
			return nil, err
		}
		file, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		defer file.Close()
		return decodeSchema(file, path, "")
	}
	if parsed.Scheme == BuiltinScheme {
		return loadBuiltin(location)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, fmt.Errorf("unsupported schema URL scheme %q", parsed.Scheme)
	}
	request, err := http.NewRequestWithContext(l.ctx, http.MethodGet, location, nil)
	if err != nil {
		return nil, err
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("schema %s: HTTP %s", location, response.Status)
	}
	return decodeSchema(response.Body, parsed.Path, response.Header.Get("Content-Type"))
}

func decodeSchema(reader io.Reader, path, contentType string) (any, error) {
	extension := strings.ToLower(filepath.Ext(path))
	yamlEncoded := extension == ".yaml" || extension == ".yml"
	if !yamlEncoded && contentType != "" {
		mediaType, _, err := mime.ParseMediaType(contentType)
		if err != nil {
			return nil, fmt.Errorf("schema content type %q: %w", contentType, err)
		}
		switch mediaType {
		case "application/yaml", "application/x-yaml", "text/yaml", "text/x-yaml":
			yamlEncoded = true
		}
	}
	if !yamlEncoded {
		return jsonschema.UnmarshalJSON(reader)
	}
	decoder := yaml.NewDecoder(reader)
	var document yaml.Node
	if err := decoder.Decode(&document); err != nil {
		return nil, err
	}
	var extra yaml.Node
	if err := decoder.Decode(&extra); err != io.EOF {
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("schema must contain exactly one YAML document")
	}
	return yamlValue(&document, map[*yaml.Node]bool{})
}

func schemaDeclaration(value any) (string, error) {
	object, ok := value.(map[string]any)
	if !ok {
		return "", nil
	}
	declared, ok := object["$schema"]
	if !ok {
		return "", nil
	}
	reference, ok := declared.(string)
	if !ok || strings.TrimSpace(reference) == "" {
		return "", fmt.Errorf("$schema must be a nonempty path or URL")
	}
	return reference, nil
}

func yamlSchema(node *yaml.Node) (string, error) {
	if node.Kind == yaml.DocumentNode && len(node.Content) > 0 {
		node = node.Content[0]
	}
	if node.Kind != yaml.MappingNode {
		return "", nil
	}
	for i := 0; i < len(node.Content); i += 2 {
		if node.Content[i].Value != "$schema" {
			continue
		}
		var value any
		if err := node.Content[i+1].Decode(&value); err != nil {
			return "", err
		}
		return schemaDeclaration(map[string]any{"$schema": value})
	}
	return "", nil
}
