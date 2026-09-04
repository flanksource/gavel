package verify

import (
	"fmt"
	"sort"
	"strings"

	yamlv3 "gopkg.in/yaml.v3"
)

// removedKey is one .gavel.yaml key that no longer exists, and what replaced it.
type removedKey struct {
	// Path is the dotted key path, matched from the document root.
	Path string
	// Replacement completes the sentence "use ..." — it must name the exact key
	// or setting that took over, not a vague direction.
	Replacement string
}

// todosRemovedKeys are the keys the todo lifecycle stopped reading. An unknown
// key is silently dropped by yaml.Unmarshal, so a project still declaring one
// of these ran on built-in defaults with nothing saying so — the whole point of
// naming them here is that the config now fails loudly instead.
var todosRemovedKeys = []removedKey{
	{Path: "todos.driver", Replacement: `ai.model with the compact "mode:model:effort" form, e.g. ai.model: "cli:opus:high"`},
	{Path: "todos.prompts", Replacement: "a lifecycle step under todos.<step>"},
	{Path: "todos.groupBy", Replacement: "nothing — grouping was removed; runs dispatch per todo"},
	{Path: "todos.approvals", Replacement: `permissions.mode: default on the step (the dashboard brokers each tool call)`},
	{Path: "checks.maxIterations", Replacement: "todos.run.workflow.verify.maxIterations"},
}

// checkRemovedKeys reports every removed key present in one parsed .gavel.yaml
// document, as a single error listing them all so a config carrying three of
// them is fixed in one pass. Presence of the key is what fails — its value may
// be a mapping, a scalar, or null. An empty or absent document is not an error.
func checkRemovedKeys(path string, doc *yamlv3.Node) error {
	var found []removedKey
	for _, key := range todosRemovedKeys {
		if hasYAMLPath(doc, strings.Split(key.Path, ".")) {
			found = append(found, key)
		}
	}
	if len(found) == 0 {
		return nil
	}

	sort.Slice(found, func(i, j int) bool { return found[i].Path < found[j].Path })
	lines := make([]string, 0, len(found))
	for _, key := range found {
		lines = append(lines, fmt.Sprintf("%s: %s is no longer supported; use %s", path, key.Path, key.Replacement))
	}
	return fmt.Errorf("%s", strings.Join(lines, "\n"))
}

// hasYAMLPath walks mapping nodes segment by segment from the document root, so
// a removed name is only matched at the exact depth it was removed from:
// `commit.groupBy` and `todos.run.prompts` are untouched by `todos.groupBy` and
// `todos.prompts`.
func hasYAMLPath(node *yamlv3.Node, segments []string) bool {
	if node == nil || len(segments) == 0 {
		return false
	}
	if node.Kind == yamlv3.DocumentNode {
		for _, child := range node.Content {
			if hasYAMLPath(child, segments) {
				return true
			}
		}
		return false
	}
	if node.Kind != yamlv3.MappingNode {
		return false
	}
	for i := 0; i+1 < len(node.Content); i += 2 {
		if node.Content[i].Value != segments[0] {
			continue
		}
		if len(segments) == 1 {
			return true
		}
		return hasYAMLPath(node.Content[i+1], segments[1:])
	}
	return false
}
