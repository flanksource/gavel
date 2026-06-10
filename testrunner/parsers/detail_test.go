package parsers

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
)

type providerDetail struct {
	Source string `json:"source,omitempty"`
	Result string `json:"result,omitempty"`
}

func (d providerDetail) Pretty() api.Text {
	return clicky.Text("provider detail: " + d.Source + " -> " + d.Result)
}

func TestDetailUsesProviderJSON(t *testing.T) {
	in := Test{
		Name:   "doc-1",
		Passed: true,
		Detail: providerDetail{
			Source: "kind: TestPlan\n",
			Result: "POL-123",
		},
	}

	raw, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(raw), `"detail":{"source":"kind: TestPlan\n","result":"POL-123"}`) {
		t.Fatalf("detail should marshal as provider JSON, got %s", raw)
	}
	if strings.Contains(string(raw), `"node"`) || strings.Contains(string(raw), `"version"`) {
		t.Fatalf("detail should not marshal as a ClickyDocument, got %s", raw)
	}
}

func TestRenderDetailUsesPretty(t *testing.T) {
	rendered := RenderDetail(providerDetail{Source: "source", Result: "result"}).String()
	if !strings.Contains(rendered, "provider detail: source -> result") {
		t.Fatalf("RenderDetail should use Pretty(), got %q", rendered)
	}
}

// TestPretty_RendersDetailOnFailure: a failed node surfaces its provider Detail
// inline beneath the summary line, so the rich body (e.g. an activity trace's
// Input Fields / Errors block) reaches the terminal instead of staying buried in
// the JSON report. providerDetail.Pretty() returns a known marker string.
func TestPretty_RendersDetailOnFailure(t *testing.T) {
	rendered := Test{
		Name:   "step",
		Failed: true,
		Detail: providerDetail{Source: "source", Result: "result"},
	}.Pretty().String()

	if !strings.Contains(rendered, "provider detail: source -> result") {
		t.Fatalf("a failed node must render its Detail inline, got %q", rendered)
	}
}

// TestPretty_OmitsDetailOnPassWithoutVerbose: a passing node does NOT dump its
// Detail at default verbosity (logger.V(2) is off in the test process) — the
// summary line stays scannable on a green run.
func TestPretty_OmitsDetailOnPassWithoutVerbose(t *testing.T) {
	rendered := Test{
		Name:   "step",
		Passed: true,
		Detail: providerDetail{Source: "source", Result: "result"},
	}.Pretty().String()

	if strings.Contains(rendered, "provider detail:") {
		t.Fatalf("a passing node must not render Detail without -vv, got %q", rendered)
	}
}

// TestPretty_NoDetailLabelWhenEmpty: a failed node with no Detail renders no
// "detail" label — the IsEmpty guard keeps the empty block out.
func TestPretty_NoDetailLabelWhenEmpty(t *testing.T) {
	rendered := Test{Name: "step", Failed: true}.Pretty().String()

	if strings.Contains(rendered, "detail") {
		t.Fatalf("a node with no Detail must not render a detail label, got %q", rendered)
	}
}

// TestDetailOmittedWhenNil keeps the field invisible for tests that never set
// provider-owned detail.
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
		t.Fatalf("unset Detail should be omitted, got %s", raw)
	}
}
