package todos

import (
	"context"
	"strings"
	"testing"

	"github.com/flanksource/gavel/todos/types"
)

// backlogTODO is one other TODO the triaged one can name.
func backlogTODO(shortID, title string, mutate func(*types.TODO)) *types.TODO {
	todo := &types.TODO{
		ID:              "id-" + shortID,
		ShortID:         shortID,
		TODOFrontmatter: types.TODOFrontmatter{Title: title, Status: types.StatusPending, Priority: types.PriorityLow},
	}
	if mutate != nil {
		mutate(todo)
	}
	return todo
}

func mergeIntoTriage(merges []string, mutate func(*types.TriageEnvelope)) *types.TriageEnvelope {
	return completedTriage(func(e *types.TriageEnvelope) {
		e.Verdict = types.VerdictMergeInto
		e.Merges = merges
		e.Body = "## Acceptance Criteria\n\n- [ ] one combined behaviour"
		if mutate != nil {
			mutate(e)
		}
	})
}

// A fold rewrites the survivor once, then closes each absorbed TODO. The order
// matters: the comment and the link must survive a delete that fails.
func TestApplyTriageMergeIntoFoldsAndRetiresEachTODO(t *testing.T) {
	first := backlogTODO("ff0011", "Parser crash on bad input", nil)
	second := backlogTODO("cc3344", "Parser panics on EOF", nil)
	provider := &triageRecorder{known: map[string]*types.TODO{"ff0011": first, "cc3344": second}}
	todo := triageTODO()
	env := mergeIntoTriage([]string{"ff0011", "cc3344"}, func(e *types.TriageEnvelope) {
		e.Comment = "one parser fix, three reports"
	})

	if err := ApplyTriage(context.Background(), provider, todo, env, TriageOptions{}); err != nil {
		t.Fatalf("ApplyTriage: %v", err)
	}

	if len(provider.edits) != 1 || provider.edits[0].Body == nil {
		t.Fatalf("the survivor's merged body was not written: %+v", provider.edits)
	}
	// state lands because the survivor had no priority and inherits the folded low.
	want := []string{
		"edit:ab12cd", "state:ab12cd", "comment:ab12cd",
		"comment:ff0011", "link:ff0011", "delete:ff0011",
		"comment:cc3344", "link:cc3344", "delete:cc3344",
	}
	if strings.Join(provider.writes, " ") != strings.Join(want, " ") {
		t.Errorf("write order =\n  %v\nwant\n  %v", provider.writes, want)
	}
	for _, link := range provider.links {
		if link != "related_to:ab12cd" {
			t.Errorf("link = %q, want every folded TODO linked to the survivor as related_to", link)
		}
	}
	if !strings.Contains(provider.comments[1], "Merged into ab12cd") {
		t.Errorf("retirement comment = %q, should name the survivor", provider.comments[1])
	}
	if !strings.Contains(provider.comments[1], "one parser fix, three reports") {
		t.Errorf("retirement comment = %q, should carry the rationale", provider.comments[1])
	}
}

// Absorbing work never makes it less urgent, and the agent judging the survivor
// has not necessarily read every folded TODO's priority.
func TestApplyTriageMergeIntoRaisesThePriorityToTheHighestFolded(t *testing.T) {
	urgent := backlogTODO("ff0011", "Parser crash", func(todo *types.TODO) { todo.Priority = types.PriorityHigh })
	provider := &triageRecorder{known: map[string]*types.TODO{"ff0011": urgent}}
	todo := triageTODO()
	todo.Priority = types.PriorityLow

	if err := ApplyTriage(context.Background(), provider, todo, mergeIntoTriage([]string{"ff0011"}, nil), TriageOptions{}); err != nil {
		t.Fatalf("ApplyTriage: %v", err)
	}
	if len(provider.states) != 1 || provider.states[0].Priority == nil {
		t.Fatalf("priority was not raised: %+v", provider.states)
	}
	if got := *provider.states[0].Priority; got != types.PriorityHigh {
		t.Errorf("priority = %q, want high", got)
	}
}

