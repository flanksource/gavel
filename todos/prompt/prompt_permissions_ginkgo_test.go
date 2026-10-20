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
	It("declares auto approval with AskUserQuestion as its only tool policy", func() {
		req, _, err := renderResolvedForTest([]*types.TODO{newTestTODO("edit-parser", "Update the parser")}, Options{Mode: types.ModeRun})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(req.Permissions).To(gomega.Equal(api.Permissions{Mode: api.PermissionAuto, Tools: api.Tools{"AskUserQuestion": api.ToolPolicyDeny}}))
	})

	for _, runtime := range api.AllRuntimes() {
		// API runtimes honour no posture at all, so they never ran todos; every
		// runtime that honoured the former acceptEdits posture must honour this one.
		It("is honoured by "+runtime.String()+" wherever acceptEdits was", func() {
			req, _, err := renderResolvedForTest([]*types.TODO{newTestTODO("edit-parser", "Update the parser")}, Options{Mode: types.ModeRun})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			capabilities := api.PermissionCapabilitiesFor(runtime)
			support := capabilities.ModeSupport(req.Permissions.Mode)
			gomega.Expect(support.Honoured()).To(gomega.Equal(capabilities.ModeSupport(api.PermissionAcceptEdits).Honoured()), "%s: %+v", runtime, support)
		})

		It("passes the tool-policy guard for "+runtime.String(), func() {
			req, _, err := renderResolvedForTest([]*types.TODO{newTestTODO("edit-parser", "Update the parser")}, Options{Mode: types.ModeRun})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			provider, ok := runtime.ModelProvider()
			gomega.Expect(ok).To(gomega.BeTrue())
			gomega.Expect(api.RequireToolPolicySupport(provider, runtime.Mode, req.Permissions)).To(gomega.Succeed())
		})
	}
})

// Triage stays read-only on the code; every TODO read and write goes through
// the todo tools, which validate and record it, never through the gavel CLI.
var _ = Describe("the triage prompt permission posture", func() {
	var permissions api.Permissions

	BeforeEach(func() {
		req, _, err := renderResolvedForTest(
			[]*types.TODO{newTestTODO("triage-parser", "Triage the parser")},
			Options{Prompt: "triage", RawFixture: true},
		)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		permissions = req.Permissions
	})

	It("stays in plan mode", func() {
		gomega.Expect(permissions.Mode).To(gomega.Equal(api.PermissionPlan))
	})

	It("reads backlog candidates through the todo tools, not the gavel CLI", func() {
		for _, allowed := range permissions.Tools.AllowList() {
			gomega.Expect(allowed).NotTo(gomega.ContainSubstring("gavel todos"))
		}
	})

	// PR candidates come from the persisted PR index, which `gavel pr search`
	// reads without touching GitHub. The allow is scoped to that subcommand: an
	// allow on `gavel pr:*` would also cover create, close and merge.
	It("allows searching the PR index", func() {
		gomega.Expect(permissions.Tools.AllowList()).To(gomega.ContainElement("Bash(gavel pr search:*)"))
	})

	It("never allows a TODO- or PR-mutating command", func() {
		for _, allowed := range permissions.Tools.AllowList() {
			gomega.Expect(allowed).NotTo(gomega.SatisfyAny(
				gomega.ContainSubstring("todos edit"),
				gomega.ContainSubstring("todos delete"),
				gomega.ContainSubstring("todos run"),
				gomega.Equal("Bash(gavel todos:*)"),
				gomega.Equal("Bash(gavel pr:*)"),
				gomega.ContainSubstring("pr create"),
				gomega.ContainSubstring("pr close"),
				gomega.Equal("Bash"),
			), "triage writes nothing itself: %s", allowed)
		}
	})
})

// todo prompts ask a human through todo_ask, which parks the question on the
// dashboard; the native AskUserQuestion has no answer path on a todo run.
var _ = DescribeTable("every todo prompt denies the native AskUserQuestion",
	func(name string) {
		req, _, err := renderResolvedForTest([]*types.TODO{newTestTODO("ask-parser", "Ask about the parser")}, Options{Prompt: name, RawFixture: true})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(req.Permissions.Tools.DenyList()).To(gomega.ContainElement("AskUserQuestion"))
	},
	Entry("run", "run"), Entry("plan", "plan"), Entry("triage", "triage"), Entry("triage.new", "triage.new"),
)
