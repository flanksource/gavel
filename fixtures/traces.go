package fixtures

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"os"
	"time"

	"github.com/flanksource/commons-db/tracing/terminal"

	"github.com/flanksource/gavel/testrunner/parsers"
	"github.com/flanksource/gavel/traces"
)

const terminalTraceKind = terminal.Kind

// DeclaredTraces is the fixture's own `traces:`, nil when it declared none: a
// typed step's, else the test's or its file's front matter.
func (fixture FixtureTest) DeclaredTraces() *TraceSpecs {
	if fixture.Step != nil && fixture.Step.Traces != nil {
		return fixture.Step.Traces
	}
	return fixture.ExecBase().Traces
}

// effectiveTraces is what a fixture arms: its own declaration, else the
// run-wide --traces default.
func (r *Runner) effectiveTraces(fixture FixtureTest) TraceSpecs {
	if declared := fixture.DeclaredTraces(); declared != nil {
		return *declared
	}
	if r.options.DefaultTraces != nil {
		return *r.options.DefaultTraces
	}
	return nil
}

// prepareTraces opens the run's trace host when any fixture arms a trace
// (unless the caller handed one in), and starts the run- and file-scope
// captures before the first fixture: a fixture's file scope is seeded with
// their exports. Runs after the daemon, as recorders did.
func (r *Runner) prepareTraces() error {
	shared := map[string]sharedTrace{}
	armed := false
	r.tree.Walk(func(node *FixtureNode) {
		if node.Test == nil {
			return
		}
		armed = armed || r.wantsGoProfiles(*node.Test)
		for _, spec := range r.effectiveTraces(*node.Test) {
			armed = true
			if spec.Scope != TraceScopeTest {
				shared[sharedTraceKey(originFile(node), spec)] = sharedTrace{spec: spec, file: originFile(node)}
			}
		}
	})
	if !armed {
		return nil
	}
	if err := r.openTraceHost(); err != nil {
		return err
	}
	r.sharedTraces = map[string]*traces.Capture{}
	for key, trace := range shared {
		capture, err := r.traceHost.Start(trace.spec.Kind, trace.spec.Params)
		if err != nil {
			return fmt.Errorf("%s: %w", trace.file, err)
		}
		r.sharedTraces[key] = capture
	}
	return nil
}

type sharedTrace struct {
	spec TraceSpec
	file string
}

func (r *Runner) openTraceHost() error {
	if r.traceHost != nil {
		return nil
	}
	if r.options.TraceHost != nil {
		r.traceHost = r.options.TraceHost
		return nil
	}
	if r.options.WorkDir == "" {
		wd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("open trace host: %w", err)
		}
		r.options.WorkDir = wd
	}
	host, err := traces.Open(r.options.WorkDir, traces.RunID(time.Now()))
	if err != nil {
		return err
	}
	r.traceHost, r.ownsTraceHost = host, true
	return nil
}

// closeTraces stops the run- and file-scope captures after every file's
// teardown, so a service stopping is still traced, reads the --profile
// samples, then closes a host the runner opened itself, copies its store to
// TraceArchive and prunes old runs' stores. A later call does nothing.
func (r *Runner) closeTraces() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	errs := []error{r.finishProfile(ctx)}
	if r.traceHost == nil {
		return errors.Join(errs...)
	}
	for _, capture := range r.sharedTraces {
		_, err := capture.Stop(ctx)
		errs = append(errs, err)
	}
	r.sharedTraces = nil
	if r.ownsTraceHost {
		errs = append(errs, r.traceHost.Close(ctx), r.traceHost.Persist(ctx, r.options.TraceArchive), traces.Prune(r.options.WorkDir, traces.KeepRuns))
	}
	r.traceHost, r.ownsTraceHost = nil, false
	return errors.Join(errs...)
}

// sharedTraceKey names the capture a file- or run-scope entry shares: one per
// file for file scope, one per run for run scope.
func sharedTraceKey(file string, spec TraceSpec) string {
	if spec.Scope == TraceScopeFile {
		return file + "|" + spec.key()
	}
	return spec.key()
}

// fileTraceEnv is what the file's step scope is seeded with: the exports of
// every run- and file-scope capture one of its fixtures arms.
func (r *Runner) fileTraceEnv(file *FixtureNode) map[string]string {
	env := map[string]string{}
	file.Walk(func(node *FixtureNode) {
		if node.Test == nil {
			return
		}
		for _, spec := range r.effectiveTraces(*node.Test) {
			if capture := r.sharedTraces[sharedTraceKey(originFile(node), spec)]; capture != nil {
				maps.Copy(env, capture.Exports())
			}
		}
	})
	return env
}

