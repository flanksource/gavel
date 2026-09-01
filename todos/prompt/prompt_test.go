package prompt

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	captainai "github.com/flanksource/captain/pkg/ai"
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/commons-db/shell"
	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/todos/types"
	"github.com/flanksource/gavel/verify"
)

func newTestTODO(name, impl string) *types.TODO {
	return &types.TODO{
		FilePath: ".todos/" + name + ".md",
		TODOFrontmatter: types.TODOFrontmatter{
			Title: name,
		},
		Implementation: impl,
		StepsToReproduce: []*fixtures.FixtureNode{
			{Test: &fixtures.FixtureTest{Name: name + "-repro", ExecFixtureBase: fixtures.ExecFixtureBase{Exec: "go test ./..."}}},
		},
		Verification: []*fixtures.FixtureNode{
			{Test: &fixtures.FixtureTest{Name: name + "-verify", ExecFixtureBase: fixtures.ExecFixtureBase{Exec: "go test -run TestFoo"}}},
		},
	}
}

func renderUser(t *testing.T, todoList []*types.TODO, opts Options) string {
	t.Helper()
	if opts.Mode == "" {
		opts.Mode = types.ModeRun
	}
	req, _, err := Render(todoList, opts)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	return req.Prompt.User
}

func TestRenderRunGroup(t *testing.T) {
	todoList := []*types.TODO{
		newTestTODO("fix-auth", "Fix the auth handler"),
		newTestTODO("fix-db", "Fix the database query"),
		newTestTODO("fix-cache", "Fix the cache invalidation"),
	}
	req, _, err := Render(todoList, Options{Mode: types.ModeRun})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}

	for _, want := range []string{
		"implementing the 3 todo items listed below",
		"## 1. fix-auth",
		"## 2. fix-db",
		"## 3. fix-cache",
		"Fix the auth handler",
		"Fix the database query",
		"Fix the cache invalidation",
		"Implement ALL todo items",
		"Do NOT run git add or git commit",
	} {
		if !strings.Contains(req.Prompt.User, want) {
			t.Errorf("run prompt missing %q", want)
		}
	}

	// The envelope schema rides on SchemaJSON as a native field, not appended to
	// the prompt text — so the driver, not gavel, decides how to deliver it.
	if !strings.Contains(string(req.Prompt.SchemaJSON), `"endStatus"`) {
		t.Errorf("run envelope schema missing endStatus: %s", req.Prompt.SchemaJSON)
	}
	if req.Prompt.SchemaStrictness != api.SchemaStrictnessRetry {
		t.Errorf("run schema strictness = %q, want retry", req.Prompt.SchemaStrictness)
	}
	if strings.Contains(req.Prompt.User, "conforms to this JSON Schema") {
		t.Error("schema instruction text must not be appended to the prompt body")
	}
}

func TestRenderSingleTODO(t *testing.T) {
	prompt := renderUser(t, []*types.TODO{newTestTODO("solo", "Single task")}, Options{})
	if !strings.Contains(prompt, "## solo") {
		t.Error("single-element group should contain title heading")
	}
	if !strings.Contains(prompt, "Single task") {
		t.Error("should contain implementation text")
	}
	if strings.Contains(prompt, "## 1. solo") {
		t.Error("single todo should not be numbered")
	}
}

// TestRenderFoldsFrontmatter pins the captain-engine contract: the plan
// template's frontmatter declares the request options, so the rendered request
// itself carries plan permissions and the default model — no Go code sets them.
func TestRenderFoldsFrontmatter(t *testing.T) {
	req, _, err := Render([]*types.TODO{newTestTODO("solo", "task")}, Options{Mode: types.ModePlan})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if req.Permissions.Mode != api.PermissionPlan {
		t.Errorf("plan template Permissions.Mode = %q, want %q", req.Permissions.Mode, api.PermissionPlan)
	}
	// The frontmatter says `model: claude`. Rendering resolves that family alias
	// to the exact id a driver receives — which id is a catalog decision, so the
	// assertion is that it resolved at all and stayed in the family.
	if req.Model.Name == "claude" || req.Model.Provider != api.Anthropic {
		t.Errorf("plan template Model = %q on %v, want a resolved anthropic model",
			req.Model.Name, req.Model.Provider)
	}
	if !strings.Contains(string(req.Prompt.SchemaJSON), `"planStatus"`) || !strings.Contains(string(req.Prompt.SchemaJSON), `"planPath"`) {
		t.Errorf("plan envelope schema missing planStatus/planPath: %s", req.Prompt.SchemaJSON)
	}
	if req.Prompt.SchemaStrictness != api.SchemaStrictnessRetry {
		t.Errorf("plan schema strictness = %q, want retry", req.Prompt.SchemaStrictness)
	}
	if req.Prompt.Source != "todos.plan" {
		t.Errorf("Source = %q, want todos.plan", req.Prompt.Source)
	}
}

