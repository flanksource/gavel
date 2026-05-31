package parsers

import (
	"encoding/json"
	"testing"

	"github.com/flanksource/clicky/api"
	"github.com/flanksource/clicky/formatters"
)

// TestDetailLiveBuildAndRoundTrip exercises the live builder path
// (test.Detail().Add(...)) and the snapshot reload contract: a Test whose Detail
// was built incrementally must survive json.Marshal → json.Unmarshal →
// json.Marshal unchanged, so reloaded .gavel snapshots keep the rich view.
func TestDetailLiveBuildAndRoundTrip(t *testing.T) {
	source := api.Text{Content: "source"}.NewLine().Add(api.CodeBlock("yaml", "kind: TestPlan\n"))

	in := Test{Name: "doc-1", Passed: true}
	in.Detail().
		Add(api.Collapsed{Label: "source", Content: source}).
		Add(api.Collapsed{Label: "result", Content: api.Text{Content: "POL-1"}})

	raw, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var out Test
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if out.DetailDoc == nil {
		t.Fatal("Detail dropped on round-trip")
	}

	// Re-marshalling the reloaded Test must be byte-stable with the original
	// detail JSON, proving the verbatim re-emit path.
	reMarshaled, err := json.Marshal(out)
	if err != nil {
		t.Fatalf("re-marshal: %v", err)
	}
	if string(reMarshaled) != string(raw) {
		t.Fatalf("re-marshal not byte-stable:\n first: %s\n again: %s", raw, reMarshaled)
	}

	if !hasKind(out.DetailDoc.doc.Node, "code") {
		t.Fatal("reloaded Detail lost its code block")
	}
}

// TestDetailOmittedWhenNil keeps the field invisible for tests that never touch
// Detail().
func TestDetailOmittedWhenNil(t *testing.T) {
	raw, err := json.Marshal(Test{Name: "doc-1", Passed: true})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var generic map[string]any
	if err := json.Unmarshal(raw, &generic); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if _, present := generic["detail"]; present {
		t.Fatalf("untouched Detail should be omitted, got %s", raw)
	}
}

func hasKind(node formatters.ClickyNode, kind string) bool {
	if node.Kind == kind {
		return true
	}
	for _, c := range node.Children {
		if hasKind(c, kind) {
			return true
		}
	}
	for _, c := range node.Items {
		if hasKind(c, kind) {
			return true
		}
	}
	if node.Content != nil && hasKind(*node.Content, kind) {
		return true
	}
	if node.Label != nil && hasKind(*node.Label, kind) {
		return true
	}
	return false
}
