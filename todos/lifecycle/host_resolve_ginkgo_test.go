package lifecycle_test

import (
	"context"
	"time"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/lifecycle"
	"github.com/flanksource/gavel/todos/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Host Resolve", func() {
	var (
		provider *fakeProvider
		host     *lifecycle.Host
		ctx      context.Context
	)
	BeforeEach(func() {
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		provider = &fakeProvider{plan: todos.PlanState{Exists: true, Approved: true, Content: "# plan", Revision: 2}}
		host = newHost(provider)
		ctx = context.Background()
	})

	It("renders the prompt and folds the spec a run step would dispatch", func() {
		todo := hostTodo()
		step := stepNamed(host.Def, "run")

		resolution, err := host.Resolve(ctx, todo, step, lifecycle.RunOptions{})

		Expect(err).NotTo(HaveOccurred())
		Expect(resolution.Step.Name).To(Equal("run"))
		Expect(resolution.Class).To(Equal(types.ModeRun))
		Expect(resolution.Prompt).To(ContainSubstring("Implement the thing"))
		Expect(resolution.Spec.Prompt.User).To(Equal(resolution.Prompt))
		Expect(resolution.Timeout).To(BeNumerically(">", time.Duration(0)))
		Expect(resolution.WorkDir).To(Equal(host.WorkDir))
		// The step's own `spec:` block reached the fold: todos.yaml gives the run
		// step a worktree checkout and a commit policy.
		Expect(resolution.Spec.Setup).NotTo(BeNil())
		Expect(resolution.Spec.Workflow).NotTo(BeNil())
		Expect(resolution.Spec.Workflow.Commits).NotTo(BeEmpty())
		// The verify fixture placeholder expanded to the todo's own document.
		Expect(resolution.Spec.Workflow.Verify).NotTo(BeNil())
		Expect(resolution.Spec.Workflow.Verify.Fixture).To(ContainSubstring("it works"))
	})

	It("reports the layer trace so a caller can say which layer supplied what", func() {
		resolution, err := host.Resolve(ctx, hostTodo(), stepNamed(host.Def, "run"), lifecycle.RunOptions{})

		Expect(err).NotTo(HaveOccurred())
		var names []string
		for _, layer := range resolution.Trace {
			names = append(names, layer.Name)
		}
		Expect(names).To(ContainElement(".gavel.yaml ai"))
		Expect(names).To(ContainElement("lifecycle step run"))
	})

	It("folds the caller's request as the top layer", func() {
		resolution, err := host.Resolve(ctx, hostTodo(), stepNamed(host.Def, "run"), lifecycle.RunOptions{
			Request: api.Spec{Budget: api.Budget{MaxTurns: 7, Timeout: "12m"}},
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(resolution.Spec.Budget.MaxTurns).To(Equal(7))
		Expect(resolution.Timeout).To(Equal(12 * time.Minute))
	})

	It("resolves a verify step to the definition of done without an agent prompt", func() {
		resolution, err := host.Resolve(ctx, hostTodo(), stepNamed(host.Def, lifecycle.StepVerify), lifecycle.RunOptions{})

		Expect(err).NotTo(HaveOccurred())
		Expect(resolution.Class).To(Equal(types.ModeVerify))
		Expect(resolution.Prompt).To(BeEmpty())
		Expect(resolution.Spec.Workflow).NotTo(BeNil())
		Expect(resolution.Spec.Workflow.Verify).NotTo(BeNil())
		Expect(resolution.Spec.Workflow.Verify.Fixture).To(ContainSubstring("it works"))
		// A verify step grades what exists; it must never commit.
		Expect(resolution.Spec.Workflow.Commits).To(BeEmpty())
	})

	It("rejects a step the lifecycle does not declare rather than resolving it", func() {
		_, err := host.Resolve(ctx, hostTodo(), lifecycle.Step{Name: "nonesuch"}, lifecycle.RunOptions{})

		Expect(err).To(MatchError(ContainSubstring(`step "nonesuch" is not part of lifecycle todos`)))
	})
})
