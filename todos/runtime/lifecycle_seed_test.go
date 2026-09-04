package runtime

import (
	"testing"

	"github.com/flanksource/gavel/todos/native"
	"github.com/flanksource/gavel/todos/types"
)

// The dispatch identity itself — the seed and the UUIDs it hashes to — is
// lifecycle.Seed/IdentityFor, pinned against golden literals in
// todos/lifecycle. What stays here is the runtime's own mapping: which prompt
// name seeds a run, and which step kind it is recorded under.

// TestStepForRecordsTheLifecycleStepName covers the recording side: a run is
// RECORDED under the step a backlog groups by, and that step is the lifecycle
// step's own name — there is no second vocabulary to map into. A project that
// declares its own step gets runs recorded under it.
func TestStepForRecordsTheLifecycleStepName(t *testing.T) {
	for _, name := range []string{"triage", "plan", "run", "verify", "security-review", "shape-it"} {
		t.Run(name, func(t *testing.T) {
			got, err := stepFor(name)
			if err != nil {
				t.Fatalf("stepFor(%q): %v", name, err)
			}
			if got != native.StepKind(name) {
				t.Errorf("stepFor(%q) = %q, want %q", name, got, name)
			}
		})
	}
}

// A run with no step names nothing a link row, an ordinal sequence or a resume
// could match against, so it is refused rather than defaulted onto one.
func TestStepForRejectsAnUnnamedStep(t *testing.T) {
	for _, name := range []string{"", "   "} {
		if _, err := stepFor(name); err == nil {
			t.Errorf("stepFor(%q) must reject a run that names no lifecycle step", name)
		}
	}
}

// TestPromptNameOrDefault pins the seed's step component: a run dispatched
// without a prompt name seeds as its behaviour class, which is what keeps the
// built-in run/plan/verify identities byte-identical to their pre-lifecycle
// values.
func TestPromptNameOrDefault(t *testing.T) {
	for _, tc := range []struct {
		name string
		mode types.RunMode
		want string
	}{
		{name: "", mode: types.ModeRun, want: "run"},
		{name: "   ", mode: types.ModePlan, want: "plan"},
		{name: "triage", mode: types.ModePlan, want: "triage"},
		{name: "security", mode: types.ModePlan, want: "security"},
	} {
		if got := promptNameOrDefault(tc.name, tc.mode); got != tc.want {
			t.Errorf("promptNameOrDefault(%q, %q) = %q, want %q", tc.name, tc.mode, got, tc.want)
		}
	}
}
