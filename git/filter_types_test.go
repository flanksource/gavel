package git_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flanksource/gavel/git"
	"github.com/flanksource/gavel/models"
)

var _ = Describe("Argument Detection", func() {
	Describe("ParseArgs", func() {
		It("should separate commits, ranges, and paths correctly", func() {
			opts := git.HistoryOptions{
				Args: []string{
					"abc1234567890abcdef1234567890abcdef12345", // full SHA
					"abc1234",            // short SHA
					"main..feature",      // range
					"*.go",               // path
					"cmd/*.go",           // path with directory
					"main...develop",     // triple dot range
					"**/*kubernetes*.go", // glob path
				},
			}

			err := opts.ParseArgs()
			Expect(err).ToNot(HaveOccurred())

			Expect(opts.CommitShas).To(ConsistOf(
				"abc1234567890abcdef1234567890abcdef12345",
				"abc1234",
			))

			Expect(opts.CommitRanges).To(ConsistOf(
				"main..feature",
				"main...develop",
			))

			Expect(opts.FilePaths).To(ConsistOf(
				"*.go",
				"cmd/*.go",
				"**/*kubernetes*.go",
			))
		})

		It("should handle branch names as commits", func() {
			opts := git.HistoryOptions{
				Args: []string{"main", "feature-branch", "v1.0.0"},
			}

			err := opts.ParseArgs()
			Expect(err).ToNot(HaveOccurred())

			Expect(opts.CommitShas).To(ConsistOf("main", "feature-branch", "v1.0.0"))
		})

		It("should handle empty Args", func() {
			opts := git.HistoryOptions{
				Args: []string{},
			}

			err := opts.ParseArgs()
			Expect(err).ToNot(HaveOccurred())

			Expect(opts.CommitShas).To(BeEmpty())
			Expect(opts.CommitRanges).To(BeEmpty())
			Expect(opts.FilePaths).To(BeEmpty())
		})

		It("should handle paths with wildcards", func() {
			opts := git.HistoryOptions{
				Args: []string{
					"**/*.yaml",
					"cmd/**/*.go",
					"*.md",
					"test?file.go",
				},
			}

			err := opts.ParseArgs()
			Expect(err).ToNot(HaveOccurred())

			Expect(opts.FilePaths).To(HaveLen(4))
			Expect(opts.CommitShas).To(BeEmpty())
			Expect(opts.CommitRanges).To(BeEmpty())
		})

		It("should handle mixed content", func() {
			opts := git.HistoryOptions{
				Args: []string{
					"abc123",         // commit
					"main..branch",   // range
					"git/*.go",       // path
					"def456",         // commit
					"test/data/*.md", // path
				},
			}

			err := opts.ParseArgs()
			Expect(err).ToNot(HaveOccurred())

			Expect(opts.CommitShas).To(ConsistOf("abc123", "def456"))
			Expect(opts.CommitRanges).To(ConsistOf("main..branch"))
			Expect(opts.FilePaths).To(ConsistOf("git/*.go", "test/data/*.md"))
		})
	})

	Describe("detectArgType", func() {
		DescribeTable("should correctly detect argument types",
			func(arg string, expectedType string) {
				// We can't call detectArgType directly as it's not exported,
				// so we test via ParseArgs
				opts := git.HistoryOptions{Args: []string{arg}}
				err := opts.ParseArgs()
				Expect(err).ToNot(HaveOccurred())

				switch expectedType {
				case "commit":
					Expect(opts.CommitShas).To(ContainElement(arg))
				case "range":
					Expect(opts.CommitRanges).To(ContainElement(arg))
				case "path":
					Expect(opts.FilePaths).To(ContainElement(arg))
				}
			},
			// Commit SHAs
			Entry("full SHA (40 chars)", "abc1234567890abcdef1234567890abcdef12345", "commit"),
			Entry("short SHA (7 chars)", "abc1234", "commit"),
			Entry("short SHA (10 chars)", "abc1234567", "commit"),
			Entry("branch name", "main", "commit"),
			Entry("branch with dash", "feature-branch", "commit"),
			Entry("tag", "v1.0.0", "commit"),

			// Commit ranges
			Entry("two-dot range", "main..feature", "range"),
			Entry("three-dot range", "main...develop", "range"),
			Entry("SHA range", "abc123..def456", "range"),

			// File paths
			Entry("wildcard", "*.go", "path"),
			Entry("glob pattern", "**/*.yaml", "path"),
			Entry("path with directory", "cmd/test.go", "path"),
			Entry("path with wildcard", "cmd/*.go", "path"),
			Entry("question mark wildcard", "test?.go", "path"),
			Entry("nested path", "git/kubernetes/*.go", "path"),
		)
	})
})

