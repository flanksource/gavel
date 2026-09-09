package ui

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/flanksource/commons/logger"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func captureTodoRequestWarnings(t *testing.T) *bytes.Buffer {
	t.Helper()
	warnings := &bytes.Buffer{}
	logger.SetOutput(warnings)
	t.Cleanup(func() { logger.SetOutput(nil) })
	return warnings
}

var _ = Describe("todo request decoding", func() {
	var warnings bytes.Buffer
	BeforeEach(func() {
		warnings.Reset()
		logger.SetOutput(&warnings)
		DeferCleanup(func() { logger.SetOutput(nil) })
	})

	It("warns about unknown run fields with the HTTP source and keeps declared options", func() {
		request := httptest.NewRequest(http.MethodPost, "/api/todos/run?ref=from-query",
			strings.NewReader(`{"ref":"todo-1","step":"plan","runMode":"run","futureSetting":true,"spec":{"budget":{"maxTurns":4,"futureBudget":true}}}`))
		var payload todoRunPayload
		Expect(decodeTodoRequest(request, &payload)).To(Succeed())
		Expect(payload.Ref).To(Equal("todo-1"))
		Expect(payload.Step).To(Equal("plan"))
		Expect(payload.Spec.Budget.MaxTurns).To(Equal(4))
		Expect(warnings.String()).To(SatisfyAll(
			ContainSubstring("POST /api/todos/run request body: unknown field runMode"),
			ContainSubstring("POST /api/todos/run request body: unknown field futureSetting"),
			ContainSubstring("POST /api/todos/run request body: unknown field spec.budget.futureBudget"),
		))
	})

	It("warns about nested continuation options through the same decoder", func() {
		request := httptest.NewRequest(http.MethodPost, "/api/todos/review",
			strings.NewReader(`{"options":{"step":"run","driver":"retired"}}`))
		var payload struct {
			Options *todoRunPayload `json:"options"`
		}
		Expect(decodeTodoRequest(request, &payload)).To(Succeed())
		Expect(payload.Options.Step).To(Equal("run"))
		Expect(warnings.String()).To(ContainSubstring("POST /api/todos/review request body: unknown field options.driver"))
	})

	DescribeTable("rejects malformed input",
		func(body string) {
			request := httptest.NewRequest(http.MethodPost, "/api/todos/run", strings.NewReader(body))
			var payload todoRunPayload
			Expect(decodeTodoRequest(request, &payload)).To(MatchError(ContainSubstring("invalid request body")))
			Expect(warnings.String()).To(BeEmpty())
		},
		Entry("incomplete JSON", `{"step":`),
		Entry("field type", `{"resume":"yes"}`),
		Entry("trailing object", `{} {}`),
		Entry("empty body", ``),
		Entry("null body", `null`),
		Entry("array body", `[]`),
		Entry("sandbox typo", `{"spec":{"sandbox":{"mode":"native","typo":true}}}`),
	)

	It("uses the shared decoder for run requests before resolving targets", func() {
		request := httptest.NewRequest(http.MethodPost, "/api/todos/run", strings.NewReader(`{"runMode":"plan"}`))
		_, _, _, _, status, err := (&Server{}).resolveTodoRunRequest(request)
		Expect(status).To(Equal(http.StatusBadRequest))
		Expect(err).To(MatchError("ref is required"))
		Expect(warnings.String()).To(ContainSubstring("POST /api/todos/run request body: unknown field runMode"))
	})
})
