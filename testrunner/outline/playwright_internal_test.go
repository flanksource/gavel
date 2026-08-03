package outline

import (
	. "github.com/onsi/gomega"

	"github.com/flanksource/gavel/testrunner/parsers"
)

var _ = Describe("Playwright list report", func() {
	It("returns listed tests and visible collection errors", func() {
		data := []byte(`{
  "suites": [{
    "title": "login.spec.ts",
    "file": "tests/login.spec.ts",
    "line": 1,
    "suites": [{
      "title": "login",
      "file": "tests/login.spec.ts",
      "line": 3,
      "specs": [{
        "title": "accepts valid credentials",
        "file": "tests/login.spec.ts",
        "line": 7,
        "tests": [{"projectName": "chromium", "results": []}]
      }]
    }]
  }],
  "errors": [{"message": "configuration warning"}]
}`)

		entries, errors, err := parsePlaywrightList(data, "/repo/site", "site", "/repo")
		Expect(err).NotTo(HaveOccurred())
		Expect(entries).To(HaveLen(1))
		Expect(entries[0].Framework).To(Equal(parsers.Playwright))
		Expect(entries[0].File).To(Equal("site/tests/login.spec.ts"))
		Expect(entries[0].Line).To(Equal(7))
		Expect(entries[0].Name).To(Equal("accepts valid credentials"))
		Expect(errors).To(Equal([]string{"configuration warning"}))
	})
})
