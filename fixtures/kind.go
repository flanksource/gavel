package fixtures

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"
)

// Kind is one typed fixture step: exec, ai, test, lint, services, todo,
// todo_prompt, browser, playwright, agent-browser, …. Every step of a YAML/JSON
// fixture document — and every markdown fence, once lowered — is dispatched by
// its Kind name through KindRegistry, so a kind added here runs from every
// frontend and every caller (gavel fixtures run, a TODO's verification, the DoD
// verifier).
//
// Kinds that would make this package import heavy dependencies (captain,
// testrunner, deps, HTTP clients for the todo API) live in their own packages
// and call RegisterKind from init().
type Kind interface {
	// Name is the `kind:` value a step selects this kind with.
	Name() string
	// Schema is the JSON schema of the step's kind-specific config (the keys
	// next to `kind:` minus the common StepSpec keys). A step is validated
	// against it before anything runs.
	Schema() json.RawMessage
	// Execute performs the step's action once and returns its outcome. With a
	// `wait:` policy, the generic waiter then calls Observer.Observe until the
	// policy's CEL verdict holds; with `retry:`, a failed Execute is re-run.
	Execute(ctx context.Context, step *StepContext) (StepOutcome, error)
}

// Observer is implemented by kinds whose state can be polled after Execute —
// a todo's status, a service's readiness, a browser snapshot. Only kinds that
// implement it may carry a `wait:` policy; validation rejects the rest.
type Observer interface {
	Observe(ctx context.Context, step *StepContext, handle any) (map[string]any, error)
}

// DefaultWaiter supplies the `wait:` policy a kind applies when the step
// declares none (e.g. todo_prompt waits for its attempt to be terminal).
// Fixture-declared fields override it field by field.
type DefaultWaiter interface {
	DefaultWait(step StepSpec) *WaitPolicy
}

// FailureCauser is implemented by observing kinds that can name, from a wait's
// last outputs, the domain reason it did not succeed (a todo run's recorded
// error). The step's failure message leads with it, ahead of the wait's own
// condition; an empty cause adds nothing.
type FailureCauser interface {
	FailureCause(outputs map[string]any) string
}

// BackgroundKind is implemented by kinds that can run for the rest of the
// fixture (`background: true`) — a page watcher, a long-lived browser. Start
// returns once the step is running; its verdict is collected at teardown.
//
// ctx bounds the start only: it is cancelled when the step returns. Work that
// outlives the step binds to step.Scope.Context(), which lives until teardown.
// The same holds for a process an Execute leaves running behind a Cleanup.
type BackgroundKind interface {
	Start(ctx context.Context, step *StepContext) (BackgroundHandle, error)
}

// SelfTemplating is implemented by kinds that render their own config, because
// some of its values only exist while Execute runs (a service's allocated
// `{{.port}}`). Every other kind's config string values are rendered as Go
// templates against StepContext's variables before Execute sees them.
type SelfTemplating interface {
	TemplatesOwnConfig()
}

// RawConfigKeys is implemented by kinds with top-level config keys that are
// left unrendered, because the kind templates them itself against variables
// that only exist while it runs (exec's goProfiles args take `{{.path}}`).
type RawConfigKeys interface {
	RawConfigKeys() []string
}

// BackgroundHandle is a running background step. Stop ends it and returns its
// verdict, which is reported as that step's result.
type BackgroundHandle interface {
	Stop(ctx context.Context) FixtureResult
}

// StepSpec is the kind-agnostic envelope of a fixture step. Config holds the
// kind-specific keys, decoded by the kind itself (KindConfig).
type StepSpec struct {
	ID         string            `yaml:"id,omitempty" json:"id,omitempty"`
	Name       string            `yaml:"name,omitempty" json:"name,omitempty"`
	Kind       string            `yaml:"kind" json:"kind"`
	CEL        string            `yaml:"cel,omitempty" json:"cel,omitempty"`
	Export     map[string]string `yaml:"export,omitempty" json:"export,omitempty"`
	Override   bool              `yaml:"override,omitempty" json:"override,omitempty"`
	Background bool              `yaml:"background,omitempty" json:"background,omitempty"`
	// ContinueOnError reports the step's failure without skipping the steps
	// after it: they do not depend on it.
	ContinueOnError bool         `yaml:"continueOnError,omitempty" json:"continueOnError,omitempty"`
	Wait            *WaitPolicy  `yaml:"wait,omitempty" json:"wait,omitempty"`
	Retry           *RetryPolicy `yaml:"retry,omitempty" json:"retry,omitempty"`
	Timeout         *Duration    `yaml:"timeout,omitempty" json:"timeout,omitempty"`
	// Traces arms traces around this step, replacing the document's.
	Traces *TraceSpecs     `yaml:"traces,omitempty" json:"traces,omitempty"`
	Config json.RawMessage `yaml:"-" json:"config,omitempty"`
	// Defaults are the document's `defaults:`, merged under Wait/Retry/Timeout
	// field by field when the step runs. Set by the parser, never by a step.
	Defaults *StepDefaults `yaml:"-" json:"defaults,omitempty"`
}

