package testui_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"

	"github.com/flanksource/clicky/entity"
	"github.com/flanksource/clicky/prompt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/fixtures/kinds/dbconn"
	"github.com/flanksource/gavel/fixtures/kinds/form"
	"github.com/flanksource/gavel/fixtures/kinds/lookup"
	testui "github.com/flanksource/gavel/testrunner/ui"
)

var _ = Describe("run prompts", func() {
	var srv *testui.Server
	var handler http.Handler
	answer := func(id, body string, header map[string]string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/api/prompts/"+id+"/answer", strings.NewReader(body))
		for k, v := range header {
			req.Header.Set(k, v)
		}
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		return rr
	}
	BeforeEach(func() {
		srv = testui.NewServer()
		handler = srv.Handler()
	})

	It("lists the run's pending form, validates the answer and records who answered", func() {
		srv.BeginRun("initial")
		owner := srv.PromptOwner()
		Expect(owner).NotTo(BeEmpty())
		answered := make(chan prompt.Answer, 1)
		go func() {
			defer GinkgoRecover()
			ans, err := srv.Prompter().Ask(context.Background(), prompt.Prompt{ID: "approval", Kind: "form", Owner: owner,
				Schema: json.RawMessage(`{"type":"object","required":["count"],"properties":{"count":{"type":"integer"}}}`)})
			Expect(err).NotTo(HaveOccurred())
			answered <- ans
		}()
		var listed []prompt.PromptSnapshot
		Eventually(func() []prompt.PromptSnapshot {
			Expect(json.Unmarshal(get(handler, "/api/prompts?owner="+owner).Body.Bytes(), &listed)).To(Succeed())
			return listed
		}).Should(HaveLen(1))
		Expect(listed[0].State).To(Equal("pending"))
		Expect(get(handler, "/api/prompts?owner=someone-else").Body.String()).To(MatchJSON(`[]`))

		Expect(answer("approval", `{"values":{"count":"three"}}`, nil).Code).To(Equal(http.StatusBadRequest))
		Expect(answer("approval", `{"values":{"count":3}}`, map[string]string{"X-Gavel-User": "ops@example.com"}).Code).To(Equal(http.StatusOK))
		Eventually(answered).Should(Receive(And(
			HaveField("Values", HaveKeyWithValue("count", BeNumerically("==", 3))),
			HaveField("By", "ops@example.com"),
		)))
		Expect(answer("approval", `{"cancelled":true}`, nil).Code).To(Equal(http.StatusConflict))
		Expect(answer("unknown", `{"cancelled":true}`, nil).Code).To(Equal(http.StatusConflict))
	})

	It("gives each run its own prompt owner", func() {
		srv.BeginRun("initial")
		first := srv.PromptOwner()
		srv.BeginRun("rerun")
		Expect(srv.PromptOwner()).NotTo(Equal(first))
	})
})

