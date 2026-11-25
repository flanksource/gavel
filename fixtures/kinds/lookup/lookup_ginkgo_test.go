package lookup_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/flanksource/clicky/schemaform"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/fixtures/kinds/dbconn"
	"github.com/flanksource/gavel/fixtures/kinds/lookup"
)

func TestLookup(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "lookup Suite")
}

// tables are the rows every spec's database holds.
var tables = [][]any{
	{"orders", "Orders", "postgres", `["pk","fk"]`, "success"},
	{"order_items", "Order items", "ph:table-thin", "pk, wide", ""},
	{"customers", "Customers", "", `[{"label":"pii","icon":"ph:lock-thin","tone":"danger"}]`, "warning"},
}

// openConnections is a fixture declaring a sqlite `lab` of tables (plus
// count extra rows) and a postgres `warehouse` whose URL carries a password.
func openConnections(extra int) *dbconn.Connections {
	GinkgoHelper()
	GinkgoT().Setenv("XDG_CONFIG_HOME", GinkgoT().TempDir())
	dir := GinkgoT().TempDir()
	path := filepath.Join(dir, "lab.db")
	db, err := sql.Open("sqlite", path)
	Expect(err).NotTo(HaveOccurred())
	defer db.Close()
	_, err = db.Exec("CREATE TABLE tables(name TEXT, title TEXT, kind TEXT, tags TEXT, tone TEXT)")
	Expect(err).NotTo(HaveOccurred())
	for _, row := range tables {
		_, err = db.Exec("INSERT INTO tables VALUES (?, ?, ?, ?, ?)", row...)
		Expect(err).NotTo(HaveOccurred())
	}
	for i := range extra {
		_, err = db.Exec("INSERT INTO tables(name) VALUES (?)", fmt.Sprintf("extra_%03d", i))
		Expect(err).NotTo(HaveOccurred())
	}
	conns, err := dbconn.OpenConnections(context.Background(), dbconn.OpenOptions{WorkDir: dir, Set: &fixtures.ConnectionSet{Named: map[string]any{
		"lab":       map[string]any{"type": "sqlite", "url": path},
		"warehouse": map[string]any{"type": "postgres", "url": "postgres://reader:s3cret@db.example.com:6543/sales"},
	}}})
	Expect(err).NotTo(HaveOccurred())
	return conns
}

func resolve(binding lookup.Binding, req lookup.Request) (lookup.Result, error) {
	GinkgoHelper()
	token, err := lookup.Register("/table", binding)
	Expect(err).NotTo(HaveOccurred())
	DeferCleanup(lookup.Release, token)
	req.Token, req.Pointer = token, "/table"
	return lookup.Resolve(context.Background(), req)
}

var _ = Describe("RowOptions", func() {
	DescribeTable("maps a row's columns onto an option",
		func(row map[string]any, columns map[string]string, want schemaform.Option) {
			Expect(lookup.RowOptions([]map[string]any{row}, columns)).To(Equal([]schemaform.Option{want}))
		},
		Entry("value and label", map[string]any{"value": "orders", "label": "Orders", "other": 1}, nil,
			schemaform.Option{Value: "orders", Label: "Orders"}),
		Entry("a lone column is value and label", map[string]any{"name": "orders"}, nil,
			schemaform.Option{Value: "orders", Label: "orders"}),
		Entry("the one column no role reads is the value", map[string]any{"name": "orders", "label": "Orders"}, nil,
			schemaform.Option{Value: "orders", Label: "Orders"}),
		Entry("a numeric value", map[string]any{"value": int64(7)}, nil, schemaform.Option{Value: "7", Label: "7"}),
		Entry("a comma list of badges", map[string]any{"value": "a", "badges": "pk, wide,"}, nil,
			schemaform.Option{Value: "a", Label: "a", Badges: []schemaform.Badge{{Label: "pk"}, {Label: "wide"}}}),
		Entry("a JSON array of badge labels", map[string]any{"value": "a", "badges": `["pk"]`}, nil,
			schemaform.Option{Value: "a", Label: "a", Badges: []schemaform.Badge{{Label: "pk"}}}),
		Entry("a driver's decoded badge objects", map[string]any{"value": "a", "badges": []any{map[string]any{"label": "pii", "tone": "danger"}}}, nil,
			schemaform.Option{Value: "a", Label: "a", Badges: []schemaform.Badge{{Label: "pii", Tone: "danger"}}}),
		Entry("columns renamed by x-query.columns", map[string]any{"id": "o", "name": "Orders", "kind": "postgres", "tags": "pk", "state": "success", "note": "rows"},
			map[string]string{"value": "id", "label": "name", "icon": "kind", "badges": "tags", "tone": "state", "description": "note"},
			schemaform.Option{Value: "o", Label: "Orders", Icon: "postgres", Badges: []schemaform.Badge{{Label: "pk"}}, Tone: "success", Description: "rows"}),
	)

	DescribeTable("refuses a row it cannot map, naming the row",
		func(row map[string]any, columns map[string]string, message string) {
			_, err := lookup.RowOptions([]map[string]any{{"value": "ok", "title": "OK"}, row}, columns)
			Expect(err).To(MatchError(ContainSubstring("lookup row 1: " + message)))
		},
		Entry("no columns", map[string]any{}, nil, "the query returned no columns"),
		Entry("no value column among several", map[string]any{"a": 1, "b": 2}, nil, "no value column among a, b"),
		Entry("a null value", map[string]any{"value": nil}, nil, "value is null"),
		Entry("malformed JSON badges", map[string]any{"value": "a", "badges": `["pk"`}, nil, "badges: JSON array"),
		Entry("a badge object without a label", map[string]any{"value": "a", "badges": `[{"icon":"x"}]`}, nil, "badges: item 0: a badge object needs a label"),
		Entry("a non-text icon", map[string]any{"value": "a", "icon": []any{"x"}}, nil, "icon: want text"),
		Entry("a renamed column the query lacks", map[string]any{"value": "a"}, map[string]string{"label": "title"}, `x-query.columns.label names column "title"`),
	)
})

