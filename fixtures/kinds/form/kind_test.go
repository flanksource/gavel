package form_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/flanksource/clicky/prompt"
	"github.com/flanksource/clicky/task"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flanksource/gavel/fixtures"
	_ "github.com/flanksource/gavel/fixtures/kinds/form"
	_ "github.com/flanksource/gavel/fixtures/types"
	"github.com/flanksource/gavel/processruntime"
)

// fakePrompter answers with answer, recording every prompt it was asked.
type fakePrompter struct {
	mu     sync.Mutex
	asked  []prompt.Prompt
	answer func(context.Context, prompt.Prompt) (prompt.Answer, error)
}

func (f *fakePrompter) Ask(ctx context.Context, p prompt.Prompt) (prompt.Answer, error) {
	f.mu.Lock()
	f.asked = append(f.asked, p)
	f.mu.Unlock()
	return f.answer(ctx, p)
}

func answering(values map[string]any, cancelled bool) *fakePrompter {
	return &fakePrompter{answer: func(context.Context, prompt.Prompt) (prompt.Answer, error) {
		return prompt.Answer{Values: values, Cancelled: cancelled, At: time.Date(2026, 11, 8, 9, 0, 0, 0, time.UTC), By: "operator"}, nil
	}}
}

// recordKind passes, recording the ids of the steps it ran.
type recordKind struct{ ran *[]string }

func (recordKind) Name() string            { return "record" }
func (recordKind) Schema() json.RawMessage { return json.RawMessage(`{"type":"object"}`) }
func (k recordKind) Execute(_ context.Context, step *fixtures.StepContext) (fixtures.StepOutcome, error) {
	*k.ran = append(*k.ran, step.Spec.ID)
	return fixtures.StepOutcome{Result: fixtures.FixtureResult{Status: task.StatusPASS}}, nil
}

// rowsKind passes with the rows its config lists, standing in for a query step.
type rowsKind struct{}

func (rowsKind) Name() string            { return "rows" }
func (rowsKind) Schema() json.RawMessage { return json.RawMessage(`{"type":"object"}`) }
func (rowsKind) Execute(_ context.Context, step *fixtures.StepContext) (fixtures.StepOutcome, error) {
	cfg, err := fixtures.KindConfig[struct {
		Rows []any `json:"rows"`
	}](step)
	if err != nil {
		return fixtures.StepOutcome{}, err
	}
	return fixtures.StepOutcome{Result: fixtures.FixtureResult{Status: task.StatusPASS}, Outputs: map[string]any{"rows": cfg.Rows}}, nil
}

func runDocument(ctx context.Context, document string, opts fixtures.RunOptions) []fixtures.FixtureResult {
	GinkgoHelper()
	path := filepath.Join(GinkgoT().TempDir(), "form.fixture.yaml")
	Expect(os.WriteFile(path, []byte(document), 0o600)).To(Succeed())
	file, err := fixtures.ParseFixtureFile(path)
	Expect(err).ToNot(HaveOccurred())
	opts.WorkDir = filepath.Dir(path)
	results, _, err := fixtures.RunNodes(ctx, []*fixtures.FixtureNode{file}, opts)
	Expect(err).ToNot(HaveOccurred())
	return results
}

// formOf is the form record a form step published, as plain JSON.
func formOf(result fixtures.FixtureResult) map[string]any {
	GinkgoHelper()
	raw, err := json.Marshal(result.Metadata[fixtures.FormMetadataKey])
	Expect(err).ToNot(HaveOccurred())
	var record map[string]any
	Expect(json.Unmarshal(raw, &record)).To(Succeed())
	return record
}

func outputsOf(result fixtures.FixtureResult) map[string]any {
	GinkgoHelper()
	outputs, ok := result.Metadata["outputs"].(map[string]any)
	Expect(ok).To(BeTrue(), "result %s has no outputs: %v", result.Name, result.Metadata)
	return outputs
}

const approvalDocument = `kind: Fixture
env: {DB: lab}
steps:
  - id: approval
    kind: form
    title: Purge staged activities?
    description: "on {{.env.DB}}"
    schema:
      type: object
      required: [expected]
      properties:
        expected: {type: integer, minimum: 1}
        reload: {type: boolean, default: false}
  - id: purge
    kind: record
    if: approval.approved && approval.expected == 5
`