func TestRenderPlanExistingPlan(t *testing.T) {
	existing := "1. Change the parser\n2. Add tests"
	with := renderUser(t, []*types.TODO{newTestTODO("solo", "task")}, Options{Mode: types.ModePlan, ExistingPlan: existing})
	if !strings.Contains(with, "## Existing Plan") || !strings.Contains(with, existing) {
		t.Error("existing plan content should be embedded")
	}
	if !strings.Contains(with, `"unchanged"`) {
		t.Error("existing-plan framing should explain the unchanged status")
	}

	without := renderUser(t, []*types.TODO{newTestTODO("solo", "task")}, Options{Mode: types.ModePlan})
	if strings.Contains(without, "## Existing Plan") {
		t.Error("no existing plan → no Existing Plan section")
	}
	if !strings.Contains(without, `"new"`) {
		t.Error("fresh-plan framing should name the new status")
	}
}

// TestRenderRunApprovedPlan pins that a run (implement) prompt carries the
// approved/edited plan when one is supplied, so human plan review actually
// steers the run — and omits the section otherwise.
func TestRenderRunApprovedPlan(t *testing.T) {
	plan := "1. Edit the parser\n2. Wire the flag\n3. Add tests"
	with := renderUser(t, []*types.TODO{newTestTODO("solo", "task")}, Options{Mode: types.ModeRun, ExistingPlan: plan})
	if !strings.Contains(with, "## Approved Plan") || !strings.Contains(with, plan) {
		t.Errorf("run prompt should embed the approved plan; got:\n%s", with)
	}

	without := renderUser(t, []*types.TODO{newTestTODO("solo", "task")}, Options{Mode: types.ModeRun})
	if strings.Contains(without, "## Approved Plan") {
		t.Error("no plan → no Approved Plan section")
	}
}

// TestRenderOverridesKeepSchema pins the output contract: neither a hostile
// template override nor a verbatim body override can touch the envelope schema,
// which now rides on a separate SchemaJSON field rather than the prompt text.
func TestRenderOverridesKeepSchema(t *testing.T) {
	todoList := []*types.TODO{newTestTODO("solo", "Single task")}

	tmplReq, _, err := Render(todoList, Options{
		Mode:     types.ModeRun,
		Template: "CUSTOM FRAMING for {{count}} item(s)\n\n{{{body}}}END",
		Spec:     api.Spec{Model: api.Model{Name: "claude-sonnet-5"}},
	})
	if err != nil {
		t.Fatalf("Render(template override): %v", err)
	}
	tmpl := tmplReq.Prompt.User
	if !strings.Contains(tmpl, "CUSTOM FRAMING for 1 item(s)") {
		t.Errorf("override framing not rendered")
	}
	if !strings.Contains(tmpl, "## solo") || !strings.Contains(tmpl, "Single task") {
		t.Errorf("per-TODO body not injected into override template")
	}
	if strings.Contains(tmpl, "## Instructions") {
		t.Error("override should replace the default instructions block")
	}
	if !strings.Contains(string(tmplReq.Prompt.SchemaJSON), `"endStatus"`) {
		t.Error("template override must not affect the native envelope schema")
	}

	bodyReq, _, err := Render(todoList, Options{Mode: types.ModeRun, Spec: api.Spec{Prompt: api.Prompt{User: "Just do the thing."}}})
	if err != nil {
		t.Fatalf("Render(body override): %v", err)
	}
	body := bodyReq.Prompt.User
	if !strings.Contains(body, "Just do the thing.") {
		t.Error("body override not applied")
	}
	if strings.Contains(body, "## solo") {
		t.Error("body override should replace the rendered body")
	}
	if !strings.Contains(string(bodyReq.Prompt.SchemaJSON), `"endStatus"`) {
		t.Error("body override must not affect the native envelope schema")
	}
}

