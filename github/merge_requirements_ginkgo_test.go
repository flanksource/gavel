package github

import (
	"encoding/json"

	"github.com/flanksource/gavel/pr/model"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("PR merge requirements", func() {
	It("retains auto-merge state and branch requirements in status output", func() {
		var pr graphQLPR
		Expect(json.Unmarshal([]byte(`{"autoMergeRequest":{"mergeMethod":"REBASE","enabledAt":"2026-11-03T04:00:00Z"},"baseRef":{"branchProtectionRule":{"requiredStatusCheckContexts":["test"],"requiredApprovingReviewCount":2,"requiresApprovingReviews":true,"requiresCommitSignatures":true,"requiresConversationResolution":true,"requiresStrictStatusChecks":true}}}`), &pr)).To(Succeed())
		info := pr.toPRInfo()
		Expect(info.AutoMerge).To(Equal(&model.AutoMergeInfo{Method: "REBASE", EnabledAt: "2026-11-03T04:00:00Z"}))
		Expect(info.BranchProtection).To(Equal(&model.BranchProtectionInfo{RequiredContexts: []string{"test"}, RequiredApprovals: 2, RequiresApprovingReviews: true, RequiresCommitSignatures: true, RequiresConversationResolution: true, RequiresStrictStatusChecks: true}))
	})
})
