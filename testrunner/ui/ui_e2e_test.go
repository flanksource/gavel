package testui_test

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/flanksource/gavel/testrunner/parsers"
	testui "github.com/flanksource/gavel/testrunner/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestUI(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Test UI E2E")
}

func startServer(srv *testui.Server) (string, func()) {
	listener, err := net.Listen("tcp", "localhost:0")
	Expect(err).ToNot(HaveOccurred())
	port := listener.Addr().(*net.TCPAddr).Port
	url := fmt.Sprintf("http://localhost:%d", port)
	go http.Serve(listener, srv.Handler()) //nolint:errcheck
	return url, func() { listener.Close() }
}

func sampleTests() []parsers.Test {
	return []parsers.Test{
		{
			Name:      "testrunner/",
			Framework: "go test",
			Children: parsers.Tests{
				{
					Name:      "TestParser",
					Framework: "go test",
					Passed:    true,
					Duration:  150 * time.Millisecond,
					File:      "parser_test.go",
					Line:      42,
					Stdout:    "=== RUN TestParser\n--- PASS: TestParser (0.15s)",
				},
				{
					Name:      "TestBuildFailed",
					Framework: "go test",
					Failed:    true,
					Duration:  2 * time.Second,
					Message:   "expected 3 but got 5",
					Stdout:    "error output here",
					Stderr:    "\x1b[31mfatal error\x1b[0m",
				},
				{
					Name:      "TestRegistry/DetectsGinkgo",
					Framework: "go test",
					Passed:    true,
					Duration:  50 * time.Millisecond,
					File:      "registry_test.go",
					Line:      18,
					Context: parsers.GoTestContext{
						ParentTest: "TestRegistry",
					},
				},
				{
					Name:      "TestRegistry/DetectsGoTest",
					Framework: "go test",
					Failed:    true,
					Duration:  80 * time.Millisecond,
					File:      "registry_test.go",
					Line:      25,
					Message:   "runner not found for framework: gotest",
					Context: parsers.GoTestContext{
						ParentTest: "TestRegistry",
					},
				},
			},
		},
		{
			Name:      "parsers/",
			Framework: "ginkgo",
			Children: parsers.Tests{
				{
					Name:      "parses valid JSON",
					Framework: "ginkgo",
					Suite:     []string{"GoTestJSON Parser", "Parse"},
					Passed:    true,
					Duration:  30 * time.Millisecond,
					File:      "gotest_json_test.go",
					Line:      55,
					Context: parsers.GinkgoContext{
						SuiteDescription: "Parsers Suite",
						SuitePath:        "./testrunner/parsers",
					},
				},
				{
					Name:      "handles malformed input",
					Framework: "ginkgo",
					Suite:     []string{"GoTestJSON Parser", "Parse"},
					Failed:    true,
					Duration:  15 * time.Millisecond,
					File:      "gotest_json_test.go",
					Line:      72,
					Message:   "Expected\n    <error>: unexpected EOF\nto be nil",
					Context: parsers.GinkgoContext{
						SuiteDescription: "Parsers Suite",
						SuitePath:        "./testrunner/parsers",
						FailureLocation:  "gotest_json_test.go:78",
					},
				},
				{
					Name:      "skips empty lines",
					Framework: "ginkgo",
					Suite:     []string{"GoTestJSON Parser", "Edge Cases"},
					Skipped:   true,
					File:      "gotest_json_test.go",
					Line:      90,
					Context: parsers.GinkgoContext{
						SuiteDescription: "Parsers Suite",
						SuitePath:        "./testrunner/parsers",
					},
				},
			},
		},
		{
			Name:      "filters.md",
			Framework: "fixture",
			Children: parsers.Tests{
				{
					Name:      "filter by type",
					Framework: "fixture",
					Passed:    true,
					Duration:  300 * time.Millisecond,
					Stdout:    `{"result": true}`,
					Context: parsers.FixtureContext{
						Command:       `gavel analyze --type feat`,
						ExitCode:      0,
						CWD:           "/tmp/test-repo",
						CELExpression: `results.size() > 0`,
						CELVars: map[string]any{
							"results": []any{"commit1", "commit2"},
							"stdout":  `{"result": true}`,
						},
					},
				},
				{
					Name:      "CEL eval fails",
					Framework: "fixture",
					Failed:    true,
					Duration:  500 * time.Millisecond,
					Message:   "CEL expression evaluated to false",
					Context: parsers.FixtureContext{
						Command:       `gavel analyze --type chore`,
						ExitCode:      0,
						CELExpression: `results.size() == 0`,
						CELVars: map[string]any{
							"results": []any{"unexpected"},
						},
						Expected: 0,
						Actual:   1,
					},
				},
			},
		},
	}
}

