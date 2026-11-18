package editor

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/flanksource/gavel/editor/preview"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("pinned editor preview", func() {
	It("returns the pinned build even when the project and URL request other versions", func() {
		runtime := &preview.NamedRuntime{
			Name: "clicky-ui", Version: "local",
			Versions: map[string]string{
				"@flanksource/facet":     "1.2.0",
				"@flanksource/clicky-ui": "3.4.0",
				"example-components":     "2.0.0",
			},
		}
		host := &previewHost{cache: &preview.Cache{Pinned: runtime}}
		server := &Server{previews: &previewSessions{options: PreviewOptions{Runtime: "clicky-ui"}, defaultHost: host}}
		deps := map[string]string{"@flanksource/facet": "0.0.1", "example-components": "1.0.0", "unavailable": "*"}
		Expect(host.resolve(GinkgoT().Context(), deps, "0.0.2")).To(Succeed())
		Expect(deps).To(Equal(runtime.Versions))
		response := httptest.NewRecorder()
		server.previewCatalog(response, httptest.NewRequest(http.MethodGet, "/api/preview/runtimes?refresh=1", nil))
		Expect(response.Code).To(Equal(http.StatusOK))
		var catalog map[string]string
		Expect(json.Unmarshal(response.Body.Bytes(), &catalog)).To(Succeed())
		Expect(catalog).To(Equal(map[string]string{"pinnedRuntime": "clicky-ui@local"}))
	})
})

var _ = Describe("installed editor previews", func() {
	It("lists installed builds and routes previews through the selected build's origin", func() {
		cache := GinkgoT().TempDir()
		GinkgoT().Setenv("FACET_CACHE_DIR", cache)
		const earlier, latest = "facet@2.0.0", "reports@1.0.0"
		writePreviewRuntime(cache, "facet", "2.0.0", "579fe4e5-c9e1-493d-9199-96d8effc4413", "2026-11-16T10:00:00Z")
		writePreviewRuntime(cache, "reports", "1.0.0", "679fe4e5-c9e1-493d-9199-96d8effc4413", "2026-11-17T10:00:00Z")
		page := filepath.Join(GinkgoT().TempDir(), "Report.tsx")
		Expect(os.WriteFile(page, []byte("export default () => null"), 0600)).To(Succeed())
		store, err := NewStore(StoreOptions{Paths: []string{page}})
		Expect(err).NotTo(HaveOccurred())
		server := &Server{Store: store}
		stop, err := server.StartPreview(GinkgoT().Context(), PreviewOptions{HostOrigin: "http://127.0.0.1:9000", CacheRoot: cache, Directory: cache})
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(stop)
		handler := server.Handler()
		catalog := httptest.NewRecorder()
		handler.ServeHTTP(catalog, httptest.NewRequest(http.MethodGet, "/api/preview/runtimes", nil))
		Expect(catalog.Code).To(Equal(http.StatusOK))
		Expect(catalog.Body.String()).To(MatchJSON(`{"installed":["reports@1.0.0","facet@2.0.0"]}`))
		endpoint := "/api/file/" + store.Files()[0].ID + "/preview"
		requestPreview := func(query string) PreviewContext {
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, endpoint+query, nil))
			Expect(response.Code).To(Equal(http.StatusOK), response.Body.String())
			Expect(response.Header().Get("Server-Timing")).To(MatchRegexp(`^total;dur=\d+\.\d, source-index;dur=\d+\.\d, runtime;dur=\d+\.\d, dependencies;dur=\d+\.\d$`))
			var context PreviewContext
			Expect(json.Unmarshal(response.Body.Bytes(), &context)).To(Succeed())
			return context
		}
		defaultPreview := requestPreview("?facet=0.0.1")
		selected := requestPreview("?runtime=" + url.QueryEscape(earlier))
		Expect(defaultPreview.Runtime).To(Equal(latest))
		Expect(defaultPreview.RuntimeBuild).NotTo(BeNil())
		Expect(defaultPreview.RuntimeBuild.BuiltAt).To(HaveValue(Equal(time.Date(2026, 11, 17, 10, 0, 0, 0, time.UTC))))
		Expect(defaultPreview.RuntimeBuild.Git).To(HaveValue(Equal(preview.BuildGit{Commit: "0123456789abcdef0123456789abcdef01234567", Dirty: true})))
		Expect(defaultPreview.Dependencies["@flanksource/facet"]).To(Equal("1.0.0"))
		Expect(selected.Runtime).To(Equal(earlier))
		Expect(selected.RuntimeBuild).NotTo(BeNil())
		Expect(selected.RuntimeBuild.BuiltAt).To(HaveValue(Equal(time.Date(2026, 11, 16, 10, 0, 0, 0, time.UTC))))
		Expect(selected.RuntimeBuild.Git).To(BeNil())
		Expect(selected.Dependencies["@flanksource/facet"]).To(Equal("2.0.0"))
		Expect(selected.RuntimeOrigin).NotTo(Equal(defaultPreview.RuntimeOrigin))
		Expect(requestPreview("")).To(Equal(defaultPreview))
		Expect(requestPreview("?runtime=" + url.QueryEscape(earlier))).To(Equal(selected))
		client := &http.Client{Timeout: 5 * time.Second}
		for _, runtime := range []PreviewContext{defaultPreview, selected} {
			response, err := client.Get(runtime.RuntimeOrigin + "/profile.json")
			Expect(err).NotTo(HaveOccurred())
			var profile struct {
				Managed map[string]string `json:"managed"`
			}
			decodeErr := json.NewDecoder(response.Body).Decode(&profile)
			Expect(response.Body.Close()).To(Succeed())
			Expect(response.StatusCode).To(Equal(http.StatusOK))
			Expect(decodeErr).NotTo(HaveOccurred())
			Expect(profile.Managed).To(Equal(runtime.Dependencies))
		}
		Expect(stop()).To(Succeed())
		for _, build := range []string{"facet/2.0.0/579fe4e5-c9e1-493d-9199-96d8effc4413", "reports/1.0.0/679fe4e5-c9e1-493d-9199-96d8effc4413"} {
			Expect(os.ReadDir(filepath.Join(cache, "named-runtimes", build, ".leases"))).To(BeEmpty())
		}
	})
})

