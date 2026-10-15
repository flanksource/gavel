package lifecycle

import (
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/pr/model"
	"github.com/flanksource/gavel/pr/provider"
	todoprompt "github.com/flanksource/gavel/todos/prompt"
	"github.com/flanksource/gavel/todos/types"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"strings"
)

var _ = ginkgo.Describe("PR TODO lifecycle configuration", func() {
	ginkgo.DescribeTable("configures the source head without read-only commit policies", func(class types.RunMode, writes bool) {
		prepared := &preparedStep{class: class, request: api.Spec{Prompt: api.Prompt{User: "Fix the parser."}}}
		state := &prExecution{target: &provider.RunTarget{HeadSHA: strings.Repeat("a", 40), Branch: "topic", RepositoryURL: "https://github.com/acme/fork"}, snapshot: &model.StatusSnapshot{PR: &model.PRInfo{Number: 27, URL: "https://github.com/acme/service/pull/27"}}, workDir: "/repo", write: writes}
		gomega.Expect(state.configure(prepared)).To(gomega.Succeed())
		gomega.Expect(prepared.request.Prompt.User).To(gomega.Equal("Fix the parser."), "PR status is read live through gavel pr status, never dumped into the prompt")
		gomega.Expect(prepared.request.Setup).NotTo(gomega.BeNil())
		gomega.Expect(prepared.request.Setup.Checkout.Worktree.Base).To(gomega.Equal(strings.Repeat("a", 40)))
		if writes {
			gomega.Expect(prepared.request.Workflow.Commits).To(gomega.Equal([]api.Commit{{On: api.CommitOnTurn, Mode: api.CommitModeCommit, Stage: api.CommitStageWorktree, Gates: api.CommitGatesFull}}))
			gomega.Expect(string(prepared.request.Prompt.SchemaJSON)).To(gomega.ContainSubstring("addressedThreadIds"))
			gomega.Expect(prepared.request.Prompt.Metadata).To(gomega.HaveKeyWithValue(todoprompt.PRWriteMetadataKey, "true"))
		} else {
			gomega.Expect(prepared.request.Prompt.Metadata).NotTo(gomega.HaveKey(todoprompt.PRWriteMetadataKey))
			gomega.Expect(prepared.request.Workflow == nil || len(prepared.request.Workflow.Commits) == 0).To(gomega.BeTrue())
			gomega.Expect(prepared.verifyTarget.SHA).To(gomega.Equal(strings.Repeat("a", 40)))
		}
	}, ginkgo.Entry("implementation", types.ModeRun, true), ginkgo.Entry("verification", types.ModeVerify, false))
})
