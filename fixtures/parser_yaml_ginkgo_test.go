package fixtures

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/flanksource/clicky/task"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// strictKind is a fake kind with a real config schema.
type strictKind struct{ fakeKind }

func (strictKind) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","additionalProperties":false,"required":["url"],"properties":{"url":{"type":"string"},"retries":{"type":"integer"}}}`)
}

func writeDocument(name, body string) string {
	GinkgoHelper()
	path := filepath.Join(GinkgoT().TempDir(), name)
	Expect(os.WriteFile(path, []byte(body), 0o600)).To(Succeed())
	return path
}

func testsOf(node *FixtureNode) []*FixtureTest {
	var tests []*FixtureTest
	node.Walk(func(n *FixtureNode) {
		if n.Test != nil {
			tests = append(tests, n.Test)
		}
	})
	return tests
}

var _ = Describe("ParseFixtureFile (YAML/JSON documents)", func() {
	BeforeEach(func() {
		useKind(strictKind{fakeKind{name: "probe", execute: passing(nil, nil)}})
		useKind(observingKind{fakeKind: fakeKind{name: "poller", execute: passing(nil, nil)},
			observe: func(context.Context, *StepContext, any) (map[string]any, error) { return nil, nil }})
	})

	It("records the lines each step spans, leaving out the comments and separator that follow it", func() {
		path := writeDocument("spans.fixture.yaml", `kind: Fixture
steps:
  - id: api
    kind: probe
    url: http://x

  # the next one polls
  - { id: poll, kind: probe, url: http://y }
env: { GREETING: hello }
---
kind: Fixture
steps:
  - id: last
    kind: probe
    url: |
      http://z
`)
		file, err := ParseFixtureFile(path)
		Expect(err).ToNot(HaveOccurred())
		var spans [][2]int
		file.Walk(func(n *FixtureNode) {
			if n.Test != nil {
				spans = append(spans, [2]int{n.Origin.Line, n.Origin.EndLine})
			}
		})
		Expect(spans).To(Equal([][2]int{{3, 5}, {8, 8}, {13, 16}}))
	})

	It("lowers each step of each document to a test node carrying its envelope, config and the document defaults", func() {
		path := writeDocument("stack.fixture.yaml", `
kind: Fixture
name: first
defaults:
  wait: { interval: 1s, timeout: 2m }
  timeout: 5m
env: { GREETING: hello }
steps:
  - id: api
    kind: probe
    url: http://127.0.0.1:{{.env.PORT}}
    retries: 2
    export: { API_URL: "{{.outputs.url}}" }
  - kind: poller
    name: wait for api
    wait: { until: "steps.api.url != ''" }
---
kind: Fixture
name: second
steps:
  - { kind: probe, url: http://x }
`)
		file, err := ParseFixtureFile(path)
		Expect(err).ToNot(HaveOccurred())

		Expect(file.Type).To(Equal(FileNode))
		Expect(file.Origin.File).To(Equal(path))
		Expect(file.Children).To(HaveLen(2))
		Expect(file.Children[0].Name).To(Equal("first"))
		tests := testsOf(file)
		Expect(tests).To(HaveLen(3))
		Expect([]string{tests[0].Name, tests[1].Name, tests[2].Name}).To(Equal([]string{"api", "wait for api", "probe"}))
		api := tests[0].Step
		Expect(api.Kind).To(Equal("probe"))
		Expect(api.Export).To(Equal(map[string]string{"API_URL": "{{.outputs.url}}"}))
		Expect(api.Config).To(MatchJSON(`{"url":"http://127.0.0.1:{{.env.PORT}}","retries":2}`))
		Expect(api.Defaults.Timeout.Std()).To(Equal(5 * time.Minute))
		Expect(api.Defaults.Wait.Interval.Std()).To(Equal(time.Second))
		Expect(tests[0].FrontMatter.Env).To(HaveKeyWithValue("GREETING", "hello"))
		Expect(tests[0].SourceDir).To(Equal(filepath.Dir(path)))
		Expect(tests[1].Step.Wait.Until).To(Equal("steps.api.url != ''"))
		Expect(tests[2].Step.Defaults).To(BeNil())
	})

	It("resolves a document opened by relative path to an absolute fixture directory", func() {
		path := writeDocument("rel.fixture.yaml", "kind: Fixture\nsteps:\n  - { kind: probe, url: http://x }\n")
		cwd, err := os.Getwd()
		Expect(err).ToNot(HaveOccurred())
		relative, err := filepath.Rel(cwd, path)
		Expect(err).ToNot(HaveOccurred())
		Expect(filepath.IsAbs(relative)).To(BeFalse())

		file, err := ParseFixtureFile(relative)
		Expect(err).ToNot(HaveOccurred())
		Expect(testsOf(file)[0].SourceDir).To(Equal(filepath.Dir(path)))
	})

	It("keeps the authored key order of a step's config, top level and nested", func() {
		useKind(fakeKind{name: "ordered", execute: passing(nil, nil)})
		path := writeDocument("order.fixture.yaml", `
kind: Fixture
steps:
  - kind: ordered
    zeta: 1
    id: s
    alpha: { y: 1, x: [ { b: 2, a: 1 } ] }
    services:
      postgres: { deps: postgres }
      aimock: { exec: mock }
      gavel: { exec: serve }
`)
		file, err := ParseFixtureFile(path)
		Expect(err).ToNot(HaveOccurred())
		Expect(string(testsOf(file)[0].Step.Config)).To(Equal(
			`{"zeta":1,"alpha":{"y":1,"x":[{"b":2,"a":1}]},"services":{"postgres":{"deps":"postgres"},"aimock":{"exec":"mock"},"gavel":{"exec":"serve"}}}`))
	})

	It("keeps the key order of a JSON document's step config", func() {
		useKind(fakeKind{name: "ordered", execute: passing(nil, nil)})
		path := writeDocument("order.fixture.json", `{"kind":"Fixture","steps":[{"kind":"ordered","z":1,"a":{"d":1,"c":2}}]}`)
		file, err := ParseFixtureFile(path)
		Expect(err).ToNot(HaveOccurred())
		Expect(string(testsOf(file)[0].Step.Config)).To(Equal(`{"z":1,"a":{"d":1,"c":2}}`))
	})

	It("parses a JSON document the same way", func() {
		path := writeDocument("one.fixture.json", `{"kind":"Fixture","steps":[{"id":"a","kind":"probe","url":"http://x","cel":"true"}]}`)
		file, err := ParseFixtureFile(path)
		Expect(err).ToNot(HaveOccurred())
		tests := testsOf(file)
		Expect(tests).To(HaveLen(1))
		Expect(tests[0].Step.CEL).To(Equal("true"))
		Expect(tests[0].Step.Config).To(MatchJSON(`{"url":"http://x"}`))
	})

	DescribeTable("rejects an invalid document, naming the file and step",
		func(body, message string) {
			path := writeDocument("bad.fixture.yaml", body)
			_, err := ParseFixtureFile(path)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(path))
			Expect(err.Error()).To(ContainSubstring(message))
		},
		Entry("unknown kind", "kind: Fixture\nsteps:\n  - { id: s, kind: nope }\n", `step "s": unknown fixture step kind "nope"`),
		Entry("config that breaks the kind schema", "kind: Fixture\nsteps:\n  - { id: s, kind: probe, url: 3 }\n", `step "s": config does not match the probe schema`),
		Entry("config missing a required key", "kind: Fixture\nsteps:\n  - { id: s, kind: probe }\n", `step "s": config does not match`),
		Entry("wait on a kind without Observe", "kind: Fixture\nsteps:\n  - { id: s, kind: probe, url: x, wait: { until: 'true' } }\n", `kind "probe" cannot wait`),
		Entry("background on a kind without Start", "kind: Fixture\nsteps:\n  - { id: s, kind: probe, url: x, background: true }\n", `kind "probe" cannot run in the background`),
		Entry("duplicate step id", "kind: Fixture\nsteps:\n  - { id: s, kind: probe, url: x }\n  - { id: s, kind: probe, url: y }\n", `duplicate step id "s"`),
		Entry("reference to a later step", "kind: Fixture\nsteps:\n  - { id: a, kind: probe, url: '{{.steps.b.url}}' }\n  - { id: b, kind: probe, url: y }\n", `references steps.b`),
		Entry("reference to an unknown step in CEL", "kind: Fixture\nsteps:\n  - { id: a, kind: probe, url: x, cel: \"steps.ghost.ok\" }\n", `references steps.ghost`),
		Entry("export collision", "kind: Fixture\nsteps:\n  - { id: a, kind: probe, url: x, export: { URL: x } }\n  - { id: b, kind: probe, url: y, export: { URL: y } }\n", `URL is already exported by step "a"`),
		Entry("document of another kind", "kind: Pipeline\nsteps:\n  - { kind: probe, url: x }\n", `kind: Fixture`),
		Entry("unknown document key", "kind: Fixture\nstages: []\nsteps:\n  - { kind: probe, url: x }\n", `unknown field "stages"`),
		Entry("a step without kind", "kind: Fixture\nsteps:\n  - { id: s, url: x }\n", `step "s": kind is required`),
		Entry("no steps", "kind: Fixture\nsteps: []\n", "declares no steps"),
	)

	It("allows an export collision the later step marks override", func() {
		path := writeDocument("ok.fixture.yaml", "kind: Fixture\nsteps:\n  - { id: a, kind: probe, url: x, export: { URL: x } }\n  - { id: b, kind: probe, url: y, override: true, export: { URL: y } }\n")
		_, err := ParseFixtureFile(path)
		Expect(err).ToNot(HaveOccurred())
	})

	It("runs a parsed document's steps through the Runner, sequentially, with exports reaching later steps", func() {
		var seen string
		useKind(strictKind{fakeKind{name: "probe", execute: func(_ context.Context, step *StepContext) (StepOutcome, error) {
			cfg, err := KindConfig[map[string]any](step)
			if step.Spec.ID == "second" {
				seen = cfg["url"].(string)
			}
			return StepOutcome{Result: FixtureResult{Status: task.StatusPASS}, Outputs: map[string]any{"url": cfg["url"]}}, err
		}}})
		path := writeDocument("flow.fixture.yaml", `
kind: Fixture
steps:
  - { id: first, kind: probe, url: http://a, export: { BASE: "{{.outputs.url}}/v1" } }
  - { id: second, kind: probe, url: "{{.env.BASE}}/{{.steps.first.url}}" }
`)
		runner, err := NewRunner(RunnerOptions{Paths: []string{path}, WorkDir: filepath.Dir(path), MaxWorkers: 4})
		Expect(err).ToNot(HaveOccurred())
		tree, err := runner.Run()
		Expect(err).ToNot(HaveOccurred())
		Expect(tree.GetStats().Passed).To(Equal(2))
		Expect(seen).To(Equal("http://a/v1/http://a"))
	})
})