var _ = Describe("AnalyzeOptions.Matches", func() {
	commit := models.Commit{Hash: "abc123", Author: models.Author{Name: "dev"}, CommitType: models.CommitTypeFeat}

	It("should match when scope type filter matches change scope", func() {
		opts := git.AnalyzeOptions{ScopeTypes: []string{"backend"}}
		change := models.CommitChange{File: "main.go", Scope: models.Scopes{models.ScopeType("backend")}}
		Expect(opts.Matches(commit, change)).To(BeTrue())
	})

	It("should not match when scope type filter doesn't match", func() {
		opts := git.AnalyzeOptions{ScopeTypes: []string{"frontend"}}
		change := models.CommitChange{File: "main.go", Scope: models.Scopes{models.ScopeType("backend")}}
		Expect(opts.Matches(commit, change)).To(BeFalse())
	})

	It("should match when technology filter matches", func() {
		opts := git.AnalyzeOptions{Technologies: []string{"go"}}
		change := models.CommitChange{File: "main.go", Tech: []models.ScopeTechnology{models.Go}}
		Expect(opts.Matches(commit, change)).To(BeTrue())
	})

	It("should not match when technology filter doesn't match", func() {
		opts := git.AnalyzeOptions{Technologies: []string{"python"}}
		change := models.CommitChange{File: "main.go", Tech: []models.ScopeTechnology{models.Go}}
		Expect(opts.Matches(commit, change)).To(BeFalse())
	})

	It("should match commit type filter", func() {
		opts := git.AnalyzeOptions{CommitTypes: []string{"feat", "fix"}}
		change := models.CommitChange{File: "main.go"}
		Expect(opts.Matches(commit, change)).To(BeTrue())
	})

	It("should not match when commit type filter excludes", func() {
		opts := git.AnalyzeOptions{CommitTypes: []string{"chore"}}
		change := models.CommitChange{File: "main.go"}
		Expect(opts.Matches(commit, change)).To(BeFalse())
	})

	It("should support glob patterns for scope types", func() {
		opts := git.AnalyzeOptions{ScopeTypes: []string{"back*"}}
		change := models.CommitChange{File: "main.go", Scope: models.Scopes{models.ScopeType("backend")}}
		Expect(opts.Matches(commit, change)).To(BeTrue())
	})

	It("should support negation patterns for commit types", func() {
		opts := git.AnalyzeOptions{CommitTypes: []string{"!chore"}}
		change := models.CommitChange{File: "main.go"}

		Expect(opts.Matches(commit, change)).To(BeTrue())

		choreCommit := models.Commit{Hash: "def456", CommitType: models.CommitTypeChore}
		Expect(opts.Matches(choreCommit, change)).To(BeFalse())
	})

	It("should match everything when filters are empty", func() {
		opts := git.AnalyzeOptions{}
		change := models.CommitChange{File: "main.go", Scope: models.Scopes{models.ScopeType("backend")}, Tech: []models.ScopeTechnology{models.Go}}
		Expect(opts.Matches(commit, change)).To(BeTrue())
	})
})

var _ = Describe("HistoryOptions.Matches", func() {
	baseDate := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)

	It("should filter by author", func() {
		opts := git.HistoryOptions{Author: []string{"john"}}
		commit := models.Commit{Hash: "abc123", Author: models.Author{Name: "john doe", Email: "john@example.com"}}
		Expect(opts.Matches(commit)).To(BeTrue())

		commit2 := models.Commit{Hash: "def456", Author: models.Author{Name: "jane doe"}}
		Expect(opts.Matches(commit2)).To(BeFalse())
	})

	It("should filter by message glob", func() {
		opts := git.HistoryOptions{Message: "feat:*"}
		commit := models.Commit{Hash: "abc123", Author: models.Author{Name: "dev"}, Subject: "feat: new feature"}
		Expect(opts.Matches(commit)).To(BeTrue())

		commit2 := models.Commit{Hash: "def456", Author: models.Author{Name: "dev"}, Subject: "fix: bug fix"}
		Expect(opts.Matches(commit2)).To(BeFalse())
	})

	It("should filter by since date", func() {
		opts := git.HistoryOptions{Since: baseDate}
		commit := models.Commit{Hash: "abc123", Author: models.Author{Name: "dev", Date: baseDate.Add(24 * time.Hour)}}
		Expect(opts.Matches(commit)).To(BeTrue())

		commit2 := models.Commit{Hash: "def456", Author: models.Author{Name: "dev", Date: baseDate.Add(-24 * time.Hour)}}
		Expect(opts.Matches(commit2)).To(BeFalse())
	})

	It("should filter by until date", func() {
		opts := git.HistoryOptions{Until: baseDate}
		commit := models.Commit{Hash: "abc123", Author: models.Author{Name: "dev", Date: baseDate.Add(-24 * time.Hour)}}
		Expect(opts.Matches(commit)).To(BeTrue())

		commit2 := models.Commit{Hash: "def456", Author: models.Author{Name: "dev", Date: baseDate.Add(24 * time.Hour)}}
		Expect(opts.Matches(commit2)).To(BeFalse())
	})

	It("should apply all filters together", func() {
		opts := git.HistoryOptions{
			Author:  []string{"dev*"},
			Message: "feat:*",
			Since:   baseDate.Add(-7 * 24 * time.Hour),
			Until:   baseDate,
		}

		commit := models.Commit{
			Hash:    "abc123",
			Author:  models.Author{Name: "developer", Date: baseDate.Add(-2 * 24 * time.Hour)},
			Subject: "feat: add feature",
		}
		Expect(opts.Matches(commit)).To(BeTrue())

		wrongAuthor := models.Commit{
			Hash:    "def456",
			Author:  models.Author{Name: "bot", Date: baseDate.Add(-2 * 24 * time.Hour)},
			Subject: "feat: add feature",
		}
		Expect(opts.Matches(wrongAuthor)).To(BeFalse())
	})

	It("should match everything when no filters set", func() {
		opts := git.HistoryOptions{}
		commit := models.Commit{Hash: "abc123", Author: models.Author{Name: "anyone"}, Subject: "anything"}
		Expect(opts.Matches(commit)).To(BeTrue())
	})
})
