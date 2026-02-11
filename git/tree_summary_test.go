package git_test

import (
	"strings"

	. "github.com/flanksource/gavel/git"
	. "github.com/flanksource/gavel/models"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Count", func() {
	Describe("Add", func() {
		It("should merge statistics correctly", func() {
			c1 := Count{
				Adds:    100,
				Dels:    50,
				Commits: 5,
				Files:   3,
				Scopes: map[ScopeType]int{
					ScopeTypeApp:  2,
					ScopeTypeTest: 1,
				},
				CommitTypes: map[CommitType]int{
					CommitTypeFeat: 3,
					CommitTypeFix:  2,
				},
				Tech: map[ScopeTechnology]int{
					Go:     4,
					Docker: 1,
				},
			}

			c2 := Count{
				Adds:    200,
				Dels:    75,
				Commits: 3,
				Files:   2,
				Scopes: map[ScopeType]int{
					ScopeTypeApp: 1,
					ScopeTypeCI:  2,
				},
				CommitTypes: map[CommitType]int{
					CommitTypeFeat:  1,
					CommitTypeChore: 2,
				},
				Tech: map[ScopeTechnology]int{
					Go:         2,
					Kubernetes: 1,
				},
			}

			c1.Add(c2)

			Expect(c1.Adds).To(Equal(300))
			Expect(c1.Dels).To(Equal(125))
			Expect(c1.Commits).To(Equal(8))
			Expect(c1.Files).To(Equal(5))
			Expect(c1.Scopes[ScopeTypeApp]).To(Equal(3))
			Expect(c1.Scopes[ScopeTypeTest]).To(Equal(1))
			Expect(c1.Scopes[ScopeTypeCI]).To(Equal(2))
			Expect(c1.CommitTypes[CommitTypeFeat]).To(Equal(4))
			Expect(c1.CommitTypes[CommitTypeFix]).To(Equal(2))
			Expect(c1.CommitTypes[CommitTypeChore]).To(Equal(2))
			Expect(c1.Tech[Go]).To(Equal(6))
			Expect(c1.Tech[Docker]).To(Equal(1))
			Expect(c1.Tech[Kubernetes]).To(Equal(1))
		})
	})

	Describe("Total", func() {
		It("should return sum of adds and dels", func() {
			c := Count{
				Adds: 150,
				Dels: 75,
			}

			Expect(c.Total()).To(Equal(225))
		})

		It("should return zero for empty count", func() {
			c := Count{}
			Expect(c.Total()).To(Equal(0))
		})
	})
})