const paramsDocument = `kind: Fixture
name: purge
params:
  type: object
  required: [db, n]
  properties:
    db: {type: string, format: connection, x-connection-types: [sql_server]}
    n: {type: integer}
    dry: {type: boolean, default: true}
steps:
  - id: inspect
    kind: record
    if: params.n == 5 && params.dry && params.db == "lab"
`

var _ = Describe("form kind", func() {
	var ran []string
	BeforeEach(func() {
		ran = nil
		DeferCleanup(fixtures.DefaultKinds.Replace(recordKind{ran: &ran}))
	})

	It("asks the run's prompter, publishes waiting, and gates a later step on the answer", func() {
		prompter := answering(map[string]any{"expected": float64(5), "reload": true}, false)
		var states []fixtures.ExecutionState
		ctx := fixtures.WithProgressSink(context.Background(), func(_ context.Context, snapshot fixtures.ExecutionSnapshot) error {
			states = append(states, snapshot.State)
			return nil
		})

		results := runDocument(ctx, approvalDocument, fixtures.RunOptions{Prompter: prompter, RunID: "run-1"})

		Expect(results[0].Status).To(Equal(task.StatusPASS), results[0].Error)
		Expect(outputsOf(results[0])).To(Equal(map[string]any{
			"expected": int64(5), "reload": true, "approved": true, "answeredBy": "operator", "answeredAt": "2026-11-08T09:00:00Z",
			"values": map[string]any{"expected": int64(5), "reload": true},
		}))
		Expect(prompter.asked).To(HaveLen(1))
		Expect(prompter.asked[0].Owner).To(Equal("run-1"))
		Expect(prompter.asked[0].Description).To(Equal("on lab"))
		Expect(prompter.asked[0].Labels).To(HaveKeyWithValue("step", "approval"))
		Expect(prompter.asked[0].Labels).To(HaveKeyWithValue("node", "approval"))
		Expect(prompter.asked[0].Default).To(Equal(map[string]any{"reload": false}))
		Expect(states).To(ContainElement(fixtures.ExecutionWaiting))
		Expect(ran).To(Equal([]string{"purge"}))
	})

	It("passes a cancelled form as approved=false, and the gated step is skipped", func() {
		results := runDocument(context.Background(), approvalDocument, fixtures.RunOptions{Prompter: answering(nil, true)})

		Expect(results[0].Status).To(Equal(task.StatusPASS), results[0].Error)
		Expect(outputsOf(results[0])).To(HaveKeyWithValue("approved", false))
		Expect(results[1].Status).To(Equal(task.StatusSKIP))
		Expect(results[1].Error).To(Equal("skipped (if: approval.approved && approval.expected == 5 = false)"))
		Expect(ran).To(BeEmpty())
	})

	DescribeTable("publishes the form it showed and how it was answered, for a read-only view",
		func(opts fixtures.RunOptions, document, step, state string, values map[string]any, answeredBy string) {
			results := runDocument(context.Background(), document, opts)
			record := formOf(results[0])
			Expect(record).To(HaveKeyWithValue("state", state))
			Expect(record).To(HaveKeyWithValue("values", values))
			Expect(record["schema"]).To(HaveKeyWithValue("type", "object"))
			if answeredBy == "" {
				Expect(record).NotTo(HaveKey("answeredBy"))
			} else {
				Expect(record).To(HaveKeyWithValue("answeredBy", answeredBy))
			}
			Expect(record).To(HaveKeyWithValue("title", step))
		},
		Entry("an answered form", fixtures.RunOptions{Prompter: answering(map[string]any{"expected": float64(5), "reload": true}, false)},
			approvalDocument, "Purge staged activities?", "answered", map[string]any{"expected": float64(5), "reload": true}, "operator"),
		Entry("a rejected form", fixtures.RunOptions{Prompter: answering(nil, true)},
			approvalDocument, "Purge staged activities?", "rejected", map[string]any{}, "operator"),
		Entry("params satisfied by supplied values", fixtures.RunOptions{Params: map[string]any{"db": "lab", "n": "5"}},
			paramsDocument, "purge", "supplied", map[string]any{"db": "lab", "n": float64(5), "dry": true}, "supplied"),
		Entry("a form nothing could ask, with its prefill", fixtures.RunOptions{},
			approvalDocument, "Purge staged activities?", "unanswered", map[string]any{"reload": false}, ""),
	)

	It("fails a form whose answer does not arrive before the step's timeout", func() {
		blocking := &fakePrompter{answer: func(ctx context.Context, _ prompt.Prompt) (prompt.Answer, error) {
			<-ctx.Done()
			return prompt.Answer{}, ctx.Err()
		}}
		document := strings.Replace(approvalDocument, "    kind: form\n", "    kind: form\n    timeout: 200ms\n", 1)

		results := runDocument(context.Background(), document, fixtures.RunOptions{Prompter: blocking})

		Expect(results[0].Status).To(Equal(task.StatusFAIL))
		Expect(results[0].TimedOut).To(BeTrue())
		Expect(results[0].Error).To(ContainSubstring("step approval timed out after 200ms"))
		Expect(results[1].Error).To(Equal("skipped: step approval failed"))
	})

	It("fails a form that must ask when the run has no prompter", func() {
		results := runDocument(context.Background(), approvalDocument, fixtures.RunOptions{})

		Expect(results[0].Status).To(Equal(task.StatusFAIL))
		Expect(results[0].Error).To(Equal(`form "approval" needs an interactive prompter (--ui or a TTY); missing required: expected`))
	})

	It("resolves params: from supplied values typed by the schema, without asking", func() {
		results := runDocument(context.Background(), paramsDocument, fixtures.RunOptions{Params: map[string]any{"db": "lab", "n": "5"}})

		Expect(results[0].Status).To(Equal(task.StatusPASS), results[0].Error)
		outputs := outputsOf(results[0])
		Expect(outputs).To(HaveKeyWithValue("n", int64(5)))
		Expect(outputs).To(HaveKeyWithValue("dry", true))
		Expect(outputs).To(HaveKeyWithValue("answeredBy", "supplied"))
		Expect(ran).To(Equal([]string{"inspect"}))
	})

	It("asks params: prefilled with what was supplied and the schema defaults, over a connection picker", func() {
		prompter := answering(map[string]any{"db": "lab", "n": float64(5), "dry": true}, false)

		results := runDocument(context.Background(), paramsDocument, fixtures.RunOptions{Prompter: prompter, Params: map[string]any{"db": "lab"}})

		Expect(results[1].Status).To(Equal(task.StatusPASS), results[1].Error)
		Expect(prompter.asked).To(HaveLen(1))
		Expect(prompter.asked[0].Title).To(Equal("purge"))
		Expect(prompter.asked[0].Default).To(Equal(map[string]any{"db": "lab", "dry": true}))
		var schema map[string]any
		Expect(json.Unmarshal(prompter.asked[0].Schema, &schema)).To(Succeed())
		Expect(schema["x-order"]).To(Equal([]any{"db", "n", "dry"}))
		Expect(schema["properties"].(map[string]any)["db"]).To(HaveKeyWithValue("x-clicky-lookup", map[string]any{
			"url": "/api/v1/connection?types=sql_server", "filter": "connection", "searchParam": "__lookup_q", "multi": false,
		}))
	})

	It("fails params: naming the missing keys before any other step runs when nothing can ask", func() {
		results := runDocument(context.Background(), paramsDocument, fixtures.RunOptions{Params: map[string]any{"dry": "false"}})

		Expect(results[0].Status).To(Equal(task.StatusFAIL))
		Expect(results[0].Error).To(Equal(`form "params" needs an interactive prompter (--ui or a TTY); missing required: db, n`))
		Expect(results[1].Error).To(Equal("skipped: step params failed"))
		Expect(ran).To(BeEmpty())
	})

	It("refuses a supplied param the schema does not declare", func() {
		results := runDocument(context.Background(), paramsDocument, fixtures.RunOptions{Params: map[string]any{"db": "lab", "n": "5", "typo": "x"}})

		Expect(results[0].Status).To(Equal(task.StatusFAIL))
		Expect(results[0].Error).To(ContainSubstring(`unknown value "typo" (declared: db, dry, n)`))
	})

	Describe("an enum computed by $cel from an earlier step", func() {
		BeforeEach(func() { DeferCleanup(fixtures.DefaultKinds.Replace(rowsKind{})) })

		pickerDocument := func(rows, enum string) string {
			return `kind: Fixture
steps:
  - id: databases
    kind: rows
    rows: ` + rows + `
  - id: target
    kind: form
    schema:
      type: object
      required: [database]
      properties:
        database: {type: string, title: Database, enum: ` + enum + `}
  - id: use
    kind: record
    if: target.database == "uat"
`
		}
		const names = `{$cel: "databases.rows.map(r, r.name)"}`

		It("offers the rows as the property's options and gates on the pick", func() {
			prompter := answering(map[string]any{"database": "uat"}, false)

			results := runDocument(context.Background(), pickerDocument(`[{name: master}, {name: uat}]`, names), fixtures.RunOptions{Prompter: prompter})

			Expect(results[1].Status).To(Equal(task.StatusPASS), results[1].Error)
			var schema map[string]any
			Expect(json.Unmarshal(prompter.asked[0].Schema, &schema)).To(Succeed())
			Expect(schema["properties"].(map[string]any)["database"]).To(Equal(map[string]any{
				"type": "string", "title": "Database", "enum": []any{"master", "uat"},
			}))
			Expect(ran).To(Equal([]string{"use"}))
		})

		It("refuses an answer outside the options", func() {
			results := runDocument(context.Background(), pickerDocument(`[{name: master}]`, names),
				fixtures.RunOptions{Prompter: answering(map[string]any{"database": "prod"}, false)})

			Expect(results[1].Status).To(Equal(task.StatusFAIL))
			Expect(results[1].Error).To(ContainSubstring(`form "target": answer does not match its schema`))
			Expect(ran).To(BeEmpty())
		})

		DescribeTable("fails the form, naming the property, when the options cannot be picked from",
			func(rows, enum, message string) {
				results := runDocument(context.Background(), pickerDocument(rows, enum), fixtures.RunOptions{Prompter: answering(nil, false)})

				Expect(results[1].Status).To(Or(Equal(task.StatusFAIL), Equal(task.StatusERR)))
				Expect(results[1].Error).To(ContainSubstring(message))
				Expect(ran).To(BeEmpty())
			},
			Entry("no rows", `[]`, names, "schema property database: enum has no options"),
			Entry("not a list", `[{name: master}]`, `{$cel: "databases.rows[0].name"}`, "schema property database: enum must be a list, got string(master)"),
			Entry("rows, not names", `[{name: master}]`, `{$cel: "databases.rows"}`, "schema property database: enum option 0 must be a string, number or boolean"),
			Entry("a repeated name", `[{name: uat}, {name: uat}]`, names, "schema property database: enum option 1 (uat) is repeated"),
			Entry("$cel beside another key", `[{name: uat}]`, `{$cel: "databases.rows.map(r, r.name)", default: uat}`,
				"step target: schema: properties: database: enum: $cel must be the only key of its object, got 2 keys"),
		)
	})

	It("relays a markdown params: form from an isolated worker to the parent's prompter", func() {
		dir := GinkgoT().TempDir()
		document := "---\nparams:\n  type: object\n  required: [name]\n  properties:\n    name: {type: string}\n---\n\n" +
			"| Name | Command | CEL |\n| --- | --- | --- |\n| greet | echo {{.params.name}} | stdout == 'acme\\n' |\n"
		tree, err := fixtures.ParseMarkdownDocument("greeting.md", document, dir)
		Expect(err).ToNot(HaveOccurred())
		executable, err := os.Executable()
		Expect(err).ToNot(HaveOccurred())
		timeout := "60s"
		prompter := answering(map[string]any{"name": "acme"}, false)

		results, _, err := fixtures.RunNodes(context.Background(), tree.Children, fixtures.RunOptions{
			WorkDir: dir, ExecutablePath: executable, Prompter: prompter, RunID: "run-2",
			ProcessRuntime: processruntime.Options{Timeout: &timeout},
		})

		Expect(err).ToNot(HaveOccurred())
		Expect(results).To(HaveLen(2))
		Expect(results[0].Test.Name).To(Equal("params"))
		Expect(outputsOf(results[0])).To(HaveKeyWithValue("answeredBy", "operator"))
		Expect(results[1].Status).To(Equal(task.StatusPASS), results[1].Error)
		Expect(prompter.asked).To(HaveLen(1))
		Expect(prompter.asked[0].Owner).To(Equal("run-2"))
		Expect(prompter.asked[0].Title).To(Equal("greeting"))
	})
})
