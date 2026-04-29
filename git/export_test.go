package git

import (
	"io"

	"github.com/flanksource/gavel/models"
)

// CommitMetaForTest mirrors the unexported commitMeta used by the streaming
// numstat parser, exposed here purely for white-box testing from git_test.
type CommitMetaForTest struct {
	Hash    string
	Subject string
}

// NumstatRowForTest mirrors the unexported numstatRow.
type NumstatRowForTest struct {
	Adds     int
	Dels     int
	File     string
	IsBinary bool
}

// AggForTest is an opaque handle to the unexported per-type accumulator map.
type AggForTest struct {
	inner map[models.CommitType]*typeAccumulator
}

func NewAggForTest() *AggForTest {
	return &AggForTest{inner: make(map[models.CommitType]*typeAccumulator)}
}

func ParseNumstatLineForTest(line string) (adds int, dels int, file string, isBinary bool, ok bool) {
	return parseNumstatLine(line)
}

func AggregateByTypeForTest(meta CommitMetaForTest, rows []NumstatRowForTest, agg *AggForTest) {
	internalRows := make([]numstatRow, len(rows))
	for i, r := range rows {
		internalRows[i] = numstatRow{adds: r.Adds, dels: r.Dels, file: r.File, isBinary: r.IsBinary}
	}
	aggregateByType(commitMeta{hash: meta.Hash, subject: meta.Subject}, internalRows, agg.inner)
}

func FinalizeForTest(agg *AggForTest) TypeSummaries {
	return finalizeTypeSummaries(agg.inner)
}

func ParseNumstatStreamForTest(r io.Reader, onCommit func(CommitMetaForTest, []NumstatRowForTest) error) error {
	return parseNumstatStream(r, func(meta commitMeta, rows []numstatRow) error {
		external := make([]NumstatRowForTest, len(rows))
		for i, row := range rows {
			external[i] = NumstatRowForTest{Adds: row.adds, Dels: row.dels, File: row.file, IsBinary: row.isBinary}
		}
		return onCommit(CommitMetaForTest{Hash: meta.hash, Subject: meta.subject}, external)
	})
}
