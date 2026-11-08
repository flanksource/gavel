package testui_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"

	"github.com/flanksource/clicky/prompt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

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
			"warehouse": {Plain: "warehouse (sql_server)"},
		}))
	})

	It("refuses writes", func() {
		srv := testui.NewServer()
		srv.SetConnectionCatalog(GinkgoT().TempDir())
		req := httptest.NewRequest(http.MethodPost, "/api/v1/connection", strings.NewReader(`{}`))
		rr := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rr, req)
		Expect(rr.Code).To(Equal(http.StatusMethodNotAllowed))
	})
})