// The agent's own severity still wins when it is the higher of the two.
func TestApplyTriageMergeIntoKeepsAHigherProposedPriority(t *testing.T) {
	folded := backlogTODO("ff0011", "Parser crash", func(todo *types.TODO) { todo.Priority = types.PriorityLow })
	provider := &triageRecorder{known: map[string]*types.TODO{"ff0011": folded}}
	todo := triageTODO()
	todo.Priority = types.PriorityLow
	env := mergeIntoTriage([]string{"ff0011"}, func(e *types.TriageEnvelope) { e.Priority = "high" })

	if err := ApplyTriage(context.Background(), provider, todo, env, TriageOptions{}); err != nil {
		t.Fatalf("ApplyTriage: %v", err)
	}
	if got := *provider.states[0].Priority; got != types.PriorityHigh {
		t.Errorf("priority = %q, want the proposed high", got)
	}
}

// duplicate-of closes the TODO being triaged. Nothing survives to edit.
func TestApplyTriageDuplicateOfClosesThisTODO(t *testing.T) {
	survivor := backlogTODO("ff0011", "Fix the parser panic", nil)
	provider := &triageRecorder{known: map[string]*types.TODO{"ff0011": survivor}}
	todo := triageTODO()
	env := completedTriage(func(e *types.TriageEnvelope) {
		e.Verdict = types.VerdictDuplicateOf
		e.Body = ""
		e.DuplicateOf = "ff0011"
		e.Comment = "ff0011 already has the fixture"
	})

	if err := ApplyTriage(context.Background(), provider, todo, env, TriageOptions{}); err != nil {
		t.Fatalf("ApplyTriage: %v", err)
	}
	if len(provider.edits) != 0 {
		t.Errorf("a duplicate must not be edited on its way out: %+v", provider.edits)
	}
	want := []string{"comment:ab12cd", "comment:ab12cd", "link:ab12cd", "delete:ab12cd"}
	if strings.Join(provider.writes, " ") != strings.Join(want, " ") {
		t.Errorf("write order =\n  %v\nwant\n  %v", provider.writes, want)
	}
	if strings.Join(provider.links, ",") != "related_to:ff0011" {
		t.Errorf("links = %v, want a single related_to to the survivor", provider.links)
	}
	if !strings.Contains(provider.comments[1], "Duplicate of ff0011 — Fix the parser panic") {
		t.Errorf("retirement comment = %q, should name the survivor and its title", provider.comments[1])
	}
}

// A retire has no survivor to point at, so it is the rationale comment and the
// soft delete — and nothing else. Before this the verdict wrote the comment and
// left the TODO open in the backlog.
func TestApplyTriageRetireClosesThisTODO(t *testing.T) {
	provider := &triageRecorder{}
	todo := triageTODO()
	env := completedTriage(func(e *types.TriageEnvelope) {
		e.Verdict = types.VerdictRetire
		e.Body = ""
		e.Comment = "The /todos/new route this fixes was deleted."
	})

	if err := ApplyTriage(context.Background(), provider, todo, env, TriageOptions{}); err != nil {
		t.Fatalf("ApplyTriage: %v", err)
	}
	want := []string{"comment:ab12cd", "delete:ab12cd"}
	if strings.Join(provider.writes, " ") != strings.Join(want, " ") {
		t.Errorf("write order =\n  %v\nwant\n  %v", provider.writes, want)
	}
	if len(provider.edits) != 0 {
		t.Errorf("a retired TODO must not be edited on its way out: %+v", provider.edits)
	}
	if len(provider.links) != 0 {
		t.Errorf("a retire hands the work to nobody, so it links to nothing: %v", provider.links)
	}
	if strings.Join(provider.comments, "|") != env.Comment {
		t.Errorf("comments = %v, want only the verdict's rationale", provider.comments)
	}
}