var _ = Describe("GET /api/v1/connection", func() {
	It("lists the catalog's connections for a picker, scoped by types and search, without secrets", func() {
		GinkgoT().Setenv("XDG_CONFIG_HOME", GinkgoT().TempDir())
		workDir := GinkgoT().TempDir()
		dir := filepath.Join(workDir, ".connections")
		Expect(os.MkdirAll(dir, 0o700)).To(Succeed())
		for name, spec := range map[string]string{
			"lab":       "type: sqlite\n  url:\n    value: file:/data/lab.db",
			"warehouse": "type: sql_server\n  url:\n    value: sqlserver://db.example\n  password:\n    value: s3cret",
			"reporting": "type: sql_server\n  url:\n    value: sqlserver://reports.example",
		} {
			body := "apiVersion: mission-control.flanksource.com/v1\nkind: Connection\nmetadata:\n  name: " + name + "\nspec:\n  " + spec + "\n"
			Expect(os.WriteFile(filepath.Join(dir, name+".yaml"), []byte(body), 0o600)).To(Succeed())
		}
		srv := testui.NewServer()
		srv.SetConnectionCatalog(workDir)

		response := get(srv.Handler(), "/api/v1/connection?types=sql_server&__lookup_q=WARE")

		Expect(response.Code).To(Equal(http.StatusOK), response.Body.String())
		Expect(response.Body.String()).NotTo(ContainSubstring("s3cret"))
		var decoded struct {
			Filters map[string]struct {
				Options map[string]struct{ Plain string } `json:"options"`
				Total   int                               `json:"total"`
			} `json:"filters"`
		}
		Expect(json.Unmarshal(response.Body.Bytes(), &decoded)).To(Succeed())
		Expect(decoded.Filters["connection"].Total).To(Equal(1))
		Expect(decoded.Filters["connection"].Options).To(Equal(map[string]struct{ Plain string }{
			"connection://warehouse": {Plain: "warehouse (sql_server)"},
		}))
	})

	It("decorates each option with its type icon, group and endpoint", func() {
		GinkgoT().Setenv("XDG_CONFIG_HOME", GinkgoT().TempDir())
		workDir := GinkgoT().TempDir()
		dir := filepath.Join(workDir, ".connections")
		Expect(os.MkdirAll(dir, 0o700)).To(Succeed())
		body := "apiVersion: mission-control.flanksource.com/v1\nkind: Connection\nmetadata:\n  name: sales\nspec:\n  type: postgres\n  url:\n    value: postgres://reader:s3cret@db.example.com:6543/sales\n"
		Expect(os.WriteFile(filepath.Join(dir, "sales.yaml"), []byte(body), 0o600)).To(Succeed())
		srv := testui.NewServer()
		srv.SetConnectionCatalog(workDir)

		response := get(srv.Handler(), "/api/v1/connection?types=postgres")

		Expect(response.Code).To(Equal(http.StatusOK), response.Body.String())
		Expect(response.Body.String()).NotTo(ContainSubstring("s3cret"))
		var decoded entity.LookupResponse
		Expect(json.Unmarshal(response.Body.Bytes(), &decoded)).To(Succeed())
		Expect(decoded.Filters["connection"].Options["connection://sales"]).To(Equal(entity.LookupNode{
			Kind: "text", Text: "sales (postgres)", Plain: "sales (postgres)",
			Icon: dbconn.TypeIcon("postgres"), Description: "db.example.com:6543 · sales", Group: "Postgres",
		}))
	})

	It("serves the connection form schema", func() {
		srv := testui.NewServer()
		srv.SetConnectionCatalog(GinkgoT().TempDir())
		req := httptest.NewRequest(http.MethodGet, "/api/v1/connection", nil)
		req.Header.Set("Accept", "application/schema+json")
		rr := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rr, req)
		Expect(rr.Code).To(Equal(http.StatusOK), rr.Body.String())
		var schema struct {
			Properties map[string]json.RawMessage `json:"properties"`
		}
		Expect(json.Unmarshal(rr.Body.Bytes(), &schema)).To(Succeed())
		Expect(schema.Properties).To(HaveKey("type"))
	})

	It("creates a connection in the project's .connections and answers it without secrets", func() {
		GinkgoT().Setenv("XDG_CONFIG_HOME", GinkgoT().TempDir())
		workDir := GinkgoT().TempDir()
		srv := testui.NewServer()
		srv.SetConnectionCatalog(workDir)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/connection",
			strings.NewReader(`{"name":"scratch","type":"postgres","url":"postgres://u:s3cret@localhost:5432/scratch"}`))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rr, req)

		Expect(rr.Code).To(Equal(http.StatusOK), rr.Body.String())
		Expect(rr.Body.String()).NotTo(ContainSubstring("s3cret"))
		var created struct {
			Name string `json:"name"`
			Type string `json:"type"`
		}
		Expect(json.Unmarshal(rr.Body.Bytes(), &created)).To(Succeed())
		Expect(created).To(Equal(struct {
			Name string `json:"name"`
			Type string `json:"type"`
		}{Name: "scratch", Type: "postgres"}))
		Expect(filepath.Join(workDir, ".connections", "scratch.yaml")).To(BeAnExistingFile())
	})

	It("refuses a create body that is not a JSON object", func() {
		srv := testui.NewServer()
		srv.SetConnectionCatalog(GinkgoT().TempDir())
		req := httptest.NewRequest(http.MethodPost, "/api/v1/connection", strings.NewReader(`[1]`))
		rr := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rr, req)
		Expect(rr.Code).To(Equal(http.StatusBadRequest), rr.Body.String())
	})
})

var _ = Describe("GET /api/v1/lookup/", func() {
	var handler http.Handler
	var token string
	BeforeEach(func() {
		GinkgoT().Setenv("XDG_CONFIG_HOME", GinkgoT().TempDir())
		dir := GinkgoT().TempDir()
		conns, err := dbconn.OpenConnections(context.Background(), dbconn.OpenOptions{WorkDir: dir, Set: &fixtures.ConnectionSet{Named: map[string]any{
			"lab": map[string]any{"type": "sqlite", "url": filepath.Join(dir, "lab.db")},
		}}})
		Expect(err).NotTo(HaveOccurred())
		token, err = lookup.Register("/table", lookup.QueryBinding{From: []string{"db"}, Connections: conns,
			SQL: "select 'orders' as value, 'Orders' as label, 'ph:table-thin' as icon, 'pk' as badges"})
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(lookup.Release, token)
		handler = testui.NewServer().Handler()
	})
	decode := func(response *httptest.ResponseRecorder) entity.LookupFilter {
		GinkgoHelper()
		Expect(response.Code).To(Equal(http.StatusOK), response.Body.String())
		var decoded entity.LookupResponse
		Expect(json.Unmarshal(response.Body.Bytes(), &decoded)).To(Succeed())
		Expect(decoded.Filters).To(HaveKey(form.QueryLookupFilter))
		return decoded.Filters[form.QueryLookupFilter]
	}

	It("answers a query field's options against the connection the scope names", func() {
		filter := decode(get(handler, lookup.URL(token, "/table")+"?__lookup_filter=query&__lookup_q=ord&connection=connection://lab"))
		Expect(filter).To(Equal(entity.LookupFilter{Total: 1, Options: map[string]entity.LookupNode{
			"orders": {Kind: "text", Text: "Orders", Plain: "Orders", Icon: "ph:table-thin", Badges: []entity.LookupBadge{{Label: "pk"}}},
		}}))
	})

	It("lists nothing, labelled, until the linked connection is chosen", func() {
		filter := decode(get(handler, lookup.URL(token, "/table")+"?__lookup_filter=query"))
		Expect(filter).To(Equal(entity.LookupFilter{Label: "Choose db first"}))
	})

	It("refuses an unknown token, a request without a filter and a write", func() {
		Expect(get(handler, lookup.URL("missing", "/table")+"?__lookup_filter=query").Code).To(Equal(http.StatusNotFound))
		Expect(get(handler, lookup.URL(token, "/other")+"?__lookup_filter=query").Code).To(Equal(http.StatusNotFound))
		Expect(get(handler, lookup.URL(token, "/table")).Code).To(Equal(http.StatusBadRequest))
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, lookup.URL(token, "/table"), nil))
		Expect(rr.Code).To(Equal(http.StatusMethodNotAllowed))
	})
})