var (
	suiteSrv         *testui.Server
	suiteURL         string
	suiteSrvCleanup  func()
	suiteAllocCtx    context.Context
	suiteAllocCancel context.CancelFunc
	suiteBrowserCtx  context.Context
	suiteBrowserDone context.CancelFunc
)

var _ = BeforeSuite(func() {
	suiteSrv = testui.NewServer()
	suiteURL, suiteSrvCleanup = startServer(suiteSrv)

	tmpDir, err := os.MkdirTemp("", "testui-chrome-*")
	Expect(err).ToNot(HaveOccurred())
	DeferCleanup(os.RemoveAll, tmpDir)

	suiteAllocCtx, suiteAllocCancel = chromedp.NewExecAllocator(context.Background(),
		append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.Flag("headless", true),
			chromedp.Flag("disable-gpu", true),
			chromedp.Flag("no-sandbox", true),
			chromedp.Flag("disable-dev-shm-usage", true),
			chromedp.WindowSize(1440, 900),
			chromedp.UserDataDir(tmpDir),
		)...,
	)
	// Warm up the browser so the first tab is ready.
	suiteBrowserCtx, suiteBrowserDone = chromedp.NewContext(suiteAllocCtx)
	Expect(chromedp.Run(suiteBrowserCtx)).To(Succeed())
})

var _ = AfterSuite(func() {
	if suiteBrowserDone != nil {
		suiteBrowserDone()
	}
	if suiteAllocCancel != nil {
		suiteAllocCancel()
	}
	if suiteSrvCleanup != nil {
		suiteSrvCleanup()
	}
})

