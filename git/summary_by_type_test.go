package git_test

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	. "github.com/flanksource/gavel/git"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("SummaryByType", func() {
	Describe("ParseNumstatLine", func() {
		type tc struct {
			line     string
			adds     int
			dels     int
			file     string
			isBinary bool
			ok       bool
		}
		cases := []tc{
			{line: "10\t5\tfoo.go", adds: 10, dels: 5, file: "foo.go", isBinary: false, ok: true},
			{line: "0\t0\tbar.go", adds: 0, dels: 0, file: "bar.go", isBinary: false, ok: true},
			{line: "-\t-\timg.png", adds: 0, dels: 0, file: "img.png", isBinary: true, ok: true},
			{line: "", ok: false},
			{line: "garbage", ok: false},
			{line: "x\ty\tfoo.go", ok: false},
			{line: "10\t5\t", ok: false},
		}
		for _, c := range cases {
			It("parses "+strings.ReplaceAll(c.line, "\t", `\t`), func() {
				adds, dels, file, isBinary, ok := ParseNumstatLineForTest(c.line)
				Expect(ok).To(Equal(c.ok))
				if c.ok {
					Expect(adds).To(Equal(c.adds))
					Expect(dels).To(Equal(c.dels))
					Expect(file).To(Equal(c.file))
					Expect(isBinary).To(Equal(c.isBinary))
				}
			})
		}
	})

	Describe("NormalizeGroupBy", func() {
		It("defaults to type when input is empty", func() {
			out, err := NormalizeGroupByForTest(nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(Equal([]GroupBy{GroupByType}))
		})

		It("parses comma-separated values", func() {
			out, err := NormalizeGroupByForTest([]string{"author,type"})
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(Equal([]GroupBy{GroupByAuthor, GroupByType}))
		})

		It("parses repeated flag values", func() {
			out, err := NormalizeGroupByForTest([]string{"type", "author"})
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(Equal([]GroupBy{GroupByType, GroupByAuthor}))
		})

		It("deduplicates while preserving first-occurrence order", func() {
			out, err := NormalizeGroupByForTest([]string{"author,type,author,type"})
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(Equal([]GroupBy{GroupByAuthor, GroupByType}))
		})

		It("is case-insensitive", func() {
			out, err := NormalizeGroupByForTest([]string{"AUTHOR,Type"})
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(Equal([]GroupBy{GroupByAuthor, GroupByType}))
		})

		It("rejects unknown dimensions", func() {
			_, err := NormalizeGroupByForTest([]string{"type,garbage"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("garbage"))
		})

		It("ignores empty entries between commas", func() {
			out, err := NormalizeGroupByForTest([]string{"type,,author"})
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(Equal([]GroupBy{GroupByType, GroupByAuthor}))
		})

		It("accepts time-based dimensions in any combination", func() {
			out, err := NormalizeGroupByForTest([]string{"month,author,type"})
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(Equal([]GroupBy{GroupByCommitMonth, GroupByAuthor, GroupByType}))
		})

		DescribeTable("accepts each time-based dimension",
			func(input string, expected GroupBy) {
				out, err := NormalizeGroupByForTest([]string{input})
				Expect(err).NotTo(HaveOccurred())
				Expect(out).To(Equal([]GroupBy{expected}))
			},
			Entry("year", "year", GroupByCommitYear),
			Entry("month", "month", GroupByCommitMonth),
			Entry("week", "week", GroupByCommitWeek),
			Entry("day", "day", GroupByCommitDay),
		)
	})

	Describe("Aggregate", func() {
		meta := func(subject, author string) CommitMetaForTest {
			return CommitMetaForTest{Subject: subject, AuthorName: author}
		}
		rows := func(specs ...string) []NumstatRowForTest {
			out := make([]NumstatRowForTest, 0, len(specs))
			for _, s := range specs {
				parts := strings.SplitN(s, ":", 3)
				adds, addsOK := parseTestInt(parts[0])
				dels, delsOK := parseTestInt(parts[1])
				isBinary := !addsOK || !delsOK
				if isBinary {
					adds, dels = 0, 0
				}
				out = append(out, NumstatRowForTest{Adds: adds, Dels: dels, File: parts[2], IsBinary: isBinary})
			}
			return out
		}
		byKey := func(summaries []GroupSummary) map[string]GroupSummary {
			m := make(map[string]GroupSummary, len(summaries))
			for _, s := range summaries {
				m[strings.Join(s.Group, "/")] = s
			}
			return m
		}

		It("groups by Conventional Commit type with unique files per type", func() {
			by := []GroupBy{GroupByType}
			agg := NewAggForTest()
			AggregateForTest(meta("feat(api): add foo", "Alice"), rows("10:2:a.go", "5:1:b.go"), agg, by, false)
			AggregateForTest(meta("feat(api): add bar", "Bob"), rows("4:0:a.go"), agg, by, false) // a.go shared with prior feat
			AggregateForTest(meta("fix: bar", "Alice"), rows("1:3:c.go"), agg, by, false)
			AggregateForTest(meta("chore(deps): bump", "Alice"), rows("2:2:go.mod"), agg, by, false)
			AggregateForTest(meta("not a conventional subject", "Alice"), rows("7:7:d.go"), agg, by, false)
			AggregateForTest(meta("refactor: drop legacy", "Bob"), rows("0:50:e.go", "-:-:image.png"), agg, by, false)

			summaries := FinalizeForTest(agg)
			byGroup := byKey(summaries)

			Expect(byGroup["feat"].Commits).To(Equal(2))
			Expect(byGroup["feat"].Adds).To(Equal(19))
			Expect(byGroup["feat"].Dels).To(Equal(3))
			Expect(byGroup["feat"].Files).To(Equal(2)) // a.go + b.go, a.go deduped

			Expect(byGroup["fix"].Commits).To(Equal(1))
			Expect(byGroup["chore"].Commits).To(Equal(1))
			Expect(byGroup["unknown"].Commits).To(Equal(1))
			Expect(byGroup["unknown"].Adds).To(Equal(7))

			Expect(byGroup["refactor"].Commits).To(Equal(1))
			Expect(byGroup["refactor"].Adds).To(Equal(0))
			Expect(byGroup["refactor"].Dels).To(Equal(50))
			Expect(byGroup["refactor"].Files).To(Equal(2))

			// Unknown sorts last.
			Expect(summaries[len(summaries)-1].Group).To(Equal([]string{"unknown"}))
		})

		It("groups by author when GroupBy=[author]", func() {
			by := []GroupBy{GroupByAuthor}
			agg := NewAggForTest()
			AggregateForTest(meta("feat: x", "Alice"), rows("10:2:a.go"), agg, by, false)
			AggregateForTest(meta("fix: y", "Alice"), rows("3:1:b.go"), agg, by, false)
			AggregateForTest(meta("feat: z", "Bob"), rows("5:5:c.go"), agg, by, false)
			AggregateForTest(meta("chore: w", ""), rows("1:1:d.go"), agg, by, false)

			summaries := FinalizeForTest(agg)
			byGroup := byKey(summaries)
			Expect(byGroup["Alice"].Commits).To(Equal(2))
			Expect(byGroup["Alice"].Adds).To(Equal(13))
			Expect(byGroup["Alice"].Files).To(Equal(2))
			Expect(byGroup["Bob"].Commits).To(Equal(1))
			Expect(byGroup["unknown"].Commits).To(Equal(1)) // empty author → unknown
			Expect(summaries[len(summaries)-1].Group).To(Equal([]string{"unknown"}))
		})

		It("produces one row per (author, type) tuple when GroupBy=[author,type]", func() {
			by := []GroupBy{GroupByAuthor, GroupByType}
			agg := NewAggForTest()
			AggregateForTest(meta("feat: x", "Alice"), rows("10:2:a.go"), agg, by, false)
			AggregateForTest(meta("feat: y", "Alice"), rows("3:1:b.go"), agg, by, false)
			AggregateForTest(meta("fix: z", "Alice"), rows("4:4:c.go"), agg, by, false)
			AggregateForTest(meta("feat: q", "Bob"), rows("5:5:d.go"), agg, by, false)

			summaries := FinalizeForTest(agg)
			byGroup := byKey(summaries)

			Expect(byGroup["Alice/feat"].Commits).To(Equal(2))
			Expect(byGroup["Alice/feat"].Adds).To(Equal(13))
			Expect(byGroup["Alice/fix"].Commits).To(Equal(1))
			Expect(byGroup["Bob/feat"].Commits).To(Equal(1))
			Expect(summaries).To(HaveLen(3))

			// Each row's Group preserves dimension order.
			Expect(byGroup["Alice/feat"].Group).To(Equal([]string{"Alice", "feat"}))
			Expect(byGroup["Bob/feat"].Group).To(Equal([]string{"Bob", "feat"}))
		})

		It("sinks tuples with any unknown part to the bottom", func() {
			by := []GroupBy{GroupByAuthor, GroupByType}
			agg := NewAggForTest()
			AggregateForTest(meta("feat: x", "Alice"), rows("10:2:a.go"), agg, by, false)
			AggregateForTest(meta("not conventional", "Alice"), rows("1:1:b.go"), agg, by, false)
			AggregateForTest(meta("feat: y", ""), rows("1:1:c.go"), agg, by, false)

			summaries := FinalizeForTest(agg)
			Expect(summaries[0].Group).To(Equal([]string{"Alice", "feat"}))
			// The remaining two both contain "unknown" — order between them is by commit count then lexical.
			for _, row := range summaries[1:] {
				Expect(row.Group).To(ContainElement("unknown"))
			}
		})

		It("ignores outlier dependency-lock files when ignoreOutliers=true", func() {
			by := []GroupBy{GroupByType}
			agg := NewAggForTest()
			skipped := AggregateForTest(
				meta("chore(deps): bump go modules", "Alice"),
				rows("1:1:go.mod", "5000:3000:go.sum", "20:5:internal/foo.go"),
				agg, by, true,
			)
			Expect(skipped).To(Equal(1))

			summaries := FinalizeForTest(agg)
			Expect(summaries).To(HaveLen(1))
			chore := summaries[0]
			Expect(chore.Group).To(Equal([]string{"chore"}))
			Expect(chore.Adds).To(Equal(21))
			Expect(chore.Dels).To(Equal(6))
			Expect(chore.Files).To(Equal(2))
			Expect(chore.Commits).To(Equal(1))
		})

		It("keeps outlier files when ignoreOutliers=false", func() {
			by := []GroupBy{GroupByType}
			agg := NewAggForTest()
			skipped := AggregateForTest(
				meta("chore(deps): bump", "Alice"),
				rows("5000:3000:go.sum"),
				agg, by, false,
			)
			Expect(skipped).To(Equal(0))
			summaries := FinalizeForTest(agg)
			Expect(summaries[0].Adds).To(Equal(5000))
		})

		DescribeTable("isOutlierFile",
			func(path string, expected bool) {
				Expect(IsOutlierFileForTest(path)).To(Equal(expected))
			},
			Entry("go.sum", "go.sum", true),
			Entry("nested go.sum", "internal/foo/go.sum", true),
			Entry("package-lock.json", "frontend/package-lock.json", true),
			Entry("yarn.lock", "ui/yarn.lock", true),
			Entry("Cargo.lock", "crates/x/Cargo.lock", true),
			Entry("regular Go file", "internal/foo.go", false),
			Entry("similarly named but different", "go.sum.bak", false),
		)
	})

	Describe("ParseNumstatStream", func() {
		It("parses three commits and dispatches per-commit callbacks in order", func() {
			t1 := time.Date(2024, 6, 1, 10, 0, 0, 0, time.UTC).Format(time.RFC3339)
			t2 := time.Date(2024, 6, 2, 11, 0, 0, 0, time.UTC).Format(time.RFC3339)
			t3 := time.Date(2024, 6, 3, 12, 0, 0, 0, time.UTC).Format(time.RFC3339)
			input := "" +
				"\x1eCMT\x1faaaaaaa\x1f" + t1 + "\x1fAlice\x1ffeat(api): add foo\n" +
				"10\t2\ta.go\n" +
				"5\t1\tb.go\n" +
				"\x1eCMT\x1fbbbbbbb\x1f" + t2 + "\x1fBob\x1ffix: bar\n" +
				"1\t3\tc.go\n" +
				"\x1eCMT\x1fccccccc\x1f" + t3 + "\x1fAlice\x1fchore(deps): bump\n" +
				"-\t-\timg.png\n" +
				"2\t2\tgo.mod\n"

			var subjects []string
			var authors []string
			var rowCounts []int
			err := ParseNumstatStreamForTest(strings.NewReader(input), func(meta CommitMetaForTest, rows []NumstatRowForTest) error {
				subjects = append(subjects, meta.Subject)
				authors = append(authors, meta.AuthorName)
				rowCounts = append(rowCounts, len(rows))
				return nil
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(subjects).To(Equal([]string{
				"feat(api): add foo",
				"fix: bar",
				"chore(deps): bump",
			}))
			Expect(authors).To(Equal([]string{"Alice", "Bob", "Alice"}))
			Expect(rowCounts).To(Equal([]int{2, 1, 2}))
		})

		It("returns no commits for empty input", func() {
			calls := 0
			err := ParseNumstatStreamForTest(strings.NewReader(""), func(_ CommitMetaForTest, _ []NumstatRowForTest) error {
				calls++
				return nil
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(calls).To(Equal(0))
		})
	})

	Describe("FormatCommitDateBucket", func() {
		ts := time.Date(2025, 3, 7, 14, 30, 0, 0, time.UTC) // Friday
		DescribeTable("formats per dimension",
			func(by GroupBy, expected string) {
				Expect(FormatCommitDateBucketForTest(ts, by)).To(Equal(expected))
			},
			Entry("year", GroupByCommitYear, "2025"),
			Entry("month", GroupByCommitMonth, "2025-03"),
			Entry("week", GroupByCommitWeek, "2025-W10"),
			Entry("day", GroupByCommitDay, "2025-03-07"),
		)

		It("returns unknown for the zero time", func() {
			Expect(FormatCommitDateBucketForTest(time.Time{}, GroupByCommitMonth)).To(Equal("unknown"))
		})
	})

	Describe("Aggregate with time dimensions", func() {
		It("buckets commits into month labels via the commit date", func() {
			by := []GroupBy{GroupByCommitMonth}
			agg := NewAggForTest()
			AggregateForTest(CommitMetaForTest{
				Subject: "feat: a", AuthorName: "Alice",
				Date: time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
			}, []NumstatRowForTest{{Adds: 10, Dels: 1, File: "a.go"}}, agg, by, false)
			AggregateForTest(CommitMetaForTest{
				Subject: "feat: b", AuthorName: "Alice",
				Date: time.Date(2025, 1, 28, 0, 0, 0, 0, time.UTC),
			}, []NumstatRowForTest{{Adds: 5, Dels: 0, File: "b.go"}}, agg, by, false)
			AggregateForTest(CommitMetaForTest{
				Subject: "fix: c", AuthorName: "Bob",
				Date: time.Date(2025, 2, 3, 0, 0, 0, 0, time.UTC),
			}, []NumstatRowForTest{{Adds: 1, Dels: 1, File: "c.go"}}, agg, by, false)

			summaries := FinalizeForTest(agg)
			byKey := map[string]GroupSummary{}
			for _, s := range summaries {
				byKey[strings.Join(s.Group, "/")] = s
			}
			Expect(byKey["2025-01"].Commits).To(Equal(2))
			Expect(byKey["2025-01"].Adds).To(Equal(15))
			Expect(byKey["2025-02"].Commits).To(Equal(1))
		})
	})

	Describe("GroupSummaries.tree time-dimension sort", func() {
		It("orders time-dimension siblings chronologically descending", func() {
			gs := GroupSummaries{
				By: []GroupBy{GroupByCommitMonth, GroupByType},
				Groups: []GroupSummary{
					{Group: []string{"2025-01", "feat"}, Count: Count{Commits: 1}},
					{Group: []string{"2025-03", "feat"}, Count: Count{Commits: 1}},
					{Group: []string{"2025-02", "fix"}, Count: Count{Commits: 5}},
				},
			}
			tree := GroupSummariesTreeForTest(gs)
			Expect(tree).To(HaveLen(3))
			// Most recent month first, regardless of commit count.
			Expect(tree[0].Label).To(Equal("2025-03"))
			Expect(tree[1].Label).To(Equal("2025-02"))
			Expect(tree[2].Label).To(Equal("2025-01"))
		})

		It("still sorts non-time child dimensions by commit count", func() {
			gs := GroupSummaries{
				By: []GroupBy{GroupByCommitMonth, GroupByType},
				Groups: []GroupSummary{
					{Group: []string{"2025-01", "feat"}, Count: Count{Commits: 2}},
					{Group: []string{"2025-01", "fix"}, Count: Count{Commits: 5}},
				},
			}
			tree := GroupSummariesTreeForTest(gs)
			Expect(tree[0].Children[0].Label).To(Equal("fix"))
			Expect(tree[0].Children[1].Label).To(Equal("feat"))
		})
	})

	Describe("GroupSummaries.tree", func() {
		It("returns one leaf per row when there is a single dimension", func() {
			gs := GroupSummaries{
				By: []GroupBy{GroupByType},
				Groups: []GroupSummary{
					{Group: []string{"feat"}, Count: Count{Adds: 10, Dels: 2, Commits: 3, Files: 4}},
					{Group: []string{"fix"}, Count: Count{Adds: 1, Dels: 1, Commits: 1, Files: 1}},
				},
			}
			tree := GroupSummariesTreeForTest(gs)
			Expect(tree).To(HaveLen(2))
			Expect(tree[0].Label).To(Equal("feat"))
			Expect(tree[0].Children).To(BeEmpty())
			Expect(tree[0].Count.Commits).To(Equal(3))
			Expect(tree[1].Label).To(Equal("fix"))
		})

		It("nests rows under their leading dimension and sums child counts", func() {
			gs := GroupSummaries{
				By: []GroupBy{GroupByAuthor, GroupByType},
				Groups: []GroupSummary{
					{Group: []string{"Alice", "feat"}, Count: Count{Adds: 13, Dels: 3, Commits: 2, Files: 2}},
					{Group: []string{"Alice", "fix"}, Count: Count{Adds: 4, Dels: 4, Commits: 1, Files: 1}},
					{Group: []string{"Bob", "feat"}, Count: Count{Adds: 5, Dels: 5, Commits: 1, Files: 1}},
				},
			}
			tree := GroupSummariesTreeForTest(gs)
			Expect(tree).To(HaveLen(2))
			alice := tree[0]
			Expect(alice.Label).To(Equal("Alice"))
			Expect(alice.Count.Commits).To(Equal(3))
			Expect(alice.Count.Adds).To(Equal(17))
			Expect(alice.Count.Dels).To(Equal(7))
			Expect(alice.Children).To(HaveLen(2))
			Expect(alice.Children[0].Label).To(Equal("feat"))
			Expect(alice.Children[0].Count.Commits).To(Equal(2))
			Expect(alice.Children[1].Label).To(Equal("fix"))

			bob := tree[1]
			Expect(bob.Label).To(Equal("Bob"))
			Expect(bob.Children).To(HaveLen(1))
			Expect(bob.Children[0].Label).To(Equal("feat"))
		})

		It("sinks branches whose label is unknown to the bottom of their level", func() {
			gs := GroupSummaries{
				By: []GroupBy{GroupByAuthor, GroupByType},
				Groups: []GroupSummary{
					{Group: []string{"unknown", "feat"}, Count: Count{Commits: 5}},
					{Group: []string{"Alice", "feat"}, Count: Count{Commits: 1}},
					{Group: []string{"Alice", "unknown"}, Count: Count{Commits: 1}},
				},
			}
			tree := GroupSummariesTreeForTest(gs)
			Expect(tree[0].Label).To(Equal("Alice"))
			Expect(tree[len(tree)-1].Label).To(Equal("unknown"))

			alice := tree[0]
			Expect(alice.Children[len(alice.Children)-1].Label).To(Equal("unknown"))
		})
	})

	Describe("GroupSummaries.Total", func() {
		It("sums all rows", func() {
			gs := GroupSummaries{
				Groups: []GroupSummary{
					{Group: []string{"feat"}, Count: Count{Adds: 10, Dels: 2, Commits: 3, Files: 4}},
					{Group: []string{"fix"}, Count: Count{Adds: 1, Dels: 1, Commits: 1, Files: 1}},
				},
			}
			total := gs.Total()
			Expect(total.Adds).To(Equal(11))
			Expect(total.Dels).To(Equal(3))
			Expect(total.Commits).To(Equal(4))
			Expect(total.Files).To(Equal(5))
		})
	})

	Describe("SplitRepoPathArgs", func() {
		var tmp string
		BeforeEach(func() {
			tmp = GinkgoT().TempDir()
			Expect(os.MkdirAll(filepath.Join(tmp, "repo-a", ".git"), 0o755)).To(Succeed())
			Expect(os.MkdirAll(filepath.Join(tmp, "repo-b", ".git"), 0o755)).To(Succeed())
			Expect(os.MkdirAll(filepath.Join(tmp, "not-a-repo"), 0o755)).To(Succeed())
		})

		It("partitions args into repo dirs (with .git) and other args", func() {
			repos, rest, err := SplitRepoPathArgs([]string{
				filepath.Join(tmp, "repo-a"),
				"abcdef1",
				filepath.Join(tmp, "repo-b"),
				"main..feature",
				filepath.Join(tmp, "not-a-repo"),
				"src/foo.go",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(repos).To(Equal([]string{
				filepath.Join(tmp, "repo-a"),
				filepath.Join(tmp, "repo-b"),
			}))
			Expect(rest).To(Equal([]string{
				"abcdef1",
				"main..feature",
				filepath.Join(tmp, "not-a-repo"),
				"src/foo.go",
			}))
		})

		It("returns nil repos when no args are git work trees", func() {
			repos, rest, err := SplitRepoPathArgs([]string{"abcdef1", "main..feature"})
			Expect(err).NotTo(HaveOccurred())
			Expect(repos).To(BeEmpty())
			Expect(rest).To(Equal([]string{"abcdef1", "main..feature"}))
		})
	})

	Describe("Aggregate by repo", func() {
		It("partitions commits by the repo label stamped on each commit", func() {
			by := []GroupBy{GroupByRepo, GroupByType}
			agg := NewAggForTest()
			AggregateForTest(CommitMetaForTest{Subject: "feat: a", AuthorName: "Alice", Repo: "alpha"},
				[]NumstatRowForTest{{Adds: 10, Dels: 2, File: "a.go"}}, agg, by, false)
			AggregateForTest(CommitMetaForTest{Subject: "fix: b", AuthorName: "Alice", Repo: "alpha"},
				[]NumstatRowForTest{{Adds: 1, Dels: 1, File: "b.go"}}, agg, by, false)
			AggregateForTest(CommitMetaForTest{Subject: "feat: c", AuthorName: "Bob", Repo: "beta"},
				[]NumstatRowForTest{{Adds: 5, Dels: 5, File: "c.go"}}, agg, by, false)

			summaries := FinalizeForTest(agg)
			byKey := map[string]GroupSummary{}
			for _, s := range summaries {
				byKey[strings.Join(s.Group, "/")] = s
			}
			Expect(byKey["alpha/feat"].Commits).To(Equal(1))
			Expect(byKey["alpha/feat"].Adds).To(Equal(10))
			Expect(byKey["alpha/fix"].Commits).To(Equal(1))
			Expect(byKey["beta/feat"].Commits).To(Equal(1))
			Expect(byKey["beta/feat"].Adds).To(Equal(5))
		})

		It("buckets commits with empty Repo under unknown", func() {
			by := []GroupBy{GroupByRepo}
			agg := NewAggForTest()
			AggregateForTest(CommitMetaForTest{Subject: "feat: a", Repo: ""},
				[]NumstatRowForTest{{Adds: 1, Dels: 0, File: "a.go"}}, agg, by, false)
			summaries := FinalizeForTest(agg)
			Expect(summaries).To(HaveLen(1))
			Expect(summaries[0].Group).To(Equal([]string{"unknown"}))
		})
	})
})

func parseTestInt(s string) (int, bool) {
	if s == "" || s == "-" {
		return 0, false
	}
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, false
		}
		n = n*10 + int(c-'0')
	}
	return n, true
}
