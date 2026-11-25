package testui_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"

	"github.com/flanksource/gavel/fixtures"
	testui "github.com/flanksource/gavel/testrunner/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("GET /api/rerun/command", func() {
	const (
		fixture  = "checks/purge it.fixture.yaml"
		fromStep = "summary"
		runID    = "run-2026-11-22T13-22-21Z-2"
	)
	base := &fixtures.RerunBase{Cwd: "/work/my repo", Argv: []string{"gavel", "fixtures", "--ui=true"}}

	get := func(server *testui.Server, query url.Values) *httptest.ResponseRecorder {
		recorder := httptest.NewRecorder()
		server.Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/rerun/command?"+query.Encode(), nil))
		return recorder
	}
	query := url.Values{"fixture": {fixture}, "from_step": {fromStep}}

	It("names the saved run and the step, quoted for a shell in the run's directory", func() {
		server := testui.NewServer()
		server.SetRerunBase(base)
		server.SetSavedRun(runID)

		response := get(server, query)

		Expect(response.Code).To(Equal(http.StatusOK), response.Body.String())
		var command testui.RerunCommand
		Expect(json.Unmarshal(response.Body.Bytes(), &command)).To(Succeed())
		line := "gavel fixtures --ui=true --from-run=" + runID + " --from-step=summary 'checks/purge it.fixture.yaml'"
		Expect(command).To(Equal(testui.RerunCommand{
			Cwd: base.Cwd, Line: line, Shell: "cd '/work/my repo' && " + line,
			Argv: []string{"gavel", "fixtures", "--ui=true", "--from-run=" + runID, "--from-step=summary", fixture},
		}))
		Expect(base.Argv).To(HaveLen(3))
	})

	DescribeTable("has none", func(prepare func(*testui.Server), query url.Values, status int, reason string) {
		server := testui.NewServer()
		prepare(server)

		response := get(server, query)

		Expect(response.Code).To(Equal(status))
		Expect(strings.TrimSpace(response.Body.String())).To(ContainSubstring(reason))
	},
		Entry("before the run is saved", func(s *testui.Server) { s.SetRerunBase(base) }, query, http.StatusConflict, "not saved yet"),
		Entry("once a rerun starts", func(s *testui.Server) { s.SetRerunBase(base); s.SetSavedRun(runID); s.BeginRun("rerun") }, query, http.StatusConflict, "not saved yet"),
		Entry("for a run without fixtures", func(*testui.Server) {}, query, http.StatusNotImplemented, "no rerun command"),
		Entry("without a step", func(s *testui.Server) { s.SetRerunBase(base) }, url.Values{"fixture": {fixture}}, http.StatusBadRequest, "required"),
	)
})
