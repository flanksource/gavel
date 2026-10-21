// Package services implements the `kind: services` fixture step: it starts a
// set of services — deps catalog services, supervised exec processes and
// seeded git repositories — in declaration order, exports their URLs and
// ports into the fixture Env, and stops them in reverse order at teardown.
//
// Declaration order matters: an entry may reference any earlier entry as
// `{{.services.<name>.url}}`. `services:` is read either as a mapping, whose
// order comes from the JSON token stream (the YAML frontend must therefore
// keep mapping order when it lowers YAML to JSON), or as a list of entries
// carrying a `name:`.
package services

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/flanksource/clicky/task"

	"github.com/flanksource/gavel/fixtures"
)

//go:embed schema.json
var schema []byte

func init() { fixtures.RegisterKind(Kind{}) }

// Kind is the `services` fixture step kind.
type Kind struct{}

func (Kind) Name() string { return "services" }

// TemplatesOwnConfig opts out of the executor's config rendering: `.port` and
// `.services.<name>` only exist while the entries start, one after another.
func (Kind) TemplatesOwnConfig() {}

func (Kind) Schema() json.RawMessage { return schema }

// Execute starts every entry in order, rendering each against the outputs of
// the entries before it, and fails fast — stopping what it already started —
// on the first entry that cannot start or become ready. The failed outcome
// still carries the outputs recorded so far, including the failing entry's
// readiness probes.
func (Kind) Execute(ctx context.Context, step *fixtures.StepContext) (fixtures.StepOutcome, error) {
	began := time.Now()
	if step.Spec.ID == "" {
		return fixtures.StepOutcome{}, errors.New("a services step needs an id: its exports and steps.<id> outputs are keyed by it")
	}
	raws, err := parseConfig(step.Spec.Config)
	if err != nil {
		return fixtures.StepOutcome{}, fmt.Errorf("services step %q: %w", step.Spec.ID, err)
	}
	run := &running{}
	exports := newExportSet(step)
	started := map[string]any{}
	for _, raw := range raws {
		svc, err := startEntry(ctx, step, raw, started)
		if svc != nil {
			run.entries = append(run.entries, svc)
		}
		if err == nil {
			started[raw.name] = svc.outputs
			err = exports.exportEntry(raw, svc.outputs, started)
		}
		if err != nil {
			return fixtures.StepOutcome{Outputs: run.outputs(nil)},
				errors.Join(fmt.Errorf("services step %q: %w", step.Spec.ID, err), run.stop(context.WithoutCancel(ctx)))
		}
	}
	name := step.Spec.Name
	if name == "" {
		name = step.Spec.ID
	}
	return fixtures.StepOutcome{
		Result:  fixtures.FixtureResult{Name: name, Type: "services", Status: task.StatusPASS, Duration: time.Since(began)},
		Outputs: run.outputs(nil),
		Exports: exports.values,
		Handle:  run,
		Cleanup: run.stop,
	}, nil
}

// Observe re-probes every entry's readiness. Each entry's outputs gain fresh
// `ready` (and `http.status`/`error`) fields, and the top-level `ready` is
// true only when all of them are.
func (Kind) Observe(ctx context.Context, _ *fixtures.StepContext, handle any) (map[string]any, error) {
	run, ok := handle.(*running)
	if !ok {
		return nil, fmt.Errorf("services: Observe got a %T handle, want the one Execute returned", handle)
	}
	observed := make(map[string]map[string]any, len(run.entries))
	for _, svc := range run.entries {
		observed[svc.name] = svc.observe(ctx)
	}
	return run.outputs(observed), nil
}

func startEntry(ctx context.Context, step *fixtures.StepContext, raw rawEntry, started map[string]any) (*startedEntry, error) {
	entry, port, err := renderEntry(step, raw, started)
	if err != nil {
		return nil, err
	}
	svc := &startedEntry{name: entry.Name}
	switch entry.typeName() {
	case "exec":
		svc.exec, svc.outputs, err = startExec(ctx, step, entry, port)
	case "deps":
		svc.deps, svc.outputs, err = startDeps(ctx, step, entry)
	default:
		svc.outputs, err = seedGit(ctx, step, entry)
	}
	// An entry that failed after recording outputs (its readiness probes) is
	// returned with them; it holds no process, so stopping it is a no-op.
	if err != nil && svc.outputs == nil {
		return nil, err
	}
	return svc, err
}

// running is the Handle and Cleanup owner of an executed services step.
type running struct {
	entries []*startedEntry
}

type startedEntry struct {
	name    string
	exec    *execService
	deps    *depsService
	outputs map[string]any
}

// outputs is `{services: {<name>: outputs}, <name>: outputs, ready}`, with
// observed fields layered over each entry's start outputs.
func (r *running) outputs(observed map[string]map[string]any) map[string]any {
	services := make(map[string]any, len(r.entries))
	out := map[string]any{"services": services}
	ready := true
	for _, svc := range r.entries {
		entry := make(map[string]any, len(svc.outputs))
		for k, v := range svc.outputs {
			entry[k] = v
		}
		for k, v := range observed[svc.name] {
			entry[k] = v
		}
		ready = ready && entry["ready"] == true
		services[svc.name] = entry
		out[svc.name] = entry
	}
	out["ready"] = ready
	return out
}

// stop stops every started entry in reverse order; every entry is stopped
// even when an earlier stop fails.
func (r *running) stop(ctx context.Context) error {
	var errs []error
	for i := len(r.entries) - 1; i >= 0; i-- {
		if err := r.entries[i].stop(ctx); err != nil {
			errs = append(errs, err)
		}
	}
	r.entries = nil
	return errors.Join(errs...)
}

func (s *startedEntry) stop(ctx context.Context) error {
	switch {
	case s.exec != nil:
		return s.exec.stop()
	case s.deps != nil:
		return s.deps.stop(ctx)
	default:
		return nil
	}
}

func (s *startedEntry) observe(ctx context.Context) map[string]any {
	switch {
	case s.exec != nil:
		return s.exec.observe(ctx)
	case s.deps != nil:
		return s.deps.observe(ctx)
	default:
		return map[string]any{"ready": true}
	}
}
