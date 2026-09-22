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
		req, _, err := renderResolvedForTest([]*types.TODO{newTestTODO("edit-parser", "Update the parser")}, Options{Mode: types.ModeRun})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(req.Permissions).To(gomega.Equal(api.Permissions{Mode: api.PermissionAcceptEdits}))
	})

	for _, runtime := range api.AllRuntimes() {
		It("passes the tool-policy guard for "+runtime.String(), func() {
			req, _, err := renderResolvedForTest([]*types.TODO{newTestTODO("edit-parser", "Update the parser")}, Options{Mode: types.ModeRun})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			provider, ok := runtime.ModelProvider()
			gomega.Expect(ok).To(gomega.BeTrue())
			gomega.Expect(api.RequireToolPolicySupport(provider, runtime.Mode, req.Permissions)).To(gomega.Succeed())
		})
	}
})

// Triage proposes and gavel writes, so the agent stays read-only — but deciding
// two TODOs are the same work needs the other one's body, and the backlog index
// carries only an excerpt. The two lookups are scoped to their subcommands: an
// allow on `gavel todos:*` would also cover edit, delete and run.
var _ = Describe("the triage prompt permission posture", func() {
	var permissions api.Permissions

	BeforeEach(func() {
		req, _, err := renderResolvedForTest(
			[]*types.TODO{newTestTODO("triage-parser", "Triage the parser")},
			Options{Prompt: "triage", Envelope: EnvelopeTriage},
		)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		permissions = req.Permissions
	})

	It("stays in plan mode", func() {
		gomega.Expect(permissions.Mode).To(gomega.Equal(api.PermissionPlan))
	})

	It("allows reading a backlog candidate in full", func() {
		gomega.Expect(permissions.Tools.AllowList()).To(gomega.ContainElements(
			"Bash(gavel todos get:*)", "Bash(gavel todos list:*)"))
	})

	It("never allows a TODO-mutating command", func() {
		for _, allowed := range permissions.Tools.AllowList() {
			gomega.Expect(allowed).NotTo(gomega.SatisfyAny(
				gomega.ContainSubstring("todos edit"),
				gomega.ContainSubstring("todos delete"),
				gomega.ContainSubstring("todos run"),
				gomega.Equal("Bash(gavel todos:*)"),
				gomega.Equal("Bash"),
			), "triage writes nothing itself: %s", allowed)
		}
	})
})
