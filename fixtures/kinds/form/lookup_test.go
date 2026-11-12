package form_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/flanksource/clicky/prompt"
	"github.com/flanksource/clicky/schemaform"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/fixtures/kinds/dbconn"
	"github.com/flanksource/gavel/fixtures/kinds/form"
	"github.com/flanksource/gavel/fixtures/kinds/lookup"
	"github.com/flanksource/gavel/processruntime"
)

// lookupFixture is a fixture directory holding lab.db (a `tables` table) and
// the catalog connection `lab` pointing at it.
func lookupFixture() (dir, dbPath string) {
	GinkgoHelper()
	GinkgoT().Setenv("XDG_CONFIG_HOME", GinkgoT().TempDir())
	dir = GinkgoT().TempDir()
	dbPath = filepath.Join(dir, "lab.db")
	db, err := sql.Open("sqlite", dbPath)
	Expect(err).NotTo(HaveOccurred())
	defer db.Close()
	_, err = db.Exec("CREATE TABLE tables(name TEXT); INSERT INTO tables VALUES ('orders'), ('customers')")
	Expect(err).NotTo(HaveOccurred())
	Expect(os.MkdirAll(filepath.Join(dir, ".connections"), 0o700)).To(Succeed())
	Expect(os.WriteFile(filepath.Join(dir, ".connections", "lab.yaml"), []byte("apiVersion: mission-control.flanksource.com/v1\nkind: Connection\n"+
		"metadata:\n  name: lab\nspec:\n  type: sqlite\n  url:\n    value: "+dbPath+"\n"), 0o600)).To(Succeed())
	return dir, dbPath
}

// lookupDocument is a fixture whose params link `table` to the `db` picker
// and read `fixed` from the named connection `lab`.
func lookupDocument(connections, extra string) string {
	return "kind: Fixture\nname: lookup\n" + connections + `params:
  type: object
  required: [table]
  properties:
    db: {type: string, format: connection, x-connection-types: [sqlite]}
    table:
      type: string
      x-query: {connection: db, sql: "select name as value from tables where name like '%' || @q || '%'"}
` + extra + `steps:
  - id: inspect
    kind: record
`
}

// askedSchema runs document in dir and returns the params form it asked,
// with what the terminal lookup listed for table while the form was open.
func askedSchema(dir, document string, opts fixtures.RunOptions) (prompt.Prompt, schemaform.LookupResult, []fixtures.FixtureResult) {
	GinkgoHelper()
	var asked prompt.Prompt
	var listed schemaform.LookupResult
	var listErr error
	prompter := &fakePrompter{answer: func(ctx context.Context, p prompt.Prompt) (prompt.Answer, error) {
		asked = p
		var schema struct {
			Properties map[string]struct {
				Lookup schemaform.LookupSpec `json:"x-clicky-lookup"`
			} `json:"properties"`
		}
		if listErr = json.Unmarshal(p.Schema, &schema); listErr == nil {
			listed, listErr = form.TerminalLookup(dir, dbconn.Options{})(ctx, schemaform.LookupRequest{
				Spec: schema.Properties["table"].Lookup, Query: "ord", Values: map[string]any{"db": "connection://lab"},
			})
		}
		return prompt.Answer{Cancelled: true}, nil
	}}
	path := filepath.Join(dir, "lookup.fixture.yaml")
	Expect(os.WriteFile(path, []byte(document), 0o600)).To(Succeed())
	file, err := fixtures.ParseFixtureFile(path)
	Expect(err).NotTo(HaveOccurred())
	opts.WorkDir, opts.Prompter = dir, prompter
	results, _, err := fixtures.RunNodes(context.Background(), []*fixtures.FixtureNode{file}, opts)
	Expect(err).NotTo(HaveOccurred())
	Expect(asked.Schema).NotTo(BeEmpty(), "the params form was not asked: %v", results)
	Expect(listErr).NotTo(HaveOccurred())
	return asked, listed, results
}

func lookupSpecOf(schema json.RawMessage, property string) map[string]any {
	GinkgoHelper()
	var doc struct {
		Properties map[string]map[string]any `json:"properties"`
	}
	Expect(json.Unmarshal(schema, &doc)).To(Succeed())
	Expect(doc.Properties[property]).NotTo(HaveKey("x-query"))
	spec, ok := doc.Properties[property]["x-clicky-lookup"].(map[string]any)
	Expect(ok).To(BeTrue(), "property %s has no x-clicky-lookup: %v", property, doc.Properties[property])
	return spec
}

