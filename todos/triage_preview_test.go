package todos

import (
	"context"
	"strings"
	"testing"

	"github.com/flanksource/gavel/todos/types"
)

// A preview must report the fold without performing any part of it.
func TestPreviewTriageReportsAMergeWithoutWriting(t *testing.T) {
	first := backlogTODO("ff0011", "Parser crash on bad input", nil)
	second := backlogTODO("cc3344", "Parser panics on EOF", nil)
	provider := &triageRecorder{known: map[string]*types.TODO{"ff0011": first, "cc3344": second}}
	env := mergeIntoTriage([]string{"ff0011", "cc3344"}, func(e *types.TriageEnvelope) {
		e.Priority = "high"
		e.Comment = "one parser fix, three reports"
		e.AddLabels = []string{"bug"}
	})

	preview, err := PreviewTriage(context.Background(), provider, triageTODO(), env, TriageOptions{})
	if err != nil {
		t.Fatalf("PreviewTriage: %v", err)
	}
	if preview == nil {
		t.Fatal("a merge-into verdict must produce a preview")
	}
	if len(provider.writes) != 0 {
		t.Errorf("a preview must write nothing: %v", provider.writes)
	}
	if preview.Verdict != types.VerdictMergeInto {
		t.Errorf("verdict = %q, want merge-into", preview.Verdict)
	}
	if preview.Survivor != "ab12cd" {
		t.Errorf("survivor = %q, want the TODO being triaged", preview.Survivor)
	}
	if len(preview.Retires) != 2 {
		t.Fatalf("retires = %+v, want both folded TODOs", preview.Retires)
	}
	if preview.Retires[0].Ref != "ff0011" || preview.Retires[0].Title != "Parser crash on bad input" {
		t.Errorf("first retirement = %+v, want ff0011 with its title", preview.Retires[0])
	}
	// The held fields are what tells a reader the merged body did not land either.
	for _, want := range []string{"body", "priority", "comment", "labels"} {
		if !contains(preview.Withheld, want) {
			t.Errorf("withheld = %v, should name %q", preview.Withheld, want)
		}
	}
}

func TestPreviewTriageReportsADuplicate(t *testing.T) {
	survivor := backlogTODO("ff0011", "Fix the parser panic", nil)
	provider := &triageRecorder{known: map[string]*types.TODO{"ff0011": survivor}}
	env := completedTriage(func(e *types.TriageEnvelope) {
		e.Verdict, e.Body, e.DuplicateOf, e.Comment = types.VerdictDuplicateOf, "", "ff0011", "ff0011 has the fixture"
	})

	preview, err := PreviewTriage(context.Background(), provider, triageTODO(), env, TriageOptions{})
	if err != nil {
		t.Fatalf("PreviewTriage: %v", err)
	}
	if len(provider.writes) != 0 {
		t.Errorf("a preview must write nothing: %v", provider.writes)
	}
	if len(preview.Retires) != 1 || preview.Retires[0].Ref != "ab12cd" {
		t.Fatalf("retires = %+v, want the TODO being triaged", preview.Retires)
	}
	if preview.Survivor != "ff0011" {
		t.Errorf("survivor = %q, want ff0011", preview.Survivor)
	}
}

// A retire closes the TODO being triaged with nowhere for the work to go, so the
// preview names it and leaves the survivor clause off the report entirely.
func TestPreviewTriageReportsARetire(t *testing.T) {
	provider := &triageRecorder{}
	env := completedTriage(func(e *types.TriageEnvelope) {
		e.Verdict, e.Body, e.Comment = types.VerdictRetire, "", "the route it fixes was deleted"
	})

	preview, err := PreviewTriage(context.Background(), provider, triageTODO(), env, TriageOptions{})
	if err != nil {
		t.Fatalf("PreviewTriage: %v", err)
	}
	if preview == nil {
		t.Fatal("a retire closes a TODO, so it must be previewable")
	}
	if len(provider.writes) != 0 {
		t.Errorf("a preview must write nothing: %v", provider.writes)
	}
	if len(preview.Retires) != 1 || preview.Retires[0].Ref != "ab12cd" {
		t.Fatalf("retires = %+v, want the TODO being triaged", preview.Retires)
	}
	if preview.Survivor != "" {
		t.Errorf("survivor = %q, want none — a retire hands the work to nobody", preview.Survivor)
	}
	// The line ends at the title: no "— merged into <ref>" clause, because there is
	// no survivor to name.
	if report := preview.String(); !strings.Contains(report, "would close ab12cd (Fix the parser)\n") {
		t.Errorf("report should close the TODO and name no survivor:\n%s", report)
	}
}

