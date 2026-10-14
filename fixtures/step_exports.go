package fixtures

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// ExportRecord is one exported variable on a step's result:
// Metadata["exports"][NAME] = {value, step}.
type ExportRecord struct {
	Value string `json:"value"`
	Step  string `json:"step"`
}

// exportAll is the `export:` key that exports every entry of an output map.
const exportAll = "*"

var outputsPathPattern = regexp.MustCompile(`^\{\{\s*\.outputs((?:\.[A-Za-z_][A-Za-z0-9_]*)+)\s*\}\}$`)

// applyExports merges the kind's default exports with the step's `export:`
// templates and exports them into the fixture Env for every later step. A
// step entry renders over this step's `{{.outputs.*}}` (plus the usual roots);
// an empty entry suppresses the kind default of that name; `"*"` must be a
// single `{{.outputs.<path>}}` reference to a map, and exports its entries.
func applyExports(step *StepContext, defaults map[string]string, outputs map[string]any) (map[string]ExportRecord, error) {
	values := make(map[string]string, len(defaults))
	for k, v := range defaults {
		values[k] = v
	}
	if expr, ok := step.Spec.Export[exportAll]; ok {
		entries, err := exportOutputsMap(expr, outputs)
		if err != nil {
			return nil, err
		}
		for k, v := range entries {
			values[k] = v
		}
	}
	vars := step.vars()
	vars["outputs"] = outputs
	for name, expr := range step.Spec.Export {
		switch {
		case name == exportAll:
		case expr == "":
			delete(values, name)
		default:
			rendered, err := step.Scope.Render(expr, vars)
			if err != nil {
				return nil, fmt.Errorf("export %s: %w", name, err)
			}
			values[name] = rendered
		}
	}
	names := make([]string, 0, len(values))
	for name := range values {
		names = append(names, name)
	}
	sort.Strings(names)
	label := stepLabel(step)
	records := make(map[string]ExportRecord, len(values))
	for _, name := range names {
		if err := step.Scope.Env.Export(label, name, values[name], step.Spec.Override); err != nil {
			return nil, err
		}
		records[name] = ExportRecord{Value: values[name], Step: label}
	}
	return records, nil
}

func exportOutputsMap(expr string, outputs map[string]any) (map[string]string, error) {
	match := outputsPathPattern.FindStringSubmatch(strings.TrimSpace(expr))
	if match == nil {
		return nil, fmt.Errorf(`export "*": %q must be a single {{.outputs.<path>}} reference to a map`, expr)
	}
	var value any = outputs
	for _, key := range strings.Split(strings.TrimPrefix(match[1], "."), ".") {
		parent, ok := value.(map[string]any)
		if !ok {
			return nil, fmt.Errorf(`export "*": %s is not a map`, expr)
		}
		if value, ok = parent[key]; !ok {
			return nil, fmt.Errorf(`export "*": outputs have no %q (in %s)`, key, expr)
		}
	}
	entries := map[string]string{}
	switch m := value.(type) {
	case map[string]string:
		for k, v := range m {
			entries[k] = v
		}
	case map[string]any:
		for k, v := range m {
			entries[k] = fmt.Sprint(v)
		}
	default:
		return nil, fmt.Errorf(`export "*": %s is a %T, not a map`, expr, value)
	}
	return entries, nil
}

// renderConfig renders every string value of the step's config as a Go
// template against the step's variables (`{{.env.X}}`, `{{.steps.<id>.…}}`,
// `{{.tmp}}`, `{{.fixture.dir}}`). `$X` is deliberately left alone here: a
// shell script in a config spells its own variables that way, and the exported
// Env already reaches the subprocess.
func renderConfig(step *StepContext) (json.RawMessage, error) {
	config, err := decodeOrderedJSON(step.Spec.Config)
	if err != nil {
		return nil, fmt.Errorf("decode config: %w", err)
	}
	rendered, err := renderValue(config, step.vars())
	if err != nil {
		return nil, err
	}
	return encodeOrderedJSON(rendered)
}

func renderValue(value any, vars map[string]any) (any, error) {
	switch v := value.(type) {
	case string:
		return renderGoTemplate(v, vars)
	case orderedObject:
		for i, member := range v {
			rendered, err := renderValue(member.Value, vars)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", member.Key, err)
			}
			v[i].Value = rendered
		}
		return v, nil
	case []any:
		for i, item := range v {
			rendered, err := renderValue(item, vars)
			if err != nil {
				return nil, fmt.Errorf("[%d]: %w", i, err)
			}
			v[i] = rendered
		}
		return v, nil
	default:
		return value, nil
	}
}
