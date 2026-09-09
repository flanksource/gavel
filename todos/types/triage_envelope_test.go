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
