package git_test

import (
	"time"

	. "github.com/flanksource/gavel/git"
	. "github.com/flanksource/gavel/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Summarize", func() {
	var (
		now       = time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC) // Monday, Jan 15
		weekStart = time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)  // Start of week
		weekEnd   = time.Date(2024, 1, 21, 23, 59, 59, 999999999, time.UTC)
	)

	Describe("calculateTimeWindows", func() {
		It("should create daily windows", func() {
			from := time.Date(2024, 1, 1, 10, 30, 0, 0, time.UTC)
			until := time.Date(2024, 1, 3, 14, 45, 0, 0, time.UTC)

			windows := CalculateTimeWindows(from, until, GroupByDay)

			Expect(windows).To(HaveLen(3))
			Expect(windows[0].Start).To(Equal(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)))
			Expect(windows[0].End).To(Equal(time.Date(2024, 1, 1, 23, 59, 59, 999999999, time.UTC)))
			Expect(windows[2].Start).To(Equal(time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC)))
		})

		It("should create weekly windows aligned to Monday", func() {
			from := time.Date(2024, 1, 10, 10, 0, 0, 0, time.UTC)  // Wednesday
			until := time.Date(2024, 1, 25, 14, 0, 0, 0, time.UTC) // Thursday

			windows := CalculateTimeWindows(from, until, GroupByWeek)

			Expect(windows).To(HaveLen(3))
			// First window starts on Monday Jan 8
			Expect(windows[0].Start).To(Equal(time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC)))
			Expect(windows[0].End).To(Equal(time.Date(2024, 1, 14, 23, 59, 59, 999999999, time.UTC)))
			// Second window: Jan 15-21
			Expect(windows[1].Start).To(Equal(time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)))
			// Third window: Jan 22-28
			Expect(windows[2].Start).To(Equal(time.Date(2024, 1, 22, 0, 0, 0, 0, time.UTC)))
		})

		It("should create monthly windows", func() {
			from := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
			until := time.Date(2024, 3, 20, 14, 0, 0, 0, time.UTC)

			windows := CalculateTimeWindows(from, until, GroupByMonth)

			Expect(windows).To(HaveLen(3))
			Expect(windows[0].Start).To(Equal(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)))
			Expect(windows[0].End).To(Equal(time.Date(2024, 1, 31, 23, 59, 59, 999999999, time.UTC)))
			Expect(windows[1].Start).To(Equal(time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)))
			Expect(windows[2].Start).To(Equal(time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)))
		})
	})

	Describe("getWindowForCommit", func() {
		It("should find correct window for commit", func() {
			windows := []TimeWindow{
				{Start: weekStart, End: weekEnd},
			}
			commit := CommitAnalysis{
				Commit: Commit{
					Author: Author{Date: now},
				},
			}

			window := GetWindowForCommit(commit, windows)

			Expect(window).NotTo(BeNil())
			Expect(window.Start).To(Equal(weekStart))
		})

		It("should return nil for commit outside windows", func() {
			windows := []TimeWindow{
				{Start: weekStart, End: weekEnd},
			}
			commit := CommitAnalysis{
				Commit: Commit{
					Author: Author{Date: time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)},
				},
			}

			window := GetWindowForCommit(commit, windows)

			Expect(window).To(BeNil())
		})
	})

	Describe("selectTopScopes", func() {
		It("should select top N scopes by commit count", func() {
			commits := CommitAnalyses{
				{Commit: Commit{Scope: ScopeTypeApp}},
				{Commit: Commit{Scope: ScopeTypeApp}},
				{Commit: Commit{Scope: ScopeTypeApp}},
				{Commit: Commit{Scope: ScopeTypeTest}},
				{Commit: Commit{Scope: ScopeTypeTest}},
				{Commit: Commit{Scope: ScopeTypeCI}},
			}

			scopes := SelectTopScopes(commits, 2)

			Expect(scopes).To(HaveLen(2))
			Expect(scopes[0]).To(Equal(ScopeTypeApp))  // 3 commits
			Expect(scopes[1]).To(Equal(ScopeTypeTest)) // 2 commits
		})

		It("should return all scopes when less than max", func() {
			commits := CommitAnalyses{
				{Commit: Commit{Scope: ScopeTypeApp}},
				{Commit: Commit{Scope: ScopeTypeTest}},
			}

			scopes := SelectTopScopes(commits, 5)

			Expect(scopes).To(HaveLen(2))
		})
	})

	Describe("aggregateCommitGroup", func() {
		It("should aggregate statistics correctly", func() {
			commits := CommitAnalyses{
				{
					Commit: Commit{
						CommitType: CommitTypeFeat,
						Scope:      ScopeTypeApp,
					},
					Tech: []ScopeTechnology{Go, Docker},
					Changes: []CommitChange{
						{Adds: 100, Dels: 50, File: "main.go"},
						{Adds: 50, Dels: 25, File: "test.go"},
					},
				},
				{
					Commit: Commit{
						CommitType: CommitTypeFix,
						Scope:      ScopeTypeApp,
					},
					Tech: []ScopeTechnology{Go},
					Changes: []CommitChange{
						{Adds: 75, Dels: 30, File: "main.go"},
					},
				},
			}

			count := AggregateCommitGroup(commits)

			Expect(count.Adds).To(Equal(225))
			Expect(count.Dels).To(Equal(105))
			Expect(count.Commits).To(Equal(2))
			Expect(count.Files).To(Equal(2)) // main.go, test.go unique
			Expect(count.Scopes[ScopeTypeApp]).To(Equal(2))
			Expect(count.CommitTypes[CommitTypeFeat]).To(Equal(1))
			Expect(count.CommitTypes[CommitTypeFix]).To(Equal(1))
			Expect(count.Tech[Go]).To(Equal(2))
			Expect(count.Tech[Docker]).To(Equal(1))
		})
	})

	Describe("generateFallbackDescription", func() {
		It("should generate deterministic description", func() {
			commits := CommitAnalyses{
				{Commit: Commit{CommitType: CommitTypeFeat}},
				{Commit: Commit{CommitType: CommitTypeFeat}},
				{Commit: Commit{CommitType: CommitTypeFeat}},
				{Commit: Commit{CommitType: CommitTypeFix}},
				{Commit: Commit{CommitType: CommitTypeFix}},
			}

			name, desc := GenerateFallbackDescription(ScopeTypeApp, commits)

			Expect(name).To(Equal("app changes"))
			Expect(desc).To(ContainSubstring("3 feat"))
			Expect(desc).To(ContainSubstring("2 fix"))
			Expect(desc).To(ContainSubstring("5 commits"))
		})
	})

	Describe("Summarize", func() {
		It("should create summaries grouped by time window and scope", func() {
			commits := CommitAnalyses{
				{
					Commit: Commit{
						CommitType: CommitTypeFeat,
						Scope:      ScopeTypeApp,
						Author:     Author{Date: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)},
					},
					Changes: []CommitChange{{Adds: 100, Dels: 50, File: "app.go"}},
				},
				{
					Commit: Commit{
						CommitType: CommitTypeFix,
						Scope:      ScopeTypeApp,
						Author:     Author{Date: time.Date(2024, 1, 16, 10, 0, 0, 0, time.UTC)},
					},
					Changes: []CommitChange{{Adds: 50, Dels: 25, File: "app.go"}},
				},
				{
					Commit: Commit{
						CommitType: CommitTypeFeat,
						Scope:      ScopeTypeTest,
						Author:     Author{Date: time.Date(2024, 1, 22, 10, 0, 0, 0, time.UTC)}, // Next week
					},
					Changes: []CommitChange{{Adds: 75, Dels: 30, File: "test.go"}},
				},
			}

			options := SummaryOptions{
				Window:        GroupByWeek,
				MaxCategories: 5,
			}

			summaries, err := Summarize(commits, options)

			Expect(err).To(BeNil())
			Expect(summaries).To(HaveLen(2)) // 2 windows with commits

			// Should be sorted newest first
			Expect(*summaries[0].From).To(Equal(time.Date(2024, 1, 22, 0, 0, 0, 0, time.UTC)))
			Expect(*summaries[1].From).To(Equal(time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)))

			// Week 1: app scope
			Expect(summaries[1].Scopes).To(ContainElement(ScopeTypeApp))
			Expect(summaries[1].Commits.Commits).To(Equal(2))
			Expect(summaries[1].Commits.Adds).To(Equal(150))
			Expect(summaries[1].Commits.Dels).To(Equal(75))

			// Week 2: test scope
			Expect(summaries[0].Scopes).To(ContainElement(ScopeTypeTest))
			Expect(summaries[0].Commits.Commits).To(Equal(1))
		})

		It("should limit scopes per window to MaxCategories with Other group", func() {
			commits := CommitAnalyses{
				{Commit: Commit{Scope: ScopeTypeApp, Author: Author{Date: now}}},
				{Commit: Commit{Scope: ScopeTypeApp, Author: Author{Date: now}}},
				{Commit: Commit{Scope: ScopeTypeApp, Author: Author{Date: now}}},
				{Commit: Commit{Scope: ScopeTypeTest, Author: Author{Date: now}}},
				{Commit: Commit{Scope: ScopeTypeTest, Author: Author{Date: now}}},
				{Commit: Commit{Scope: ScopeTypeCI, Author: Author{Date: now}}},
				{Commit: Commit{Scope: ScopeTypeDocs, Author: Author{Date: now}}},
			}

			options := SummaryOptions{
				Window:        GroupByMonth,
				MaxCategories: 2,
			}

			summaries, err := Summarize(commits, options)

			Expect(err).To(BeNil())
			// Should have 2 summaries: top 1 scope (App) + Other group
			Expect(summaries).To(HaveLen(2))

			// Find the App and Other summaries
			var appSummary, otherSummary *GitSummary
			for i := range summaries {
				if len(summaries[i].Scopes) > 0 && summaries[i].Scopes[0] == ScopeTypeApp {
					appSummary = &summaries[i]
				}
				// Other summary will contain Test, CI, and Docs scopes
				if len(summaries[i].Scopes) > 1 {
					otherSummary = &summaries[i]
				}
			}

			Expect(appSummary).NotTo(BeNil())
			Expect(appSummary.Commits.Commits).To(Equal(3))

			Expect(otherSummary).NotTo(BeNil())
			Expect(otherSummary.Commits.Commits).To(Equal(4))
			Expect(otherSummary.Scopes).To(ContainElements(ScopeTypeTest, ScopeTypeCI, ScopeTypeDocs))
		})

		It("should not create Other group when all commits fit in top categories", func() {
			commits := CommitAnalyses{
				{Commit: Commit{Scope: ScopeTypeApp, Author: Author{Date: now}}},
				{Commit: Commit{Scope: ScopeTypeTest, Author: Author{Date: now}}},
			}

			options := SummaryOptions{
				Window:        GroupByMonth,
				MaxCategories: 5,
			}

			summaries, err := Summarize(commits, options)

			Expect(err).To(BeNil())
			// Should have 2 summaries (no Other group needed)
			Expect(summaries).To(HaveLen(2))

			// Verify no "Other" summary exists
			for _, summary := range summaries {
				Expect(summary.Scopes).NotTo(ContainElement(ScopeTypeOther))
			}
		})

		It("should limit categories per window not globally", func() {
			// Create commits across two months
			month1 := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
			month2 := time.Date(2024, 2, 15, 0, 0, 0, 0, time.UTC)

			commits := CommitAnalyses{
				// Month 1: App (3), Test (2), CI (1)
				{Commit: Commit{Scope: ScopeTypeApp, Author: Author{Date: month1}}},
				{Commit: Commit{Scope: ScopeTypeApp, Author: Author{Date: month1}}},
				{Commit: Commit{Scope: ScopeTypeApp, Author: Author{Date: month1}}},
				{Commit: Commit{Scope: ScopeTypeTest, Author: Author{Date: month1}}},
				{Commit: Commit{Scope: ScopeTypeTest, Author: Author{Date: month1}}},
				{Commit: Commit{Scope: ScopeTypeCI, Author: Author{Date: month1}}},
				// Month 2: Docs (3), Security (2), Database (1)
				{Commit: Commit{Scope: ScopeTypeDocs, Author: Author{Date: month2}}},
				{Commit: Commit{Scope: ScopeTypeDocs, Author: Author{Date: month2}}},
				{Commit: Commit{Scope: ScopeTypeDocs, Author: Author{Date: month2}}},
				{Commit: Commit{Scope: ScopeTypeSecurity, Author: Author{Date: month2}}},
				{Commit: Commit{Scope: ScopeTypeSecurity, Author: Author{Date: month2}}},
				{Commit: Commit{Scope: ScopeTypeDatabase, Author: Author{Date: month2}}},
			}

			options := SummaryOptions{
				Window:        GroupByMonth,
				MaxCategories: 2,
			}

			summaries, err := Summarize(commits, options)

			Expect(err).To(BeNil())
			// Should have 4 summaries: 2 per month (top 1 + Other)
			Expect(summaries).To(HaveLen(4))

			// Group summaries by month
			month1Summaries := []GitSummary{}
			month2Summaries := []GitSummary{}
			for _, summary := range summaries {
				if summary.From.Month() == time.January {
					month1Summaries = append(month1Summaries, summary)
				} else {
					month2Summaries = append(month2Summaries, summary)
				}
			}

			// Each month should have 2 summaries
			Expect(month1Summaries).To(HaveLen(2))
			Expect(month2Summaries).To(HaveLen(2))
		})

		It("should handle empty commit list", func() {
			commits := CommitAnalyses{}
			options := SummaryOptions{
				Window:        GroupByMonth,
				MaxCategories: 5,
			}

			summaries, err := Summarize(commits, options)

			Expect(err).To(BeNil())
			Expect(summaries).To(BeEmpty())
		})

		It("should track repositories in summaries from multiple repos", func() {
			commits := CommitAnalyses{
				{
					Commit: Commit{
						Repository: "repo1",
						CommitType: CommitTypeFeat,
						Scope:      ScopeTypeApp,
						Author:     Author{Date: now},
					},
					Changes: []CommitChange{{Adds: 100, Dels: 50, File: "app.go"}},
				},
				{
					Commit: Commit{
						Repository: "repo2",
						CommitType: CommitTypeFix,
						Scope:      ScopeTypeApp,
						Author:     Author{Date: now},
					},
					Changes: []CommitChange{{Adds: 50, Dels: 25, File: "app.go"}},
				},
				{
					Commit: Commit{
						Repository: "repo1",
						CommitType: CommitTypeFeat,
						Scope:      ScopeTypeTest,
						Author:     Author{Date: now},
					},
					Changes: []CommitChange{{Adds: 75, Dels: 30, File: "test.go"}},
				},
			}

			options := SummaryOptions{
				Window:        GroupByMonth,
				MaxCategories: 5,
			}

			summaries, err := Summarize(commits, options)

			Expect(err).To(BeNil())
			Expect(summaries).To(HaveLen(2)) // app and test scopes

			// Find app and test summaries
			var appSummary, testSummary *GitSummary
			for i := range summaries {
				if len(summaries[i].Scopes) > 0 {
					if summaries[i].Scopes[0] == ScopeTypeApp {
						appSummary = &summaries[i]
					} else if summaries[i].Scopes[0] == ScopeTypeTest {
						testSummary = &summaries[i]
					}
				}
			}

			Expect(appSummary).NotTo(BeNil())
			Expect(appSummary.Repositories).To(HaveLen(2))
			Expect(appSummary.Repositories).To(ConsistOf("repo1", "repo2"))

			Expect(testSummary).NotTo(BeNil())
			Expect(testSummary.Repositories).To(HaveLen(1))
			Expect(testSummary.Repositories).To(ContainElement("repo1"))
		})

		It("should have empty repositories list when no repository field", func() {
			commits := CommitAnalyses{
				{
					Commit: Commit{
						CommitType: CommitTypeFeat,
						Scope:      ScopeTypeApp,
						Author:     Author{Date: now},
					},
					Changes: []CommitChange{{Adds: 100, Dels: 50, File: "app.go"}},
				},
			}

			options := SummaryOptions{
				Window:        GroupByMonth,
				MaxCategories: 5,
			}

			summaries, err := Summarize(commits, options)

			Expect(err).To(BeNil())
			Expect(summaries).To(HaveLen(1))
			Expect(summaries[0].Repositories).To(BeEmpty())
		})
	})
})
