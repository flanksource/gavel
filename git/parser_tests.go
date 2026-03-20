package git

import (
	"github.com/flanksource/gavel/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Commit Parser", func() {

	tests := []struct {
		message string
		commit  models.Commit
	}{
		{
			message: "feat(api): subject",
			commit: models.Commit{
				CommitType: models.CommitTypeFeat,
				Scope:      models.ScopeType("api"),
				Subject:    "subject",
			},
		},
		{
			message: "fix: subject",
			commit: models.Commit{
				CommitType: models.CommitTypeFix,
				Scope:      models.ScopeTypeUnknown,
				Subject:    "subject",
			},
		},
		{
			message: "chore(deps): update dependencies",
			commit: models.Commit{
				CommitType: models.CommitTypeChore,
				Scope:      models.ScopeType("deps"),
				Subject:    "update dependencies",
			},
		},
		{
			message: "    fix: corect;y scope fsgroup instruction (#22738)",
			commit: models.Commit{
				CommitType: models.CommitTypeFix,
				Scope:      models.ScopeTypeUnknown,
				Subject:    "corect;y scope fsgroup instruction",
				Reference:  "22738",
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		It("should parse: "+tt.message, func() {
			commit := NewCommit(tt.message)
			Expect(commit.CommitType).To(Equal(tt.commit.CommitType))
			Expect(commit.Scope).To(Equal(tt.commit.Scope))
			Expect(commit.Subject).To(Equal(tt.commit.Subject))
			Expect(commit.Reference).To(Equal(tt.commit.Reference))
		})
	}
})
