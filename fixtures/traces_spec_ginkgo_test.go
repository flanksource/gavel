package fixtures_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flanksource/gavel/fixtures"
)

// tests are the parsed fixtures of a tree, in order.
func tests(tree *fixtures.FixtureNode) []fixtures.FixtureTest {
	var found []fixtures.FixtureTest
	tree.Walk(func(node *fixtures.FixtureNode) {
		if node.Test != nil {
			found = append(found, *node.Test)
		}
	})
	return found
}

func parseMarkdown(markdown string) []fixtures.FixtureTest {
	tree, err := fixtures.ParseMarkdownDocument("traces.md", markdown, GinkgoT().TempDir())
	Expect(err).ToNot(HaveOccurred())
	return tests(tree)
}

func declared(test fixtures.FixtureTest) fixtures.TraceSpecs {
	specs := test.DeclaredTraces()
	Expect(specs).ToNot(BeNil(), "%s declared no traces", test.Name)
	return *specs
}

var httpProxy = fixtures.TraceSpec{Kind: "http_proxy", Scope: fixtures.TraceScopeTest}

var _ = Describe("traces: declarations", func() {
	It("reads the compact form: kinds, none, and nothing", func() {
		specs, err := fixtures.ParseTraces("http_proxy, terminal")
		Expect(err).ToNot(HaveOccurred())
		Expect(*specs).To(Equal(fixtures.TraceSpecs{httpProxy, {Kind: "terminal", Scope: fixtures.TraceScopeTest}}))

		optOut, err := fixtures.ParseTraces("none")
		Expect(err).ToNot(HaveOccurred())
		Expect(optOut).ToNot(BeNil(), "none is a declaration")
		Expect(*optOut).To(BeEmpty())

		nothing, err := fixtures.ParseTraces("  ")
		Expect(err).ToNot(HaveOccurred())
		Expect(nothing).To(BeNil())
	})

	It("refuses an unknown kind by naming the registered ones", func() {
		_, err := fixtures.ParseTraces("smtp")
		Expect(err).To(MatchError(ContainSubstring(`unknown trace kind "smtp" (registered: goprof, http, http_proxy, pg_proxy, process_samples, process_tree, sql_statements, stacks, terminal)`)))
	})

	DescribeTable("refuses an entry the host could not arm",
		func(entries, message string) {
			var specs fixtures.TraceSpecs
			Expect(json.Unmarshal([]byte(entries), &specs)).To(MatchError(ContainSubstring(message)))
		},
		Entry("params the kind refuses", `[{"kind": "http_proxy", "params": {"mode": "sniff"}}]`, "trace http_proxy params"),
		Entry("an unknown scope", `[{"kind": "http_proxy", "scope": "suite"}]`, `scope "suite" is not one of test, file, run`),
		Entry("a terminal beyond its test", `[{"kind": "terminal", "scope": "file"}]`, "its scope must be test"),
		Entry("one kind twice", `["http_proxy", "http_proxy"]`, "http_proxy is declared twice"),
		Entry("an unknown key", `[{"kind": "http_proxy", "requires": "count > 0"}]`, `unknown field "requires"`),
	)

	It("binds front matter traces to every test, with params, scope and require", func() {
		found := parseMarkdown(`---
traces:
  - kind: http_proxy
    params: {mode: mitm}
    scope: file
    require: count > 0
---

# Calls

| Name | Command |
|------|---------|
| one  | true    |
`)
		Expect(found).To(HaveLen(1))
		Expect(declared(found[0])).To(Equal(fixtures.TraceSpecs{{
			Kind: "http_proxy", Params: json.RawMessage(`{"mode":"mitm"}`), Scope: fixtures.TraceScopeFile, Require: "count > 0",
		}}))
		Expect(found[0].FrontMatter.Metadata).ToNot(HaveKey("traces"), "traces: leaked into the template variables")
	})

	It("lets a table row's traces column replace the file's, and none opt out", func() {
		// pg_proxy resolves its upstream from the environment when declared
		// without a dsn; pin one so the spec does not depend on the host's.
		GinkgoT().Setenv("GAVEL_DB_DSN", "postgres://tester@127.0.0.1:5432/app")
		found := parseMarkdown(`---
traces: [pg_proxy]
---

# Rows

| Name    | Command | Traces     |
|---------|---------|------------|
| own     | true    | http_proxy |
| none    | true    | none       |
| file    | true    |            |
`)
		Expect(found).To(HaveLen(3))
		Expect(declared(found[0])).To(Equal(fixtures.TraceSpecs{httpProxy}))
		Expect(declared(found[1])).To(BeEmpty())
		Expect(declared(found[2])).To(Equal(fixtures.TraceSpecs{{Kind: "pg_proxy", Scope: fixtures.TraceScopeTest}}))
	})

	It("binds a command block's own traces", func() {
		found := parseMarkdown(`
### command: echo hi

` + "```yaml\ntraces: terminal\n```" + `

` + "```bash\necho hi\n```" + `
`)
		Expect(found).To(HaveLen(1))
		Expect(declared(found[0])).To(Equal(fixtures.TraceSpecs{{Kind: "terminal", Scope: fixtures.TraceScopeTest}}))
	})

	It("binds a fixture document's traces to its steps, and a step's own over them", func() {
		tree, err := fixtures.ParseFixtureDocuments("doc.fixture.yaml", []byte(`
kind: Fixture
traces: [http_proxy]
steps:
  - id: inherits
    kind: exec
    script: "true"
  - id: own
    kind: exec
    script: "true"
    traces: []
`))
		Expect(err).ToNot(HaveOccurred())
		found := tests(tree)
		Expect(found).To(HaveLen(2))
		Expect(declared(found[0])).To(Equal(fixtures.TraceSpecs{httpProxy}))
		Expect(declared(found[1])).To(BeEmpty())
	})

	It("defers a test-scope trace's templated params to its start, and refuses them at file scope", func() {
		const templatedDSN = `dsn: "postgres://tester@127.0.0.1:{{.env.GAVEL_SPEC_UNSET_PG_PORT}}/app"`
		found := parseMarkdown(`---
traces:
  - kind: pg_proxy
    params: {` + templatedDSN + `}
---

| Name | Command |
|------|---------|
| psql | true    |
`)
		Expect(declared(found[0])).To(ConsistOf(HaveField("Kind", "pg_proxy")))

		var specs fixtures.TraceSpecs
		err := json.Unmarshal([]byte(`[{"kind": "pg_proxy", "scope": "file", "params": {"dsn": "postgres://tester@127.0.0.1:{{.env.PG_PORT}}/app"}}]`), &specs)
		Expect(err).To(MatchError(ContainSubstring("pg_proxy params are templated, so its scope must be test")))
	})

	DescribeTable("parses the repository's traced fixtures",
		func(path, kind string) {
			tree, err := fixtures.ParseFixtureFile(path)
			Expect(err).ToNot(HaveOccurred())
			found := tests(tree)
			Expect(found).ToNot(BeEmpty())
			for _, test := range found {
				Expect(declared(test)).To(ContainElement(HaveField("Kind", kind)))
			}
		},
		Entry("the PR status terminal audit", "testdata/pr-status-ansi.md", "terminal"),
		Entry("the short-screen terminal audit", "testdata/pr-status-ansi-short-screen.md", "terminal"),
		Entry("the session tail proxy audit", "../pr/ui/testdata/session-tail.fixture.md", "http_proxy"),
	)

	It("keeps an opt-out through the JSON a fixture worker is sent", func() {
		found := parseMarkdown(`
| Name | Command | Traces |
|------|---------|--------|
| none | true    | none   |
`)
		encoded, err := json.Marshal(found[0])
		Expect(err).ToNot(HaveOccurred())
		var decoded fixtures.FixtureTest
		Expect(json.Unmarshal(encoded, &decoded)).To(Succeed())
		Expect(declared(decoded)).To(BeEmpty())
	})
})