func TestRenderEffortDirective(t *testing.T) {
	got := renderUser(t, []*types.TODO{newTestTODO("solo", "task")}, Options{Spec: api.Spec{Model: api.Model{Effort: api.EffortHigh}}})
	if !strings.HasPrefix(got, EffortDirective("high")) {
		t.Errorf("prompt should lead with the effort directive, got %q…", got[:60])
	}
}

func TestRenderMergesCanonicalSpecWithoutDroppingRuntimeFields(t *testing.T) {
	spec := api.Spec{
		Model:  api.Model{Name: "gpt-5.6-sol", Mode: api.ModeAgent, Effort: api.EffortHigh},
		Prompt: api.Prompt{User: "Use the reviewed implementation instructions.", System: "Keep changes surgical."},
		Budget: api.Budget{Cost: 2.5, MaxTurns: 8, Timeout: "12m"},
		Memory: api.Memory{Skills: []string{"gavel-todos"}},
		Permissions: api.Permissions{
			Mode:    api.PermissionAcceptEdits,
			Tools:   api.Tools{"Bash": api.ToolPolicyAsk},
			MCP:     api.MCP{Servers: []string{"postgres"}},
			Plugins: api.ResourcePolicies{"review": api.ResourceEnabled},
			Skills:  api.ResourcePolicies{"gavel-todos": api.ResourceEnabled},
		},
		Setup:     &shell.Setup{Cwd: "workspace"},
		Workflow:  &api.Workflow{Verify: &api.Verify{Commands: []string{"go test ./todos"}, MaxIterations: 4}},
		SessionID: "session-123",
		CLIArgs:   map[string]any{"fullAuto": true},
	}

	req, _, err := Render([]*types.TODO{newTestTODO("solo", "task")}, Options{
		Mode: types.ModeRun,
		Spec: spec,
	})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}

	if !strings.Contains(req.Prompt.User, spec.Prompt.User) || strings.Contains(req.Prompt.User, "## solo") {
		t.Fatalf("rendered body did not use Spec.Prompt.User: %q", req.Prompt.User)
	}
	if !strings.HasPrefix(req.Prompt.User, EffortDirective("high")) {
		t.Fatalf("rendered body omitted the resolved effort directive: %q", req.Prompt.User)
	}
	if !strings.Contains(string(req.Prompt.SchemaJSON), `"endStatus"`) {
		t.Fatalf("rendered request lost the TODO envelope schema: %s", req.Prompt.SchemaJSON)
	}

	// want is the caller's spec plus the fields the template itself contributes.
	// Those are not drift: a .prompt declaring schemaStrictness is the template
	// doing its job, and the caller's spec never modelled it.
	//
	// The model is resolved too — rendering hands back a driver-ready one — so
	// the expectation resolves the caller's model rather than restating the
	// capability flags resolution fills in.
	want := spec
	resolvedModel, err := captainai.Resolve(spec.Model)
	if err != nil {
		t.Fatalf("resolve the caller's model: %v", err)
	}
	want.Model = resolvedModel
	want.Prompt.User = req.Prompt.User
	want.Prompt.Source = "todos.run"
	want.Prompt.SchemaJSON = req.Prompt.SchemaJSON
	want.Prompt.SchemaStrictness = "retry"
	if !reflect.DeepEqual(req, want) {
		t.Fatalf("rendered request dropped or changed Spec fields:\n got: %#v\nwant: %#v", req, want)
	}
}

func TestRenderRejectsVerifyMode(t *testing.T) {
	if _, _, err := Render([]*types.TODO{newTestTODO("solo", "task")}, Options{Mode: types.ModeVerify}); err == nil {
		t.Fatal("verify mode has no agent prompt and must error")
	}
}