var _ = Describe("GitSummary", func() {
	Describe("NewGitSummary", func() {
		It("should create root node with initialized maps", func() {
			root := NewGitSummary(".")

			Expect(root.Path).To(Equal("."))
			Expect(root.Scopes).ToNot(BeNil())
			Expect(root.CommitTypes).ToNot(BeNil())
			Expect(root.Tech).ToNot(BeNil())
			Expect(root.Children).ToNot(BeNil())
			Expect(root.Children).To(BeEmpty())
		})
	})

	Describe("AddFile", func() {
		It("should add file node with statistics", func() {
			root := NewGitSummary(".")
			changes := []CommitChange{
				{
					File:  "main.go",
					Adds:  100,
					Dels:  50,
					Scope: Scopes{ScopeTypeApp},
					Tech:  []ScopeTechnology{Go},
				},
			}

			root.AddFile("main.go", changes)

			Expect(root.Children).To(HaveLen(1))
			Expect(root.Children[0].Path).To(Equal("main.go"))
			Expect(root.Children[0].Adds).To(Equal(100))
			Expect(root.Children[0].Dels).To(Equal(50))
			Expect(root.Children[0].Commits).To(Equal(1))
			Expect(root.Children[0].Files).To(Equal(1))
		})

		It("should create nested directory structure", func() {
			root := NewGitSummary(".")
			changes := []CommitChange{
				{
					File:  "src/api/handler.go",
					Adds:  100,
					Dels:  50,
					Scope: Scopes{ScopeTypeApp},
					Tech:  []ScopeTechnology{Go},
				},
			}

			root.AddFile("src/api/handler.go", changes)

			Expect(root.Children).To(HaveLen(1))
			Expect(root.Children[0].Path).To(Equal("src"))
			Expect(root.Children[0].Children).To(HaveLen(1))
			Expect(root.Children[0].Children[0].Path).To(Equal("api"))
			Expect(root.Children[0].Children[0].Children).To(HaveLen(1))
			Expect(root.Children[0].Children[0].Children[0].Path).To(Equal("handler.go"))
		})

		It("should aggregate statistics to parent directories", func() {
			root := NewGitSummary(".")
			changes := []CommitChange{
				{
					File:  "src/main.go",
					Adds:  100,
					Dels:  50,
					Scope: Scopes{ScopeTypeApp},
					Tech:  []ScopeTechnology{Go},
				},
			}

			root.AddFile("src/main.go", changes)

			// Check file node stats
			srcDir := &root.Children[0]
			Expect(srcDir.Path).To(Equal("src"))

			// Parent directory should have aggregated stats
			Expect(srcDir.Adds).To(Equal(100))
			Expect(srcDir.Dels).To(Equal(50))
			Expect(srcDir.Commits).To(Equal(1))
		})
	})

	Describe("BuildFromAnalyses", func() {
		It("should build tree from single commit", func() {
			root := NewGitSummary(".")
			analyses := CommitAnalyses{
				{
					Commit: Commit{
						Hash:       "abc123",
						CommitType: CommitTypeFeat,
						Scope:      ScopeTypeApp,
					},
					Changes: []CommitChange{
						{
							File:  "main.go",
							Adds:  100,
							Dels:  50,
							Scope: Scopes{ScopeTypeApp},
							Tech:  []ScopeTechnology{Go},
						},
					},
				},
			}

			root.BuildFromAnalyses(analyses)

			Expect(root.Children).To(HaveLen(1))
			Expect(root.Children[0].Path).To(Equal("main.go"))
			Expect(root.Children[0].Adds).To(Equal(100))
			Expect(root.Children[0].Dels).To(Equal(50))
		})

		It("should aggregate multiple files in same directory", func() {
			root := NewGitSummary(".")
			analyses := CommitAnalyses{
				{
					Changes: []CommitChange{
						{
							File:  "src/file1.go",
							Adds:  100,
							Dels:  50,
							Scope: Scopes{ScopeTypeApp},
							Tech:  []ScopeTechnology{Go},
						},
						{
							File:  "src/file2.go",
							Adds:  200,
							Dels:  75,
							Scope: Scopes{ScopeTypeApp},
							Tech:  []ScopeTechnology{Go},
						},
					},
				},
			}

			root.BuildFromAnalyses(analyses)

			Expect(root.Children).To(HaveLen(1))
			srcDir := &root.Children[0]
			Expect(srcDir.Path).To(Equal("src"))
			Expect(srcDir.Adds).To(Equal(300))
			Expect(srcDir.Dels).To(Equal(125))
			Expect(srcDir.Children).To(HaveLen(2))
		})

		It("should sort children by change volume", func() {
			root := NewGitSummary(".")
			analyses := CommitAnalyses{
				{
					Changes: []CommitChange{
						{
							File: "small.go",
							Adds: 10,
							Dels: 5,
						},
						{
							File: "large.go",
							Adds: 500,
							Dels: 200,
						},
						{
							File: "medium.go",
							Adds: 100,
							Dels: 50,
						},
					},
				},
			}

			root.BuildFromAnalyses(analyses)

			Expect(root.Children).To(HaveLen(3))
			Expect(root.Children[0].Path).To(Equal("large.go"))
			Expect(root.Children[1].Path).To(Equal("medium.go"))
			Expect(root.Children[2].Path).To(Equal("small.go"))
		})

		It("should handle nested directory structures with collapsing", func() {
			root := NewGitSummary(".")
			analyses := CommitAnalyses{
				{
					Changes: []CommitChange{
						{
							File:  "src/api/handlers/user.go",
							Adds:  150,
							Dels:  75,
							Scope: Scopes{ScopeTypeApp},
							Tech:  []ScopeTechnology{Go},
						},
					},
				},
			}

			root.BuildFromAnalyses(analyses)

			Expect(root.Children).To(HaveLen(1))

			// After collapse, path should be "src/api/handlers"
			collapsed := &root.Children[0]
			Expect(collapsed.Path).To(Equal("src/api/handlers"))
			Expect(collapsed.Children).To(HaveLen(1))

			user := &collapsed.Children[0]
			Expect(user.Path).To(Equal("user.go"))
			Expect(user.Adds).To(Equal(150))
			Expect(user.Dels).To(Equal(75))

			// Verify aggregation to collapsed parent
			Expect(collapsed.Adds).To(Equal(150))
			Expect(collapsed.Dels).To(Equal(75))
		})
	})

	Describe("sorting by change volume", func() {
		It("should sort files by total changes descending", func() {
			root := NewGitSummary(".")
			analyses := CommitAnalyses{
				{
					Changes: []CommitChange{
						{
							File: "small.go",
							Adds: 10,
							Dels: 5,
						},
						{
							File: "large.go",
							Adds: 500,
							Dels: 200,
						},
						{
							File: "medium.go",
							Adds: 100,
							Dels: 50,
						},
					},
				},
			}

			root.BuildFromAnalyses(analyses)

			Expect(root.Children).To(HaveLen(3))
			Expect(root.Children[0].Path).To(Equal("large.go"))
			Expect(root.Children[1].Path).To(Equal("medium.go"))
			Expect(root.Children[2].Path).To(Equal("small.go"))
		})
	})

	Describe("CollapseChains", func() {
		It("should automatically collapse single-child chains during BuildFromAnalyses", func() {
			root := NewGitSummary(".")
			analyses := CommitAnalyses{
				{
					Changes: []CommitChange{
						{
							File:  "src/api/handlers/user.go",
							Adds:  150,
							Dels:  75,
							Scope: Scopes{ScopeTypeApp},
							Tech:  []ScopeTechnology{Go},
						},
					},
				},
			}

			root.BuildFromAnalyses(analyses)

			// After BuildFromAnalyses calls CollapseChains, should have collapsed path
			Expect(root.Children).To(HaveLen(1))
			Expect(root.Children[0].Path).To(Equal("src/api/handlers"))
			Expect(root.Children[0].Children).To(HaveLen(1))
			Expect(root.Children[0].Children[0].Path).To(Equal("user.go"))
		})

		It("should not collapse when directory has multiple children", func() {
			root := NewGitSummary(".")
			analyses := CommitAnalyses{
				{
					Changes: []CommitChange{
						{
							File:  "src/api/handler1.go",
							Adds:  100,
							Dels:  50,
							Scope: Scopes{ScopeTypeApp},
							Tech:  []ScopeTechnology{Go},
						},
						{
							File:  "src/api/handler2.go",
							Adds:  200,
							Dels:  75,
							Scope: Scopes{ScopeTypeApp},
							Tech:  []ScopeTechnology{Go},
						},
					},
				},
			}

			root.BuildFromAnalyses(analyses)

			// Should have src/ and api/ as separate levels because api has 2 children
			Expect(root.Children).To(HaveLen(1))
			Expect(root.Children[0].Path).To(Equal("src/api"))
			Expect(root.Children[0].Children).To(HaveLen(2))
		})

		It("should preserve statistics during collapse", func() {
			root := NewGitSummary(".")
			analyses := CommitAnalyses{
				{
					Changes: []CommitChange{
						{
							File:  "deep/nested/path/file.go",
							Adds:  100,
							Dels:  50,
							Scope: Scopes{ScopeTypeApp},
							Tech:  []ScopeTechnology{Go},
						},
					},
				},
			}

			root.BuildFromAnalyses(analyses)

			// Verify collapsed path has correct stats
			Expect(root.Children).To(HaveLen(1))
			collapsed := &root.Children[0]
			Expect(collapsed.Adds).To(Equal(100))
			Expect(collapsed.Dels).To(Equal(50))
			Expect(collapsed.Commits).To(Equal(1))
		})
	})

	Describe("parsing commit with parentheses in subject", func() {
		It("should preserve parentheses in subject text", func() {
			message := "feat(api): add endpoint (with middleware)"
			commit := NewCommit(message)

			Expect(commit.CommitType).To(Equal(CommitTypeFeat))
			Expect(commit.Scope).To(Equal(ScopeType("api")))
			Expect(commit.Subject).To(Equal("add endpoint (with middleware)"))
			// Verify closing parenthesis is preserved
			Expect(strings.Count(commit.Subject, "(")).To(Equal(strings.Count(commit.Subject, ")")))
		})

		It("should handle multiple parentheses pairs", func() {
			message := "fix(db): update query (for users) and cache (for sessions)"
			commit := NewCommit(message)

			Expect(commit.CommitType).To(Equal(CommitTypeFix))
			Expect(commit.Scope).To(Equal(ScopeType("db")))
			Expect(commit.Subject).To(Equal("update query (for users) and cache (for sessions)"))
			Expect(strings.Count(commit.Subject, "(")).To(Equal(2))
			Expect(strings.Count(commit.Subject, ")")).To(Equal(2))
		})
	})

	Describe("Pretty", func() {
		It("should display file with inline stats", func() {
			root := NewGitSummary(".")
			analyses := CommitAnalyses{
				{
					Commit: Commit{
						CommitType: CommitTypeFeat,
					},
					Changes: []CommitChange{
						{
							File:  "main.go",
							Adds:  100,
							Dels:  50,
							Scope: Scopes{ScopeTypeApp},
							Tech:  []ScopeTechnology{Go},
						},
					},
				},
			}

			root.BuildFromAnalyses(analyses)

			pretty := root.Children[0].Pretty().String()

			Expect(pretty).To(ContainSubstring("main.go"))
			Expect(pretty).To(ContainSubstring("+100"))
			Expect(pretty).To(ContainSubstring("-50"))
			// Note: CommitTypes are not populated from analysis.Commit.CommitType
			// and "commits" only appears when gs.Commits > 1
		})

		It("should display directory with inline stats", func() {
			root := NewGitSummary(".")
			analyses := CommitAnalyses{
				{
					Changes: []CommitChange{
						{
							File:  "src/file1.go",
							Adds:  100,
							Dels:  50,
							Scope: Scopes{ScopeTypeApp},
							Tech:  []ScopeTechnology{Go},
						},
						{
							File:  "src/file2.go",
							Adds:  200,
							Dels:  75,
							Scope: Scopes{ScopeTypeApp},
							Tech:  []ScopeTechnology{Go},
						},
					},
				},
			}

			root.BuildFromAnalyses(analyses)

			srcDir := &root.Children[0]
			pretty := srcDir.Pretty().String()

			Expect(pretty).To(ContainSubstring("src/"))
			Expect(pretty).To(ContainSubstring("+300"))
			Expect(pretty).To(ContainSubstring("-125"))
			Expect(pretty).To(ContainSubstring("files"))
		})
	})
})
