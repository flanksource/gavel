package types

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/flanksource/captain/pkg/api"
)

func validEnvelope() ResultEnvelope {
	return ResultEnvelope{Summary: "implemented the change", EndStatus: EndCompleted}
}

func TestResultEnvelopeValidate(t *testing.T) {
	cases := []struct {
		name    string
		mutate  func(*ResultEnvelope)
		wantErr string
	}{
		{name: "completed with summary is valid", mutate: func(e *ResultEnvelope) {}},
		{name: "failed is valid", mutate: func(e *ResultEnvelope) { e.EndStatus = EndFailed }},
		{name: "empty summary rejected", mutate: func(e *ResultEnvelope) { e.Summary = "  " }, wantErr: "summary"},
		{name: "unknown end status rejected", mutate: func(e *ResultEnvelope) { e.EndStatus = "done" }, wantErr: "endStatus"},
		{name: "ask without questions rejected", mutate: func(e *ResultEnvelope) { e.EndStatus = EndAsk }, wantErr: "question"},
		{
			name: "ask with a question is valid",
			mutate: func(e *ResultEnvelope) {
				e.EndStatus = EndAsk
				e.Questions = []AgentQuestion{{Text: "which auth provider should be used?"}}
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := validEnvelope()
			tc.mutate(&e)
			err := e.Validate()
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("Validate() = %v, want nil", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("Validate() = %v, want error containing %q", err, tc.wantErr)
			}
		})
	}
}

func TestPlanEnvelopeValidate(t *testing.T) {
	planPath := "/home/user/.claude/plans/plan-for-fix-auth.md"
	cases := []struct {
		name    string
		env     PlanEnvelope
		wantErr string
	}{
		{
			name: "new plan with path is valid",
			env: PlanEnvelope{
				ResultEnvelope: validEnvelope(),
				PlanStatus:     PlanNew,
				PlanPath:       planPath,
			},
		},
		{
			name: "new inline plan content is valid",
			env: PlanEnvelope{
				ResultEnvelope: validEnvelope(),
				PlanStatus:     PlanNew,
				PlanContent:    "- [x] inspect\n- [ ] implement",
			},
		},
		{
			name: "unchanged plan needs no path",
			env: PlanEnvelope{
				ResultEnvelope: validEnvelope(),
				PlanStatus:     PlanUnchanged,
			},
		},
		{
			name: "new plan without path or content rejected",
			env: PlanEnvelope{
				ResultEnvelope: validEnvelope(),
				PlanStatus:     PlanNew,
			},
			wantErr: "planPath or planContent",
		},
		{
			name: "completed without plan status rejected",
			env: PlanEnvelope{
				ResultEnvelope: validEnvelope(),
			},
			wantErr: "planStatus",
		},
		{
			name: "ask needs no plan",
			env: PlanEnvelope{
				ResultEnvelope: ResultEnvelope{
					Summary:   "blocked on a question",
					EndStatus: EndAsk,
					Questions: []AgentQuestion{{Text: "monorepo or split repos?"}},
				},
			},
		},
		{
			name: "failed needs no plan",
			env: PlanEnvelope{
				ResultEnvelope: ResultEnvelope{Summary: "could not explore the repo", EndStatus: EndFailed},
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.env.Validate()
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("Validate() = %v, want nil", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("Validate() = %v, want error containing %q", err, tc.wantErr)
			}
		})
	}
}

// TestPlanEnvelopeSchema pins the schema contract the prompts embed: field names,
// required set, and enum values. A change here changes what agents are asked to
// emit and must be deliberate.
func TestPlanEnvelopeSchema(t *testing.T) {
	raw, err := api.SchemaJSON(&PlanEnvelope{})
	if err != nil {
		t.Fatalf("SchemaJSON: %v", err)
	}
	schema := schemaObject(t, raw)
	root := resolveRef(t, schema, schema)

	props, required := schemaProps(t, root)
	for _, field := range []string{"summary", "endStatus", "planStatus", "planPath", "planContent"} {
		if _, ok := props[field]; !ok {
			t.Errorf("schema properties missing %q (have %v)", field, keys(props))
		}
		if (field == "summary" || field == "endStatus" || field == "planStatus") && !contains(required, field) {
			t.Errorf("schema required missing %q (have %v)", field, required)
		}
	}

	if got := enumValues(t, props, "endStatus"); !equalStrings(got, []string{"ask", "completed", "failed"}) {
		t.Errorf("endStatus enum = %v, want ask/completed/failed", got)
	}

	if got := enumValues(t, props, "planStatus"); !equalStrings(got, []string{"new", "unchanged", "updated"}) {
		t.Errorf("planStatus enum = %v, want new/unchanged/updated", got)
	}
	if _, ok := props["plan"]; ok {
		t.Errorf("plan schema must not contain nested plan property (have %v)", keys(props))
	}
	summary := props["summary"].(map[string]any)
	if summary["maxLength"] != float64(1000) {
		t.Errorf("summary maxLength = %v, want 1000", summary["maxLength"])
	}
}

func schemaObject(t *testing.T, raw []byte) map[string]any {
	t.Helper()
	var schema map[string]any
	if err := json.Unmarshal(raw, &schema); err != nil {
		t.Fatalf("schema is not valid JSON: %v\n%s", err, raw)
	}
	return schema
}

// schemaProps returns the properties and required list of a schema node,
// following a top-level $ref into $defs when invopop emits one.
func schemaProps(t *testing.T, node map[string]any) (map[string]any, []string) {
	t.Helper()
	props, _ := node["properties"].(map[string]any)
	if props == nil {
		t.Fatalf("schema node has no properties: %v", node)
	}
	var required []string
	if list, ok := node["required"].([]any); ok {
		for _, v := range list {
			required = append(required, v.(string))
		}
	}
	return props, required
}

// resolveRef follows a {$ref: "#/$defs/X"} node against the root schema; a node
// without $ref is returned as-is.
func resolveRef(t *testing.T, root, node map[string]any) map[string]any {
	t.Helper()
	ref, _ := node["$ref"].(string)
	if ref == "" {
		return node
	}
	name := ref[strings.LastIndex(ref, "/")+1:]
	defs, _ := root["$defs"].(map[string]any)
	resolved, _ := defs[name].(map[string]any)
	if resolved == nil {
		t.Fatalf("cannot resolve %q in $defs (have %v)", ref, keys(defs))
	}
	return resolved
}

func enumValues(t *testing.T, props map[string]any, field string) []string {
	t.Helper()
	node, _ := props[field].(map[string]any)
	if node == nil {
		t.Fatalf("property %q missing", field)
	}
	list, _ := node["enum"].([]any)
	if list == nil {
		t.Fatalf("property %q has no enum: %v", field, node)
	}
	var got []string
	for _, v := range list {
		got = append(got, v.(string))
	}
	return got
}

func equalStrings(got, wantSorted []string) bool {
	if len(got) != len(wantSorted) {
		return false
	}
	seen := map[string]bool{}
	for _, g := range got {
		seen[g] = true
	}
	for _, w := range wantSorted {
		if !seen[w] {
			return false
		}
	}
	return true
}

func contains(list []string, v string) bool {
	for _, s := range list {
		if s == v {
			return true
		}
	}
	return false
}

func keys(m map[string]any) []string {
	var out []string
	for k := range m {
		out = append(out, k)
	}
	return out
}