func TestPlanEnvelopeSchemaUsesFlatScalarFields(t *testing.T) {
	raw, err := EnvelopeSchemaJSON(EnvelopePlan)
	if err != nil {
		t.Fatalf("EnvelopeSchemaJSON: %v", err)
	}
	got, err := captainai.SchemaJSONForRuntime(api.Anthropic, captainai.ModeAgent, api.Prompt{SchemaJSON: raw})
	if err != nil {
		t.Fatalf("SchemaJSONForRuntime: %v", err)
	}

	var root map[string]any
	if err := json.Unmarshal(got, &root); err != nil {
		t.Fatalf("decode transformed schema: %v", err)
	}
	defs, ok := root["$defs"].(map[string]any)
	if !ok {
		t.Fatalf("transformed schema has no $defs: %s", got)
	}
	planEnvelope, ok := defs["PlanEnvelope"].(map[string]any)
	if !ok {
		t.Fatalf("transformed schema has no PlanEnvelope definition: %s", got)
	}
	properties, ok := planEnvelope["properties"].(map[string]any)
	if !ok {
		t.Fatalf("PlanEnvelope has no properties: %#v", planEnvelope)
	}
	for _, field := range []string{"summary", "endStatus", "planStatus", "planPath", "planContent"} {
		property, ok := properties[field].(map[string]any)
		if !ok {
			t.Fatalf("PlanEnvelope.%s is %T, want schema object", field, properties[field])
		}
		if property["$ref"] != nil {
			t.Fatalf("PlanEnvelope.%s = %#v, want a top-level scalar", field, property)
		}
	}
	if properties["plan"] != nil {
		t.Fatalf("PlanEnvelope.plan = %#v, want no nested plan object", properties["plan"])
	}
}

// TestTriageEnvelopeSchemaUsesFlatScalarFields mirrors the plan envelope's
// guard: a nested object becomes a $ref node, which some backends refuse to
// emit, so every triage field must be a top-level scalar or scalar array.
func TestTriageEnvelopeSchemaUsesFlatScalarFields(t *testing.T) {
	raw, err := EnvelopeSchemaJSON(EnvelopeTriage)
	if err != nil {
		t.Fatalf("EnvelopeSchemaJSON: %v", err)
	}
	got, err := captainai.SchemaJSONForRuntime(api.Anthropic, captainai.ModeAgent, api.Prompt{SchemaJSON: raw})
	if err != nil {
		t.Fatalf("SchemaJSONForRuntime: %v", err)
	}

	var root map[string]any
	if err := json.Unmarshal(got, &root); err != nil {
		t.Fatalf("decode transformed schema: %v", err)
	}
	defs, ok := root["$defs"].(map[string]any)
	if !ok {
		t.Fatalf("transformed schema has no $defs: %s", got)
	}
	triage, ok := defs["TriageEnvelope"].(map[string]any)
	if !ok {
		t.Fatalf("transformed schema has no TriageEnvelope definition: %s", got)
	}
	properties, ok := triage["properties"].(map[string]any)
	if !ok {
		t.Fatalf("TriageEnvelope has no properties: %#v", triage)
	}
	for _, field := range []string{
		"summary", "endStatus", "verdict", "title", "body",
		"verification", "priority", "status", "duplicateOf", "comment",
	} {
		property, ok := properties[field].(map[string]any)
		if !ok {
			t.Fatalf("TriageEnvelope.%s is %T, want schema object", field, properties[field])
		}
		if property["$ref"] != nil {
			t.Fatalf("TriageEnvelope.%s = %#v, want a top-level scalar", field, property)
		}
	}
	related, ok := properties["related"].(map[string]any)
	if !ok {
		t.Fatalf("TriageEnvelope.related is %T, want schema object", properties["related"])
	}
	if related["type"] != "array" {
		t.Fatalf("TriageEnvelope.related = %#v, want an array of scalars", related)
	}
}

