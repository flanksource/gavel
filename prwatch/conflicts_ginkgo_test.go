package prwatch

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flanksource/gavel/github"
)

// greenChecks is an all-passing rollup, so every conflict assertion below is
// about the conflict alone and never rides on a failing check.
var greenChecks = github.StatusChecks{
	{Name: "lint", Status: "COMPLETED", Conclusion: "SUCCESS"},
	{Name: "test", Status: "COMPLETED", Conclusion: "SUCCESS"},
}

func conflictingResult(report *github.MergeConflictReport) *PRWatchResult {
	return &PRWatchResult{
		PR: &github.PRInfo{
			Number: 7, Title: "feat: widget", Author: github.PRAuthor{Login: "alice"},
			BaseRefName: "main", HeadRefName: "feat/widget",
			State: "OPEN", Mergeable: "CONFLICTING", MergeState: "DIRTY",
			StatusCheckRollup: greenChecks,
		},
		Conflicts: report,
	}
}

var _ = Describe("a PR that conflicts with its base", func() {
	report := &github.MergeConflictReport{
		BaseRefName: "main", HeadRefName: "feat/widget",
		BaseOID: "aaaaaaaa", HeadOID: "bbbbbbbb", MergeBase: "cccccccc",
		Files: []github.MergeConflict{
			{Path: "go.mod", Kind: "content"},
			{Path: "docs/api.md", Kind: "modify/delete"},
		},
	}

	It("exits 1 even when every check passed", func() {
		Expect(statusExitCode(conflictingResult(report))).To(Equal(1),
			"an all-green PR that cannot merge is not a success")
	})

	It("still exits 0 once the conflict is resolved", func() {
		result := conflictingResult(nil)
		result.PR.Mergeable = "MERGEABLE"
		result.PR.MergeState = "CLEAN"

		Expect(statusExitCode(result)).To(Equal(0))
	})

	It("ignores a stale CONFLICTING verdict on a PR that already merged", func() {
		result := conflictingResult(nil)
		result.PR.State = "MERGED"

		Expect(result.HasMergeConflict()).To(BeFalse())
		Expect(statusExitCode(result)).To(Equal(0))
	})

	Describe("the rendered status", func() {
		var rendered string

		BeforeEach(func() {
			rendered = conflictingResult(report).Pretty().String()
		})

		It("names the blocker beside the mergeable verdict", func() {
			Expect(rendered).To(ContainSubstring("Mergeable: CONFLICTING (conflicts with base)"))
		})

		It("lists each conflicting path with its conflict kind", func() {
			Expect(rendered).To(ContainSubstring("go.mod (content)"))
			Expect(rendered).To(ContainSubstring("docs/api.md (modify/delete)"))
			Expect(rendered).To(ContainSubstring("Merge conflicts"))
		})

		It("shows the commands that resolve it locally", func() {
			Expect(rendered).To(ContainSubstring("Resolve locally"))
			Expect(rendered).To(ContainSubstring("git merge origin/main"))
		})

		It("keeps rendering the checks below the conflict", func() {
			Expect(rendered).To(ContainSubstring("Workflows:"))
			Expect(rendered).To(ContainSubstring("lint"))
		})
	})

	It("explains itself when the conflicting paths could not be listed", func() {
		unavailable := &github.MergeConflictReport{
			BaseRefName: "main", HeadRefName: "feat/widget",
			Unavailable: "no git remote points at acme/widget",
		}

		rendered := conflictingResult(unavailable).Pretty().String()

		Expect(rendered).To(ContainSubstring("no git remote points at acme/widget"))
	})

	It("does not stop --fail-fast, which waits on checks and jobs only", func() {
		Expect(conflictingResult(report).HasTerminalFailure()).To(BeFalse(),
			"a conflict is fixed by pushing a merge, so --follow keeps watching the checks")
	})

	It("renders no conflict section for a mergeable PR", func() {
		result := &PRWatchResult{PR: &github.PRInfo{
			Number: 8, BaseRefName: "main", HeadRefName: "feat/other",
			State: "OPEN", Mergeable: "MERGEABLE", StatusCheckRollup: greenChecks,
		}}

		Expect(result.Pretty().String()).ToNot(ContainSubstring("Merge conflicts"))
	})
})
