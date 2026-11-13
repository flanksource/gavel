package ui

import (
	"net/http"
	"net/http/httptest"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("todo labels batch API", func() {
	DescribeTable("rejects invalid workspace sets",
		func(body string) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, "/api/todos/labels/batch", strings.NewReader(body))
			(&Server{}).Handler().ServeHTTP(recorder, request)

			Expect(recorder.Code).To(Equal(http.StatusBadRequest), recorder.Body.String())
		},
		Entry("when dirs is missing", `{}`),
		Entry("when dirs is null", `{"dirs":null}`),
		Entry("when a dir is blank", `{"dirs":[" "]}`),
		Entry("when a dir is relative", `{"dirs":["relative/workspace"]}`),
		Entry("when a dir is not clean", `{"dirs":["/work/../workspace"]}`),
		Entry("when a dir is duplicated", `{"dirs":["/work/a","/work/a"]}`),
	)
})
