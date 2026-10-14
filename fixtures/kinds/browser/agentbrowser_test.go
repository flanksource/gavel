package browser

import (
	"context"
	"os"

	"github.com/flanksource/clicky/task"
	"github.com/flanksource/gavel/fixtures"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("agent-browser kind", func() {
	It("fails fast without a browser session to attach to", func() {
		scope := fixtures.NewStepScope([]string{"PATH=" + os.Getenv("PATH")})
		step := newStep(scope, "ab", "agent-browser", map[string]any{"commands": []string{"snapshot"}})
		_, err := AgentBrowser{}.Execute(context.Background(), step)
		Expect(err).To(MatchError(ContainSubstring("BROWSER_CDP_URL")))
	})

	It("opens a local page in the shared Chrome, snapshots it and re-observes", func() {
		server := newPageServer(page(`<h1>Hello agent</h1><button>Remove enhancement</button>`))
		scope := fixtures.NewStepScope(os.Environ())
		_, chrome := startChrome(scope, nil)
		Expect(scope.Env.Export("todo", "TODO_URL", server.URL, false)).To(Succeed())

		step := newStep(scope, "ui", "agent-browser", map[string]any{
			"commands": []string{"open $TODO_URL", "snapshot"},
			"observe":  "snapshot",
		})
		outcome, err := AgentBrowser{}.Execute(context.Background(), step)
		Expect(err).ToNot(HaveOccurred())
		Expect(outcome.Cleanup).ToNot(BeNil(), "the first step in a session owns its daemon teardown")
		Expect(outcome.Result.Status).To(Equal(task.StatusPASS), outcome.Result.Error)
		Expect(outcome.Outputs["snapshot"]).To(And(ContainSubstring("Hello agent"), ContainSubstring("Remove enhancement")))
		results := outcome.Outputs["results"].([]map[string]any)
		Expect(results).To(HaveLen(2))
		Expect(results[0]).To(HaveKeyWithValue("exitCode", 0))
		Expect(server.userAgents()).ToNot(BeEmpty())

		observed, err := AgentBrowser{}.Observe(context.Background(), step, outcome.Handle)
		Expect(err).ToNot(HaveOccurred())
		Expect(observed["snapshot"]).To(ContainSubstring("Remove enhancement"))

		again, err := AgentBrowser{}.Execute(context.Background(), newStep(scope, "ui-again", "agent-browser", map[string]any{"commands": []string{"get title"}}))
		Expect(err).ToNot(HaveOccurred())
		Expect(again.Result.Status).To(Equal(task.StatusPASS), again.Result.Error)
		Expect(again.Outputs["stdout"]).To(ContainSubstring("gavel fixture page"), "a later step reuses the session's open page")
		Expect(again.Cleanup).To(BeNil(), "only the first step in a session tears it down")

		Expect(outcome.Cleanup(context.Background())).To(Succeed())
		Expect(probeError(chrome.Exports["BROWSER_CDP_URL"])).To(Succeed(), "closing the session must not close the fixture's Chrome")
	})

	It("fails the step with the command and its stderr when a command exits non-zero", func() {
		scope := fixtures.NewStepScope(os.Environ())
		startChrome(scope, nil)
		step := newStep(scope, "ui", "agent-browser", map[string]any{"commands": []string{"no-such-command"}})
		outcome, err := AgentBrowser{}.Execute(context.Background(), step)
		Expect(err).ToNot(HaveOccurred())
		if outcome.Cleanup != nil {
			DeferCleanup(func() { Expect(outcome.Cleanup(context.Background())).To(Succeed()) })
		}
		Expect(outcome.Result.Status).To(Equal(task.StatusFAIL))
		Expect(outcome.Result.Error).To(ContainSubstring("no-such-command"))
	})
})
