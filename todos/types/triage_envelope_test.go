package types

import (
	"strings"
	"testing"
)

func validTriage() TriageEnvelope {
	return TriageEnvelope{
		ResultEnvelope: ResultEnvelope{Summary: "triaged the issue", EndStatus: EndCompleted},
		Verdict:        VerdictReady,
	}
}

func TestTriageEnvelopeValidate(t *testing.T) {
	for _, tc := range []struct {
		name    string
		mutate  func(*TriageEnvelope)
		wantErr string
	}{
		{name: "ready with a verdict is valid", mutate: func(*TriageEnvelope) {}},
		{
			name:   "shape with a body is valid",
			mutate: func(e *TriageEnvelope) { e.Verdict = VerdictShape; e.Body = "## Acceptance Criteria" },
		},
		{
			name:   "retire with a comment is valid",
			mutate: func(e *TriageEnvelope) { e.Verdict = VerdictRetire; e.Comment = "superseded by #12" },
		},
		{
			name:    "missing verdict rejected",
			mutate:  func(e *TriageEnvelope) { e.Verdict = "" },
			wantErr: "verdict",
		},
		{
			name:    "unknown verdict rejected",
			mutate:  func(e *TriageEnvelope) { e.Verdict = "archive" },
			wantErr: "verdict",
		},
		{
			// shape means "rewrite it"; without a body nothing would be rewritten and
			// the verdict would silently do nothing.
			name:    "shape without a body rejected",
			mutate:  func(e *TriageEnvelope) { e.Verdict = VerdictShape },
			wantErr: "body",
		},
		{
			// Retiring work without saying why is unreviewable.
			name:    "retire without a comment rejected",
			mutate:  func(e *TriageEnvelope) { e.Verdict = VerdictRetire },
			wantErr: "comment",
		},
		{
			// Projected statuses are owned by run history; storage would decline the
			// write, so accepting it here would silently drop the agent's intent.
			name:    "projected status rejected",
			mutate:  func(e *TriageEnvelope) { e.Status = "review" },
			wantErr: "projected",
		},
		{
			name:    "unknown status rejected",
			mutate:  func(e *TriageEnvelope) { e.Status = "archived" },
			wantErr: "unknown status",
		},
		{
			name:   "assignable status accepted",
			mutate: func(e *TriageEnvelope) { e.Status = "completed" },
		},
		{
			name:    "unknown priority rejected",
			mutate:  func(e *TriageEnvelope) { e.Priority = "urgent" },
			wantErr: "priority",
		},
		{
			name:   "known priority accepted",
			mutate: func(e *TriageEnvelope) { e.Priority = "high" },
		},
		{
			// An agent that stopped to ask a question has no verdict to honour, so the
			// verdict contract must not apply.
			name: "ask without a verdict is valid",
			mutate: func(e *TriageEnvelope) {
				e.EndStatus = EndAsk
				e.Verdict = ""
				e.Questions = []AgentQuestion{{Text: "retire this?"}}
			},
		},
		{
			name:    "empty summary rejected",
			mutate:  func(e *TriageEnvelope) { e.Summary = "  " },
			wantErr: "summary",
		},
		{
			name:   "a label delta is valid",
			mutate: func(e *TriageEnvelope) { e.AddLabels = []string{"bug"}; e.RemoveLabels = []string{"docs"} },
		},
		{
			// Applying both would depend on which side won, so the intent is
			// unknowable rather than merely odd.
			name:    "a label both added and removed rejected",
			mutate:  func(e *TriageEnvelope) { e.AddLabels = []string{"Bug"}; e.RemoveLabels = []string{"bug"} },
			wantErr: `"Bug" is both added and removed`,
		},
		{
			name:    "a blank added label rejected",
			mutate:  func(e *TriageEnvelope) { e.AddLabels = []string{"bug", "  "} },
			wantErr: "addLabels contains a blank label",
		},
		{
			name:    "a blank removed label rejected",
			mutate:  func(e *TriageEnvelope) { e.RemoveLabels = []string{""} },
			wantErr: "removeLabels contains a blank label",
		},
		{
			// The taxonomy check needs the provider, so it belongs to ApplyTriage —
			// Validate must not reject an unlisted label and hide the real message.
			name:   "an unlisted label is not rejected here",
			mutate: func(e *TriageEnvelope) { e.AddLabels = []string{"nothing-defines-this"} },
		},
		{
			name: "merge-into with merges and a body is valid",
			mutate: func(e *TriageEnvelope) {
				e.Verdict, e.Merges, e.Body = VerdictMergeInto, []string{"ab12cd"}, "## Acceptance Criteria"
			},
		},
		{
			// A fold with nothing to fold is a no-op wearing a destructive verdict.
			name:    "merge-into without merges rejected",
			mutate:  func(e *TriageEnvelope) { e.Verdict, e.Body = VerdictMergeInto, "## Acceptance Criteria" },
			wantErr: "requires merges",
		},
		{
			// Folding without merged content discards the absorbed descriptions, which
			// is the whole point of the verdict.
			name:    "merge-into without a body rejected",
			mutate:  func(e *TriageEnvelope) { e.Verdict, e.Merges = VerdictMergeInto, []string{"ab12cd"} },
			wantErr: "requires a body",
		},
		{
			name: "merge-into with duplicateOf rejected",
			mutate: func(e *TriageEnvelope) {
				e.Verdict, e.Merges, e.Body, e.DuplicateOf = VerdictMergeInto, []string{"ab12cd"}, "body", "ff0011"
			},
			wantErr: "uses merges, not duplicateOf",
		},
		{
			name: "merges naming the same TODO twice rejected",
			mutate: func(e *TriageEnvelope) {
				e.Verdict, e.Merges, e.Body = VerdictMergeInto, []string{"ab12cd", "AB12CD"}, "body"
			},
			wantErr: "twice",
		},
		{
			name: "a blank merges id rejected",
			mutate: func(e *TriageEnvelope) {
				e.Verdict, e.Merges, e.Body = VerdictMergeInto, []string{"ab12cd", " "}, "body"
			},
			wantErr: "blank id",
		},
		{
			name: "duplicate-of with a target and a comment is valid",
			mutate: func(e *TriageEnvelope) {
				e.Verdict, e.DuplicateOf, e.Comment = VerdictDuplicateOf, "ff0011", "ff0011 covers it"
			},
		},
		{
			name:    "duplicate-of without a target rejected",
			mutate:  func(e *TriageEnvelope) { e.Verdict, e.Comment = VerdictDuplicateOf, "why" },
			wantErr: "requires duplicateOf",
		},
		{
			name:    "duplicate-of without a comment rejected",
			mutate:  func(e *TriageEnvelope) { e.Verdict, e.DuplicateOf = VerdictDuplicateOf, "ff0011" },
			wantErr: "requires a comment",
		},
		{
			// The TODO is on its way out; an edit or a status write would be applied to
			// something that is about to be closed.
			name: "duplicate-of carrying a body rejected",
			mutate: func(e *TriageEnvelope) {
				e.Verdict, e.DuplicateOf, e.Comment, e.Body = VerdictDuplicateOf, "ff0011", "why", "rewritten"
			},
			wantErr: "cannot also set body",
		},
		{
			name: "duplicate-of carrying a status rejected",
			mutate: func(e *TriageEnvelope) {
				e.Verdict, e.DuplicateOf, e.Comment, e.Status = VerdictDuplicateOf, "ff0011", "why", "completed"
			},
			wantErr: "cannot also set status",
		},
		{
			// Duplicates belong to the verdicts that record where the work went.
			name:    "retire with duplicateOf rejected",
			mutate:  func(e *TriageEnvelope) { e.Verdict, e.Comment, e.DuplicateOf = VerdictRetire, "obsolete", "ff0011" },
			wantErr: `duplicateOf is only accepted with verdict "duplicate-of"`,
		},
		{
			// retire closes this TODO too, so the same guard applies: an edit written
			// alongside the close lands on something already on its way out.
			name:    "retire carrying a status rejected",
			mutate:  func(e *TriageEnvelope) { e.Verdict, e.Comment, e.Status = VerdictRetire, "obsolete", "completed" },
			wantErr: "cannot also set status",
		},
		{
			name:    "retire carrying a body rejected",
			mutate:  func(e *TriageEnvelope) { e.Verdict, e.Comment, e.Body = VerdictRetire, "obsolete", "rewritten" },
			wantErr: "cannot also set body",
		},
		{
			name:    "retire carrying a title rejected",
			mutate:  func(e *TriageEnvelope) { e.Verdict, e.Comment, e.Title = VerdictRetire, "obsolete", "Renamed" },
			wantErr: "cannot also set title",
		},
		{
			name: "retire carrying a fixture rejected",
			mutate: func(e *TriageEnvelope) {
				e.Verdict, e.Comment, e.Verification = VerdictRetire, "obsolete", "```exec\ntrue\n```"
			},
			wantErr: "cannot also set verification",
		},
		{
			name:    "merges under another verdict rejected",
			mutate:  func(e *TriageEnvelope) { e.Merges = []string{"ff0011"} },
			wantErr: `merges is only accepted with verdict "merge-into"`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			env := validTriage()
			tc.mutate(&env)
			err := env.Validate()
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("Validate() = %v, want nil", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("Validate() = nil, want error containing %q", tc.wantErr)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("Validate() = %v, want error containing %q", err, tc.wantErr)
			}
		})
	}
}

