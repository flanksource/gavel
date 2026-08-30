package runtime

import (
	"fmt"
	"testing"

	"github.com/flanksource/gavel/todos/native"
	"github.com/flanksource/gavel/todos/types"
	"github.com/google/uuid"
)

// goldenIssue and goldenVersion pin the inputs the literal UUIDs below were
// derived from.
//
// To regenerate a value, evaluate the pre-naming formula by hand:
//
//	seed      = "<issue>:<step>:<version>"
//	session   = uuid.NewSHA1(uuid.NameSpaceOID, "gavel-todo-session:"+seed)
//	promptRun = uuid.NewSHA1(uuid.NameSpaceOID, "gavel-todo-prompt-run:"+seed)
var goldenIssue = uuid.MustParse("11111111-2222-3333-4444-555555555555")

const goldenVersion = int64(7)

// seedIdentities derives the identities the way production does, so the literal
// expectations below catch a change to promptRunSeed itself.
func seedIdentities(step native.StepKind, prompt string) (session, promptRun string) {
	seed := promptRunSeed(goldenIssue, step, prompt, goldenVersion)
	return uuid.NewSHA1(uuid.NameSpaceOID, []byte("gavel-todo-session:"+seed)).String(),
		uuid.NewSHA1(uuid.NameSpaceOID, []byte("gavel-todo-prompt-run:"+seed)).String()
}

// TestPromptRunSeedBuiltinIdentitiesAreStable pins the durable identity of the
// three built-in operations against literal UUIDs.
//
// Naming prompts meant the seed had to grow a prompt component, and any change
// to the seed silently re-keys every in-flight run: a resumed dispatch would no
// longer recognise its own prompt run and would be rejected as a stale mutation.
// Built-ins must therefore hash exactly as they did before prompts had names.
func TestPromptRunSeedBuiltinIdentitiesAreStable(t *testing.T) {
	for _, tc := range []struct {
		step              native.StepKind
		prompt            string
		wantSession       string
		wantPromptRunUUID string
	}{
		{native.StepRun, "run", "676cd2c8-f81e-50dc-855f-6fe23e6ab8ea", "ffda6967-d576-5e0e-ab65-7b7e56ee8d66"},
		{native.StepPlan, "plan", "4fe86321-f9f8-537a-90df-53e83683f9d5", "9685e778-a709-556f-ac6d-351c1cb11941"},
		{native.StepVerify, "verify", "3a92756b-29ec-5009-be4c-33577760fc47", "a7894403-c2e1-5464-9883-79798192fe55"},
	} {
		t.Run(string(tc.step), func(t *testing.T) {
			session, promptRun := seedIdentities(tc.step, tc.prompt)
			if session != tc.wantSession {
				t.Errorf("session id changed: got %s, want %s", session, tc.wantSession)
			}
			if promptRun != tc.wantPromptRunUUID {
				t.Errorf("prompt run id changed: got %s, want %s", promptRun, tc.wantPromptRunUUID)
			}

			// An unnamed dispatch must hash identically to the explicitly named one,
			// so callers that predate the prompt axis keep their identity.
			bare := promptRunSeed(goldenIssue, tc.step, "", goldenVersion)
			if want := fmt.Sprintf("%s:%s:%d", goldenIssue, tc.step, goldenVersion); bare != want {
				t.Errorf("unnamed seed = %q, want %q", bare, want)
			}
		})
	}
}

// TestPromptRunSeedDistinguishesPromptsOfOneClass is the reason the prompt name
// joins the seed at all: triage and plan are both plan-class, so without it they
// would collide at one issue version and the second dispatch would be refused as
// already claimed.
func TestPromptRunSeedDistinguishesPromptsOfOneClass(t *testing.T) {
	_, planRun := seedIdentities(native.StepPlan, "plan")
	_, triageRun := seedIdentities(native.StepPlan, "triage")
	_, securityRun := seedIdentities(native.StepPlan, "security")

	if planRun == triageRun {
		t.Fatalf("plan and triage share prompt run id %s at one issue version", planRun)
	}
	if triageRun == securityRun {
		t.Fatalf("triage and security share prompt run id %s at one issue version", triageRun)
	}
	if triageRun != "e3250879-4282-531a-a7e3-506ab92c90fb" {
		t.Errorf("triage prompt run id changed: got %s", triageRun)
	}
}

