package git

import (
	"io"
	"time"
)

// CommitMetaForTest mirrors the unexported commitMeta used by the streaming
// numstat parser, exposed here purely for white-box testing from git_test.
type CommitMetaForTest struct {
	Hash       string
	Date       time.Time
	AuthorName string
	Subject    string
	Repo       string
}

// NumstatRowForTest mirrors the unexported numstatRow.
type NumstatRowForTest struct {
	Adds     int
	Dels     int
	File     string
	IsBinary bool
}

// AggForTest is an opaque handle to the unexported per-group accumulator map.
type AggForTest struct {
	inner map[string]*groupAccumulator
}

func NewAggForTest() *AggForTest {
	return &AggForTest{inner: make(map[string]*groupAccumulator)}
}

func ParseNumstatLineForTest(line string) (adds int, dels int, file string, isBinary bool, ok bool) {
	return parseNumstatLine(line)
}

// AggregateForTest dispatches to the internal aggregate() with the chosen
// grouping dimensions. Returns the number of rows skipped due to ignoreOutliers.
func AggregateForTest(meta CommitMetaForTest, rows []NumstatRowForTest, agg *AggForTest, by []GroupBy, ignoreOutliers bool) int {
	internalRows := make([]numstatRow, len(rows))
	for i, r := range rows {
		internalRows[i] = numstatRow{adds: r.Adds, dels: r.Dels, file: r.File, isBinary: r.IsBinary}
	}
	return aggregate(commitMeta{
		hash:       meta.Hash,
		date:       meta.Date,
		authorName: meta.AuthorName,
		subject:    meta.Subject,
		repo:       meta.Repo,
	}, internalRows, agg.inner, by, ignoreOutliers)
}

// NormalizeGroupByForTest exposes the unexported normalizeGroupBy.
func NormalizeGroupByForTest(raw []string) ([]GroupBy, error) {
	return normalizeGroupBy(raw)
}

func FinalizeForTest(agg *AggForTest) []GroupSummary {
	return finalizeGroupSummaries(agg.inner)
}

func ParseNumstatStreamForTest(r io.Reader, onCommit func(CommitMetaForTest, []NumstatRowForTest) error) error {
	return parseNumstatStream(r, func(meta commitMeta, rows []numstatRow) error {
		external := make([]NumstatRowForTest, len(rows))
		for i, row := range rows {
			external[i] = NumstatRowForTest{Adds: row.adds, Dels: row.dels, File: row.file, IsBinary: row.isBinary}
		}
		return onCommit(CommitMetaForTest{
			Hash:       meta.hash,
			Date:       meta.date,
			AuthorName: meta.authorName,
			Subject:    meta.subject,
		}, external)
	})
}

func IsOutlierFileForTest(path string) bool {
	return isOutlierFile(path)
}

// FormatCommitDateBucketForTest exposes the unexported formatCommitDateBucket
// helper so date-formatting can be unit-tested directly.
func FormatCommitDateBucketForTest(t time.Time, by GroupBy) string {
	return formatCommitDateBucket(t, by)
}

// GroupNodeForTest mirrors the unexported groupNode produced by
// GroupSummaries.tree(), exposing labels, totals and children for assertion.
type GroupNodeForTest struct {
	Label    string
	Count    Count
	Children []GroupNodeForTest
}

func GroupSummariesTreeForTest(gs GroupSummaries) []GroupNodeForTest {
	nodes := gs.tree()
	out := make([]GroupNodeForTest, len(nodes))
	for i, n := range nodes {
		out[i] = convertGroupNodeForTest(n)
	}
	return out
}

func convertGroupNodeForTest(n *groupNode) GroupNodeForTest {
	out := GroupNodeForTest{Label: n.Label, Count: n.Count}
	if len(n.Children) > 0 {
		out.Children = make([]GroupNodeForTest, len(n.Children))
		for i, c := range n.Children {
			out.Children[i] = convertGroupNodeForTest(c)
		}
	}
	return out
}
