package services

import (
	"encoding/json"
	"fmt"
	"maps"
	"net"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/flanksource/gavel/fixtures"
)

var (
	serviceRefPattern = regexp.MustCompile(`\.services\.([A-Za-z][A-Za-z0-9_]*)`)
	portRefPattern    = regexp.MustCompile(`\{\{[^}]*\.port\b`)
	wildcardPattern   = regexp.MustCompile(`^\{\{\s*\.outputs\.([A-Za-z][A-Za-z0-9_]*)\s*\}\}$`)
)

// verbatimKeys are rendered later (export, against the entry's outputs) or
// never (git files are written as declared).
var verbatimKeys = map[string]bool{"export": true, "git": true}

// renderEntry templates the entry's strings against the fixture scope plus
// `.services.<name>.*` of the entries already started and `.port`, a free
// port allocated when the entry references it. A reference to an entry that
// has not started yet is an error, not an empty string.
func renderEntry(step *fixtures.StepContext, raw rawEntry, started map[string]any) (Entry, int, error) {
	scanned := map[string]any{}
	for k, v := range raw.body {
		if k != "git" {
			scanned[k] = v
		}
	}
	text, err := json.Marshal(scanned)
	if err != nil {
		return Entry{}, 0, err
	}
	for _, match := range serviceRefPattern.FindAllStringSubmatch(string(text), -1) {
		if _, ok := started[match[1]]; !ok {
			return Entry{}, 0, fmt.Errorf("entry %q references services.%s, which is not declared before it", raw.name, match[1])
		}
	}
	vars := map[string]any{"services": started}
	port := 0
	if portRefPattern.Match(text) {
		if port, err = freePort(); err != nil {
			return Entry{}, 0, fmt.Errorf("entry %q: allocate port: %w", raw.name, err)
		}
		vars["port"] = port
	}
	body := make(map[string]any, len(raw.body))
	for k, v := range raw.body {
		if verbatimKeys[k] {
			body[k] = v
			continue
		}
		// `exec:` runs under `sh -c` with the fixture Env, so the shell expands
		// its $X, $?, $1 and loop variables; every other value is expanded here.
		expand := k != "exec"
		if body[k], err = renderValue(step, vars, v, expand); err != nil {
			return Entry{}, 0, fmt.Errorf("entry %q: %s: %w", raw.name, k, err)
		}
	}
	entry, err := decodeEntry(raw.name, body)
	return entry, port, err
}

func renderValue(step *fixtures.StepContext, vars map[string]any, value any, expand bool) (any, error) {
	switch v := value.(type) {
	case string:
		if expand {
			return expandText(step, vars, v)
		}
		return render(step, vars, v)
	case map[string]any:
		out := make(map[string]any, len(v))
		for k, item := range v {
			rendered, err := renderValue(step, vars, item, expand)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", k, err)
			}
			out[k] = rendered
		}
		return out, nil
	case []any:
		out := make([]any, len(v))
		for i, item := range v {
			rendered, err := renderValue(step, vars, item, expand)
			if err != nil {
				return nil, fmt.Errorf("[%d]: %w", i, err)
			}
			out[i] = rendered
		}
		return out, nil
	default:
		return value, nil
	}
}

// render runs StepContext.Template with extra variables layered over the
// test's TemplateVars, so `{{.env.X}}`, `{{.steps.…}}` and `{{.tmp}}` keep
// working next to `.services`, `.port` and `.outputs`. `$X` is left for the
// shell.
func render(step *fixtures.StepContext, extra map[string]any, text string) (string, error) {
	return renderWith(withVars(step, extra).Template, text)
}

// expandText is render plus strict `$NAME` expansion from the fixture Env, for
// values no shell will see (URLs, ready probes, env values, cwd, exports).
func expandText(step *fixtures.StepContext, extra map[string]any, text string) (string, error) {
	return renderWith(withVars(step, extra).Expand, text)
}

func withVars(step *fixtures.StepContext, extra map[string]any) *fixtures.StepContext {
	scoped := *step
	scoped.Test.TemplateVars = maps.Clone(step.Test.TemplateVars)
	if scoped.Test.TemplateVars == nil {
		scoped.Test.TemplateVars = map[string]any{}
	}
	maps.Copy(scoped.Test.TemplateVars, extra)
	return &scoped
}

func renderWith(fn func(string) (string, error), text string) (string, error) {
	out, err := fn(text)
	if err != nil {
		return "", err
	}
	if strings.Contains(out, "<no value>") {
		return "", fmt.Errorf("template %q references a value that does not exist (rendered %q)", text, out)
	}
	return out, nil
}

