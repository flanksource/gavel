package testui_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/flanksource/gavel/linters"
	testui "github.com/flanksource/gavel/testrunner/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Server.SetLintRerun", func() {
	const (
		workDir  = "/rules"
		embedded = "oipa-transaction"
	)
	var (
		srv   *testui.Server
		calls [][]string
	)
	rerun := func(body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/api/rerun", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rr, req)
		return rr
	}
	BeforeEach(func() {
		calls = nil
		srv = testui.NewServer()
		srv.SetLintResults([]*linters.LinterResult{
			{Linter: "typescript", WorkDir: workDir, Success: true},
			{Linter: embedded, WorkDir: workDir, Success: false, Error: "stale"},
		})
		srv.SetLintRerun([]string{embedded}, func(_ context.Context, names, files []string) ([]*linters.LinterResult, error) {
			calls = append(calls, append(append([]string(nil), names...), files...))
			return []*linters.LinterResult{{Linter: embedded, WorkDir: workDir, Success: true}}, nil
		})
	})

	It("runs a rerun of the linters it claims and upserts the results, keeping the others", func() {
		response := rerun(`{"lint":true,"lint_linters":["oipa-transaction"],"lint_files":["tr/Fee"]}`)

		Expect(response.Code).To(Equal(http.StatusAccepted), response.Body.String())
		Expect(calls).To(Equal([][]string{{embedded, "tr/Fee"}}))
		Expect(srv.Snapshot().Lint).To(Equal([]*linters.LinterResult{
			{Linter: "typescript", WorkDir: workDir, Success: true},
			{Linter: embedded, WorkDir: workDir, Success: true},
		}))
	})

	It("refuses a rerun mixing claimed and unclaimed linters", func() {
		response := rerun(`{"lint":true,"lint_linters":["oipa-transaction","typescript"]}`)

		Expect(response.Code).To(Equal(http.StatusBadRequest))
		Expect(response.Body.String()).To(ContainSubstring("typescript"))
		Expect(calls).To(BeEmpty())
	})
})
