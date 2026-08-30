package todos

import (
	"context"
	"strings"
	"testing"

	"github.com/flanksource/gavel/todos/types"
)

// triageRecorder captures the writes ApplyTriage performs, in order, so tests
// assert on the sequence storage actually sees rather than on the envelope.
type triageRecorder struct {
	Provider
	edits    []EditRequest
	states   []StateUpdate
	links    []string
	comments []string
	editErr  error
}

func (p *triageRecorder) Edit(_ context.Context, _ *types.TODO, edit EditRequest) error {
	if p.editErr != nil {
		return p.editErr
	}
	p.edits = append(p.edits, edit)
	return nil
}

func (p *triageRecorder) UpdateState(_ context.Context, _ *types.TODO, update StateUpdate) error {
	p.states = append(p.states, update)
	return nil
}

func (p *triageRecorder) Comment(_ context.Context, _ *types.TODO, body string) error {
	p.comments = append(p.comments, body)
	return nil
}

func (p *triageRecorder) Link(_ context.Context, _ *types.TODO, target string, relation types.RelationKind) (*Link, error) {
	p.links = append(p.links, string(relation)+":"+target)
	return &Link{Relation: relation, TargetShortID: target}, nil
}

// Unlink and Links round out RelationshipProvider; ApplyTriage only links, but
// the type assertion needs the whole interface.
func (p *triageRecorder) Unlink(context.Context, *types.TODO, string, types.RelationKind) error {
	return nil
}

func (p *triageRecorder) Links(context.Context, *types.TODO) ([]Link, error) { return nil, nil }

func triageTODO() *types.TODO {
	return &types.TODO{ID: "11111111-2222-3333-4444-555555555555", ShortID: "ab12cd", TODOFrontmatter: types.TODOFrontmatter{Title: "Fix the parser"}}
}

func completedTriage(mutate func(*types.TriageEnvelope)) *types.TriageEnvelope {
	env := &types.TriageEnvelope{
		ResultEnvelope: types.ResultEnvelope{Summary: "triaged", EndStatus: types.EndCompleted},
		Verdict:        types.VerdictShape,
		Body:           "## Acceptance Criteria\n\n- [ ] parses",
	}
	mutate(env)
	return env
}

func TestApplyTriageWritesContentStateLinksAndComment(t *testing.T) {
	provider := &triageRecorder{}
	todo := triageTODO()
	env := completedTriage(func(e *types.TriageEnvelope) {
		e.Title = "Fix the parser panic"
		e.Verification = "```yaml test\npackages: ./todos\n```"
		e.Priority = "high"
		e.Status = "pending"
		e.DuplicateOf = "ff0011"
		e.Related = []string{"cc3344"}
		e.Comment = "compacted; fixture now scoped to the diff"
	})

	if err := ApplyTriage(context.Background(), provider, todo, env, TriageOptions{}); err != nil {
		t.Fatalf("ApplyTriage: %v", err)
	}

	if len(provider.edits) != 1 {
		t.Fatalf("content writes = %d, want exactly 1 (each refreshes the optimistic lock)", len(provider.edits))
	}
	edit := provider.edits[0]
	if edit.Title == nil || *edit.Title != "Fix the parser panic" {
		t.Errorf("title not written: %+v", edit.Title)
	}
	if edit.Body == nil || !strings.Contains(*edit.Body, "Acceptance Criteria") {
		t.Errorf("body not written: %+v", edit.Body)
	}
	if edit.Verification == nil || !strings.Contains(*edit.Verification, "yaml test") {
		t.Errorf("fixture not written: %+v", edit.Verification)
	}

	if len(provider.states) != 1 {
		t.Fatalf("state writes = %d, want exactly 1", len(provider.states))
	}
	if provider.states[0].Status == nil || *provider.states[0].Status != types.StatusPending {
		t.Errorf("status = %+v, want pending", provider.states[0].Status)
	}
	if provider.states[0].Priority == nil || *provider.states[0].Priority != types.PriorityHigh {
		t.Errorf("priority = %+v, want high", provider.states[0].Priority)
	}

	wantLinks := []string{"related_to:ff0011", "related_to:cc3344"}
	if strings.Join(provider.links, ",") != strings.Join(wantLinks, ",") {
		t.Errorf("links = %v, want %v", provider.links, wantLinks)
	}
	if len(provider.comments) != 1 || provider.comments[0] != env.Comment {
		t.Errorf("comments = %v, want the rationale", provider.comments)
	}
}

// A fixture is the only thing that can prove a TODO done. Storing one that
// cannot be parsed would replace a working definition of done with a silent one,
// and the failure would not surface until someone ran the check.
func TestApplyTriageRejectsAnUnparseableFixtureBeforeWriting(t *testing.T) {
	provider := &triageRecorder{}
	env := completedTriage(func(e *types.TriageEnvelope) {
		e.Verification = "---\nthis: [is not: valid yaml\n---\n"
	})

	err := ApplyTriage(context.Background(), provider, triageTODO(), env, TriageOptions{})
	if err == nil {
		t.Fatal("an unparseable fixture must fail the triage")
	}
	if !strings.Contains(err.Error(), "ab12cd") {
		t.Errorf("error %q should name the TODO", err)
	}
	if len(provider.edits) != 0 || len(provider.states) != 0 || len(provider.comments) != 0 {
		t.Errorf("nothing may be written before the fixture validates: %+v", provider)
	}
}