// StepDefaults is a fixture document's `defaults:` block.
type StepDefaults struct {
	Wait    *WaitPolicy  `yaml:"wait,omitempty" json:"wait,omitempty"`
	Retry   *RetryPolicy `yaml:"retry,omitempty" json:"retry,omitempty"`
	Timeout *Duration    `yaml:"timeout,omitempty" json:"timeout,omitempty"`
}

// StepSpecKeys are the envelope keys; everything else on a step is Config.
var StepSpecKeys = []string{"id", "name", "kind", "cel", "export", "override", "background", "continueOnError", "wait", "retry", "timeout", "traces"}

// EffectiveTimeout is the step's hard cap: its own `timeout:`, else the
// document default; nil when neither declares one.
func (s StepSpec) EffectiveTimeout() *Duration {
	if s.Timeout != nil {
		return s.Timeout
	}
	if s.Defaults != nil {
		return s.Defaults.Timeout
	}
	return nil
}

// StepOutcome is what one Execute (or a background Start) produced.
type StepOutcome struct {
	Result FixtureResult
	// Outputs are this step's CEL/template variables: `steps.<id>.*` for later
	// steps, and the roots its own `cel:` / `wait.until` evaluate against.
	Outputs map[string]any
	// Exports are the kind's default env exports (e.g. TODO_ID, GAVEL_URL); the
	// step's `export:` renames or adds to them.
	Exports map[string]string
	// Handle is passed back to Observer.Observe.
	Handle any
	// Cleanup, when set, runs at fixture teardown in reverse registration order
	// (a started service, a browser process).
	Cleanup func(context.Context) error
}

// StepContext is everything a kind sees when it runs.
type StepContext struct {
	Test    FixtureTest
	Spec    StepSpec
	Opts    RunOptions
	Scope   *StepScope
	WorkDir string
	TmpDir  string
}

// Template renders text as a gomplate template against the fixture scope:
// `{{.env.X}}`, `{{.steps.<id>.…}}`, `{{.fixture.dir}}`, `{{.tmp}}` and the
// test's TemplateVars. `$` is left alone — use it for shell-bound text, whose
// process gets the fixture Env (Environ) and expands $X itself.
func (s *StepContext) Template(text string) (string, error) {
	return s.Scope.Render(text, s.vars())
}

// Expand is Template followed by strict `$NAME` / `${NAME}` expansion from the
// fixture Env (unset is an error; `$?`, `$1`, `$$` stay as they are) — for
// strings no shell sees: URLs, refs, paths, command arguments.
func (s *StepContext) Expand(text string) (string, error) {
	return s.Scope.Expand(text, s.vars())
}

// ChildEnv is the environment a process this step starts inherits: the
// fixture Env with its armed traces' exports (RunOptions.ChildEnv).
func (s *StepContext) ChildEnv() map[string]string { return childEnv(s.Scope, s.Opts.Traces) }

// Environ is ChildEnv as sorted KEY=VALUE pairs, for os/exec.
func (s *StepContext) Environ() []string {
	env := s.ChildEnv()
	pairs := make([]string, 0, len(env))
	for key, value := range env {
		pairs = append(pairs, key+"="+value)
	}
	sort.Strings(pairs)
	return pairs
}

func (s *StepContext) vars() map[string]any {
	vars := map[string]any{}
	for k, v := range s.Test.TemplateVars {
		vars[k] = v
	}
	vars["env"] = s.Scope.Env.Map()
	vars["steps"] = s.Scope.StepOutputs()
	vars["tmp"] = s.TmpDir
	vars["fixture"] = map[string]any{"dir": s.Test.SourceDir}
	return vars
}

