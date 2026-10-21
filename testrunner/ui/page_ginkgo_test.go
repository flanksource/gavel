package testui_test

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	testui "github.com/flanksource/gavel/testrunner/ui"
)

var _ = Describe("the test UI page", func() {
	It("inlines its bundle so no sequence in it ends or escapes the script element", func() {
		request := httptest.NewRequest(http.MethodGet, "/", nil)
		request.Header.Set("Accept", "text/html")
		response := httptest.NewRecorder()
		testui.NewServer().Handler().ServeHTTP(response, request)

		page := response.Body.String()
		opening := `<div id="root"></div><script >`
		start := strings.Index(page, opening)
		end := strings.LastIndex(page, "</script>")
		Expect(start).To(BeNumerically(">=", 0), "the page inlines its bundle after the root")
		script := page[start+len(opening) : end]
		Expect(len(script)).To(BeNumerically(">", 1000))
		Expect(regexp.MustCompile(`(?i)<(!--|/?script)`).FindString(script)).To(BeEmpty(),
			"an HTML comment opener or a script tag inside the inline script changes where the browser ends it")
		Expect(testui.BundleJS()).To(Equal(script), "hosts that inline the bundle get the same safe form")
	})
})
