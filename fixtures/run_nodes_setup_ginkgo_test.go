package fixtures

import (
	"context"
	"path/filepath"

	"github.com/flanksource/clicky/task"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("RunNodes prepares the document's setup", func() {
	const greeting = "from-setup"

	// probe records the GREETING it was rendered with and the prepared setup
	// it ran under.
	type probed struct{ rendered, prepared string }
	useProbe := func() *probed {
		seen := &probed{}
		useKind(fakeKind{name: "probe", execute: func(_ context.Context, step *StepContext) (StepOutcome, error) {
			cfg, err := KindConfig[map[string]any](step)
			seen.rendered, _ = cfg["greeting"].(string)
			if step.Opts.Setup != nil {
				seen.prepared = step.Opts.Setup.Env["GREETING"]
			}
			return StepOutcome{Result: FixtureResult{Status: task.StatusPASS}}, err
		}})
		return seen
	}
	runDocument := func(body string, opts RunOptions) []FixtureResult {
		GinkgoHelper()
		path := writeDocument("setup.fixture.yaml", body)
		file, err := ParseFixtureFile(path)
		Expect(err).ToNot(HaveOccurred())
		opts.WorkDir = filepath.Dir(path)
		results, _, err := RunNodes(context.Background(), []*FixtureNode{file}, opts)
		Expect(err).ToNot(HaveOccurred())
		return results
	}
	statusOf := func(results []FixtureResult) map[string]string {
		byName := map[string]string{}
		for _, result := range results {
			byName[result.Name] = string(result.Status) + ": " + result.Error
		}
		return byName
	}

	It("prepares an eager setup before the steps run", func() {
		seen := useProbe()
		results := runDocument(`
kind: Fixture
setup:
  envVars: [{name: GREETING, value: `+greeting+`}]
steps:
  - { id: use, kind: probe, greeting: "{{.env.GREETING}}" }
`, RunOptions{})
		Expect(statusOf(results)).To(Equal(map[string]string{"use": "PASS: "}))
		Expect(*seen).To(Equal(probed{rendered: greeting, prepared: greeting}))
	})

	It("prepares a params-reading setup once the params form answered", func() {
		useKind(fakeKind{name: FormKind, execute: passing(map[string]any{"greeting": greeting}, nil)})
		seen := useProbe()
		results := runDocument(`
kind: Fixture
params: {type: object, properties: {greeting: {type: string}}}
setup:
  envVars: [{name: GREETING, value: "{{.params.greeting}}"}]
steps:
  - { id: use, kind: probe, greeting: "{{.env.GREETING}}" }
`, RunOptions{})
		Expect(statusOf(results)).To(Equal(map[string]string{"params": "PASS: ", "use": "PASS: "}))
		Expect(*seen).To(Equal(probed{rendered: greeting, prepared: greeting}))
	})

	It("keeps a caller-prepared setup instead of preparing the document's", func() {
		seen := useProbe()
		results := runDocument(`
kind: Fixture
setup:
  envVars: [{name: GREETING, value: from-document}]
steps:
  - { id: use, kind: probe, greeting: "{{.env.GREETING}}" }
`, RunOptions{Setup: &PreparedSetup{Env: map[string]string{"GREETING": greeting}}})
		Expect(statusOf(results)).To(Equal(map[string]string{"use": "PASS: "}))
		Expect(*seen).To(Equal(probed{rendered: greeting, prepared: greeting}))
	})

	DescribeTable("errors the first step on a setup that fails to prepare and skips the rest",
		func(body, first, cause string) {
			useKind(fakeKind{name: FormKind, execute: passing(map[string]any{"cluster": "acme"}, nil)})
			seen := useProbe()
			results := runDocument(body, RunOptions{})
			Expect(results).To(HaveLen(2))
			Expect(results[0].Name).To(Equal(first))
			Expect(results[0].Status).To(Equal(task.StatusERR))
			Expect(results[0].Error).To(ContainSubstring(cause))
			Expect(statusOf(results)).To(HaveKeyWithValue("use", "SKIP: skipped: step "+first+" failed"))
			Expect(*seen).To(Equal(probed{}))
		},
		Entry("an eager setup", `
kind: Fixture
setup:
  connections: {kubernetes: {connection: "connection://acme"}}
steps:
  - { id: first, kind: probe, greeting: x }
  - { id: use, kind: probe, greeting: x }
`, "first", "resolving a connection:// reference needs a database"),
		Entry("a params-reading setup", `
kind: Fixture
params: {type: object, properties: {cluster: {type: string}}}
setup:
  connections: {kubernetes: {connection: "connection://{{.params.cluster}}"}}
steps:
  - { id: use, kind: probe, greeting: x }
`, "params", `connection "connection://acme" not found`),
	)
})