// The verdicts that close nothing have nothing to preview, so the caller applies
// them as usual rather than holding a shape rewrite back for no reason.
func TestPreviewTriageIgnoresNonClosingVerdicts(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*types.TriageEnvelope)
	}{
		{name: "shape", mutate: func(*types.TriageEnvelope) {}},
		{name: "ready", mutate: func(e *types.TriageEnvelope) { e.Verdict, e.Body = types.VerdictReady, "" }},
		{name: "investigate", mutate: func(e *types.TriageEnvelope) {
			e.Verdict, e.Body, e.Comment = types.VerdictInvestigate, "", "needs a spike"
		}},
		{name: "an ask session", mutate: func(e *types.TriageEnvelope) {
			e.EndStatus, e.Verdict = types.EndAsk, ""
			e.Questions = []types.AgentQuestion{{Text: "retire this?"}}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			preview, err := PreviewTriage(context.Background(), &triageRecorder{}, triageTODO(), completedTriage(tc.mutate), TriageOptions{})
			if err != nil {
				t.Fatalf("PreviewTriage: %v", err)
			}
			if preview != nil {
				t.Errorf("preview = %+v, want nil for a verdict that closes nothing", preview)
			}
		})
	}
}

// A preview that accepted what the real apply would refuse would be worse than no
// preview: it would report a fold that then failed.
func TestPreviewTriageFailsOnWhatApplyWouldRefuse(t *testing.T) {
	closed := backlogTODO("dd5566", "Already folded", func(todo *types.TODO) { todo.Status = types.StatusCompleted })
	provider := &triageRecorder{known: map[string]*types.TODO{"dd5566": closed}}

	for _, tc := range []struct {
		name    string
		env     *types.TriageEnvelope
		wantErr string
	}{
		{name: "missing target", env: mergeIntoTriage([]string{"nope99"}, nil), wantErr: "could not be resolved"},
		{name: "closed target", env: mergeIntoTriage([]string{"dd5566"}, nil), wantErr: "already closed"},
		{name: "itself", env: mergeIntoTriage([]string{"ab12cd"}, nil), wantErr: "names itself"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := PreviewTriage(context.Background(), provider, triageTODO(), tc.env, TriageOptions{}); err == nil ||
				!strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("PreviewTriage = %v, want an error containing %q", err, tc.wantErr)
			}
		})
	}
}

// The report is what a reader acts on, so it has to name the TODOs and the way
// to apply it.
func TestTriagePreviewStringNamesTheClosuresAndTheCommand(t *testing.T) {
	preview := &TriagePreview{
		Verdict:  types.VerdictMergeInto,
		Survivor: "ab12cd",
		Retires:  []TriagePreviewRetirement{{Ref: "ff0011", Title: "Parser crash", Reason: "Merged into"}},
		Withheld: []string{"body", "priority"},
	}

	report := preview.String()
	for _, want := range []string{
		"merge-into", "not applied", "would close ff0011", "Parser crash",
		"merged into ab12cd", "would also write: body, priority", "--preview",
	} {
		if !strings.Contains(report, want) {
			t.Errorf("report is missing %q:\n%s", want, report)
		}
	}
	if (*TriagePreview)(nil).String() != "" {
		t.Error("a nil preview should render empty")
	}
}

// An applied verdict used to log nothing at all, while a preview of the same
// verdict logged everything it would have done.
func TestRenderTriageAppliedNamesTheFieldsAndTheClosures(t *testing.T) {
	env := mergeIntoTriage([]string{"ff0011", "cc3344"}, func(e *types.TriageEnvelope) {
		e.Priority = "high"
		e.AddLabels = []string{"bug"}
	})

	report := RenderTriageApplied(triageTODO(), env)
	for _, want := range []string{"merge-into", "applied to ab12cd", "body", "priority", "labels", "closed ff0011, cc3344"} {
		if !strings.Contains(report, want) {
			t.Errorf("report is missing %q:\n%s", want, report)
		}
	}
}

// retire and duplicate-of close the TODO being triaged, not another one, so the
// report must name it rather than a survivor. A retire that logged only
// "applied: comment" read as though the TODO were still open — it was.
func TestRenderTriageAppliedNamesTheTriagedTODOWhenItCloses(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*types.TriageEnvelope)
	}{
		{name: "duplicate-of", mutate: func(e *types.TriageEnvelope) {
			e.Verdict, e.Body, e.DuplicateOf, e.Comment = types.VerdictDuplicateOf, "", "ff0011", "ff0011 has the fixture"
		}},
		{name: "retire", mutate: func(e *types.TriageEnvelope) {
			e.Verdict, e.Body, e.Comment = types.VerdictRetire, "", "obsolete"
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			report := RenderTriageApplied(triageTODO(), completedTriage(tc.mutate))
			if !strings.Contains(report, "closed ab12cd") {
				t.Errorf("report should say the triaged TODO was closed:\n%s", report)
			}
		})
	}
	if RenderTriageApplied(triageTODO(), nil) != "" {
		t.Error("a nil envelope should render empty")
	}
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
