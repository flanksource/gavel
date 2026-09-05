package prompt

import (
	"testing"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/todos/types"
	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

func TestPromptPermissions(t *testing.T) {
	gomega.RegisterFailHandler(Fail)
	RunSpecs(t, "TODO Prompt Permissions")
}

var _ = Describe("the run prompt permission posture", func() {
	It("declares native edit approval without a preset or an explicit tool policy", func() {
		req, _, err := Render([]*types.TODO{newTestTODO("edit-parser", "Update the parser")}, Options{Mode: types.ModeRun})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(req.Permissions).To(gomega.Equal(api.Permissions{Mode: api.PermissionAcceptEdits}))
	})

	for _, runtime := range api.AllRuntimes() {
		It("passes the tool-policy guard for "+runtime.String(), func() {
			req, _, err := Render([]*types.TODO{newTestTODO("edit-parser", "Update the parser")}, Options{Mode: types.ModeRun})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			provider, ok := runtime.ModelProvider()
			gomega.Expect(ok).To(gomega.BeTrue())
			gomega.Expect(api.RequireToolPolicySupport(provider, runtime.Mode, req.Permissions)).To(gomega.Succeed())
		})
	}
})