var _ = Describe("Resolve", func() {
	It("reads a query's options on a fixed connection, filtered by @q, with icon, badges and tone", func() {
		conns := openConnections(0)
		lab, err := conns.Named("lab")
		Expect(err).NotTo(HaveOccurred())
		result, err := resolve(lookup.QueryBinding{Connection: lab, SQL: "select name as value, title as label, kind as icon, tags as badges, tone " +
			"from tables where name like '%' || @q || '%' order by name"}, lookup.Request{Query: "order"})
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal(lookup.Result{LookupResult: schemaform.LookupResult{Total: 2, Options: []schemaform.Option{
			{Value: "order_items", Label: "Order items", Icon: "ph:table-thin", Badges: []schemaform.Badge{{Label: "pk"}, {Label: "wide"}}},
			{Value: "orders", Label: "Orders", Icon: "postgres", Tone: "success", Badges: []schemaform.Badge{{Label: "pk"}, {Label: "fk"}}},
		}}}))
	})

	It("caps the options at clicky's lookup ceiling and says so", func() {
		lab, err := openConnections(lookup.MaxOptions).Named("lab")
		Expect(err).NotTo(HaveOccurred())
		result, err := resolve(lookup.QueryBinding{Connection: lab, SQL: "select name as value from tables"}, lookup.Request{})
		Expect(err).NotTo(HaveOccurred())
		Expect(result.Options).To(HaveLen(lookup.MaxOptions))
		Expect(result.Truncated).To(BeTrue())
	})

	It("lists nothing, labelled, until the linked connection param is chosen, then queries it", func() {
		binding := lookup.QueryBinding{From: []string{"db"}, Connections: openConnections(0), SQL: "select name as value from tables where name = @q"}
		empty, err := resolve(binding, lookup.Request{Values: map[string]any{"db": ""}})
		Expect(err).NotTo(HaveOccurred())
		Expect(empty).To(Equal(lookup.Result{Label: "Choose db first"}))

		fromValues, err := resolve(binding, lookup.Request{Query: "customers", Values: map[string]any{"db": "connection://lab"}})
		Expect(err).NotTo(HaveOccurred())
		Expect(fromValues.Options).To(Equal([]schemaform.Option{{Value: "customers", Label: "customers"}}))

		fromScope, err := resolve(binding, lookup.Request{Query: "orders", Scope: map[string]any{"connection": "lab"}})
		Expect(err).NotTo(HaveOccurred())
		Expect(fromScope.Options).To(Equal([]schemaform.Option{{Value: "orders", Label: "orders"}}))
	})

	It("refuses a released token and a pointer it was not registered for", func() {
		lab, err := openConnections(0).Named("lab")
		Expect(err).NotTo(HaveOccurred())
		token, err := lookup.Register("/table", lookup.QueryBinding{Connection: lab, SQL: "select 1 as value"})
		Expect(err).NotTo(HaveOccurred())
		_, err = lookup.Resolve(context.Background(), lookup.Request{Token: token, Pointer: "/other"})
		Expect(err).To(MatchError(ContainSubstring("is for field /table, not /other")))
		lookup.Release(token)
		_, err = lookup.Resolve(context.Background(), lookup.Request{Token: token, Pointer: "/table"})
		Expect(err).To(MatchError(ContainSubstring("unknown or its form step has ended")))
	})

	It("lists the fixture's named connections with type icon, endpoint, group and no credentials", func() {
		result, err := resolve(lookup.ConnectionBinding{Types: []string{"postgres"}, Connections: openConnections(0), WorkDir: GinkgoT().TempDir()}, lookup.Request{})
		Expect(err).NotTo(HaveOccurred())
		Expect(result.Options).To(Equal([]schemaform.Option{{
			Value: "connection://warehouse", Label: "warehouse (postgres)", Icon: "postgres",
			Description: "db.example.com:6543 · sales", Group: "Postgres",
		}}))
		wire, err := json.Marshal(result)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(wire)).NotTo(ContainSubstring("s3cret"))
	})
})

var _ = Describe("lookup URLs", func() {
	It("round-trips a token and an escaped field pointer", func() {
		pointer := lookup.Pointer([]string{"source", "a/b c"})
		token, parsed, ok, err := lookup.ParseURL(lookup.URL("tok", pointer) + "?__lookup_q=x")
		Expect(err).NotTo(HaveOccurred())
		Expect([]any{token, parsed, ok}).To(Equal([]any{"tok", "/source/a~1b c", true}))
		_, _, ok, err = lookup.ParseURL("/api/v1/connection?types=postgres")
		Expect([]any{ok, err}).To(Equal([]any{false, nil}))
	})
})
