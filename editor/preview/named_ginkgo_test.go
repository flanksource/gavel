package preview

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("named Facet runtimes", func() {
	It("lists installed builds with the most recently built runtime first", func() {
		cache := GinkgoT().TempDir()
		GinkgoT().Setenv("FACET_CACHE_DIR", cache)
		Expect(ListNamedRuntimes(cache)).To(BeEmpty())
		indexes := map[string]string{
			"facet": `[{"version":"2.0.0","id":"579fe4e5-c9e1-493d-9199-96d8effc4413","builtAt":"2026-11-16T10:00:00Z"},
				{"version":"1.0.0","id":"779fe4e5-c9e1-493d-9199-96d8effc4413","builtAt":"2026-11-15T10:00:00Z"}]`,
			"reports": `[{"version":"1.0.0","id":"679fe4e5-c9e1-493d-9199-96d8effc4413","builtAt":"2026-11-17T10:00:00Z"}]`,
		}
		for name, index := range indexes {
			root := filepath.Join(cache, "named-runtimes", name)
			Expect(os.MkdirAll(root, 0700)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(root, "index.json"), []byte(index), 0600)).To(Succeed())
		}
		Expect(ListNamedRuntimes(cache)).To(Equal([]string{"reports@1.0.0", "facet@2.0.0", "facet@1.0.0"}))
	})

	It("resolves exact versions and browser packages through the Facet cache", func() {
		cache := GinkgoT().TempDir()
		GinkgoT().Setenv("FACET_CACHE_DIR", cache)
		const name, version, id = "clicky-ui", "1.2.0", "579fe4e5-c9e1-493d-9199-96d8effc4413"
		dir := filepath.Join(cache, "named-runtimes", name, version, id)
		packages := map[string]map[string]any{}
		for _, library := range []string{"@flanksource/facet", "@flanksource/clicky-ui"} {
			path := filepath.Join("node_modules", library)
			pkg := filepath.Join(dir, path)
			Expect(os.MkdirAll(filepath.Join(pkg, "dist", "runtime"), 0700)).To(Succeed())
			manifest, err := json.Marshal(map[string]string{"name": library, "version": version})
			Expect(err).NotTo(HaveOccurred())
			Expect(os.WriteFile(filepath.Join(pkg, "package.json"), manifest, 0600)).To(Succeed())
			for _, file := range []string{"importmap.json", "barrels.json"} {
				Expect(os.WriteFile(filepath.Join(pkg, "dist", "runtime", file), []byte(`{}`), 0600)).To(Succeed())
			}
			packages[library+"@"+version] = map[string]any{"name": library, "version": version, "path": filepath.ToSlash(path), "direct": true}
		}
		builtAt := time.Date(2026, 11, 15, 10, 0, 0, 0, time.UTC)
		git := BuildGit{Commit: "0123456789abcdef0123456789abcdef01234567"}
		manifest, err := json.Marshal(map[string]any{"schema": 1, "name": name, "version": version, "id": id, "browserPackages": packages, "builtAt": builtAt, "git": git})
		Expect(err).NotTo(HaveOccurred())
		Expect(os.WriteFile(filepath.Join(dir, "facet-runtime.json"), manifest, 0600)).To(Succeed())
		index := filepath.Join(cache, "named-runtimes", name, "index.json")
		Expect(os.WriteFile(index, []byte(`[{"version":"`+version+`","id":"`+id+`","builtAt":"2026-11-15T10:00:00Z"}]`), 0600)).To(Succeed())

		runtime, err := OpenNamedRuntime(GinkgoT().Context(), name+"@"+version, cache)
		Expect(err).NotTo(HaveOccurred())
		Expect(runtime.Label()).To(Equal(name + "@" + version))
		Expect(runtime.Build).To(Equal(BuildInfo{BuiltAt: &builtAt, Git: &git}))
		entries, err := os.ReadDir(filepath.Join(dir, ".leases"))
		Expect(err).NotTo(HaveOccurred())
		Expect(entries).To(HaveLen(1))
		DeferCleanup(runtime.Close)
		cacheForPreview := &Cache{Pinned: runtime}
		Expect(cacheForPreview.Package(GinkgoT().Context(), "@flanksource/facet", version)).To(Equal(filepath.Join(dir, "node_modules", "@flanksource", "facet")))
		_, err = cacheForPreview.Metadata(GinkgoT().Context(), MetadataRequest{Name: "@example/missing"})
		Expect(err).To(MatchError(ContainSubstring("absent from Facet runtime " + name + "@" + version)))
		Expect(runtime.BrowserCatalog).To(BeEmpty())
		catalog := []byte(`{"imports":{},"scopes":{},"remaps":{},"integrity":{},"versions":{"@flanksource/facet":"1.2.0"}}`)
		Expect(os.WriteFile(filepath.Join(dir, "runtime-layers.json"), append(append([]byte(`{"browserCatalog":`), catalog...), '}'), 0600)).To(Succeed())
		layered, err := OpenNamedRuntime(GinkgoT().Context(), name+"@"+version, cache)
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(layered.Close)
		Expect(layered.BrowserCatalog).To(MatchJSON(catalog))
	})
})