// TestStepForRecordsTriageAsItsOwnKind covers the two-step split: a run is
// RECORDED under the kind a backlog groups by, but SEEDED from its behaviour
// class. Triage is the only prompt where the two differ.
func TestStepForRecordsTriageAsItsOwnKind(t *testing.T) {
	for _, tc := range []struct {
		prompt string
		mode   types.RunMode
		want   native.StepKind
	}{
		{"triage", types.ModePlan, native.StepTriage},
		{"plan", types.ModePlan, native.StepPlan},
		{"run", types.ModeRun, native.StepRun},
		{"verify", types.ModeVerify, native.StepVerify},
		// A configured prompt of plan class is still a planning pass: only the
		// built-in triage name earns its own kind.
		{"security", types.ModePlan, native.StepPlan},
	} {
		t.Run(tc.prompt, func(t *testing.T) {
			got, err := stepFor(tc.prompt, tc.mode)
			if err != nil {
				t.Fatalf("stepFor(%q, %q): %v", tc.prompt, tc.mode, err)
			}
			if got != tc.want {
				t.Errorf("stepFor(%q, %q) = %q, want %q", tc.prompt, tc.mode, got, tc.want)
			}
		})
	}

	// The class step is what seeds identity, so it must NOT follow triage into
	// its own kind — see TestTriageIdentitySurvivesItsOwnStepKind.
	classStep, err := classStepForMode(types.ModePlan)
	if err != nil {
		t.Fatalf("classStepForMode(plan): %v", err)
	}
	if classStep != native.StepPlan {
		t.Errorf("classStepForMode(plan) = %q, want %q", classStep, native.StepPlan)
	}
}

// TestTriageIdentitySurvivesItsOwnStepKind is the regression guard for giving
// triage a step kind. Seeding from the recorded kind would have re-keyed every
// triage dispatch — an in-flight run would stop recognising its own prompt run
// and be rejected as a stale mutation. The literal below is the same UUID
// TestPromptRunSeedDistinguishesPromptsOfOneClass has always pinned.
func TestTriageIdentitySurvivesItsOwnStepKind(t *testing.T) {
	classStep, err := classStepForMode(types.ModePlan)
	if err != nil {
		t.Fatalf("classStepForMode(plan): %v", err)
	}
	_, promptRun := seedIdentities(classStep, "triage")
	if promptRun != "e3250879-4282-531a-a7e3-506ab92c90fb" {
		t.Errorf("triage prompt run id changed: got %s", promptRun)
	}

	// Seeding from the recorded kind instead is exactly the mistake this guards.
	recorded, err := stepFor("triage", types.ModePlan)
	if err != nil {
		t.Fatalf("stepFor(triage, plan): %v", err)
	}
	if _, wrong := seedIdentities(recorded, "triage"); wrong == promptRun {
		t.Fatal("recorded-kind seed collides with class seed; the guard cannot fail")
	}
}

func TestPromptNameOrDefault(t *testing.T) {
	for _, tc := range []struct {
		name string
		mode types.RunMode
		want string
	}{
		{name: "", mode: types.ModeRun, want: "run"},
		{name: "   ", mode: types.ModePlan, want: "plan"},
		{name: "triage", mode: types.ModePlan, want: "triage"},
	} {
		if got := promptNameOrDefault(tc.name, tc.mode); got != tc.want {
			t.Errorf("promptNameOrDefault(%q, %q) = %q, want %q", tc.name, tc.mode, got, tc.want)
		}
	}
}