var _ = Describe("x-query lowering", func() {
	var ran []string
	BeforeEach(func() {
		ran = nil
		DeferCleanup(fixtures.DefaultKinds.Replace(recordKind{ran: &ran}))
	})

	It("binds a linked query to a step-scoped token the terminal resolves with the form's values", func() {
		dir, dbPath := lookupFixture()
		asked, listed, _ := askedSchema(dir, lookupDocument("connections:\n  lab: {type: sqlite, url: "+dbPath+"}\n",
			"    fixed:\n      type: string\n      x-query: {connection: lab, sql: \"select 'x' as value\"}\n"), fixtures.RunOptions{})

		table := lookupSpecOf(asked.Schema, "table")
		Expect(table["url"]).To(HavePrefix(lookup.URLPrefix))
		Expect(table["url"]).To(HaveSuffix("/table"))
		Expect(table).To(HaveKeyWithValue("filter", form.QueryLookupFilter))
		Expect(table).To(HaveKeyWithValue("scope", map[string]any{"param": "connection", "from": "db"}))
		Expect(lookupSpecOf(asked.Schema, "fixed")).NotTo(HaveKey("scope"))
		Expect(listed.Options).To(Equal([]schemaform.Option{{Value: "orders", Label: "orders"}}))

		db := lookupSpecOf(asked.Schema, "db")
		Expect(db["url"]).To(HavePrefix(lookup.URLPrefix), "a fixture declaring connections lists them through its own token")
		token, pointer, _, err := lookup.ParseURL(table["url"].(string))
		Expect(err).NotTo(HaveOccurred())
		_, err = lookup.Resolve(context.Background(), lookup.Request{Token: token, Pointer: pointer})
		Expect(err).To(MatchError(ContainSubstring("form step has ended")))
	})

	It("prefills a one-type picker with its declared default, still asking", func() {
		dir, _ := lookupFixture()
		asked, _, _ := askedSchema(dir, lookupDocument("connections: [lab]\n", ""), fixtures.RunOptions{})
		Expect(asked.Default).To(Equal(map[string]any{"db": "connection://lab"}))
	})

	DescribeTable("lists the example's tables matching the search, decorated, from its inline sqlite connection, then ends the field",
		func(isolated bool) {
			GinkgoT().Setenv("XDG_CONFIG_HOME", GinkgoT().TempDir())
			example, err := os.ReadFile(filepath.Join("..", "..", "..", "examples", "query-lookup.fixture.yaml"))
			Expect(err).NotTo(HaveOccurred())
			var opts fixtures.RunOptions
			if isolated {
				executable, err := os.Executable()
				Expect(err).NotTo(HaveOccurred())
				timeout := "60s"
				opts = fixtures.RunOptions{ExecutablePath: executable, ProcessRuntime: processruntime.Options{Timeout: &timeout}}
			}

			asked, listed, _ := askedSchema(GinkgoT().TempDir(), string(example), opts)

			Expect(listed.Options).To(Equal([]schemaform.Option{
				{Value: "order_items", Label: "Order items", Icon: "ph:table-thin", Badges: []schemaform.Badge{{Label: "fk"}}},
				{Value: "orders", Label: "Orders", Icon: "ph:table-thin", Badges: []schemaform.Badge{{Label: "pk"}, {Label: "fk"}}},
			}))
			token, pointer, _, err := lookup.ParseURL(lookupSpecOf(asked.Schema, "table")["url"].(string))
			Expect(err).NotTo(HaveOccurred())
			_, err = lookup.Resolve(context.Background(), lookup.Request{Token: token, Pointer: pointer})
			Expect(err).To(MatchError(lookup.ErrNotFound))
		},
		Entry("in the run's process", false),
		Entry("in an isolated worker, relayed through the parent", true),
	)

	It("fails a terminal connection picker with nothing to pick, naming where connections live", func() {
		GinkgoT().Setenv("XDG_CONFIG_HOME", GinkgoT().TempDir())
		dir := GinkgoT().TempDir()
		_, err := form.TerminalLookup(dir, dbconn.Options{})(context.Background(), schemaform.LookupRequest{
			Spec: schemaform.LookupSpec{URL: form.ConnectionLookupURL + "?types=postgres", Filter: form.ConnectionLookupFilter},
		})
		Expect(err).To(MatchError(ContainSubstring("no postgres connection in the fixture's connections or the gavel connection catalog")))
	})

	It("keeps the global catalog lookup for a picker in a fixture declaring no connections", func() {
		dir, _ := lookupFixture()
		document := strings.Replace(lookupDocument("", ""), "{connection: db, sql:", "{connection: lab, sql:", 1)
		asked, _, _ := askedSchema(dir, document, fixtures.RunOptions{})
		Expect(lookupSpecOf(asked.Schema, "db")).To(HaveKeyWithValue("url", form.ConnectionLookupURL+"?types=sqlite"))
	})

	DescribeTable("fails the form on an x-query it cannot bind",
		func(replace, with, message string) {
			dir, _ := lookupFixture()
			results := runDocumentIn(dir, strings.Replace(lookupDocument("", ""), replace, with, 1))
			Expect(results[0].Error).To(ContainSubstring(message))
		},
		Entry("a sibling that is not a connection picker", "format: connection, x-connection-types: [sqlite]", "format: email",
			"schema property table: x-query.connection: sibling property db is not format: connection"),
		Entry("no sql", `sql: "select name as value from tables where name like '%' || @q || '%'"`, "sql: \"\"",
			"schema property table: x-query.sql is required"),
		Entry("a connection nothing resolves", "{connection: db, sql:", "{connection: missing, sql:",
			`schema property table: x-query.connection: connection "missing": connection not found`),
		Entry("no connection and no default", "{connection: db, sql:", "{sql:",
			"no default connection"),
	)
})

func runDocumentIn(dir, document string) []fixtures.FixtureResult {
	GinkgoHelper()
	path := filepath.Join(dir, "lookup.fixture.yaml")
	Expect(os.WriteFile(path, []byte(document), 0o600)).To(Succeed())
	file, err := fixtures.ParseFixtureFile(path)
	Expect(err).NotTo(HaveOccurred())
	results, _, err := fixtures.RunNodes(context.Background(), []*fixtures.FixtureNode{file}, fixtures.RunOptions{WorkDir: dir, Prompter: answering(nil, true)})
	Expect(err).NotTo(HaveOccurred())
	return results
}
