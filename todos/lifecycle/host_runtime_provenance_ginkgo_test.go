package lifecycle_test

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/captain/pkg/captainconfig"
	"github.com/flanksource/gavel/todos/lifecycle"
	"github.com/flanksource/gavel/todos/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("lifecycle final runtime provenance", func() {
	var host *lifecycle.Host
	BeforeEach(func() {
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		host = newHost(&fakeProvider{})
		host.Saved = &captainconfig.Config{}
	})

	It("retains authored timeout and empty cwd origins when runtime context fills or normalizes them", func() {
		var request api.Spec
		Expect(json.Unmarshal([]byte(`{"budget":{"timeout":"90s"},"setup":{"cwd":""}}`), &request)).To(Succeed())
		resolution, err := host.Resolve(context.Background(), hostTodo(), stepNamed(host.Def, "plan"), lifecycle.RunOptions{Request: request})
		Expect(err).NotTo(HaveOccurred())
		Expect(resolution.Spec.Budget.Timeout).To(Equal("1m30s"))
		Expect(resolution.Spec.Setup.Cwd).To(Equal(host.WorkDir))
		for _, path := range []string{"/budget/timeout", "/setup/cwd"} {
			Expect(resolution.Provenance[path]).To(Equal(api.FieldProvenance{
				Source:       api.FieldSource{Kind: api.FieldSourceLayer, Name: "request", Key: path},
				NormalizedBy: &api.FieldSource{Kind: api.FieldSourceContext, Name: "lifecycle runtime", Key: path},
			}))
		}
		Expect(resolution.Trace[len(resolution.Trace)-1].Spec.Budget.Timeout).To(Equal("90s"))
		Expect(resolution.Trace[len(resolution.Trace)-1].Spec.Setup.Cwd).To(BeEmpty())
	})

	DescribeTable("fills a missing deadline only after saved defaults and project limits",
		func(saved, limit, timeout string, source api.FieldSourceKind) {
			host.Saved.AI.Timeout = saved
			host.Config.Todos.Timeout = limit
			resolution, err := host.Resolve(context.Background(), hostTodo(), stepNamed(host.Def, "plan"), lifecycle.RunOptions{})
			Expect(err).NotTo(HaveOccurred())
			Expect(resolution.Spec.Budget.Timeout).To(Equal(timeout))
			expected, err := time.ParseDuration(timeout)
			Expect(err).NotTo(HaveOccurred())
			Expect(resolution.Timeout).To(Equal(expected))
			Expect(resolution.Provenance["/budget/timeout"].Source.Kind).To(Equal(source))
		},
		Entry("host default", "", "", "30m0s", api.FieldSourceContext),
		Entry("saved deadline", "40m", "", "40m0s", api.FieldSourceSaved),
		Entry("project cap", "40m", "10m", "10m0s", api.FieldSourceSaved),
	)

	It("removes forbidden commit values and presence from plan provenance without altering the authored trace", func() {
		var request api.Spec
		Expect(json.Unmarshal([]byte(`{"workflow":{"commits":[{"mode":"commit","dryRun":false}]}}`), &request)).To(Succeed())
		resolution, err := host.Resolve(context.Background(), hostTodo(), stepNamed(host.Def, "plan"), lifecycle.RunOptions{Request: request})
		Expect(err).NotTo(HaveOccurred())
		Expect(resolution.Spec.Workflow.Commits).To(BeNil())
		for path := range resolution.Provenance {
			Expect(strings.HasPrefix(path, "/workflow/commits")).To(BeFalse(), path)
		}
		for path := range resolution.Spec.Fields() {
			Expect(strings.HasPrefix(path, "/workflow/commits")).To(BeFalse(), path)
		}
		raw, err := json.Marshal(resolution.Spec)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(raw)).NotTo(ContainSubstring(`"commits"`))
		Expect(resolution.Trace[len(resolution.Trace)-1].Spec.Workflow.Commits).To(HaveLen(1))
	})

	It("attributes the generated conversation and schema while preserving an authored body origin", func() {
		resolution, err := host.Resolve(context.Background(), hostTodo(), stepNamed(host.Def, "plan"), lifecycle.RunOptions{
			Request: api.Spec{Prompt: api.Prompt{User: "Review the scoped change"}},
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(resolution.Prompt).To(ContainSubstring("Review the scoped change"))
		Expect(resolution.Provenance["/prompt/user"].Source.Name).To(Equal("request"))
		Expect(resolution.Provenance["/prompt/user"].NormalizedBy).To(Equal(&api.FieldSource{
			Kind: api.FieldSourceContext, Name: "lifecycle prompt plan", Key: "/prompt/user",
		}))
		for _, path := range []string{"/prompt/source", "/prompt/schemaJSON"} {
			Expect(resolution.Provenance[path].Source).To(Equal(api.FieldSource{
				Kind: api.FieldSourceContext, Name: "lifecycle prompt plan", Key: path,
			}))
		}
	})

	It("attributes a continuation message to the supplied continuation rather than the replaced prompt", func() {
		resolution, err := host.Resolve(context.Background(), hostTodo(), stepNamed(host.Def, "plan"), lifecycle.RunOptions{
			Resume: true, Message: "Continue with the updated criteria", Request: api.Spec{Prompt: api.Prompt{User: "Original request"}},
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(resolution.Prompt).To(Equal("Continue with the updated criteria"))
		Expect(resolution.Provenance["/prompt/user"]).To(Equal(api.FieldProvenance{Source: api.FieldSource{
			Kind: api.FieldSourceContext, Name: "lifecycle continuation", Key: "/prompt/user",
		}}))
	})

	It("keeps the original verification layer while attributing its final grader document", func() {
		host.Saved.AI.DefaultModel = "api:claude-haiku-4-5"
		todo := hostTodo()
		todo.AcceptanceCriteria = []types.AcceptanceCriterion{{Text: "The requested behavior is implemented"}}
		initial, err := host.Context(context.Background(), todo)
		Expect(err).NotTo(HaveOccurred())
		original := initial.Subject["verification"].(map[string]any)["document"].(string)
		resolution, err := host.Resolve(context.Background(), todo, stepNamed(host.Def, "verify"), lifecycle.RunOptions{
			Request: api.Spec{Budget: api.Budget{Cost: 2}},
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(resolution.Spec.Workflow.Verify.Fixture).NotTo(Equal(original))
		Expect(resolution.Trace).To(ContainElement(SatisfyAll(
			HaveField("Name", "lifecycle step verify"),
			HaveField("Spec.Workflow.Verify.Fixture", original),
		)))
		Expect(resolution.Provenance["/workflow/verify/fixture"].Source.Name).To(Equal("lifecycle step verify"))
		Expect(resolution.Provenance["/workflow/verify/fixture"].NormalizedBy).To(Equal(&api.FieldSource{
			Kind: api.FieldSourceContext, Name: "lifecycle verification", Key: "/workflow/verify/fixture",
		}))
	})

	It("accepts catalog effort normalization and preserves the explicit request origin", func() {
		resolution, err := host.Resolve(context.Background(), hostTodo(), stepNamed(host.Def, "plan"), lifecycle.RunOptions{
			Request: api.Spec{Model: api.Model{Name: "claude-sonnet-5", Mode: api.ModeAgent, Effort: api.EffortUltra}},
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(resolution.Spec.Effort).To(Equal(api.EffortMax))
		Expect(resolution.Provenance["/effort"].Source.Name).To(Equal("request"))
		Expect(resolution.Provenance["/effort"].NormalizedBy.Kind).To(Equal(api.FieldSourceCatalog))
	})

	It("rejects a conflicting effort declared inside the same raw request selector", func() {
		_, err := host.Resolve(context.Background(), hostTodo(), stepNamed(host.Def, "plan"), lifecycle.RunOptions{
			Request: api.Spec{Model: api.Model{Name: "agent:claude-sonnet-4-6:high", Effort: api.EffortLow}},
		})
		Expect(err).To(MatchError(ContainSubstring("conflicts with model")))
	})

	It("lets explicit request effort override a compact selector from a lower layer", func() {
		host.Config.Todos.Plan.Spec.Model = api.Model{Name: "agent:claude-sonnet-5:high"}
		resolution, err := host.Resolve(context.Background(), hostTodo(), stepNamed(host.Def, "plan"), lifecycle.RunOptions{
			Request: api.Spec{Model: api.Model{Effort: api.EffortLow}},
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(resolution.Spec.Effort).To(Equal(api.EffortLow))
		Expect(resolution.Provenance["/model"].Source.Name).To(Equal(".gavel.yaml todos.plan"))
		Expect(resolution.Provenance["/effort"].Source.Name).To(Equal("request"))
	})
})
