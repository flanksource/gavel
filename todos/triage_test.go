package todos

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/flanksource/gavel/todos/labels"
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
	deletes  []string
	editErr  error
	// known is the backlog Get resolves against, keyed by short id. A ref that is
	// absent resolves to an error, which is what an agent naming a TODO that does
	// not exist produces.
	known map[string]*types.TODO
	// writes records every mutating call in order, so a test can assert that a
	// retirement's comment and link landed before its delete.
	writes []string
}

func (p *triageRecorder) Edit(_ context.Context, todo *types.TODO, edit EditRequest) error {
	if p.editErr != nil {
		return p.editErr
	}
	p.edits = append(p.edits, edit)
	p.record("edit", todo)
	return nil
}

func (p *triageRecorder) UpdateState(_ context.Context, todo *types.TODO, update StateUpdate) error {
	p.states = append(p.states, update)
	p.record("state", todo)
	return nil
}

func (p *triageRecorder) Comment(_ context.Context, todo *types.TODO, body string) error {
	p.comments = append(p.comments, body)
	p.record("comment", todo)
	return nil
}

func (p *triageRecorder) Link(_ context.Context, todo *types.TODO, target string, relation types.RelationKind) (*Link, error) {
	p.links = append(p.links, string(relation)+":"+target)
	p.record("link", todo)
	return &Link{Relation: relation, TargetShortID: target}, nil
}

func (p *triageRecorder) Delete(_ context.Context, todo *types.TODO) error {
	p.deletes = append(p.deletes, triageRef(todo))
	p.record("delete", todo)
	return nil
}

func (p *triageRecorder) Get(_ context.Context, ref string) (*types.TODO, error) {
	if todo, ok := p.known[ref]; ok {
		return todo, nil
	}
	return nil, fmt.Errorf("no TODO matched %q", ref)
}

