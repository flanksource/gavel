package fixtures

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/flanksource/gomplate/v3"
)

// StepScope is the state one fixture run shares across its steps: the
// fixture-scoped environment, every completed step's outputs, the teardown
// stack and the running background steps. One is created per fixture file
// (Runner) or per RunNodes call, and torn down when it ends.
type StepScope struct {
	Env *Env

	ctx        context.Context
	cancel     context.CancelFunc
	mu         sync.Mutex
	outputs    map[string]map[string]any
	cleanups   []namedCleanup
	background []namedBackground
	tmpDir     string
	failed     string
}

// MarkFailed records the first typed step that failed; later steps are skipped.
func (s *StepScope) MarkFailed(step string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.failed == "" {
		s.failed = step
	}
}

// FailedStep is the first typed step that failed, empty while all passed.
func (s *StepScope) FailedStep() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.failed
}

type namedCleanup struct {
	step string
	fn   func(context.Context) error
}

type namedBackground struct {
	step   string
	handle BackgroundHandle
}

// NewStepScope seeds the Env from base (typically os.Environ plus the file's
// setup/env).
func NewStepScope(base []string) *StepScope {
	ctx, cancel := context.WithCancel(context.Background())
	return &StepScope{Env: NewEnv(base), outputs: map[string]map[string]any{}, ctx: ctx, cancel: cancel}
}

// Context lives until Teardown: bind processes that outlive the step that
// started them (a service, a background watcher) to it, not to the step's ctx.
func (s *StepScope) Context() context.Context { return s.ctx }

// TempDir is the fixture-scoped scratch directory (`{{.tmp}}`), created on
// first use and removed at teardown after every other cleanup.
func (s *StepScope) TempDir() (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.tmpDir != "" {
		return s.tmpDir, nil
	}
	dir, err := os.MkdirTemp("", "gavel-fixture-")
	if err != nil {
		return "", fmt.Errorf("create fixture temp dir: %w", err)
	}
	s.tmpDir = dir
	s.cleanups = append([]namedCleanup{{step: "tmp", fn: func(context.Context) error { return os.RemoveAll(dir) }}}, s.cleanups...)
	return dir, nil
}

// SetOutputs records a completed step's outputs under steps.<id>.
func (s *StepScope) SetOutputs(id string, outputs map[string]any) error {
	if id == "" {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.outputs[id]; exists {
		return fmt.Errorf("duplicate step id %q", id)
	}
	s.outputs[id] = outputs
	return nil
}

// StepOutputs is a snapshot of steps.<id>.* for templates and CEL.
func (s *StepScope) StepOutputs() map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make(map[string]any, len(s.outputs))
	for id, outputs := range s.outputs {
		out[id] = outputs
	}
	return out
}

// AddCleanup registers fn to run at teardown, after every later one.
func (s *StepScope) AddCleanup(step string, fn func(context.Context) error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cleanups = append(s.cleanups, namedCleanup{step: step, fn: fn})
}

// AddBackground registers a running background step.
func (s *StepScope) AddBackground(step string, handle BackgroundHandle) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.background = append(s.background, namedBackground{step: step, handle: handle})
}

// Teardown stops background steps (newest first) and returns their verdicts,
// then runs cleanups in reverse registration order. Every cleanup runs even
// when an earlier one fails; their errors are joined.
func (s *StepScope) Teardown(ctx context.Context) ([]FixtureResult, error) {
	s.mu.Lock()
	background, cleanups := s.background, s.cleanups
	s.background, s.cleanups = nil, nil
	s.mu.Unlock()

	results := make([]FixtureResult, 0, len(background))
	for i := len(background) - 1; i >= 0; i-- {
		verdict := background[i].handle.Stop(ctx)
		if verdict.Name == "" {
			verdict.Name = background[i].step
		}
		if verdict.Metadata == nil {
			verdict.Metadata = map[string]interface{}{}
		}
		verdict.Metadata["background"] = true
		results = append(results, verdict)
	}
	var errs []error
	for i := len(cleanups) - 1; i >= 0; i-- {
		if err := cleanups[i].fn(ctx); err != nil {
			errs = append(errs, fmt.Errorf("teardown %s: %w", cleanups[i].step, err))
		}
	}
	s.cancel()
	return results, errors.Join(errs...)
}

// Render runs a gomplate template over vars and nothing else: every `$` is
// left alone, because shell-bound text gets the fixture Env as its process
// environment and the shell expands $X, $?, $1 and loop variables itself.
func (s *StepScope) Render(text string, vars map[string]any) (string, error) {
	return renderGoTemplate(text, vars)
}

// Expand renders text like Render, then expands `$NAME` / `${NAME}` (NAME an
// identifier) from the Env — for strings no shell will see: URLs, refs,
// paths, command arguments. An unset NAME is an error rather than an empty
// string, since a URL built from a missing export would otherwise fail far
// from its cause. `$?`, `$1`, `$$` and a lone `$` are left untouched.
func (s *StepScope) Expand(text string, vars map[string]any) (string, error) {
	rendered, err := renderGoTemplate(text, vars)
	if err != nil {
		return "", err
	}
	expanded, missing := expandIdentifiers(rendered, s.Env.Get)
	if len(missing) > 0 {
		return "", fmt.Errorf("unset environment variable(s) %s in %q", strings.Join(missing, ", "), text)
	}
	return expanded, nil
}

