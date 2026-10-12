package fixturebenchtest

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"

	"github.com/flanksource/gavel/testrunner/parsers"
	testui "github.com/flanksource/gavel/testrunner/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("fixture Go profile downloads", func() {
	It("serves only a captured artifact from the configured profile root", func() {
		root := GinkgoT().TempDir()
		profile := filepath.Join(root, ".gavel", "profiles", "fixtures", "run-a", "cpu.pprof")
		Expect(os.MkdirAll(filepath.Dir(profile), 0o755)).To(Succeed())
		Expect(os.WriteFile(profile, []byte("profile-data"), 0o600)).To(Succeed())
		server := testui.NewServer()
		server.SetProfileRoot(root)
		server.SetResults([]parsers.Test{{Name: "fixture", GoProfiles: []parsers.GoProfileArtifact{{Name: "cpu", ID: "run-a/cpu.pprof", Path: profile, Status: "captured"}}}})

		response := httptest.NewRecorder()
		server.Handler().ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/tests/fixture-profile?id=run-a%2Fcpu.pprof", nil))
		Expect(response.Code).To(Equal(http.StatusOK))
		Expect(response.Body.String()).To(Equal("profile-data"))

		response = httptest.NewRecorder()
		server.Handler().ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/tests/fixture-profile?id=..%2Fsecret", nil))
		Expect(response.Code).To(Equal(http.StatusBadRequest))
	})
})