// KindConfig decodes the step's kind-specific config into T.
func KindConfig[T any](step *StepContext) (T, error) {
	var cfg T
	if len(step.Spec.Config) == 0 {
		return cfg, nil
	}
	if err := json.Unmarshal(step.Spec.Config, &cfg); err != nil {
		return cfg, fmt.Errorf("%s step %q: decode config: %w", step.Spec.Kind, step.Spec.ID, err)
	}
	return cfg, nil
}

// Duration is a time.Duration that (un)marshals as a Go duration string.
type Duration time.Duration

func (d Duration) Std() time.Duration { return time.Duration(d) }

func (d Duration) MarshalJSON() ([]byte, error) { return json.Marshal(time.Duration(d).String()) }

func (d *Duration) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return fmt.Errorf("duration must be a string like \"2s\": %w", err)
	}
	parsed, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	*d = Duration(parsed)
	return nil
}

// WaitPolicy polls a step's observable state until a CEL verdict.
type WaitPolicy struct {
	// Until is a CEL expression over the step's outputs (+ env, steps); the
	// wait succeeds the first observation it is true.
	Until string `yaml:"until,omitempty" json:"until,omitempty"`
	// FailWhen ends the wait early, loudly, when true.
	FailWhen string    `yaml:"failWhen,omitempty" json:"failWhen,omitempty"`
	Interval *Duration `yaml:"interval,omitempty" json:"interval,omitempty"`
	Backoff  *Backoff  `yaml:"backoff,omitempty" json:"backoff,omitempty"`
	// Jitter is a percentage ("20%") or an absolute duration ("500ms").
	Jitter  string    `yaml:"jitter,omitempty" json:"jitter,omitempty"`
	Timeout *Duration `yaml:"timeout,omitempty" json:"timeout,omitempty"`
}

// RetryPolicy re-runs a failed Execute while When holds.
type RetryPolicy struct {
	Attempts int `yaml:"attempts,omitempty" json:"attempts,omitempty"`
	// When is a CEL expression over the failed result (error, exitCode, outputs);
	// empty never retries.
	When    string    `yaml:"when,omitempty" json:"when,omitempty"`
	Delay   *Duration `yaml:"delay,omitempty" json:"delay,omitempty"`
	Backoff *Backoff  `yaml:"backoff,omitempty" json:"backoff,omitempty"`
	Jitter  string    `yaml:"jitter,omitempty" json:"jitter,omitempty"`
}

// Backoff grows a delay geometrically up to Max.
type Backoff struct {
	Factor float64   `yaml:"factor,omitempty" json:"factor,omitempty"`
	Max    *Duration `yaml:"max,omitempty" json:"max,omitempty"`
}

// KindRegistry maps `kind:` names to implementations.
type KindRegistry struct {
	mu    sync.RWMutex
	kinds map[string]Kind
}

// DefaultKinds is the registry every frontend dispatches through.
var DefaultKinds = &KindRegistry{kinds: map[string]Kind{}}

// RegisterKind adds a kind to DefaultKinds; a duplicate name panics, since two
// packages claiming one kind is a wiring bug, not a runtime condition.
func RegisterKind(kind Kind) {
	DefaultKinds.mu.Lock()
	defer DefaultKinds.mu.Unlock()
	if _, exists := DefaultKinds.kinds[kind.Name()]; exists {
		panic(fmt.Sprintf("fixtures: kind %q registered twice", kind.Name()))
	}
	DefaultKinds.kinds[kind.Name()] = kind
}

// Replace swaps the kind registered under kind.Name() (registering it when
// absent) and returns the function that restores the previous state. It is
// the seam tests use to stand in for a kind whose real implementation reaches
// outside the process (an AI provider, a test engine).
func (r *KindRegistry) Replace(kind Kind) (restore func()) {
	r.mu.Lock()
	defer r.mu.Unlock()
	name := kind.Name()
	previous, existed := r.kinds[name]
	r.kinds[name] = kind
	return func() {
		r.mu.Lock()
		defer r.mu.Unlock()
		if existed {
			r.kinds[name] = previous
			return
		}
		delete(r.kinds, name)
	}
}

// Lookup returns the kind registered under name.
func (r *KindRegistry) Lookup(name string) (Kind, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	kind, ok := r.kinds[name]
	if !ok {
		return nil, fmt.Errorf("unknown fixture step kind %q (registered: %v); import the package that registers it", name, r.namesLocked())
	}
	return kind, nil
}

// Names lists the registered kind names, sorted.
func (r *KindRegistry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.namesLocked()
}

func (r *KindRegistry) namesLocked() []string {
	names := make([]string, 0, len(r.kinds))
	for name := range r.kinds {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