// expandIdentifiers replaces `$NAME` and `${NAME}` where NAME matches
// [A-Za-z_][A-Za-z0-9_]*, returning the names lookup did not find. `$$`
// passes through as two characters.
func expandIdentifiers(text string, lookup func(string) (string, bool)) (string, []string) {
	var out strings.Builder
	var missing []string
	for i := 0; i < len(text); i++ {
		if text[i] != '$' || i+1 >= len(text) {
			out.WriteByte(text[i])
			continue
		}
		name, width := shellIdentifier(text[i+1:])
		if name == "" {
			out.WriteByte('$')
			if text[i+1] == '$' {
				out.WriteByte('$')
				i++
			}
			continue
		}
		if value, ok := lookup(name); ok {
			out.WriteString(value)
		} else {
			missing = append(missing, name)
		}
		i += width
	}
	return out.String(), missing
}

// shellIdentifier reads `NAME` or `{NAME}` at the start of s, returning the
// name and how many bytes it spans; empty when s starts with neither.
func shellIdentifier(s string) (string, int) {
	if strings.HasPrefix(s, "{") {
		end := strings.IndexByte(s, '}')
		if end > 1 && isIdentifier(s[1:end]) {
			return s[1:end], end + 1
		}
		return "", 0
	}
	n := 0
	for n < len(s) && (s[n] == '_' || isASCIILetter(s[n]) || (n > 0 && s[n] >= '0' && s[n] <= '9')) {
		n++
	}
	return s[:n], n
}

func isIdentifier(s string) bool {
	name, width := shellIdentifier(s)
	return name != "" && width == len(s)
}

func isASCIILetter(c byte) bool { return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') }

// missingValue is what text/template prints for a key absent from a map.
const missingValue = "<no value>"

// renderGoTemplate runs text as a gomplate Go template over vars. A reference
// to a key the variables do not hold is an error: text/template renders it as
// "<no value>", which would otherwise reach a command line or a URL silently.
func renderGoTemplate(text string, vars map[string]any) (string, error) {
	if !strings.Contains(text, "{{") {
		return text, nil
	}
	out, err := gomplate.RunTemplate(vars, gomplate.Template{Template: text})
	if err != nil {
		return "", fmt.Errorf("render %q: %w", text, err)
	}
	if strings.Count(out, missingValue) > strings.Count(text, missingValue) {
		return "", fmt.Errorf("render %q: references a value that is not set (available roots: %s)", text, strings.Join(sortedKeys(vars), ", "))
	}
	return out, nil
}

// Env is the fixture-scoped environment. Exports record which step set each
// variable so a run shows where a value came from, and so a second step
// cannot silently clobber the first.
type Env struct {
	mu     sync.RWMutex
	vars   map[string]string
	origin map[string]string
}

// NewEnv builds an Env from KEY=VALUE pairs; seeded variables have no origin
// step and may be re-exported freely.
func NewEnv(base []string) *Env {
	env := &Env{vars: map[string]string{}, origin: map[string]string{}}
	for _, kv := range base {
		if key, value, ok := strings.Cut(kv, "="); ok {
			env.vars[key] = value
		}
	}
	return env
}

// Export sets name=value on behalf of step. Re-exporting a name another step
// already exported fails unless override is set.
func (e *Env) Export(step, name, value string, override bool) error {
	if name == "" {
		return fmt.Errorf("step %q: export with an empty name", step)
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if prior, ok := e.origin[name]; ok && prior != step && !override {
		return fmt.Errorf("step %q: %s was already exported by step %q; set override: true to replace it", step, name, prior)
	}
	e.vars[name] = value
	e.origin[name] = step
	return nil
}

// Get returns the variable's value.
func (e *Env) Get(name string) (string, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	value, ok := e.vars[name]
	return value, ok
}

// Origin returns the step that exported name, empty for seeded variables.
func (e *Env) Origin(name string) string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.origin[name]
}

// Map is a copy of every variable.
func (e *Env) Map() map[string]string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	out := make(map[string]string, len(e.vars))
	for k, v := range e.vars {
		out[k] = v
	}
	return out
}

// Exported lists the variables exported by steps (not seeded), sorted, as
// name → origin step.
func (e *Env) Exported() map[string]string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	out := make(map[string]string, len(e.origin))
	for k, v := range e.origin {
		out[k] = v
	}
	return out
}

// Environ renders the Env as sorted KEY=VALUE pairs for a subprocess.
func (e *Env) Environ() []string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	out := make([]string, 0, len(e.vars))
	for k, v := range e.vars {
		out = append(out, k+"="+v)
	}
	sort.Strings(out)
	return out
}