// The triage prompt exists to review a fixture, and the command projection used
// by run/plan discards the frontmatter, CEL predicate and yaml step structure —
// exactly what is being assessed. So a triage render must carry the fixture
// source verbatim.
func TestRenderTriageIncludesRawVerificationFixture(t *testing.T) {
	todo := newTestTODO("solo", "task")
	todo.VerificationMarkdown = "---\nverify:\n  scope: diff\n---\n```yaml test\npackages: ./todos\n```"

	user := renderUser(t, []*types.TODO{todo}, Options{Prompt: "triage", Envelope: EnvelopeTriage})
	for _, want := range []string{"verify:", "scope: diff", "```yaml test", "packages: ./todos"} {
		if !strings.Contains(user, want) {
			t.Errorf("triage prompt omitted %q from the fixture source:\n%s", want, user)
		}
	}
}

// A TODO with no fixture is the most important case for triage to act on, so the
// absence must be stated rather than rendered as an empty section.
func TestRenderTriageStatesAMissingFixture(t *testing.T) {
	user := renderUser(t, []*types.TODO{newTestTODO("solo", "task")}, Options{Prompt: "triage", Envelope: EnvelopeTriage})
	if !strings.Contains(user, "NO verification fixture") {
		t.Errorf("triage prompt did not flag the missing fixture:\n%s", user)
	}
}

func TestRenderTriageIncludesBacklogForDedupe(t *testing.T) {
	backlog := "- ab12cd Fix the parser panic (pending, high)"
	user := renderUser(t, []*types.TODO{newTestTODO("solo", "task")}, Options{
		Prompt: "triage", Envelope: EnvelopeTriage, Backlog: backlog,
	})
	if !strings.Contains(user, backlog) {
		t.Errorf("triage prompt omitted the backlog index:\n%s", user)
	}

	without := renderUser(t, []*types.TODO{newTestTODO("solo", "task")}, Options{Prompt: "triage", Envelope: EnvelopeTriage})
	if strings.Contains(without, "## Backlog") {
		t.Errorf("empty backlog should omit the section entirely:\n%s", without)
	}
}

// Run and plan must keep the executable command projection: an implementation
// run needs to know what to execute, not what the fixture source looks like.
func TestRenderRunKeepsFixtureCommandProjection(t *testing.T) {
	todo := newTestTODO("solo", "task")
	todo.VerificationMarkdown = "---\nverify:\n  scope: diff\n---\n"

	user := renderUser(t, []*types.TODO{todo}, Options{Mode: types.ModeRun})
	if strings.Contains(user, "scope: diff") {
		t.Errorf("run prompt leaked raw fixture source:\n%s", user)
	}
}

func TestRenderIncludesComments(t *testing.T) {
	todo := newTestTODO("fix-auth", "Fix the auth handler")
	todo.ProviderEvents = []types.ProviderEvent{
		{Kind: "IssueCreated", Actor: "agent", Body: "initial body"},
		{Kind: "CommentAdded", Actor: "reviewer", Body: "Please include history"},
		{Kind: "LabelChanged", Actor: "agent", OldLabel: "status:pending", NewLabel: "status:in-progress"},
		{Kind: "CommentAdded", Actor: "maintainer", Body: "Reuse the existing helper"},
		{Kind: "CommentAdded", Actor: "bot", Body: "   "},
	}
	prompt := renderUser(t, []*types.TODO{todo}, Options{})

	for _, want := range []string{"## Comments", "**reviewer:**", "Please include history", "**maintainer:**", "Reuse the existing helper"} {
		if !strings.Contains(prompt, want) {
			t.Errorf("prompt missing comment content %q", want)
		}
	}
	if strings.Contains(prompt, "status:in-progress") {
		t.Error("non-comment events should not leak into the prompt")
	}
	if strings.Contains(prompt, "**bot:**") {
		t.Error("blank-body comments should be skipped")
	}
}

func TestRenderExcludesPRButIncludesSource(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "pkg/auth.go", 50)

	todo := newTestTODO("fix-auth", "Fix the auth handler")
	todo.PR = &types.PR{Number: 42, URL: "https://example.com/pull/42", Head: "feat/x", Base: "main"}
	todo.Path = types.StringOrSlice{"pkg/auth.go:25"}

	prompt := renderUser(t, []*types.TODO{todo}, Options{WorkDir: dir})
	if strings.Contains(prompt, "## PR Context") {
		t.Error("grouped prompt should not contain PR Context section")
	}
	if !strings.Contains(prompt, "```go file=pkg/auth.go:25") {
		t.Error("grouped prompt should contain annotated code block")
	}
	if !strings.Contains(prompt, "line 25 content") {
		t.Error("should contain actual source code lines")
	}
}

