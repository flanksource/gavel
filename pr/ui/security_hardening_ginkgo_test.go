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

	"github.com/flanksource/gavel/verify"
)

// These specs cover the CodeQL findings on the pr/ui HTTP surface: reflected
// XSS through the client-controlled Host header, unsafe quoting of the build
// metadata script, and path traversal through the prompt-override and test-run
// endpoints.
var _ = Describe("pr/ui security hardening", func() {
	Describe("react grab install page", func() {
		// The Host header and X-Forwarded-Proto are entirely client-controlled and
		// are reflected into the served page, so a payload must never survive as
		// executable markup or as a JS string-literal break-out.
		hostPayloads := map[string]string{
			"script tag":     `a"><script>alert(1)</script>`,
			"quote breakout": `a'+alert(1)+'b`,
			"attribute exit": `a" onload="alert(1)`,
		}

		for name, payload := range hostPayloads {
			It("neutralizes a "+name+" injected via the Host header", func() {
				req := httptest.NewRequest(http.MethodGet, "/react-grab-install", nil)
				req.Host = payload
				rec := httptest.NewRecorder()

				handleReactGrabInstall(rec, req)

				body := rec.Body.String()
				Expect(rec.Code).To(Equal(http.StatusOK))
				Expect(body).ToNot(ContainSubstring(payload))
				Expect(body).ToNot(ContainSubstring("<script>alert(1)</script>"))
				Expect(body).ToNot(ContainSubstring(`onload="alert(1)`))
				Expect(rec.Header().Get("X-Content-Type-Options")).To(Equal("nosniff"))
			})

			It("neutralizes a "+name+" injected into the plugin script", func() {
				req := httptest.NewRequest(http.MethodGet, "/react-grab-plugin.js", nil)
				req.Host = payload
				rec := httptest.NewRecorder()

				handleReactGrabPlugin(rec, req)

				body := rec.Body.String()
				Expect(rec.Code).To(Equal(http.StatusOK))
				Expect(body).ToNot(ContainSubstring(payload))
				// The placeholder sits inside a JS string literal; no raw quote may
				// reach it.
				Expect(body).To(ContainSubstring(`var GAVEL_ORIGIN = "http://`))
			})
		}

		It("passes an ordinary origin through unchanged", func() {
			req := httptest.NewRequest(http.MethodGet, "/react-grab-install", nil)
			req.Host = "localhost:9092"
			rec := httptest.NewRecorder()

			handleReactGrabInstall(rec, req)

			Expect(rec.Body.String()).To(ContainSubstring("http://localhost:9092/react-grab-plugin.js"))
			Expect(rec.Body.String()).ToNot(ContainSubstring("__GAVEL_ORIGIN__"))
		})
	})

	Describe("dashboard build-info script", func() {
		// A build stamped with a quote must not be able to terminate the JSON
		// string or the <script> element it is embedded in.
		It("escapes quotes and script terminators in the build metadata", func() {
			saved := Build
			DeferCleanup(func() { Build = saved })
			Build = BuildInfo{
				Version: `1.0"</script><script>alert(1)</script>`,
				Commit:  `de"ad`,
				Date:    "2026-01-01",
			}

			html := pageHTML()

			Expect(html).ToNot(ContainSubstring("<script>alert(1)</script>"))
			Expect(html).ToNot(ContainSubstring(`1.0"</script>`))
			Expect(strings.Count(html, "</script>")).To(Equal(2), "one per legitimate script element")

			raw := html[strings.Index(html, "window.__GAVEL__=")+len("window.__GAVEL__="):]
			raw = raw[:strings.Index(raw, ";</script>")]
			var decoded BuildInfo
			Expect(json.Unmarshal([]byte(raw), &decoded)).To(Succeed())
			Expect(decoded.Version).To(Equal(Build.Version), "the value must round-trip, not be mangled")
			Expect(decoded.Commit).To(Equal(Build.Commit))
		})
	})

	Describe("prompt override file writes", func() {
		// The config layer sits one level below the spec's temp root so a `..`
		// escape lands somewhere the spec fully owns and can assert on.
		newConfigDir := func() (base, dir string) {
			base = GinkgoT().TempDir()
			dir = filepath.Join(base, "workspace")
			Expect(os.MkdirAll(dir, 0o755)).To(Succeed())
			return base, dir
		}

		escapes := []string{
			"../escaped.prompt",
			"nested/../../escaped.prompt",
			"./../escaped.prompt",
		}

		for _, path := range escapes {
			It("rejects the traversing path "+path, func() {
				base, dir := newConfigDir()
				outside := filepath.Join(base, "escaped.prompt")

				cfg := verify.GavelConfig{}
				err := persistPromptOverride(&cfg, &verify.PromptSpec{}, promptWrite{
					Dir: dir, Source: "file", Path: path, Text: "body",
				})

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(path))
				_, statErr := os.Stat(outside)
				Expect(os.IsNotExist(statErr)).To(BeTrue(), "no file may be written outside the config dir")
			})
		}

		It("rejects an absolute path", func() {
			base, dir := newConfigDir()
			target := filepath.Join(base, "abs.prompt")

			cfg := verify.GavelConfig{}
			err := persistPromptOverride(&cfg, &verify.PromptSpec{}, promptWrite{
				Dir: dir, Source: "file", Path: target, Text: "body",
			})

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(target))
			_, statErr := os.Stat(target)
			Expect(os.IsNotExist(statErr)).To(BeTrue())
		})

		It("writes a contained relative path", func() {
			_, dir := newConfigDir()
			Expect(os.MkdirAll(filepath.Join(dir, "prompts"), 0o755)).To(Succeed())

			cfg := verify.GavelConfig{}
			ov := &verify.PromptSpec{}
			Expect(persistPromptOverride(&cfg, ov, promptWrite{
				Dir: dir, Source: "file", Path: "prompts/custom.prompt", Text: "hello",
			})).To(Succeed())

			data, err := os.ReadFile(filepath.Join(dir, "prompts", "custom.prompt"))
			Expect(err).ToNot(HaveOccurred())
			Expect(string(data)).To(Equal("hello"))
			Expect(ov.File).To(Equal("prompts/custom.prompt"))
		})
	})

	Describe("overlayFrontmatter", func() {
		It("merges the prompt block one level deep without losing sibling keys", func() {
			base := map[string]any{
				"model":  "claude",
				"prompt": map[string]any{"system": "base-system", "schema": "base-schema"},
			}
			over := map[string]any{
				"effort": "high",
				"prompt": map[string]any{"system": "over-system"},
			}

			Expect(overlayFrontmatter(base, over)).To(Equal(map[string]any{
				"model":  "claude",
				"effort": "high",
				"prompt": map[string]any{"system": "over-system", "schema": "base-schema"},
			}))
		})
	})

	Describe("test run snapshot reads", func() {
		newWorkspace := func() string {
			dir := GinkgoT().TempDir()
			Expect(os.MkdirAll(filepath.Join(dir, ".gavel"), 0o755)).To(Succeed())
			return dir
		}

		It("reads a contained run snapshot", func() {
			dir := newWorkspace()
			path := filepath.Join(dir, ".gavel", "run-2026-01-01T00-00-00Z.json")
			Expect(os.WriteFile(path, []byte(`{"ok":true}`), 0o644)).To(Succeed())

			data, err := readRunSnapshot(dir, path)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(data)).To(Equal(`{"ok":true}`))
		})

		It("rejects a path outside the workspace .gavel directory", func() {
			dir := newWorkspace()
			outside := filepath.Join(GinkgoT().TempDir(), "secret.json")
			Expect(os.WriteFile(outside, []byte(`{"secret":true}`), 0o644)).To(Succeed())

			_, err := readRunSnapshot(dir, outside)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(outside))
		})

		It("rejects a symlink inside .gavel that escapes it", func() {
			dir := newWorkspace()
			outside := filepath.Join(GinkgoT().TempDir(), "secret.json")
			Expect(os.WriteFile(outside, []byte(`{"secret":true}`), 0o644)).To(Succeed())
			link := filepath.Join(dir, ".gavel", "run-link.json")
			Expect(os.Symlink(outside, link)).To(Succeed())

			_, err := readRunSnapshot(dir, link)
			Expect(err).To(HaveOccurred())
		})
	})
})