// armTraces arms what a fixture traces: it starts its test-scope captures and
// marks where its window of each shared one begins. Nil when it traces nothing.
func (r *Runner) armTraces(ctx context.Context, fixture FixtureTest, env fixtureEnv) (*FixtureTraces, error) {
	specs := r.effectiveTraces(fixture)
	if len(specs) == 0 && !r.wantsGoProfiles(fixture) {
		return nil, nil
	}
	armed := &FixtureTraces{host: r.traceHost, env: map[string]string{}}
	for _, spec := range specs {
		spec, err := renderTraceParams(spec, fixture, env.scope)
		if err != nil {
			return nil, errors.Join(err, armed.abort())
		}
		if err := armed.arm(ctx, spec, r.sharedTraces[sharedTraceKey(env.file, spec)]); err != nil {
			return nil, errors.Join(err, armed.abort())
		}
	}
	for key := range declaredEnvKeys(env.setup, fixture.FrontMatter) {
		delete(armed.env, key)
	}
	return armed, nil
}

// renderTraceParams renders a test-scope trace's templated params against the
// fixture's scope as it starts: its Env (setup, front matter, the exports of
// the steps before it) and step outputs, beside the test's own variables.
func renderTraceParams(spec TraceSpec, fixture FixtureTest, scope *StepScope) (TraceSpec, error) {
	if !spec.templated() {
		return spec, nil
	}
	if scope == nil {
		return spec, fmt.Errorf("trace %s params are templated, but %s has no step scope to render them against", spec.Kind, fixture.Name)
	}
	vars := map[string]any{}
	maps.Copy(vars, fixture.TemplateVars)
	vars["env"], vars["steps"] = scope.Env.Map(), scope.StepOutputs()
	return spec.rendered(vars)
}

// declaredEnvKeys are the variables the file's setup and front matter set,
// which a trace's exports never override.
func declaredEnvKeys(setup *PreparedSetup, frontMatter FrontMatter) map[string]bool {
	keys := map[string]bool{}
	if setup != nil {
		for key := range setup.Env {
			keys[key] = true
		}
	}
	for key := range frontMatter.Env {
		keys[key] = true
	}
	return keys
}

// FixtureTraces is what one fixture traces while it runs: the overlay its
// test-scope captures put on its children's environment, and the windows it
// is summarized over once its work is done.
type FixtureTraces struct {
	host     *traces.Host
	env      map[string]string
	armed    []armedTrace
	terminal *TerminalTrace
	// profiles are the goprof windows of the profiles the fixture's
	// `goProfiles:` wrote under --profile, one each.
	profiles []parsers.TraceResult

	collected bool
	harvest   TraceHarvest
}

type armedTrace struct {
	spec    TraceSpec
	capture *traces.Capture
	// after is where a shared capture's window begins; a test-scope capture's
	// window is its whole stream.
	after int64
}

func (t *FixtureTraces) arm(ctx context.Context, spec TraceSpec, shared *traces.Capture) error {
	if spec.Scope != TraceScopeTest {
		if shared == nil {
			return fmt.Errorf("trace %s (%s scope) was not started for this fixture", spec.Kind, spec.Scope)
		}
		after, err := shared.Mark(ctx)
		if err != nil {
			return err
		}
		t.armed = append(t.armed, armedTrace{spec: spec, capture: shared, after: after})
		return nil
	}
	capture, err := t.host.Start(spec.Kind, spec.Params)
	if err != nil {
		return err
	}
	t.armed = append(t.armed, armedTrace{spec: spec, capture: capture})
	maps.Copy(t.env, capture.Exports())
	if spec.Kind == terminalTraceKind {
		return t.armTerminal(spec, capture)
	}
	return nil
}

func (t *FixtureTraces) armTerminal(spec TraceSpec, capture *traces.Capture) error {
	params := terminal.Params{Width: terminal.DefaultWidth, Height: terminal.DefaultHeight}
	if len(spec.Params) > 0 {
		if err := json.Unmarshal(spec.Params, &params); err != nil {
			return fmt.Errorf("terminal trace params: %w", err)
		}
	}
	t.terminal = &TerminalTrace{Stream: capture.Stream(), Width: params.Width, Height: params.Height}
	return nil
}