// TestEnvelopeSchemaBytesStable pins the schema identity shared by initial,
// retry, and feedback turns in a claude-agent session.
func TestEnvelopeSchemaBytesStable(t *testing.T) {
	for _, name := range []string{"run", "plan", "triage"} {
		req, _, err := Render([]*types.TODO{newTestTODO("solo", "task")}, Options{Prompt: name, Envelope: envelopeForPrompt(t, name)})
		if err != nil {
			t.Fatalf("Render(%s): %v", name, err)
		}
		initial := string(req.Prompt.SchemaJSON)
		if initial == "" {
			t.Fatalf("Render(%s) produced no envelope schema", name)
		}

		recomputed, err := EnvelopeSchemaJSON(envelopeForPrompt(t, name))
		if err != nil {
			t.Fatalf("EnvelopeSchemaJSON(%s): %v", name, err)
		}
		if string(recomputed) != initial {
			t.Errorf("%s: recomputed schema differs from Render's:\n initial=%s\n recomputed=%s", name, initial, recomputed)
		}
	}
}

// envelopeForPrompt resolves a built-in prompt's envelope through the catalog, so
// the test asserts against the same wiring production uses rather than a literal.
func envelopeForPrompt(t *testing.T, name string) EnvelopeKind {
	t.Helper()
	catalog, err := NewCatalog(verify.TodosConfig{})
	if err != nil {
		t.Fatalf("NewCatalog: %v", err)
	}
	def, err := catalog.Lookup(name)
	if err != nil {
		t.Fatalf("Lookup(%s): %v", name, err)
	}
	return def.Envelope
}

func TestPromptsRegistered(t *testing.T) {
	got := Prompts()
	if len(got) != 3 {
		t.Fatalf("Prompts() returned %d entries, want run/plan/triage", len(got))
	}
	for _, p := range got {
		if p.ID == "" || strings.TrimSpace(p.Default) == "" || p.ConfigPath == "" {
			t.Errorf("descriptor %+v incomplete", p)
		}
	}
}

func TestInputSchema(t *testing.T) {
	raw, err := InputSchema(types.ModeRun)
	if err != nil {
		t.Fatalf("InputSchema: %v", err)
	}
	// Field names follow the structs' YAML tags — the fixture-fence wire contract.
	for _, want := range []string{`"test"`, `"lint"`, "show-passed", "linters"} {
		if !strings.Contains(string(raw), want) {
			t.Errorf("run input schema missing %q", want)
		}
	}
	if raw, err := InputSchema(types.ModePlan); err != nil || raw != nil {
		t.Errorf("plan mode inputs = (%s, %v), want none", raw, err)
	}
}

func TestLangFromExt(t *testing.T) {
	tests := []struct {
		ext, expected string
	}{
		{".go", "go"}, {".ts", "typescript"}, {".tsx", "typescript"}, {".js", "javascript"},
		{".py", "python"}, {".sql", "sql"}, {".yaml", "yaml"}, {".yml", "yaml"},
		{".json", "json"}, {".sh", "bash"}, {".rs", "rust"}, {".rb", "ruby"},
		{".unknown", ""}, {"", ""},
	}
	for _, tc := range tests {
		if got := langFromExt(tc.ext); got != tc.expected {
			t.Errorf("langFromExt(%q) = %q, want %q", tc.ext, got, tc.expected)
		}
	}
}

func TestStripFileRefLine(t *testing.T) {
	tests := []struct {
		name, input, expected string
	}{
		{"removes file:line ref", "File: `pkg/handler.go:42`\n\nSome description", "Some description"},
		{"removes file ref without line", "File: `pkg/handler.go`\n\nSome description", "Some description"},
		{"no-op when absent", "Just a description", "Just a description"},
		{"strips from middle", "Before\nFile: `foo.go:10`\nAfter", "Before\nAfter"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := stripFileRefLine(tc.input); got != tc.expected {
				t.Errorf("got %q, want %q", got, tc.expected)
			}
		})
	}
}
