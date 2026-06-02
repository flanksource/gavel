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
