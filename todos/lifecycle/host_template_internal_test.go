package lifecycle

import (
	"github.com/flanksource/captain/pkg/api"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("step spec placeholders", func() {
	vars := map[string]any{VarSubject: map[string]any{
		"verification": map[string]any{"exists": true, "document": "# Definition of done\n\n```bash\ntrue\n```"},
		"plan":         map[string]any{"revision": 3},
	}}

	ginkgo.It("substitutes a whole-string placeholder with the value itself", func() {
		spec := &api.Spec{Workflow: &api.Workflow{Verify: &api.Verify{Fixture: "{{subject.verification.document}}"}}}

		expanded, err := expandSpec(spec, vars)

		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(expanded.Workflow.Verify.Fixture).To(gomega.Equal("# Definition of done\n\n```bash\ntrue\n```"))
	})

	ginkgo.It("renders an embedded placeholder as text", func() {
		spec := &api.Spec{Prompt: api.Prompt{System: "plan revision {{ subject.plan.revision }} applies"}}

		expanded, err := expandSpec(spec, vars)

		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(expanded.Prompt.System).To(gomega.Equal("plan revision 3 applies"))
	})

	ginkgo.It("rejects a path the subject does not carry", func() {
		spec := &api.Spec{Workflow: &api.Workflow{Verify: &api.Verify{Fixture: "{{subject.verification.fixture}}"}}}

		_, err := expandSpec(spec, vars)

		gomega.Expect(err).To(gomega.MatchError(gomega.ContainSubstring(`"fixture" is not a known field`)))
	})

	ginkgo.It("leaves a spec without placeholders untouched", func() {
		spec := &api.Spec{Permissions: api.Permissions{Mode: api.PermissionPlan}}

		expanded, err := expandSpec(spec, vars)

		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(expanded).To(gomega.Equal(*spec))
	})

	ginkgo.It("expands a nil spec to the zero spec", func() {
		expanded, err := expandSpec(nil, vars)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(api.IsEmpty(expanded)).To(gomega.BeTrue())
	})
})
