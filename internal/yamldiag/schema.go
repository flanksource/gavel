package yamldiag

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/flanksource/gavel/models"
	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/santhosh-tekuri/jsonschema/v6/kind"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"gopkg.in/yaml.v3"
)

// Violations resolves schema leaves against the authored document, not decoded JSON.
func Violations(path string, data []byte, node *yaml.Node, offset int, invalid *jsonschema.ValidationError) []models.Violation {
	var out []models.Violation
	if len(invalid.Causes) > 0 {
		for _, cause := range invalid.Causes {
			out = append(out, Violations(path, data, node, offset, cause)...)
		}
		return out
	}
	locations := [][]string{invalid.InstanceLocation}
	if extra, ok := invalid.ErrorKind.(*kind.AdditionalProperties); ok {
		locations = nil
		for _, property := range slices.Sorted(slices.Values(extra.Properties)) {
			locations = append(locations, append(slices.Clone(invalid.InstanceLocation), property))
		}
	}
	for _, location := range locations {
		_, key := invalid.ErrorKind.(*kind.AdditionalProperties)
		source := SourceNode(node, location, key)
		line, column := max(1, source.Line)+offset, max(1, source.Column)
		pointer := ""
		for _, part := range location {
			pointer += "/" + strings.ReplaceAll(strings.ReplaceAll(part, "~", "~0"), "/", "~1")
		}
		text := invalid.ErrorKind.LocalizedString(message.NewPrinter(language.English))
		if pointer != "" {
			text = fmt.Sprintf("%s: %s", pointer, text)
		}
		v := models.NewViolationBuilder().WithFile(path).WithLocation(line, column).WithSource("json").WithRuleFromLinter("json", "schema").WithMessage(text).Build()
		v.Severity = models.SeverityError
		v.Excerpt = models.NewSourceExcerpt(data, line, column)
		out = append(out, v)
	}
	return out
}

func SourceNode(node *yaml.Node, location []string, key bool) *yaml.Node {
	for node.Kind == yaml.DocumentNode || node.Kind == yaml.AliasNode {
		if node.Kind == yaml.AliasNode {
			node = node.Alias
		} else if len(node.Content) > 0 {
			node = node.Content[0]
		} else {
			return node
		}
	}
	if len(location) == 0 {
		return node
	}
	switch node.Kind {
	case yaml.MappingNode:
		for i := 0; i < len(node.Content); i += 2 {
			if node.Content[i].Value == location[0] {
				if key && len(location) == 1 {
					return node.Content[i]
				}
				return SourceNode(node.Content[i+1], location[1:], key)
			}
		}
	case yaml.SequenceNode:
		index, err := strconv.Atoi(location[0])
		if err == nil && index >= 0 && index < len(node.Content) {
			return SourceNode(node.Content[index], location[1:], key)
		}
	}
	return node
}
