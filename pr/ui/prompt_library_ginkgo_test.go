package ui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func libraryDetailCall(method, id, query, body string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	(&Server{}).Handler().ServeHTTP(rec, httptest.NewRequest(method, "/api/settings/prompts/"+id+"?"+query, strings.NewReader(body)))
	return rec
}

var _ = Describe("discovered prompt library", func() {
	var dir, scope string
	BeforeEach(func() {
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		dir = newProjectDir(catalogProject, "acme/"+catalogProject)
		scope = "project=" + catalogProject
	})

	It("lists library prompts and discovers new workspace templates without registration", func() {
		root := filepath.Join(dir, ".captain", "prompts")
		Expect(os.MkdirAll(root, 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(root, "custom-review.prompt"), []byte("Review {{{change}}}"), 0o600)).To(Succeed())
		entries := catalogCall(scope)
		Expect(catalogEntry(entries, "library.branch-merge").Raw).To(ContainSubstring("Rebase"))
		Expect(catalogEntry(entries, "library.branch-rebase").ParseError).To(BeEmpty())
		Expect(catalogEntry(entries, "library.pr-rebase").Variables).To(ContainElement("branch"))
		Expect(catalogEntry(entries, "library.custom-review").Body).To(Equal("Review {{{change}}}"))
		global := catalogCall("scope=global")
		Expect(global).NotTo(ContainElement(HaveField("ID", "library.custom-review")))
	})

	It("saves an override through the shared editor and uses it for preview and discovery", func() {
		entry := catalogEntry(catalogCall(scope), "library.branch-merge")
		layer := entry.Layers[len(entry.Layers)-1]
		Expect(layer.Editable).To(BeTrue())
		loaded := libraryDetailCall(http.MethodGet, entry.ID, layer.Scope, "")
		Expect(loaded.Code).To(Equal(http.StatusOK), loaded.Body.String())
		detail := decodePromptDetail(loaded)
		body := "Resolve conflicts for {{{branch}}}"
		spec := *detail.Spec
		write := promptDetailRequest{Source: "file", Path: detail.Path, Spec: &spec, Body: &body, BaseRaw: &detail.Raw}
		saved := libraryDetailCall(http.MethodPut, entry.ID, layer.Scope, mustJSON(write))
		Expect(saved.Code).To(Equal(http.StatusOK), saved.Body.String())
		Expect(decodePromptDetail(saved).Raw).To(ContainSubstring(body))
		Expect(decodePromptDetail(saved).Raw).To(ContainSubstring("workflow:"))
		entry = catalogEntry(catalogCall(scope), entry.ID)
		Expect(entry.Source).To(Equal("file"))
		Expect(entry.Body).To(Equal(body))
		preview := renderCall(entry.ID, scope, `{"variables":{"branch":"topic"}}`)
		Expect(preview.Code).To(Equal(http.StatusOK), preview.Body.String())
		var rendered promptRenderResponse
		Expect(json.Unmarshal(preview.Body.Bytes(), &rendered)).To(Succeed())
		Expect(rendered.User).To(Equal("Resolve conflicts for topic"))

		stale := libraryDetailCall(http.MethodPut, entry.ID, layer.Scope, mustJSON(write))
		Expect(stale.Code).To(Equal(http.StatusConflict), stale.Body.String())
		Expect(libraryDetailCall(http.MethodPut, entry.ID, layer.Scope, mustJSON(promptDetailRequest{Source: "default", BaseRaw: &entry.Raw})).Code).To(Equal(http.StatusOK))
		Expect(catalogEntry(catalogCall(scope), entry.ID).Source).To(Equal("builtin"))
	})

	It("removes a workspace template with no built-in default", func() {
		root := filepath.Join(dir, ".captain", "prompts")
		Expect(os.MkdirAll(root, 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(root, "custom-review.prompt"), []byte("Review the change."), 0o600)).To(Succeed())
		entry := catalogEntry(catalogCall(scope), "library.custom-review")
		layer := entry.Layers[len(entry.Layers)-1]
		request := promptDetailRequest{Source: "default", BaseRaw: &entry.Raw}
		response := libraryDetailCall(http.MethodPut, entry.ID, layer.Scope, mustJSON(request))
		Expect(response.Code).To(Equal(http.StatusOK), response.Body.String())
		Expect(catalogCall(scope)).NotTo(ContainElement(HaveField("ID", entry.ID)))
	})

	It("refuses inline saves and paths outside the discovered prompt", func() {
		entry := catalogEntry(catalogCall(scope), "library.branch-merge")
		layer := entry.Layers[len(entry.Layers)-1]
		raw := "Resolve {{{branch}}}"
		request := promptDetailRequest{Source: "inline", Raw: &raw, BaseRaw: &entry.Raw}
		Expect(libraryDetailCall(http.MethodPut, entry.ID, layer.Scope, mustJSON(request)).Code).To(Equal(http.StatusBadRequest))
		request.Source, request.Path = "file", "../outside.prompt"
		Expect(libraryDetailCall(http.MethodPut, entry.ID, layer.Scope, mustJSON(request)).Code).To(Equal(http.StatusBadRequest))
	})

	It("retains malformed templates for repair and refuses invalid writes", func() {
		root := filepath.Join(dir, ".captain", "prompts")
		Expect(os.MkdirAll(root, 0o755)).To(Succeed())
		invalid := "---\nmodel: [broken\n---\nrepair me"
		Expect(os.WriteFile(filepath.Join(root, "broken.prompt"), []byte(invalid), 0o600)).To(Succeed())
		entry := catalogEntry(catalogCall(scope), "library.broken")
		Expect(entry.ParseError).NotTo(BeEmpty())
		layer := entry.Layers[len(entry.Layers)-1]
		loaded := libraryDetailCall(http.MethodGet, entry.ID, layer.Scope, "")
		Expect(loaded.Code).To(Equal(http.StatusOK), loaded.Body.String())
		Expect(decodePromptDetail(loaded).Raw).To(Equal(invalid))
		write := promptDetailRequest{Source: "file", Path: "broken.prompt", Raw: &invalid, BaseRaw: &invalid}
		Expect(libraryDetailCall(http.MethodPut, entry.ID, layer.Scope, mustJSON(write)).Code).To(Equal(http.StatusBadRequest))
		fixed := "Repaired {{{input}}}"
		write.Raw = &fixed
		Expect(libraryDetailCall(http.MethodPut, entry.ID, layer.Scope, mustJSON(write)).Code).To(Equal(http.StatusOK))
		Expect(catalogEntry(catalogCall(scope), entry.ID).ParseError).To(BeEmpty())
	})
})
