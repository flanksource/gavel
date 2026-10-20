package fixtures

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/goccy/go-yaml"
)

// FixtureDocumentKind is the `kind:` of a YAML/JSON fixture document.
const FixtureDocumentKind = "Fixture"

// IsStructuredFixtureFile reports whether path is a YAML/JSON fixture document
// rather than markdown.
func IsStructuredFixtureFile(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".yaml", ".yml", ".json":
		return true
	}
	return false
}

// ParseFixtureFile parses one fixture file into a FileNode tree, choosing the
// frontend by extension: markdown, or a YAML/JSON fixture document.
func ParseFixtureFile(path string) (*FixtureNode, error) {
	if !IsStructuredFixtureFile(path) {
		return ParseMarkdownFixturesWithTree(path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read fixture file: %w", err)
	}
	return ParseFixtureDocuments(path, data)
}

// fixtureDocument is one `kind: Fixture` document.
type fixtureDocument struct {
	Kind     string            `json:"kind"`
	Name     string            `json:"name,omitempty"`
	Defaults *StepDefaults     `json:"defaults,omitempty"`
	Setup    *SetupSpec        `json:"setup,omitempty"`
	Env      map[string]any    `json:"env,omitempty"`
	Traces   *TraceSpecs       `json:"traces,omitempty"`
	Steps    []json.RawMessage `json:"steps"`
}

// ParseFixtureDocuments parses a (multi-document) YAML or JSON fixture file
// into the node tree every runner walks: a FileNode, a SectionNode per
// document when there are several, and a TestNode per step whose FixtureTest
// carries the step (Step) and the document's env/setup (FrontMatter). Every
// step is validated against its kind before anything runs.
func ParseFixtureDocuments(path string, data []byte) (*FixtureNode, error) {
	docs, err := decodeFixtureDocuments(data)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	if len(docs) == 0 {
		return nil, fmt.Errorf("%s: no fixture documents", path)
	}
	spans, err := stepSpans(data)
	if err != nil {
		return nil, fmt.Errorf("%s: locate steps: %w", path, err)
	}
	if len(spans) != len(docs) {
		return nil, fmt.Errorf("%s: located steps in %d documents, decoded %d", path, len(spans), len(docs))
	}
	file := &FixtureNode{Name: filepath.Base(path), Type: FileNode, Origin: &FixtureOrigin{File: path, Kind: "file"}}
	// Absolute: `{{.fixture.dir}}` is handed to processes that run in other
	// directories (a service's `cwd:`, a mock started from the repo root).
	sourceDir, err := filepath.Abs(filepath.Dir(path))
	if err != nil {
		return nil, fmt.Errorf("%s: resolve fixture directory: %w", path, err)
	}
	var steps []parsedStep
	setups := 0
	for i, doc := range docs {
		label := fmt.Sprintf("document %d", i+1)
		if doc.Name != "" {
			label = doc.Name
		}
		if doc.Setup != nil {
			setups++
		}
		parent := file
		if len(docs) > 1 {
			parent = &FixtureNode{Name: label, Type: SectionNode, Level: 1, Origin: &FixtureOrigin{File: path, Kind: "document"}}
			file.AddChild(parent)
		}
		if len(spans[i]) != len(doc.Steps) {
			return nil, fmt.Errorf("%s: %s: located %d steps, decoded %d", path, label, len(spans[i]), len(doc.Steps))
		}
		docSteps, err := lowerDocument(doc, label, path, sourceDir, parent, spans[i])
		if err != nil {
			return nil, fmt.Errorf("%s: %s: %w", path, label, err)
		}
		steps = append(steps, docSteps...)
	}
	if setups > 1 {
		return nil, fmt.Errorf("%s: %d documents declare setup:, a file prepares one", path, setups)
	}
	if err := validateSteps(steps); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return file, nil
}

func decodeFixtureDocuments(data []byte) ([]fixtureDocument, error) {
	decoder := yaml.NewDecoder(bytes.NewReader(data), yaml.UseOrderedMap())
	var docs []fixtureDocument
	for i := 1; ; i++ {
		var raw any
		if err := decoder.Decode(&raw); err != nil {
			if errors.Is(err, io.EOF) {
				return docs, nil
			}
			return nil, fmt.Errorf("document %d: %w", i, err)
		}
		if raw == nil {
			continue
		}
		encoded, err := encodeOrderedJSON(raw)
		if err != nil {
			return nil, fmt.Errorf("document %d: %w", i, err)
		}
		var doc fixtureDocument
		strict := json.NewDecoder(bytes.NewReader(encoded))
		strict.DisallowUnknownFields()
		if err := strict.Decode(&doc); err != nil {
			return nil, fmt.Errorf("document %d: %w", i, err)
		}
		if doc.Kind != FixtureDocumentKind {
			return nil, fmt.Errorf("document %d: want kind: %s, got %q", i, FixtureDocumentKind, doc.Kind)
		}
		docs = append(docs, doc)
	}
}

// parsedStep is one step with the label validation errors name it by.
type parsedStep struct {
	label string
	spec  *StepSpec
}

func lowerDocument(doc fixtureDocument, docName, path, sourceDir string, parent *FixtureNode, spans []lineSpan) ([]parsedStep, error) {
	if len(doc.Steps) == 0 {
		return nil, fmt.Errorf("declares no steps")
	}
	var steps []parsedStep
	for i, raw := range doc.Steps {
		spec, err := splitStep(raw)
		label := stepIndexLabel(i, spec)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", label, err)
		}
		spec.Defaults = doc.Defaults
		name := firstNonEmpty(spec.Name, spec.ID, spec.Kind)
		test := &FixtureTest{Name: name, SourceDir: sourceDir, Step: &spec}
		test.FrontMatter.Env = doc.Env
		test.FrontMatter.Traces = doc.Traces
		test.Setup = doc.Setup
		parent.AddChild(&FixtureNode{Name: name, Type: TestNode, Test: test,
			Origin: &FixtureOrigin{File: path, Kind: "step", SectionPath: docName, RowIndex: i + 1,
				Line: spans[i].start, EndLine: spans[i].end}})
		steps = append(steps, parsedStep{label: label, spec: &spec})
	}
	return steps, nil
}

