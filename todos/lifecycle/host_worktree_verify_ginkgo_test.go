package lifecycle_test

import (
	"context"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/commons-db/shell"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/lifecycle"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// The definition of done is built from the main checkout, but a run step works
// in a worktree of it. The document must run where the step works: pinned to
// the main checkout, the in-loop verification judged a tree without the agent's
// edits.
var _ = Describe("RunStep in-loop verification in a worktree", func() {
	// Passes only inside a linked worktree, whose git dir differs from the
	// repository's common one.
	const inLinkedWorktree = "### command: runs in the run's worktree\n\n```bash\n" +
		`test "$(git rev-parse --git-dir)" != "$(git rev-parse --git-common-dir)"` + "\n```"

	It("runs a default todo fixture inside the run step's worktree", func() {
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		provider := &fakeProvider{plan: todos.PlanState{Exists: true, Approved: true, Content: "# plan", Revision: 2}}
		host := newHost(provider)
		gitIn(host.WorkDir, "init", "-b", "main")
		commitFile(host.WorkDir, "README.md", "widgets\n")

		// The built-in run step's checkout, without its commit pipeline: the
		// scripted agent edits nothing, so there is nothing to commit.
		def := host.Def.Definition()
		for i := range def.Steps {
			if def.Steps[i].Name == "run" {
				def.Steps[i].Spec = &api.Spec{
					Setup: &shell.Setup{Checkout: &shell.Checkout{
						Mode: shell.CheckoutLocal, Worktree: &shell.Worktree{Mode: shell.WorktreeNew},
					}},
					Workflow: &api.Workflow{Verify: &api.Verify{
						Fixture: "{{subject.verification.document}}", Scope: api.VerifyScopeAll,
					}},
				}
			}
		}
		engine, err := lifecycle.New(def)
		Expect(err).NotTo(HaveOccurred())
		host.Def = engine
		todo := hostTodo()
		todo.VerificationMarkdown = inLinkedWorktree
		agent := &scriptedProvider{events: scriptedTurn(`{"summary":"Built it.","endStatus":"completed"}`, true)}

		outcome, err := host.RunStep(context.Background(), todo, stepNamed(engine, "run"), lifecycle.RunOptions{Provider: agent})

		Expect(err).NotTo(HaveOccurred())
		Expect(outcome.Result.Verify).NotTo(BeNil())
		Expect([]bool{outcome.Result.Verify.Ran, outcome.Result.Verify.Passed}).To(Equal([]bool{true, true}),
			"the definition of done ran in the worktree: %s", outcome.Result.Verify.Reason)
		Expect(outcome.Status).To(Equal("verified"))
	})
})
