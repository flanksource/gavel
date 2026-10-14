package browser

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/chromedp/chromedp"
	"github.com/flanksource/gavel/fixtures"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestBrowserKinds(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Browser Kinds Suite")
}

// repoRoot is the gavel checkout; the suite runs from fixtures/kinds/browser.
func repoRoot() string {
	root, err := filepath.Abs("../../..")
	Expect(err).ToNot(HaveOccurred())
	return root
}

// uiFixtureDir stands in for a fixture file living in pr/ui, so the
// playwright kind discovers pr/ui (the package with @playwright/test).
func uiFixtureDir() string { return filepath.Join(repoRoot(), "pr", "ui", "testdata") }

func newStep(scope *fixtures.StepScope, id, kind string, cfg map[string]any) *fixtures.StepContext {
	raw, err := json.Marshal(cfg)
	Expect(err).ToNot(HaveOccurred())
	return &fixtures.StepContext{
		Test:    fixtures.FixtureTest{Name: id, SourceDir: uiFixtureDir()},
		Spec:    fixtures.StepSpec{ID: id, Kind: kind, Config: raw},
		Scope:   scope,
		WorkDir: repoRoot(),
		TmpDir:  GinkgoT().TempDir(),
	}
}

// applyExports does what the fixture runner does after a successful step.
func applyExports(scope *fixtures.StepScope, step *fixtures.StepContext, outcome fixtures.StepOutcome) {
	for name, value := range outcome.Exports {
		Expect(scope.Env.Export(step.Spec.ID, name, value, false)).To(Succeed())
	}
	Expect(scope.SetOutputs(step.Spec.ID, outcome.Outputs)).To(Succeed())
}

// requireChrome skips (loudly) only when no Chrome can be resolved at all.
func requireChrome() {
	if _, err := resolveChrome(defaultChromeSearch("")); err != nil {
		fmt.Fprintf(GinkgoWriter, "\n!!! SKIPPING real-Chrome spec: %v\n", err)
		Skip("no Chrome found: " + err.Error())
	}
}

// requireAgentBrowser skips (loudly) only when no agent-browser binary is
// installed; one that is installed but cannot attach still fails the spec.
func requireAgentBrowser() {
	if _, err := resolveAgentBrowser(""); errors.Is(err, errAgentBrowserNotFound) {
		fmt.Fprintf(GinkgoWriter, "\n!!! SKIPPING agent-browser spec: %v\n", err)
		Skip(err.Error())
	}
}

// startChrome runs the browser kind and registers its teardown.
func startChrome(scope *fixtures.StepScope, cfg map[string]any) (*fixtures.StepContext, fixtures.StepOutcome) {
	requireChrome()
	step := newStep(scope, "chrome", "browser", cfg)
	outcome, err := Browser{}.Execute(context.Background(), step)
	Expect(err).ToNot(HaveOccurred())
	Expect(outcome.Cleanup).ToNot(BeNil())
	DeferCleanup(func() { Expect(outcome.Cleanup(context.Background())).To(Succeed()) })
	applyExports(scope, step, outcome)
	return step, outcome
}

// pageServer serves one HTML page at / and records the User-Agent of every
// request, which proves which Chrome loaded it.
type pageServer struct {
	*httptest.Server
	mu     sync.Mutex
	agents []string
}

func newPageServer(html string) *pageServer {
	server := &pageServer{}
	server.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		server.mu.Lock()
		server.agents = append(server.agents, r.UserAgent())
		server.mu.Unlock()
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(html))
	}))
	DeferCleanup(server.Close)
	return server
}

func (s *pageServer) userAgents() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]string(nil), s.agents...)
}

func page(body string) string {
	return `<!doctype html><html><head><title>gavel fixture page</title><link rel="icon" href="data:,"></head><body>` + body + `</body></html>`
}

func probeError(cdpURL string) error {
	resp, err := http.Get(cdpURL + "/json/version")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status)
	}
	return nil
}

// evaluateOnPage loads url in a new tab of the running Chrome and returns the
// string value of expression.
func evaluateOnPage(ctx context.Context, wsURL, url, expression string) (string, error) {
	allocCtx, cancelAlloc := chromedp.NewRemoteAllocator(ctx, wsURL, chromedp.NoModifyURL)
	defer cancelAlloc()
	tabCtx, cancelTab := chromedp.NewContext(allocCtx)
	defer cancelTab()
	var value string
	err := chromedp.Run(tabCtx, chromedp.Navigate(url), chromedp.Evaluate(expression, &value))
	return value, err
}

func writeFile(path, content string) {
	Expect(os.MkdirAll(filepath.Dir(path), 0o755)).To(Succeed())
	Expect(os.WriteFile(path, []byte(content), 0o755)).To(Succeed())
}
