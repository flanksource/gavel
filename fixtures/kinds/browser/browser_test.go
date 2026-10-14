package browser

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"

	"github.com/flanksource/gavel/fixtures"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("browser kind", func() {
	It("is registered with a schema that rejects unknown keys", func() {
		kind, err := fixtures.DefaultKinds.Lookup("browser")
		Expect(err).ToNot(HaveOccurred())
		var schema map[string]any
		Expect(json.Unmarshal(kind.Schema(), &schema)).To(Succeed())
		Expect(schema).To(HaveKeyWithValue("additionalProperties", false))
		Expect(schema["properties"]).To(HaveKey("profile"))
	})

	It("starts Chrome on the profile, exports its CDP endpoints and tears the process tree down", func() {
		scope := fixtures.NewStepScope(os.Environ())
		profileDir := filepath.Join(GinkgoT().TempDir(), "profile")
		step := newStep(scope, "chrome", "browser", map[string]any{"profile": map[string]any{"dir": profileDir}})
		requireChrome()

		outcome, err := Browser{}.Execute(context.Background(), step)
		Expect(err).ToNot(HaveOccurred())
		cdpURL := outcome.Exports["BROWSER_CDP_URL"]
		Expect(cdpURL).To(MatchRegexp(`^http://127\.0\.0\.1:\d+$`))
		Expect(outcome.Exports).To(HaveKeyWithValue("BROWSER_PROFILE_DIR", profileDir))
		Expect(outcome.Exports["BROWSER_SESSION"]).To(MatchRegexp(`^gavel-chrome-[0-9a-f]{8}$`))

		var version struct {
			WebSocketDebuggerURL string `json:"webSocketDebuggerUrl"`
		}
		resp, err := http.Get(cdpURL + "/json/version")
		Expect(err).ToNot(HaveOccurred())
		Expect(json.NewDecoder(resp.Body).Decode(&version)).To(Succeed())
		Expect(resp.Body.Close()).To(Succeed())
		Expect(outcome.Exports).To(HaveKeyWithValue("BROWSER_WS_URL", version.WebSocketDebuggerURL))
		Expect(filepath.Join(profileDir, "Local State")).To(BeAnExistingFile())

		observed, err := Browser{}.Observe(context.Background(), step, outcome.Handle)
		Expect(err).ToNot(HaveOccurred())
		Expect(observed).To(HaveKeyWithValue("cdp", HaveKeyWithValue("ready", true)))

		Expect(outcome.Cleanup(context.Background())).To(Succeed())
		Expect(probeError(cdpURL)).To(HaveOccurred())
		observed, err = Browser{}.Observe(context.Background(), step, outcome.Handle)
		Expect(err).ToNot(HaveOccurred())
		Expect(observed).To(HaveKeyWithValue("cdp", HaveKeyWithValue("ready", false)))
	})

	It("wipes a fresh profile, seeds it from another profile and honours a session override", func() {
		scope := fixtures.NewStepScope(os.Environ())
		profileDir := filepath.Join(GinkgoT().TempDir(), "profile")
		writeFile(filepath.Join(profileDir, "stale-marker"), "old run")
		seed := GinkgoT().TempDir()
		writeFile(filepath.Join(seed, "Default", "seed-marker"), "seeded")

		_, outcome := startChrome(scope, map[string]any{
			"session": "e2e-session",
			"profile": map[string]any{"dir": profileDir, "fresh": true, "from": seed},
		})
		Expect(outcome.Exports).To(HaveKeyWithValue("BROWSER_SESSION", "e2e-session"))
		Expect(filepath.Join(profileDir, "stale-marker")).ToNot(BeAnExistingFile())
		Expect(filepath.Join(profileDir, "Default", "seed-marker")).To(BeAnExistingFile())
	})

	It("seeds cookies and localStorage from a Playwright storage-state file", func() {
		server := newPageServer(page(`<h1>state</h1>`))
		state := filepath.Join(GinkgoT().TempDir(), "state.json")
		writeFile(state, `{
  "cookies": [{"name": "gavel_session", "value": "cookie-from-state", "domain": "127.0.0.1", "path": "/", "expires": -1, "httpOnly": false, "secure": false, "sameSite": "Lax"}],
  "origins": [{"origin": "`+server.URL+`", "localStorage": [{"name": "gavel_token", "value": "storage-from-state"}]}]
}`)
		scope := fixtures.NewStepScope(os.Environ())
		_, outcome := startChrome(scope, map[string]any{"profile": map[string]any{"from": state}})

		values, err := evaluateOnPage(context.Background(), outcome.Exports["BROWSER_WS_URL"], server.URL,
			`document.cookie + "|" + localStorage.getItem("gavel_token")`)
		Expect(err).ToNot(HaveOccurred())
		Expect(values).To(Equal("gavel_session=cookie-from-state|storage-from-state"))
	})

	It("fails fast when an explicit binary does not exist", func() {
		scope := fixtures.NewStepScope(os.Environ())
		step := newStep(scope, "chrome", "browser", map[string]any{"binary": "/nonexistent/chrome"})
		_, err := Browser{}.Execute(context.Background(), step)
		Expect(err).To(MatchError(ContainSubstring("/nonexistent/chrome")))
	})
})