// splitStep separates a step's envelope keys (StepSpecKeys) from its
// kind-specific config, keeping the config's authored key order.
func splitStep(raw json.RawMessage) (StepSpec, error) {
	tree, err := decodeOrderedJSON(raw)
	if err != nil {
		return StepSpec{}, err
	}
	fields, ok := tree.(orderedObject)
	if !ok {
		return StepSpec{}, fmt.Errorf("a step must be a mapping, got %s", raw)
	}
	envelope, config := orderedObject{}, orderedObject{}
	for _, field := range fields {
		if slices.Contains(StepSpecKeys, field.Key) {
			envelope = append(envelope, field)
		} else {
			config = append(config, field)
		}
	}
	encoded, err := encodeOrderedJSON(envelope)
	if err != nil {
		return StepSpec{}, err
	}
	var spec StepSpec
	strict := json.NewDecoder(bytes.NewReader(encoded))
	strict.DisallowUnknownFields()
	if err := strict.Decode(&spec); err != nil {
		return spec, err
	}
	if spec.Kind == "" {
		return spec, fmt.Errorf("kind is required")
	}
	if len(config) > 0 {
		if spec.Config, err = encodeOrderedJSON(config); err != nil {
			return spec, err
		}
	}
	return spec, nil
}

func stepIndexLabel(index int, spec StepSpec) string {
	if spec.ID != "" {
		return fmt.Sprintf("step %q", spec.ID)
	}
	if spec.Kind != "" {
		return fmt.Sprintf("step #%d (%s)", index+1, spec.Kind)
	}
	return fmt.Sprintf("step #%d", index+1)
}