func (p *triageRecorder) record(op string, todo *types.TODO) {
	p.writes = append(p.writes, op+":"+triageRef(todo))
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
		e.Related = []string{"ff0011", "cc3344"}
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

// A dropped link is how two related TODOs stay unconnected, so a provider that
// cannot link must fail loudly rather than skip.
func TestApplyTriageFailsWhenLinksAreUnsupported(t *testing.T) {
	env := completedTriage(func(e *types.TriageEnvelope) { e.Related = []string{"ff0011"} })
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

// labelledTriageProvider is a recorder whose workspace defines its own taxonomy,
// which is what decides whether a proposed label is accepted.
type labelledTriageProvider struct {
	triageRecorder
	definitions labels.Definitions
	countsErr   error
}

func (p *labelledTriageProvider) LabelDefinitions(context.Context) (labels.Definitions, error) {
	return p.definitions, nil
}

func (p *labelledTriageProvider) LabelCounts(context.Context) (map[string]int, error) {
	if p.countsErr != nil {
		return nil, p.countsErr
	}
	return map[string]int{"area/ui": 3}, nil
}

func (p *labelledTriageProvider) SetLabelDefinition(_ context.Context, definition labels.Definition, _ bool) (labels.Definition, error) {
	return definition, nil
}

func (p *labelledTriageProvider) DeleteLabelDefinition(context.Context, string, bool) (labels.Removal, error) {
	return labels.Removal{}, nil
}

// A label write replaces the whole set, so the delta has to be merged against the
// TODO's own labels — otherwise "add bug" would erase everything else it carries,
// starting with the provenance labels sync wrote.
func TestApplyTriageMergesTheLabelDeltaOntoTheTODO(t *testing.T) {
	provider := &triageRecorder{}
	todo := triageTODO()
	todo.Labels = []string{"docs", "source:todo"}
	env := completedTriage(func(e *types.TriageEnvelope) {
		e.Body = ""
		e.Verdict = types.VerdictReady
		e.AddLabels = []string{"bug"}
		e.RemoveLabels = []string{"docs"}
	})

	if err := ApplyTriage(context.Background(), provider, todo, env, TriageOptions{}); err != nil {
		t.Fatalf("ApplyTriage: %v", err)
	}
	if len(provider.edits) != 1 {
		t.Fatalf("content writes = %d, want exactly 1", len(provider.edits))
	}
	got := provider.edits[0].Labels
	if got == nil {
		t.Fatal("labels were not written")
	}
	if want := []string{"source:todo", "bug"}; strings.Join(*got, ",") != strings.Join(want, ",") {
		t.Errorf("labels = %v, want %v", *got, want)
	}
	if strings.Join(todo.Labels, ",") != strings.Join(*got, ",") {
		t.Errorf("in-memory TODO labels = %v, want the written set %v", todo.Labels, *got)
	}
}

// The vocabulary is closed so a backlog does not end up filterable by neither
// "bug" nor "bugs". A rejected label must stop the triage before any write, or the
// TODO would be left half-triaged.
func TestApplyTriageRejectsALabelOutsideTheTaxonomyBeforeWriting(t *testing.T) {
	provider := &triageRecorder{}
	env := completedTriage(func(e *types.TriageEnvelope) {
		e.AddLabels = []string{"bug", "frontend"}
	})

	err := ApplyTriage(context.Background(), provider, triageTODO(), env, TriageOptions{})
	if err == nil {
		t.Fatal("a label outside the taxonomy must fail the triage")
	}
	for _, want := range []string{`"frontend"`, "ab12cd", "bug"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q should mention %s", err, want)
		}
	}
	if len(provider.edits) != 0 || len(provider.states) != 0 || len(provider.comments) != 0 {
		t.Errorf("nothing may be written before the labels validate: %+v", provider)
	}
}

// The workspace taxonomy, not the built-in list, is what a proposal is held to.
func TestApplyTriageAcceptsAWorkspaceDefinedLabel(t *testing.T) {
	provider := &labelledTriageProvider{
		definitions: labels.Definitions{{Name: "area/ui"}, {Name: "bug"}},
	}
	todo := triageTODO()
	env := completedTriage(func(e *types.TriageEnvelope) {
		e.Body = ""
		e.Verdict = types.VerdictReady
		e.AddLabels = []string{"area/ui"}
	})

	if err := ApplyTriage(context.Background(), provider, todo, env, TriageOptions{}); err != nil {
		t.Fatalf("ApplyTriage: %v", err)
	}
	if len(provider.edits) != 1 || provider.edits[0].Labels == nil {
		t.Fatalf("workspace label was not written: %+v", provider.edits)
	}
	if got := *provider.edits[0].Labels; strings.Join(got, ",") != "area/ui" {
		t.Errorf("labels = %v, want [area/ui]", got)
	}
}

// A workspace taxonomy replaces the built-in fallback rather than extending it,
// so a builtin the workspace does not define is not silently allowed.
func TestApplyTriageRejectsABuiltinTheWorkspaceDoesNotDefine(t *testing.T) {
	provider := &labelledTriageProvider{definitions: labels.Definitions{{Name: "area/ui"}}}
	env := completedTriage(func(e *types.TriageEnvelope) { e.AddLabels = []string{"perf"} })

	if err := ApplyTriage(context.Background(), provider, triageTODO(), env, TriageOptions{}); err == nil {
		t.Fatal("a label absent from the workspace taxonomy must be rejected")
	}
}

// Every write refreshes the optimistic lock, so a delta that resolves to the set
// the TODO already has must not consume a version.
func TestApplyTriageSkipsALabelDeltaThatChangesNothing(t *testing.T) {
	provider := &triageRecorder{}
	todo := triageTODO()
	todo.Labels = []string{"bug"}
	env := completedTriage(func(e *types.TriageEnvelope) {
		e.Body = ""
		e.Verdict = types.VerdictReady
		e.AddLabels = []string{"BUG"}
		e.RemoveLabels = []string{"ci"}
	})

	if err := ApplyTriage(context.Background(), provider, todo, env, TriageOptions{}); err != nil {
		t.Fatalf("ApplyTriage: %v", err)
	}
	if len(provider.edits) != 0 {
		t.Errorf("a no-op label delta must not write: %+v", provider.edits)
	}
}

// A TODO may carry a label nothing defines, so a removal is never checked against
// the taxonomy — only additions are.
func TestApplyTriageRemovesAnUndefinedLabel(t *testing.T) {
	provider := &triageRecorder{}
	todo := triageTODO()
	todo.Labels = []string{"legacy-import", "bug"}
	env := completedTriage(func(e *types.TriageEnvelope) {
		e.Body = ""
		e.Verdict = types.VerdictReady
		e.RemoveLabels = []string{"legacy-import"}
	})

	if err := ApplyTriage(context.Background(), provider, todo, env, TriageOptions{}); err != nil {
		t.Fatalf("ApplyTriage: %v", err)
	}
	if len(provider.edits) != 1 || provider.edits[0].Labels == nil {
		t.Fatalf("removal was not written: %+v", provider.edits)
	}
	if got := *provider.edits[0].Labels; strings.Join(got, ",") != "bug" {
		t.Errorf("labels = %v, want [bug]", got)
	}
}

// The prompt section and the write-side check read the same taxonomy, so the list
// the agent is offered is exactly the list it is held to.
func TestLabelTaxonomySectionListsTheResolvedDefinitions(t *testing.T) {
	provider := &labelledTriageProvider{
		definitions: labels.Definitions{
			{Name: "area/ui", Description: "User interface work."},
			{Name: "bug"},
		},
	}

	section := LabelTaxonomySection(context.Background(), provider)
	for _, want := range []string{"`area/ui`", "User interface work.", "(3 in use)", "`bug`"} {
		if !strings.Contains(section, want) {
			t.Errorf("section is missing %q:\n%s", want, section)
		}
	}
	if strings.Contains(section, "`bug` (") {
		t.Errorf("a label nobody uses should carry no count:\n%s", section)
	}

	if names := AllowedLabels(context.Background(), provider); strings.Join(names, ",") != "area/ui,bug" {
		t.Errorf("AllowedLabels = %v, want the same definitions the section listed", names)
	}
}

// A provider with no definition store still offers a real vocabulary: an empty
// one would silently forbid every label instead of the ten that always resolve.
func TestLabelTaxonomyFallsBackToTheBuiltins(t *testing.T) {
	provider := &triageRecorder{}

	if got := AllowedLabels(context.Background(), provider); strings.Join(got, ",") != strings.Join(labels.BuiltinNames(), ",") {
		t.Errorf("AllowedLabels = %v, want the built-in names", got)
	}
	if section := LabelTaxonomySection(context.Background(), provider); !strings.Contains(section, "`bug`") {
		t.Errorf("fallback section should list the builtins:\n%s", section)
	}
}

// Usage counts are a nicety; losing them must not cost the section.
func TestLabelTaxonomySectionSurvivesMissingCounts(t *testing.T) {
	provider := &labelledTriageProvider{
		definitions: labels.Definitions{{Name: "bug", Description: "Something is broken."}},
		countsErr:   errors.New("counts unavailable"),
	}

	section := LabelTaxonomySection(context.Background(), provider)
	if !strings.Contains(section, "`bug` — Something is broken.") {
		t.Errorf("section lost its labels when counts failed:\n%s", section)
	}
}