// Retirement cannot be walked back, so every target is proved before the first
// write. A verdict that names something unusable must change nothing at all.
func TestApplyTriageRejectsUnusableRetirementTargets(t *testing.T) {
	closed := backlogTODO("dd5566", "Already folded elsewhere", func(todo *types.TODO) {
		todo.Status = types.StatusCompleted
	})
	open := backlogTODO("ff0011", "Parser crash", nil)

	for _, tc := range []struct {
		name    string
		env     *types.TriageEnvelope
		wantErr string
	}{
		{
			name:    "an id that resolves to nothing",
			env:     mergeIntoTriage([]string{"nope99"}, nil),
			wantErr: "could not be resolved",
		},
		{
			// Delete soft-deletes to cancelled, which projects to completed — so this
			// is also the guard against two TODOs in one batch absorbing each other.
			name:    "a TODO that is already closed",
			env:     mergeIntoTriage([]string{"dd5566"}, nil),
			wantErr: "already closed",
		},
		{
			name:    "its own short id",
			env:     mergeIntoTriage([]string{"ab12cd"}, nil),
			wantErr: "names itself",
		},
		{
			name: "a duplicate-of survivor that does not exist",
			env: completedTriage(func(e *types.TriageEnvelope) {
				e.Verdict, e.Body, e.DuplicateOf, e.Comment = types.VerdictDuplicateOf, "", "nope99", "why"
			}),
			wantErr: "could not be resolved",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			provider := &triageRecorder{known: map[string]*types.TODO{"dd5566": closed, "ff0011": open}}
			err := ApplyTriage(context.Background(), provider, triageTODO(), tc.env, TriageOptions{})
			if err == nil {
				t.Fatalf("ApplyTriage = nil, want an error containing %q", tc.wantErr)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("ApplyTriage = %v, want an error containing %q", err, tc.wantErr)
			}
			if len(provider.writes) != 0 {
				t.Errorf("nothing may be written before the targets resolve: %v", provider.writes)
			}
		})
	}
}

// A closed survivor is fine for a duplicate: "this duplicates work already
// finished" is an ordinary verdict.
func TestApplyTriageDuplicateOfAcceptsAClosedSurvivor(t *testing.T) {
	done := backlogTODO("ff0011", "Parser panic", func(todo *types.TODO) { todo.Status = types.StatusCompleted })
	provider := &triageRecorder{known: map[string]*types.TODO{"ff0011": done}}
	env := completedTriage(func(e *types.TriageEnvelope) {
		e.Verdict, e.Body, e.DuplicateOf, e.Comment = types.VerdictDuplicateOf, "", "ff0011", "shipped in ff0011"
	})

	if err := ApplyTriage(context.Background(), provider, triageTODO(), env, TriageOptions{}); err != nil {
		t.Fatalf("ApplyTriage: %v", err)
	}
	if len(provider.deletes) != 1 {
		t.Errorf("the duplicate should still be closed: %v", provider.deletes)
	}
}

// The verdicts that change nothing must not resolve or retire anything.
func TestApplyTriageLeavesRetirementAloneForOtherVerdicts(t *testing.T) {
	provider := &triageRecorder{}
	env := completedTriage(func(e *types.TriageEnvelope) { e.Verdict = types.VerdictReady; e.Body = "" })

	if err := ApplyTriage(context.Background(), provider, triageTODO(), env, TriageOptions{}); err != nil {
		t.Fatalf("ApplyTriage: %v", err)
	}
	if len(provider.deletes) != 0 {
		t.Errorf("a ready verdict must not close anything: %v", provider.deletes)
	}
}

func TestHighestPriority(t *testing.T) {
	for _, tc := range []struct {
		name  string
		input []types.Priority
		want  types.Priority
	}{
		{name: "high beats medium", input: []types.Priority{types.PriorityMedium, types.PriorityHigh}, want: types.PriorityHigh},
		{name: "medium beats low", input: []types.Priority{types.PriorityLow, types.PriorityMedium}, want: types.PriorityMedium},
		{name: "unset is ignored", input: []types.Priority{"", types.PriorityLow}, want: types.PriorityLow},
		{name: "unknown is ignored", input: []types.Priority{"urgent", types.PriorityLow}, want: types.PriorityLow},
		{name: "nothing known is empty", input: []types.Priority{"", "urgent"}, want: ""},
		{name: "no input is empty", input: nil, want: ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := types.HighestPriority(tc.input...); got != tc.want {
				t.Fatalf("HighestPriority(%v) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}
