package form_test

import (
	"context"
	"encoding/json"

	"github.com/flanksource/clicky/prompt"
	"github.com/flanksource/clicky/task"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flanksource/gavel/fixtures"
)

// scaleDocument's params show cluster and workloads only while scale is on.
const scaleDocument = `kind: Fixture
name: purge
params:
  type: object
  required: [db]
  properties:
    db: {type: string}
    scale: {type: boolean, default: true}
  allOf:
    - if: {properties: {scale: {const: true}}, required: [scale]}
      then:
        required: [cluster]
        properties:
          cluster: {type: string, format: connection, x-connection-types: [kubernetes]}
          workloads: {type: array, items: {type: string}, default: [oipa, cycle]}
steps:
  - {id: inspect, kind: record}
`

var _ = Describe("a params: form with conditional properties", func() {
	var ran []string
	BeforeEach(func() {
		ran = nil
		DeferCleanup(fixtures.DefaultKinds.Replace(recordKind{ran: &ran}))
	})

	DescribeTable("resolves supplied values through the branch their condition selects",
		func(supplied, want map[string]any) {
			results := runDocument(context.Background(), scaleDocument, fixtures.RunOptions{Params: supplied})

			Expect(results[0].Status).To(Equal(task.StatusPASS), results[0].Error)
			outputs := outputsOf(results[0])
			Expect(outputs["values"]).To(Equal(want))
			Expect(ran).To(Equal([]string{"inspect"}))
		},
		Entry("scale on: the branch's default fills workloads and cluster is typed",
			map[string]any{"db": "lab", "cluster": "connection://acme"},
			map[string]any{"db": "lab", "scale": true, "cluster": "connection://acme", "workloads": []any{"oipa", "cycle"}}),
		Entry("scale off: the branch's properties are absent, a supplied one dropped",
			map[string]any{"db": "lab", "scale": "false", "cluster": "connection://acme"},
			map[string]any{"db": "lab", "scale": false}),
	)

	It("requires the active branch's properties", func() {
		results := runDocument(context.Background(), scaleDocument, fixtures.RunOptions{Params: map[string]any{"db": "lab"}})

		Expect(results[0].Status).To(Equal(task.StatusFAIL))
		Expect(results[0].Error).To(Equal(`form "params" needs an interactive prompter (--ui or a TTY); missing required: cluster`))
	})

	It("refuses an answer missing the active branch's required property", func() {
		prompter := answering(map[string]any{"db": "lab", "scale": true}, false)

		results := runDocument(context.Background(), scaleDocument, fixtures.RunOptions{Prompter: prompter, Params: map[string]any{"db": "lab"}})

		Expect(results[0].Status).To(Equal(task.StatusFAIL))
		Expect(results[0].Error).To(ContainSubstring(`form "params": answer does not match its schema: missing required field "cluster"`))
	})

	It("keeps waiting when the prompt manager refuses a submit missing the active branch's required property, and passes on the corrected resubmit", func() {
		manager := prompt.NewManager(nil)
		done := make(chan []fixtures.FixtureResult, 1)
		go func() {
			defer GinkgoRecover()
			done <- runDocument(context.Background(), scaleDocument, fixtures.RunOptions{Prompter: manager, Params: map[string]any{"db": "lab"}})
		}()
		var pending []prompt.PromptSnapshot
		Eventually(func() []prompt.PromptSnapshot {
			pending = manager.List(prompt.Filter{State: string(prompt.StatePending)})
			return pending
		}).Should(HaveLen(1))
		id := pending[0].ID

		Expect(manager.Resolve(id, prompt.Answer{Values: map[string]any{"db": "lab", "scale": true}})).
			To(MatchError(ContainSubstring(`missing required field "cluster"`)))
		_, stillPending := manager.Pending(id)
		Expect(stillPending).To(BeTrue())
		Consistently(done, "200ms").ShouldNot(Receive())

		Expect(manager.Resolve(id, prompt.Answer{Values: map[string]any{"db": "lab", "scale": true, "cluster": "connection://acme"}})).To(Succeed())
		var results []fixtures.FixtureResult
		Eventually(done).Should(Receive(&results))
		Expect(results[0].Status).To(Equal(task.StatusPASS), results[0].Error)
		Expect(outputsOf(results[0])["cluster"]).To(Equal("connection://acme"))
		Expect(ran).To(Equal([]string{"inspect"}))
	})

	It("offers a branch's connection property through the connection picker", func() {
		prompter := answering(map[string]any{"db": "lab", "scale": true, "cluster": "connection://acme"}, false)

		results := runDocument(context.Background(), scaleDocument, fixtures.RunOptions{Prompter: prompter, Params: map[string]any{"db": "lab"}})

		Expect(results[0].Status).To(Equal(task.StatusPASS), results[0].Error)
		var schema struct {
			AllOf []struct {
				Then struct {
					Properties map[string]map[string]any `json:"properties"`
				} `json:"then"`
			} `json:"allOf"`
		}
		Expect(json.Unmarshal(prompter.asked[0].Schema, &schema)).To(Succeed())
		Expect(schema.AllOf[0].Then.Properties["cluster"]).To(HaveKeyWithValue("x-clicky-lookup", map[string]any{
			"url": "/api/v1/connection?types=kubernetes", "filter": "connection", "searchParam": "__lookup_q", "multi": false,
		}))
	})
})
