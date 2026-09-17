package ui

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("dashboard project scope", func() {
	var originalPath string

	BeforeEach(func() {
		originalPath = settingsPath
		settingsPath = filepath.Join(GinkgoT().TempDir(), "ui.settings.json")
	})

	AfterEach(func() {
		settingsPath = originalPath
	})

	It("round-trips the selected project through the config endpoint and settings", func() {
		server := &Server{}
		request := httptest.NewRequest("POST", "/api/config", strings.NewReader(`{"project":"gavel","repos":["acme/gavel"]}`))
		response := httptest.NewRecorder()

		server.handleConfig(response, request)

		Expect(response.Code).To(Equal(http.StatusOK))
		Expect(server.GetConfig()).To(Equal(SearchConfig{Project: "gavel", Repos: []string{"acme/gavel"}}))
		Eventually(LoadSettings).Should(Equal(UISettings{Project: "gavel", Repos: []string{"acme/gavel"}}))
	})
})