// abort stops the test-scope captures a fixture armed before arming failed.
func (t *FixtureTraces) abort() error {
	var errs []error
	for _, armed := range t.armed {
		if armed.spec.Scope == TraceScopeTest {
			_, err := armed.capture.Stop(context.Background())
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// Env is the overlay the fixture's test-scope captures put on its children's
// environment; nil-safe.
func (t *FixtureTraces) Env() map[string]string {
	if t == nil {
		return nil
	}
	return t.env
}

// TerminalTrace is a terminal capture the fixture's PTY feeds. Fed is set by
// the PTY once it closed the capture's writer, which ends the capture.
type TerminalTrace struct {
	Stream        string
	Width, Height int
	Fed           bool
}

// Terminal is the terminal capture armed for the fixture, nil when none is:
// a fixture that arms one runs its command under a PTY of the trace's size.
// Nil-safe.
func (t *FixtureTraces) Terminal() *TerminalTrace {
	if t == nil {
		return nil
	}
	return t.terminal
}

// terminalEndGrace is how long a fed terminal capture has to commit its final
// screen and exit once its writer closed.
const terminalEndGrace = 10 * time.Second

// endTerminal waits for a fed terminal capture to end by itself — stopping it
// first would lose its final screen — and refuses one nothing fed.
func (t *FixtureTraces) endTerminal(ctx context.Context, capture *traces.Capture) error {
	if !t.terminal.Fed {
		return fmt.Errorf("nothing fed it a PTY: only a command fixture (exec) runs under one")
	}
	ctx, cancel := context.WithTimeout(ctx, terminalEndGrace)
	defer cancel()
	return capture.Ended(ctx)
}

// TraceHarvest is what a fixture's traces came to: one result per trace, the
// CEL roots named by kind its expressions assert on, and the error of a trace
// that could not be read or whose `require:` did not hold.
type TraceHarvest struct {
	Results []parsers.TraceResult
	CELVars map[string]any
	Err     error
}

// Collect ends the fixture's windows — stopping its test-scope captures and
// marking its shared ones — and summarizes each. The first call does the
// work, once the fixture's children are done and before its CEL is evaluated;
// later calls return the same harvest. Nil-safe.
func (t *FixtureTraces) Collect(ctx context.Context) TraceHarvest {
	if t == nil {
		return TraceHarvest{}
	}
	if t.collected {
		return t.harvest
	}
	t.collected = true
	t.harvest.CELVars = map[string]any{}
	var errs []error
	for _, armed := range t.armed {
		result, err := t.collect(ctx, armed)
		if err != nil {
			result.Error = err.Error()
			errs = append(errs, fmt.Errorf("trace %s: %w", armed.spec.Kind, err))
		}
		t.harvest.Results = append(t.harvest.Results, result)
		if result.Summary != nil {
			t.harvest.CELVars[armed.spec.Kind] = result.Summary
		}
	}
	t.harvest.Results = append(t.harvest.Results, t.profiles...)
	t.harvest.Err = errors.Join(errs...)
	return t.harvest
}

func (t *FixtureTraces) collect(ctx context.Context, armed armedTrace) (parsers.TraceResult, error) {
	var endErr error
	if armed.spec.Kind == terminalTraceKind {
		endErr = t.endTerminal(ctx, armed.capture)
	}
	var to int64
	var err error
	if armed.spec.Scope == TraceScopeTest {
		to, err = armed.capture.Stop(ctx)
	} else {
		to, err = armed.capture.Mark(ctx)
	}
	result := parsers.TraceResult{RunID: t.host.RunID(), TraceRef: armed.capture.Ref(armed.after, to)}
	if err = errors.Join(endErr, err); err != nil {
		return result, err
	}
	summary, err := t.host.Summarize(ctx, result.TraceRef)
	if err != nil {
		return result, err
	}
	if result.Summary, err = plainSummary(summary); err != nil {
		return result, err
	}
	count, ok := result.Summary["count"].(int64)
	if !ok {
		return result, fmt.Errorf("summary count is %T, not a number", result.Summary["count"])
	}
	result.Count = int(count)
	return result, requireTrace(armed.spec, result.Summary)
}

// requireTrace evaluates a trace's `require:` with the kind's summary as both
// its top-level variables and the root `trace`.
func requireTrace(spec TraceSpec, summary map[string]any) error {
	if spec.Require == "" {
		return nil
	}
	vars := maps.Clone(summary)
	vars["trace"] = summary
	ok, trace, err := evalStepCEL(spec.Require, vars)
	if err != nil {
		return fmt.Errorf("require %s: %w", spec.Require, err)
	}
	if !ok {
		return fmt.Errorf("require %s is false\n%s", spec.Require, trace)
	}
	return nil
}

// plainSummary turns a kind's summary into the maps, slices and numbers CEL
// navigates; a whole number stays an integer, so `count > 0` compares ints.
func plainSummary(summary map[string]any) (map[string]any, error) {
	plain, err := PlainValue(summary)
	if err != nil {
		return nil, err
	}
	asMap, ok := integralNumbers(plain).(map[string]any)
	if !ok {
		return nil, fmt.Errorf("summary is %T, not a map", plain)
	}
	return asMap, nil
}

func integralNumbers(value any) any {
	switch v := value.(type) {
	case map[string]any:
		for key, item := range v {
			v[key] = integralNumbers(item)
		}
	case []any:
		for i, item := range v {
			v[i] = integralNumbers(item)
		}
	case float64:
		if v == float64(int64(v)) {
			return int64(v)
		}
	}
	return value
}

// applyTraces lands a fixture's traces on its result: the trace references
// always, and a trace failure only on a fixture its own assertions passed —
// those are the more specific answer.
func applyTraces(result FixtureResult, harvest TraceHarvest) FixtureResult {
	result.Traces = harvest.Results
	if harvest.Err != nil && isPassingStatus(result.Status) {
		return result.Failf("%s", harvest.Err.Error())
	}
	return result
}
