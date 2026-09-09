package lifecycle_test

import (
	"context"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/lifecycle"
	"github.com/flanksource/gavel/todos/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Host VerifyStep", func() {
	var (
		provider *fakeProvider
		host     *lifecycle.Host
		ctx      context.Context
	)
	BeforeEach(func() {
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		provider = &fakeProvider{plan: todos.PlanState{Exists: true, Approved: true}}
		host = newHost(provider)
		ctx = context.Background()
	})

	// A todo with nothing to check has judged nothing. Reporting that as a pass
	// is the one outcome a definition of done must never produce.
	It("refuses a todo with no definition of done rather than passing it", func() {
		todo := hostTodo()
		todo.VerificationMarkdown = ""
		todo.AcceptanceCriteria = nil

		_, err := host.VerifyStep(ctx, todo, api.Spec{})

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Or(
			ContainSubstring("no verification"),
			ContainSubstring("verify"),
		))
	})

	It("fails on a malformed request timeout rather than silently using the default", func() {
		_, err := host.VerifyStep(ctx, hostTodo(), api.Spec{Budget: api.Budget{Timeout: "later"}})

		Expect(err).To(MatchError(ContainSubstring("timeout")))
	})

	// The verify step runs the definition of done, not an agent turn, so a
	// project that configures no model must still be able to check a todo whose
	// definition of done is fixture steps. Requiring one here is what made
	// `todos check` unusable on a fixture-only project.
	It("resolves without a model when the definition of done has no acceptance criteria", func() {
		todo := hostTodo()
		todo.AcceptanceCriteria = nil
		step := stepNamed(host.Def, lifecycle.StepVerify)

		resolution, err := host.Resolve(ctx, todo, step, lifecycle.RunOptions{})

		Expect(err).NotTo(HaveOccurred())
		Expect(resolution.Class).To(Equal(types.ModeVerify))
		Expect(resolution.Spec.Workflow.Verify.Fixture).To(ContainSubstring("it works"))
	})
})