// Whether the work is finished is decided by running the definition of done, not
// by the agent's opinion of it.
func TestApplyTriageDoneVerdictNeverWritesStatus(t *testing.T) {
	provider := &triageRecorder{}
	env := completedTriage(func(e *types.TriageEnvelope) {
		e.Verdict = types.VerdictDone
		e.Body = ""
		e.Status = "completed"
		e.Priority = "low"
		e.Comment = "already implemented in parser.go"
	})

	if err := ApplyTriage(context.Background(), provider, triageTODO(), env, TriageOptions{}); err != nil {
		t.Fatalf("ApplyTriage: %v", err)
	}
	if len(provider.states) != 1 {
		t.Fatalf("state writes = %d, want 1 (priority only)", len(provider.states))
	}
	if provider.states[0].Status != nil {
		t.Errorf("done verdict wrote status %v; the check must decide", *provider.states[0].Status)
	}
	if provider.states[0].Priority == nil {
		t.Error("priority should still be applied")
	}
}

// An ask session reported no verdict to honour, so nothing may be applied.
func TestApplyTriageIgnoresAnAskOutcome(t *testing.T) {
	provider := &triageRecorder{}
	env := completedTriage(func(e *types.TriageEnvelope) {
		e.EndStatus = types.EndAsk
		e.Body = "should not be written"
	})

	if err := ApplyTriage(context.Background(), provider, triageTODO(), env, TriageOptions{}); err != nil {
		t.Fatalf("ApplyTriage: %v", err)
	}
	if len(provider.edits) != 0 {
		t.Errorf("ask outcome wrote content: %+v", provider.edits)
	}
}

func TestApplyTriageSkipsSelfLinks(t *testing.T) {
	provider := &triageRecorder{}
	todo := triageTODO()
	env := completedTriage(func(e *types.TriageEnvelope) {
		e.DuplicateOf = todo.ShortID
		e.Related = []string{todo.ID, "cc3344", "cc3344"}
	})

	if err := ApplyTriage(context.Background(), provider, todo, env, TriageOptions{}); err != nil {
		t.Fatalf("ApplyTriage: %v", err)
	}
	if strings.Join(provider.links, ",") != "related_to:cc3344" {
		t.Errorf("links = %v, want only the deduplicated foreign reference", provider.links)
	}
}

// A TODO without a short id still has to be nameable in an error — the fallback
// path must terminate.
func TestTriageRefFallsBackWhenThereIsNoShortID(t *testing.T) {
	for _, tc := range []struct {
		name string
		todo *types.TODO
		want string
	}{
		{name: "short id preferred", todo: triageTODO(), want: "ab12cd"},
		{name: "falls back to id", todo: &types.TODO{ID: "abc"}, want: "abc"},
		{name: "falls back to path", todo: &types.TODO{FilePath: "todo.md", ID: "abc"}, want: "todo.md"},
		{name: "nil is empty", todo: nil, want: ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := triageRef(tc.todo); got != tc.want {
				t.Fatalf("triageRef() = %q, want %q", got, tc.want)
			}
		})
	}
}

// A dropped duplicate link is how two TODOs stay duplicated, so a provider that
// cannot link must fail loudly rather than skip.
func TestApplyTriageFailsWhenLinksAreUnsupported(t *testing.T) {
	env := completedTriage(func(e *types.TriageEnvelope) { e.DuplicateOf = "ff0011" })
	err := ApplyTriage(context.Background(), &noLinkProvider{}, triageTODO(), env, TriageOptions{})
	if err == nil || !strings.Contains(err.Error(), "does not support links") {
		t.Fatalf("ApplyTriage error = %v, want an unsupported-links failure", err)
	}
}

// noLinkProvider implements Provider's write methods but not RelationshipProvider.
type noLinkProvider struct{ Provider }

func (p *noLinkProvider) Edit(context.Context, *types.TODO, EditRequest) error { return nil }
func (p *noLinkProvider) UpdateState(context.Context, *types.TODO, StateUpdate) error {
	return nil
}
func (p *noLinkProvider) Comment(context.Context, *types.TODO, string) error { return nil }

func TestApplyTriageWithNothingToApplyWritesNothing(t *testing.T) {
	provider := &triageRecorder{}
	env := completedTriage(func(e *types.TriageEnvelope) {
		e.Verdict = types.VerdictInvestigate
		e.Body = ""
	})

	if err := ApplyTriage(context.Background(), provider, triageTODO(), env, TriageOptions{}); err != nil {
		t.Fatalf("ApplyTriage: %v", err)
	}
	if len(provider.edits) != 0 || len(provider.states) != 0 || len(provider.comments) != 0 || len(provider.links) != 0 {
		t.Errorf("investigate with no payload should write nothing: %+v", provider)
	}
}
