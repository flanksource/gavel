package git_test

import (
	"strings"
	"time"

	. "github.com/flanksource/gavel/git"
	"github.com/flanksource/gavel/models"
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

	Describe("AggregateByType", func() {
		It("groups by Conventional Commit type with unique files per type", func() {
			meta := func(subject string) CommitMetaForTest {
				return CommitMetaForTest{Subject: subject}
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

			agg := NewAggForTest()
			AggregateByTypeForTest(meta("feat(api): add foo"), rows("10:2:a.go", "5:1:b.go"), agg)
			AggregateByTypeForTest(meta("feat(api): add bar"), rows("4:0:a.go"), agg) // a.go shared with prior feat
			AggregateByTypeForTest(meta("fix: bar"), rows("1:3:c.go"), agg)
			AggregateByTypeForTest(meta("chore(deps): bump"), rows("2:2:go.mod"), agg)
			AggregateByTypeForTest(meta("not a conventional subject"), rows("7:7:d.go"), agg)
			AggregateByTypeForTest(meta("refactor: drop legacy"), rows("0:50:e.go", "-:-:image.png"), agg)

			summaries := FinalizeForTest(agg)
			byType := map[models.CommitType]TypeSummary{}
			for _, s := range summaries {
				byType[s.Type] = s
			}

			Expect(byType[models.CommitTypeFeat].Commits).To(Equal(2))
			Expect(byType[models.CommitTypeFeat].Adds).To(Equal(19))
			Expect(byType[models.CommitTypeFeat].Dels).To(Equal(3))
			Expect(byType[models.CommitTypeFeat].Files).To(Equal(2)) // a.go + b.go, a.go deduped

			Expect(byType[models.CommitTypeFix].Commits).To(Equal(1))
			Expect(byType[models.CommitTypeFix].Adds).To(Equal(1))
			Expect(byType[models.CommitTypeFix].Dels).To(Equal(3))
			Expect(byType[models.CommitTypeFix].Files).To(Equal(1))

			Expect(byType[models.CommitTypeChore].Commits).To(Equal(1))
			Expect(byType[models.CommitTypeUnknown].Commits).To(Equal(1))
			Expect(byType[models.CommitTypeUnknown].Adds).To(Equal(7))

			Expect(byType[models.CommitTypeRefactor].Commits).To(Equal(1))
			Expect(byType[models.CommitTypeRefactor].Adds).To(Equal(0))
			Expect(byType[models.CommitTypeRefactor].Dels).To(Equal(50))
			Expect(byType[models.CommitTypeRefactor].Files).To(Equal(2)) // e.go + image.png (binary counted)

			// Unknown sorts last, then by commit count desc.
			Expect(summaries[len(summaries)-1].Type).To(Equal(models.CommitTypeUnknown))
		})
	})

	Describe("ParseNumstatStream", func() {
		It("parses three commits and dispatches per-commit callbacks in order", func() {
			t1 := time.Date(2024, 6, 1, 10, 0, 0, 0, time.UTC).Format(time.RFC3339)
			t2 := time.Date(2024, 6, 2, 11, 0, 0, 0, time.UTC).Format(time.RFC3339)
			t3 := time.Date(2024, 6, 3, 12, 0, 0, 0, time.UTC).Format(time.RFC3339)
			input := "" +
				"\x1eCMT\x1faaaaaaa\x1f" + t1 + "\x1ffeat(api): add foo\n" +
				"10\t2\ta.go\n" +
				"5\t1\tb.go\n" +
				"\x1eCMT\x1fbbbbbbb\x1f" + t2 + "\x1ffix: bar\n" +
				"1\t3\tc.go\n" +
				"\x1eCMT\x1fccccccc\x1f" + t3 + "\x1fchore(deps): bump\n" +
				"-\t-\timg.png\n" +
				"2\t2\tgo.mod\n"

			var subjects []string
			var rowCounts []int
			err := ParseNumstatStreamForTest(strings.NewReader(input), func(meta CommitMetaForTest, rows []NumstatRowForTest) error {
				subjects = append(subjects, meta.Subject)
				rowCounts = append(rowCounts, len(rows))
				return nil
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(subjects).To(Equal([]string{
				"feat(api): add foo",
				"fix: bar",
				"chore(deps): bump",
			}))
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

	Describe("TypeSummaries.Total", func() {
		It("sums all rows", func() {
			ts := TypeSummaries{
				{Type: models.CommitTypeFeat, Count: Count{Adds: 10, Dels: 2, Commits: 3, Files: 4}},
				{Type: models.CommitTypeFix, Count: Count{Adds: 1, Dels: 1, Commits: 1, Files: 1}},
			}
			total := ts.Total()
			Expect(total.Adds).To(Equal(11))
			Expect(total.Dels).To(Equal(3))
			Expect(total.Commits).To(Equal(4))
			Expect(total.Files).To(Equal(5))
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