// exportSet applies entry exports to the fixture Env as each entry starts, so
// later entries see `$NAME`, and keeps them for StepOutcome.Exports.
type exportSet struct {
	step   *fixtures.StepContext
	owner  map[string]string
	values map[string]string
}

func newExportSet(step *fixtures.StepContext) *exportSet {
	return &exportSet{step: step, owner: map[string]string{}, values: map[string]string{}}
}

// exportEntry exports the entry's defaults (<NAME>_URL, <NAME>_PORT when the
// entry has a url/port) and its `export:` map, which overrides a default of
// the same name.
func (s *exportSet) exportEntry(entry rawEntry, outputs map[string]any, started map[string]any) error {
	values := map[string]string{}
	if outputs["type"] == "dbtest" && outputs["default"] == true {
		for name, value := range outputs["env"].(map[string]any) {
			values[name] = value.(string)
		}
	}
	prefix := upperSnake(entry.name)
	if url, ok := outputs["url"].(string); ok && url != "" {
		values[prefix+"_URL"] = url
	}
	if port, ok := outputs["port"].(int); ok && port != 0 {
		values[prefix+"_PORT"] = strconv.Itoa(port)
	}
	declared, err := exportTemplates(entry)
	if err != nil {
		return err
	}
	vars := map[string]any{"services": started, "outputs": outputs}
	for name, tmpl := range declared {
		if name == "*" {
			if err := wildcardExport(entry.name, tmpl, outputs, values); err != nil {
				return err
			}
			continue
		}
		if values[name], err = expandText(s.step, vars, tmpl); err != nil {
			return fmt.Errorf("export %s of entry %q: %w", name, entry.name, err)
		}
	}
	names := make([]string, 0, len(values))
	for name := range values {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if err := s.add(entry.name, name, values[name]); err != nil {
			return err
		}
	}
	return nil
}

func (s *exportSet) add(entry, name, value string) error {
	if prior, ok := s.owner[name]; ok && prior != entry {
		return fmt.Errorf("entry %q exports %s, already exported by entry %q", entry, name, prior)
	}
	if err := s.step.Scope.Env.Export(s.step.Spec.ID, name, value, s.step.Spec.Override); err != nil {
		return err
	}
	s.owner[name], s.values[name] = entry, value
	return nil
}

func exportTemplates(entry rawEntry) (map[string]string, error) {
	declared, ok := entry.body["export"]
	if !ok {
		return nil, nil
	}
	mapping, ok := declared.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("entry %q: export must be a mapping of NAME: template", entry.name)
	}
	out := make(map[string]string, len(mapping))
	for name, value := range mapping {
		tmpl, ok := value.(string)
		if !ok {
			return nil, fmt.Errorf("entry %q: export %s must be a string template, got %T", entry.name, name, value)
		}
		out[name] = tmpl
	}
	return out, nil
}

// wildcardExport handles `"*": "{{.outputs.<key>}}"`: every key of that map
// output is exported under its own name.
func wildcardExport(entry, tmpl string, outputs map[string]any, into map[string]string) error {
	match := wildcardPattern.FindStringSubmatch(tmpl)
	if match == nil {
		return fmt.Errorf(`export "*" of entry %q must be exactly {{.outputs.<key>}}, got %q`, entry, tmpl)
	}
	mapping, ok := outputs[match[1]].(map[string]any)
	if !ok {
		return fmt.Errorf(`export "*" of entry %q: outputs.%s is a %s, not a map`, entry, match[1], kindOf(outputs[match[1]]))
	}
	for name, value := range mapping {
		text, ok := value.(string)
		if !ok {
			return fmt.Errorf(`export "*" of entry %q: outputs.%s.%s is a %s, not a string`, entry, match[1], name, kindOf(value))
		}
		into[name] = text
	}
	return nil
}

func kindOf(value any) string {
	switch value.(type) {
	case nil:
		return "missing value"
	case string:
		return "string"
	default:
		return fmt.Sprintf("%T", value)
	}
}

// upperSnake turns an entry name into its env prefix: aimock → AIMOCK,
// gavelServer → GAVEL_SERVER.
func upperSnake(name string) string {
	var b strings.Builder
	runes := []rune(name)
	for i, r := range runes {
		if i > 0 && unicode.IsUpper(r) && (unicode.IsLower(runes[i-1]) || unicode.IsDigit(runes[i-1])) {
			b.WriteByte('_')
		}
		b.WriteRune(unicode.ToUpper(r))
	}
	return b.String()
}

func freePort() (int, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	port := l.Addr().(*net.TCPAddr).Port
	return port, l.Close()
}