func writePreviewRuntime(cache, name, version, id, builtAt string) {
	root := filepath.Join(cache, "named-runtimes", name)
	build := filepath.Join(root, version, id)
	writeJSON := func(path string, value any) {
		data, err := json.Marshal(value)
		Expect(err).NotTo(HaveOccurred())
		Expect(os.WriteFile(path, data, 0600)).To(Succeed())
	}
	packages := map[string]any{}
	for _, name := range []string{"@flanksource/facet", "@flanksource/clicky-ui"} {
		path := filepath.Join("node_modules", name)
		pkg := filepath.Join(build, path)
		dist := filepath.Join(pkg, "dist", "runtime")
		Expect(os.MkdirAll(dist, 0700)).To(Succeed())
		writeJSON(filepath.Join(pkg, "package.json"), map[string]string{"name": name, "version": version})
		for _, file := range []string{"importmap.json", "barrels.json"} {
			writeJSON(filepath.Join(dist, file), map[string]string{})
		}
		packages[name+"@"+version] = map[string]any{"name": name, "version": version, "path": filepath.ToSlash(path), "direct": true}
	}
	manifest := map[string]any{"name": name, "version": version, "id": id, "schema": 1, "browserPackages": packages, "builtAt": builtAt}
	if name == "reports" {
		manifest["git"] = map[string]any{"commit": "0123456789abcdef0123456789abcdef01234567", "dirty": true}
	}
	writeJSON(filepath.Join(build, "facet-runtime.json"), manifest)
	writeJSON(filepath.Join(root, "index.json"), []map[string]string{{"version": version, "id": id, "builtAt": builtAt}})
}