// ChangesFixture decides which TODOs earn a verification run after a bulk
// triage, so a wrong answer either wastes a test suite or leaves a rewritten
// fixture unproven.
func TestTriageEnvelopeChangesFixture(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*TriageEnvelope)
		want   bool
	}{
		{name: "ready with no fixture change", mutate: func(*TriageEnvelope) {}, want: false},
		{
			name:   "rewritten fixture",
			mutate: func(e *TriageEnvelope) { e.Verdict = VerdictShape; e.Verification = "```yaml test\n```" },
			want:   true,
		},
		{
			// A done verdict is a claim that the fixture already passes; running it is
			// how the claim gets checked.
			name:   "done claims completion",
			mutate: func(e *TriageEnvelope) { e.Verdict = VerdictDone },
			want:   true,
		},
		{name: "investigate", mutate: func(e *TriageEnvelope) { e.Verdict = VerdictInvestigate }, want: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			env := validTriage()
			tc.mutate(&env)
			if got := env.ChangesFixture(); got != tc.want {
				t.Fatalf("ChangesFixture() = %v, want %v", got, tc.want)
			}
		})
	}
}

// RetiresTODOs is how a caller knows the TODO may not be there afterwards, and
// RetirementTargets names what else the verdict closes.
func TestTriageEnvelopeRetirements(t *testing.T) {
	for _, tc := range []struct {
		name        string
		mutate      func(*TriageEnvelope)
		wantRetires bool
		wantTargets []string
	}{
		{name: "ready retires nothing", mutate: func(*TriageEnvelope) {}},
		{
			// retire closes the TODO being triaged, which is not a target the caller
			// has to resolve — there is no survivor to hand the work to.
			name:        "retire closes this one",
			mutate:      func(e *TriageEnvelope) { e.Verdict = VerdictRetire },
			wantRetires: true,
		},
		{
			name: "merge-into closes the folded TODOs",
			mutate: func(e *TriageEnvelope) {
				e.Verdict, e.Merges = VerdictMergeInto, []string{"ff0011", " cc3344 ", ""}
			},
			wantRetires: true,
			wantTargets: []string{"ff0011", "cc3344"},
		},
		{
			// duplicate-of closes the TODO being triaged, which is not a target the
			// caller has to resolve.
			name:        "duplicate-of closes this one",
			mutate:      func(e *TriageEnvelope) { e.Verdict, e.DuplicateOf = VerdictDuplicateOf, "ff0011" },
			wantRetires: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			env := validTriage()
			tc.mutate(&env)
			if got := env.RetiresTODOs(); got != tc.wantRetires {
				t.Errorf("RetiresTODOs() = %v, want %v", got, tc.wantRetires)
			}
			if got := env.RetirementTargets(); strings.Join(got, ",") != strings.Join(tc.wantTargets, ",") {
				t.Errorf("RetirementTargets() = %v, want %v", got, tc.wantTargets)
			}
		})
	}
}
