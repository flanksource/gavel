package ui

import (
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("React Grab todo dialog", func() {
	It("serves the wider responsive dialog dimensions", func() {
		request := httptest.NewRequest(http.MethodGet, "http://gavel.example:9092/react-grab-plugin.js", nil)
		recorder := httptest.NewRecorder()

		handleReactGrabPlugin(recorder, request)

		Expect(recorder.Body.String()).To(And(
			ContainSubstring("width:min(960px,94vw)"),
			ContainSubstring("height:min(800px,92vh)"),
		))
	})
})
