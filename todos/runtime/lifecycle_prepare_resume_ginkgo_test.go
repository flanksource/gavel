package runtime

import (
	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/captain/pkg/promptrun"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/runtime/runtimetest"
	"github.com/flanksource/gavel/todos/types"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// A resume hands the prior provider session id to the new agent, so resuming
// onto another provider family is refused before anything is admitted: the
// claude conversation stays the issue's run, with no second run linked.
var _ = Describe("resuming a run on another provider family", func() {
	DescribeTable("is refused before admission",
		func(ctx SpecContext, prior promptrun.Completed) {
			provider := outcomeProvider(ctx, "resume_provider_mismatch", GinkgoT().TempDir())
			todo, err := provider.Create(ctx, todos.CreateRequest{Title: "Plan the widget", Body: "Design it", Status: types.StatusPending})
			Expect(err).NotTo(HaveOccurred())
			_, err = runtimetest.Admit(ctx, provider, todo, todos.RunPreparation{Mode: types.ModePlan, Prompt: "plan", ExecutorName: "headless-claude"}, prior)
			Expect(err).NotTo(HaveOccurred())
			issueID := uuid.MustParse(todo.ID)
			before, err := provider.Repository().GetIssue(ctx, issueID)
			Expect(err).NotTo(HaveOccurred())

			_, err = provider.PrepareRun(ctx, todo, todos.RunPreparation{
				Mode: types.ModePlan, Prompt: "plan", ExecutorName: "codex", Resume: true,
				Requested: captaindb.PromptRunRuntimeSelection{Provider: "openai", Mode: "agent", Model: "gpt-5.6-sol"},
			})

			Expect(err).To(MatchError(todos.ErrRunResumeProviderMismatch))
			after, err := provider.Repository().GetIssue(ctx, issueID)
			Expect(err).NotTo(HaveOccurred())
			Expect(after.ActivePromptRunID).To(Equal(before.ActivePromptRunID))
			links, err := provider.Repository().ListPromptRuns(ctx, issueID)
			Expect(err).NotTo(HaveOccurred())
			Expect(links).To(HaveLen(1), "no second run may be linked to the issue")
		},
		Entry("a finished claude plan", promptrun.Completed{
			Runtime: runtimetest.ClaudeCLI, Model: "claude-sonnet", ProviderSessionID: "sess-claude-plan",
			Outcome: promptrun.Outcome{State: captaindb.PromptRunStateSucceeded, Phase: captaindb.PromptRunPhaseFinished},
		}),
		Entry("a claude plan parked on a question", runtimetest.Parked("sess-claude-ask", map[string]any{"summary": "which database?"})),
	)
})