var _ = Describe("Test UI E2E", func() {
	var (
		srv    *testui.Server
		url    string
		ctx    context.Context
		cancel context.CancelFunc
	)

	BeforeEach(func() {
		srv = suiteSrv
		url = suiteURL
		srv.SetResults(sampleTests())
		srv.SetDiagnosticsManager(nil)

		// New tab per test, sharing the suite-wide browser process.
		ctx, cancel = chromedp.NewContext(suiteBrowserCtx)
	})

	AfterEach(func() {
		cancel()
	})

	It("serves the HTML page", func() {
		var title string
		err := chromedp.Run(ctx,
			chromedp.Navigate(url),
			chromedp.Title(&title),
		)
		Expect(err).ToNot(HaveOccurred())
		Expect(title).To(Equal("Test Results"))
	})

	It("renders the test tree with all frameworks", func() {
		var treeText string
		err := chromedp.Run(ctx,
			chromedp.Navigate(url),
			chromedp.WaitVisible(`body`, chromedp.ByQuery),
			chromedp.Sleep(2*time.Second),
			chromedp.Text(`body`, &treeText, chromedp.ByQuery),
		)
		Expect(err).ToNot(HaveOccurred())
		// Go tests (humanizeName strips "Test" prefix and inserts spaces)
		Expect(treeText).To(ContainSubstring("testrunner"))
		Expect(treeText).To(ContainSubstring("Parser"))
		Expect(treeText).To(ContainSubstring("Build Failed"))
		Expect(treeText).To(ContainSubstring("DetectsGinkgo"))
		// Ginkgo tests
		Expect(treeText).To(ContainSubstring("parsers"))
		Expect(treeText).To(ContainSubstring("parses valid JSON"))
		Expect(treeText).To(ContainSubstring("handles malformed input"))
		// Fixtures
		Expect(treeText).To(ContainSubstring("filters.md"))
		Expect(treeText).To(ContainSubstring("filter by type"))
	})

	It("shows summary counts for all frameworks", func() {
		var text string
		err := chromedp.Run(ctx,
			chromedp.Navigate(url),
			chromedp.Sleep(2*time.Second),
			chromedp.Text(`body`, &text, chromedp.ByQuery),
		)
		Expect(err).ToNot(HaveOccurred())
		Expect(text).To(ContainSubstring("passed"))
		Expect(text).To(ContainSubstring("failed"))
		Expect(text).To(ContainSubstring("skipped"))
		Expect(text).To(ContainSubstring("9 tests"))
	})

	It("shows detail panel when clicking a test", func() {
		var detailText string
		err := chromedp.Run(ctx,
			chromedp.Navigate(url),
			chromedp.Sleep(2*time.Second),
			// Click on the failed test
			chromedp.Click(`//span[contains(text(), "Build Failed")]`, chromedp.BySearch),
			chromedp.Sleep(500*time.Millisecond),
			chromedp.Text(`body`, &detailText, chromedp.ByQuery),
		)
		Expect(err).ToNot(HaveOccurred())
		Expect(detailText).To(ContainSubstring("expected 3 but got 5"))
	})

	It("shows diagnostics tab and captures a stack trace", func() {
		srv.SetDiagnosticsManager(testui.NewDiagnosticsManager(4242, newFakeDiagnosticsCollector()))

		var detailText string
		err := chromedp.Run(ctx,
			chromedp.Navigate(url),
			chromedp.Sleep(2*time.Second),
			chromedp.Click(`//button[contains(text(), "Diagnostics")]`, chromedp.BySearch),
			chromedp.Sleep(500*time.Millisecond),
			chromedp.Click(`//button[contains(text(), "Collect stack trace")]`, chromedp.BySearch),
			chromedp.Sleep(500*time.Millisecond),
			chromedp.Text(`body`, &detailText, chromedp.ByQuery),
		)
		Expect(err).ToNot(HaveOccurred())
		Expect(detailText).To(ContainSubstring("Diagnostics"))
		Expect(detailText).To(ContainSubstring("gavel test --ui ./testrunner/ui"))
		Expect(detailText).To(ContainSubstring("pid 4242"))
		Expect(detailText).To(ContainSubstring("main.first"))
	})

	It("shows fixture context in detail panel", func() {
		var detailText string
		err := chromedp.Run(ctx,
			chromedp.Navigate(url),
			chromedp.Sleep(2*time.Second),
			chromedp.Click(`//span[contains(text(), "CEL eval fails")]`, chromedp.BySearch),
			chromedp.Sleep(500*time.Millisecond),
			chromedp.Text(`body`, &detailText, chromedp.ByQuery),
		)
		Expect(err).ToNot(HaveOccurred())
		Expect(detailText).To(ContainSubstring("results.size() == 0"))
		Expect(detailText).To(ContainSubstring("gavel analyze --type chore"))
	})

	It("shows Go subtest context in detail panel", func() {
		var detailText string
		err := chromedp.Run(ctx,
			chromedp.Navigate(url),
			chromedp.Sleep(2*time.Second),
			chromedp.Click(`//span[contains(text(), "Registry / DetectsGoTest")]`, chromedp.BySearch),
			chromedp.Sleep(500*time.Millisecond),
			chromedp.Text(`body`, &detailText, chromedp.ByQuery),
		)
		Expect(err).ToNot(HaveOccurred())
		Expect(detailText).To(ContainSubstring("TestRegistry"))
		Expect(detailText).To(ContainSubstring("runner not found for framework: gotest"))
		Expect(detailText).To(ContainSubstring("registry_test.go"))
	})

	It("shows Ginkgo context in detail panel", func() {
		var detailText string
		err := chromedp.Run(ctx,
			chromedp.Navigate(url),
			chromedp.Sleep(2*time.Second),
			chromedp.Click(`//span[contains(text(), "handles malformed input")]`, chromedp.BySearch),
			chromedp.Sleep(500*time.Millisecond),
			chromedp.Text(`body`, &detailText, chromedp.ByQuery),
		)
		Expect(err).ToNot(HaveOccurred())
		Expect(detailText).To(ContainSubstring("Parsers Suite"))
		Expect(detailText).To(ContainSubstring("./testrunner/parsers"))
		Expect(detailText).To(ContainSubstring("gotest_json_test.go:78"))
		Expect(detailText).To(ContainSubstring("unexpected EOF"))
	})

	It("shows Ginkgo suite path in detail panel", func() {
		var detailText string
		err := chromedp.Run(ctx,
			chromedp.Navigate(url),
			chromedp.Sleep(2*time.Second),
			chromedp.Click(`//span[contains(text(), "parses valid JSON")]`, chromedp.BySearch),
			chromedp.Sleep(500*time.Millisecond),
			chromedp.Text(`body`, &detailText, chromedp.ByQuery),
		)
		Expect(err).ToNot(HaveOccurred())
		Expect(detailText).To(ContainSubstring("Parsers Suite"))
		Expect(detailText).To(ContainSubstring("GoTestJSON Parser"))
	})

	It("returns JSON from /api/tests", func() {
		var body string
		err := chromedp.Run(ctx,
			chromedp.Navigate(url+"/api/tests"),
			chromedp.Sleep(500*time.Millisecond),
			chromedp.Text(`body`, &body, chromedp.ByQuery),
		)
		Expect(err).ToNot(HaveOccurred())
		Expect(body).To(ContainSubstring(`"running":false`))
		Expect(body).To(ContainSubstring(`"testrunner/"`))
	})

	It("streams results via SSE", func() {
		streamSrv := testui.NewServer()
		streamURL, streamCleanup := startServer(streamSrv)
		defer streamCleanup()

		updates := make(chan []parsers.Test, 4)
		streamSrv.StreamFrom(updates)

		// Use a fresh tab for the streaming server so we don't disturb the shared one.
		streamCtx, streamCancel := chromedp.NewContext(suiteBrowserCtx)
		defer streamCancel()

		// Send pending first
		updates <- []parsers.Test{{Name: "pkg/", Pending: true}}

		var text1 string
		err := chromedp.Run(streamCtx,
			chromedp.Navigate(streamURL),
			chromedp.Sleep(3*time.Second),
			chromedp.Text(`body`, &text1, chromedp.ByQuery),
		)
		Expect(err).ToNot(HaveOccurred())
		Expect(text1).To(ContainSubstring("pkg/"))

		// Send completed results
		updates <- []parsers.Test{{
			Name: "pkg/", Children: parsers.Tests{{
				Name: "TestFoo", Framework: "go test", Passed: true, Duration: time.Second,
			}},
		}}
		close(updates)

		var text2 string
		err = chromedp.Run(streamCtx,
			chromedp.Sleep(3*time.Second),
			chromedp.Text(`body`, &text2, chromedp.ByQuery),
		)
		Expect(err).ToNot(HaveOccurred())
		Expect(text2).To(ContainSubstring("Foo"))
		Expect(text2).To(ContainSubstring("Test run complete"))
	})

	It("captures screenshot with detail panel showing fixture context", func() {
		screenshotDir := GinkgoT().TempDir()
		screenshotPath := filepath.Join(screenshotDir, "test-ui.png")

		var buf []byte
		err := chromedp.Run(ctx,
			chromedp.Navigate(url),
			chromedp.Sleep(2*time.Second),
			chromedp.Click(`//span[contains(text(), "CEL eval fails")]`, chromedp.BySearch),
			chromedp.Sleep(500*time.Millisecond),
			chromedp.FullScreenshot(&buf, 90),
		)
		Expect(err).ToNot(HaveOccurred())
		Expect(len(buf)).To(BeNumerically(">", 10000))

		err = os.WriteFile(screenshotPath, buf, 0644)
		Expect(err).ToNot(HaveOccurred())

		// Verify the page content has expected elements
		var text string
		err = chromedp.Run(ctx,
			chromedp.Text(`body`, &text, chromedp.ByQuery),
		)
		Expect(err).ToNot(HaveOccurred())

		// Left panel: tree with all frameworks
		Expect(text).To(ContainSubstring("testrunner/"))
		Expect(text).To(ContainSubstring("parsers/"))
		Expect(text).To(ContainSubstring("filters.md"))

		// Right panel: fixture detail
		Expect(text).To(ContainSubstring("CEL eval fails"))
		Expect(text).To(ContainSubstring("results.size() == 0"))
		Expect(text).To(ContainSubstring("gavel analyze --type chore"))

		// Summary bar
		Expect(text).To(ContainSubstring("9 tests"))
		Expect(text).To(ContainSubstring("Test run complete"))

		GinkgoWriter.Printf("Screenshot saved to: %s (%d bytes)\n", screenshotPath, len(buf))
	})

	It("expand and collapse all works", func() {
		var collapsed, expanded string
		err := chromedp.Run(ctx,
			chromedp.Navigate(url),
			chromedp.Sleep(2*time.Second),
			// Click collapse
			chromedp.Click(`//button[contains(text(), "Collapse")]`, chromedp.BySearch),
			chromedp.Sleep(500*time.Millisecond),
			chromedp.Text(`body`, &collapsed, chromedp.ByQuery),
		)
		Expect(err).ToNot(HaveOccurred())
		// After collapse, individual tests should not be visible
		Expect(strings.Contains(collapsed, "Build Failed")).To(BeFalse())

		err = chromedp.Run(ctx,
			// Click expand
			chromedp.Click(`//button[contains(text(), "Expand")]`, chromedp.BySearch),
			chromedp.Sleep(500*time.Millisecond),
			chromedp.Text(`body`, &expanded, chromedp.ByQuery),
		)
		Expect(err).ToNot(HaveOccurred())
		Expect(expanded).To(ContainSubstring("Build Failed"))
	})

	It("shows passed Go test detail with stdout and file location", func() {
		var detailText string
		err := chromedp.Run(ctx,
			chromedp.Navigate(url),
			chromedp.Sleep(2*time.Second),
			chromedp.Click(`//span[contains(text(), "Parser")]`, chromedp.BySearch),
			chromedp.Sleep(500*time.Millisecond),
			chromedp.Text(`body`, &detailText, chromedp.ByQuery),
		)
		Expect(err).ToNot(HaveOccurred())
		// Header shows test name and pass status
		Expect(detailText).To(ContainSubstring("TestParser"))
		// File location
		Expect(detailText).To(ContainSubstring("parser_test.go"))
		// Duration
		Expect(detailText).To(ContainSubstring("150ms"))
		// Framework badge
		Expect(detailText).To(ContainSubstring("go test"))
		// Stdout captured
		Expect(detailText).To(ContainSubstring("PASS: TestParser"))
	})

	It("shows passed Go subtest detail with parent test context", func() {
		var detailText string
		err := chromedp.Run(ctx,
			chromedp.Navigate(url),
			chromedp.Sleep(2*time.Second),
			chromedp.Click(`//span[contains(text(), "Registry / DetectsGinkgo")]`, chromedp.BySearch),
			chromedp.Sleep(500*time.Millisecond),
			chromedp.Text(`body`, &detailText, chromedp.ByQuery),
		)
		Expect(err).ToNot(HaveOccurred())
		Expect(detailText).To(ContainSubstring("TestRegistry/DetectsGinkgo"))
		Expect(detailText).To(ContainSubstring("registry_test.go"))
		Expect(detailText).To(ContainSubstring("TestRegistry"))
	})

	It("shows results after page reload", func() {
		// Load once to confirm it works
		var text1 string
		err := chromedp.Run(ctx,
			chromedp.Navigate(url),
			chromedp.Sleep(2*time.Second),
			chromedp.Text(`body`, &text1, chromedp.ByQuery),
		)
		Expect(err).ToNot(HaveOccurred())
		Expect(text1).To(ContainSubstring("Parser"))

		// Reload the page
		var text2 string
		err = chromedp.Run(ctx,
			chromedp.Navigate(url),
			chromedp.Sleep(2*time.Second),
			chromedp.Text(`body`, &text2, chromedp.ByQuery),
		)
		Expect(err).ToNot(HaveOccurred())
		Expect(text2).To(ContainSubstring("testrunner/"))
		Expect(text2).To(ContainSubstring("Parser"))
		Expect(text2).To(ContainSubstring("filters.md"))
		Expect(text2).To(ContainSubstring("Test run complete"))
	})
})
