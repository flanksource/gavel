package browser

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/flanksource/clicky/task"
	"github.com/flanksource/gavel/fixtures"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const sessionAlertPage = `<h1>todo</h1>
<script>
setTimeout(() => {
  const alert = document.createElement('div');
  alert.setAttribute('role', 'alert');
  alert.innerHTML = '<span>Session error</span><button type="button" aria-label="Show details">Show details</button>';
  alert.querySelector('button').addEventListener('click', () => {
    const pre = document.createElement('pre');
    pre.textContent = 'launch\nboom: no scenario rule matched';
    setTimeout(() => alert.appendChild(pre), 20);
  });
  document.body.appendChild(alert);
  setTimeout(() => alert.remove(), 400);
}, 300);
</script>`

func generatedSpecs() []string {
	matches, err := filepath.Glob(filepath.Join(repoRoot(), "pr", "ui", "e2e", ".generated", "*.ts"))
	Expect(err).ToNot(HaveOccurred())
	return matches
}

var _ = Describe("watcher events", func() {
	It("folds each raw watcher event onto the type/text/url/status/detail shape the UI renders", func() {
		const page, api = "http://127.0.0.1:1/todos", "http://127.0.0.1:1/api/todos"
		raw := []map[string]any{
			{"type": "alert", "text": "Session error", "details": "boom", "url": page, "transient": true},
			{"type": "console", "text": "Uncaught", "url": page},
			{"type": "pageerror", "message": "x is undefined", "stack": "at main.js:1"},
			{"type": "response", "method": "GET", "url": api, "status": float64(502)},
		}

		Expect(reportedEvents(raw)).To(Equal([]watcherEvent{
			{Type: "alert", Text: "Session error", URL: page, Detail: "boom"},
			{Type: "console", Text: "Uncaught", URL: page},
			{Type: "pageerror", Text: "x is undefined", Detail: "at main.js:1"},
			{Type: "response", Text: "GET", URL: api, Status: 502},
		}))
	})
})

var _ = Describe("playwright kind", func() {
	It("rejects a step with both or neither of file and script", func() {
		scope := fixtures.NewStepScope(os.Environ())
		_, err := Playwright{}.Execute(context.Background(), newStep(scope, "pw", "playwright", map[string]any{}))
		Expect(err).To(MatchError(ContainSubstring("exactly one of file or script")))
		_, err = Playwright{}.Execute(context.Background(), newStep(scope, "pw", "playwright", map[string]any{"file": "a.spec.ts", "script": "x"}))
		Expect(err).To(MatchError(ContainSubstring("exactly one of file or script")))
	})

	It("runs an inline spec over the shared Chrome and reports each test as a child", func() {
		userAgent := "gavel-cdp-probe-" + randomHex(4)
		server := newPageServer(page(`<h1>Shared profile</h1>`))
		scope := fixtures.NewStepScope(os.Environ())
		startChrome(scope, map[string]any{"args": []string{"--user-agent=" + userAgent}})

		step := newStep(scope, "session-ui", "playwright", map[string]any{
			"url": server.URL,
			"script": `import { test, expect } from '../fixtures/cdp';
test('page renders over CDP', async ({ page }) => {
  await page.goto(process.env.TARGET_URL!);
  await expect(page.getByRole('heading', { name: 'Shared profile' })).toBeVisible();
});
test('fails loudly', async ({ page }) => {
  await page.goto(process.env.TARGET_URL!);
  await expect(page.getByText('not on the page')).toBeVisible({ timeout: 500 });
});`,
		})
		outcome, err := Playwright{}.Execute(context.Background(), step)
		Expect(err).ToNot(HaveOccurred())

		Expect(outcome.Result.Status).To(Equal(task.StatusFAIL))
		Expect(outcome.Outputs).To(And(HaveKeyWithValue("passed", 1), HaveKeyWithValue("failed", 1), HaveKeyWithValue("total", 2)))
		Expect(outcome.Result.Children).To(HaveLen(2))
		statuses := map[string]task.Status{}
		for _, child := range outcome.Result.Children {
			statuses[child.Name] = child.Results.Status
		}
		Expect(statuses).To(Equal(map[string]task.Status{"page renders over CDP": task.StatusPASS, "fails loudly": task.StatusFAIL}))
		Expect(server.userAgents()).To(ContainElement(userAgent), "the spec must load the page in the fixture's Chrome")
		Expect(generatedSpecs()).To(BeEmpty(), "the inline spec is removed after the run")
	})

	It("fails a background watcher at teardown when a transient Session error alert appeared", func() {
		server := newPageServer(page(sessionAlertPage))
		scope := fixtures.NewStepScope(os.Environ())
		startChrome(scope, nil)

		step := newStep(scope, "watch", "playwright", map[string]any{"file": "../e2e/watch-session.ts", "url": server.URL})
		handle, err := Playwright{}.Start(context.Background(), step)
		Expect(err).ToNot(HaveOccurred())
		watcher := handle.(*watcherHandle)
		Eventually(watcher.eventTypes).WithTimeout(10 * time.Second).Should(ContainElement("alert"))

		result := handle.Stop(context.Background())
		Expect(result.Status).To(Equal(task.StatusFAIL))
		Expect(result.Error).To(And(ContainSubstring("Session error"), ContainSubstring("boom: no scenario rule matched")))
		Expect(result.Metadata).To(HaveKey("events"))
	})

	It("fails a background watcher on a stream error banner that has no role and logs nothing", func() {
		server := newPageServer(page(`<header><span>Gavel</span></header><script>
			setTimeout(() => {
				const banner = document.createElement('p');
				banner.textContent = 'Pull request stream received an invalid update.';
				document.querySelector('header').appendChild(banner);
			}, 300);
		</script>`))
		scope := fixtures.NewStepScope(os.Environ())
		startChrome(scope, nil)

		step := newStep(scope, "watch", "playwright", map[string]any{"file": "../e2e/watch-session.ts", "url": server.URL})
		handle, err := Playwright{}.Start(context.Background(), step)
		Expect(err).ToNot(HaveOccurred())
		Eventually(handle.(*watcherHandle).eventTypes).WithTimeout(10 * time.Second).Should(ContainElement("alert"))

		result := handle.Stop(context.Background())
		Expect(result.Status).To(Equal(task.StatusFAIL))
		Expect(result.Error).To(ContainSubstring("Pull request stream received an invalid update."))
	})

	It("passes a background watcher whose page stayed clean", func() {
		server := newPageServer(page(`<h1>clean</h1>`))
		scope := fixtures.NewStepScope(os.Environ())
		startChrome(scope, nil)

		step := newStep(scope, "watch", "playwright", map[string]any{"file": "../e2e/watch-session.ts", "url": server.URL})
		handle, err := Playwright{}.Start(context.Background(), step)
		Expect(err).ToNot(HaveOccurred())
		Expect(handle.(*watcherHandle).eventTypes()).To(BeEmpty(), "the ready handshake is not a reported event")

		result := handle.Stop(context.Background())
		Expect(result.Error).To(BeEmpty())
		Expect(result.Status).To(Equal(task.StatusPASS))
		Expect(result.Metadata["events"]).To(BeEmpty())
	})
})
