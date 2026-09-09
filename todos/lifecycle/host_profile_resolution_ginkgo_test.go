package lifecycle_test

import (
	"context"
	"os"
	"path/filepath"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/captain/pkg/runtimeprofiles"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/lifecycle"
	"github.com/flanksource/gavel/todos/types"
	"github.com/flanksource/gavel/verify"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Host profile resolution", func() {
	var host *lifecycle.Host
	BeforeEach(func() {
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		host = newHost(&fakeProvider{plan: todos.PlanState{Exists: true, Approved: true, Content: "# Plan"}})
		dir := GinkgoT().TempDir()
		for name, model := range map[string]string{"requested": "", "pinned": "", "step-default": "", "global-default": "", "grader-default": "api:haiku", "grader-request": "api:sonnet"} {
			Expect(os.WriteFile(filepath.Join(dir, name+".yaml"), []byte("name: "+name+"\nspec:\n  model: "+model+"\n  budget: {cost: 4}\n"), 0600)).To(Succeed())
		}
		source, err := runtimeprofiles.NewFileSource(runtimeprofiles.FileSourceOptions{Kind: runtimeprofiles.KindProfile, Dir: dir})
		Expect(err).NotTo(HaveOccurred())
		catalog, err := runtimeprofiles.NewCatalog(source)
		Expect(err).NotTo(HaveOccurred())
		host.Catalog = func(context.Context) (*runtimeprofiles.Catalog, error) { return catalog, nil }
		host.Config.Todos.RuntimeProfile = "global-default"
	})

	It("carries the requested profile into the prepared request and provenance", func() {
		resolution, err := host.Resolve(context.Background(), hostTodo(), stepNamed(host.Def, "run"), lifecycle.RunOptions{
			RuntimeProfile: "requested", Request: api.Spec{Budget: api.Budget{Cost: 7}},
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(resolution.RuntimeProfile).NotTo(BeNil())
		Expect(resolution.RuntimeProfile.Profile.Name).To(Equal("requested"))
		Expect(resolution.Spec.Budget.Cost).To(Equal(float64(7)))
		Expect(resolution.Trace).To(ContainElement(And(HaveField("Source", api.SpecLayerSourceProfile), HaveField("Name", "requested run spec"))))
	})

	It("selects a prompt pin above the per-step and global defaults", func() {
		path := filepath.Join(host.WorkDir, "run.prompt")
		Expect(os.WriteFile(path, []byte("---\nruntimeProfile: pinned\n---\n{{{body}}}\n"), 0600)).To(Succeed())
		host.Config.Todos.Run = verify.PromptSpec{File: path, RuntimeProfile: "step-default"}
		resolution, err := host.Resolve(context.Background(), hostTodo(), stepNamed(host.Def, "run"), lifecycle.RunOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(resolution.RuntimeProfile).NotTo(BeNil())
		Expect(resolution.RuntimeProfile.Profile.Name).To(Equal("pinned"))
		Expect(resolution.Spec.Budget.Cost).To(Equal(float64(4)))
	})

	It("uses per-step defaults for verify without requiring an agent model", func() {
		host.Config.AI = api.Spec{}
		host.Config.Todos.Verify.RuntimeProfile = "step-default"
		resolution, err := host.Resolve(context.Background(), hostTodo(), stepNamed(host.Def, "verify"), lifecycle.RunOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(resolution.RuntimeProfile).NotTo(BeNil())
		Expect(resolution.RuntimeProfile.Profile.Name).To(Equal("step-default"))
		Expect(resolution.Spec.Budget.Cost).To(Equal(float64(4)))
	})

	It("reports the same configured profile and budget in dialog defaults and resolution", func() {
		host.Config.Todos.Run.RuntimeProfile = "step-default"
		step := stepNamed(host.Def, "run")
		defaults, err := host.StepDefaults(context.Background(), step)
		Expect(err).NotTo(HaveOccurred())
		resolution, err := host.Resolve(context.Background(), hostTodo(), step, lifecycle.RunOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(defaults.RuntimeProfile).To(Equal(resolution.RuntimeProfile.Profile.ID))
		Expect(defaults.Spec.Budget.Cost).To(Equal(float64(4)))
		Expect(resolution.Spec.Budget.Cost).To(Equal(float64(4)))
	})

	It("embeds the requested verification profile in the checklist that actually runs", func() {
		host.Config.Todos.Verify.RuntimeProfile = "grader-default"
		todo := hostTodo()
		todo.AcceptanceCriteria = []types.AcceptanceCriterion{{Text: "The chart renders"}}
		resolution, err := host.Resolve(context.Background(), todo, stepNamed(host.Def, "verify"), lifecycle.RunOptions{RuntimeProfile: "grader-request"})
		Expect(err).NotTo(HaveOccurred())
		Expect(resolution.RuntimeProfile.Profile.Name).To(Equal("grader-request"))
		Expect(resolution.Spec.Workflow.Verify.Fixture).To(ContainSubstring("model: " + resolution.Spec.Name))
		Expect(resolution.Spec.Workflow.Verify.Fixture).To(ContainSubstring("mode: api"))
	})

	It("accepts a requested grader profile when the configured default has no model", func() {
		host.Config.AI = api.Spec{}
		todo := hostTodo()
		todo.AcceptanceCriteria = []types.AcceptanceCriterion{{Text: "The chart renders"}}
		resolution, err := host.Resolve(context.Background(), todo, stepNamed(host.Def, "verify"), lifecycle.RunOptions{RuntimeProfile: "grader-request"})
		Expect(err).NotTo(HaveOccurred())
		Expect(resolution.Spec.Workflow.Verify.Fixture).To(ContainSubstring("model: " + resolution.Spec.Name))
	})

	It("preserves an explicit custom verification document", func() {
		const custom = "### command: custom check\n\n```bash\ntrue\n```"
		resolution, err := host.Resolve(context.Background(), hostTodo(), stepNamed(host.Def, "verify"), lifecycle.RunOptions{
			RuntimeProfile: "grader-request", Request: api.Spec{Workflow: &api.Workflow{Verify: &api.Verify{Fixture: custom}}},
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(resolution.Spec.Workflow.Verify.Fixture).To(Equal(custom))
	})

	It("keeps generation profiles separate from the configured grader", func() {
		host.Config.Todos.Verify.RuntimeProfile = "grader-default"
		todo := hostTodo()
		todo.AcceptanceCriteria = []types.AcceptanceCriterion{{Text: "The chart renders"}}
		grader, err := host.VerifyDocument(context.Background(), todo)
		Expect(err).NotTo(HaveOccurred())
		resolution, err := host.Resolve(context.Background(), todo, stepNamed(host.Def, "run"), lifecycle.RunOptions{RuntimeProfile: "grader-request"})
		Expect(err).NotTo(HaveOccurred())
		Expect(resolution.RuntimeProfile.Profile.Name).To(Equal("grader-request"))
		Expect(resolution.Spec.Workflow.Verify.Fixture).To(Equal(grader.Fixture))
	})
})
