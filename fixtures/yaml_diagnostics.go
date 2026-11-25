package fixtures

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/flanksource/gavel/internal/yamldiag"
	"github.com/santhosh-tekuri/jsonschema/v6"
	"gopkg.in/yaml.v3"
)

func (s parsedStep) diagnostic(cause error) error {
	var invalid *jsonschema.ValidationError
	if s.source == nil || !errors.As(cause, &invalid) {
		return fmt.Errorf("%s: %w", s.label, cause)
	}
	root, err := s.sourceNode()
	if err != nil {
		return fmt.Errorf("%s: locate invalid config: %w", s.label, errors.Join(cause, err))
	}
	return &yamldiag.Error{Context: fmt.Sprintf("%s: config does not match the %s schema", s.label, s.spec.Kind),
		Violations: yamldiag.Violations(s.path, s.source, root, 0, invalid), Cause: cause}
}

func (s parsedStep) sourceNode() (*yaml.Node, error) {
	decoder := yaml.NewDecoder(bytes.NewReader(s.source))
	for {
		var doc yaml.Node
		if err := decoder.Decode(&doc); err != nil {
			if errors.Is(err, io.EOF) {
				return nil, fmt.Errorf("no step at source line %d", s.span.start)
			}
			return nil, err
		}
		steps := yamldiag.SourceNode(&doc, []string{"steps"}, false)
		if steps.Kind != yaml.SequenceNode {
			continue
		}
		for _, node := range steps.Content {
			if node.Line == s.span.start {
				return node, nil
			}
		}
	}
}
