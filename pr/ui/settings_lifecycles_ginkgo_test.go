package ui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"

	"github.com/flanksource/gavel/workflows"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("lifecycle settings API", func() {
	It("previews edits without writing and rejects invalid verification overrides", func() {
		home := GinkgoT().TempDir()
		GinkgoT().Setenv("HOME", home)
		server := &Server{}
		path := "/api/settings/lifecycles?scope=global&kind=todos"
		response := httptest.NewRecorder()
		server.handleSettingsLifecycles(response, httptest.NewRequest(http.MethodGet, path, nil))
		Expect(response.Code).To(Equal(http.StatusOK), response.Body.String())
		var view workflows.View
		Expect(json.Unmarshal(response.Body.Bytes(), &view)).To(Succeed())
		payload := `{"revision":"` + view.Revision + `","source":{"order":["run","plan"]}}`
		response = httptest.NewRecorder()
		server.handleSettingsLifecycles(response, httptest.NewRequest(http.MethodPost, path, strings.NewReader(payload)))
		Expect(response.Code).To(Equal(http.StatusOK), response.Body.String())
		_, err := os.Stat(filepath.Join(home, ".gavel.yaml"))
		Expect(os.IsNotExist(err)).To(BeTrue())
		response = httptest.NewRecorder()
		server.handleSettingsLifecycles(response, httptest.NewRequest(http.MethodPut, path, strings.NewReader(`{"revision":"`+view.Revision+`","source":{"disabled":["verify"]}}`)))
		Expect(response.Code).To(Equal(http.StatusBadRequest))
		Expect(response.Body.String()).To(ContainSubstring("cannot disable"))
	})
})
