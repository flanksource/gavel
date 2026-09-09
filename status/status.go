package status

import (
	"context"
	"errors"
	"fmt"
	"time"

	rpchttp "github.com/flanksource/clicky/rpc/http"
	"github.com/flanksource/gavel/internal/prompting"
	"github.com/flanksource/repomap"
)

type FileState string

const (
	StateStaged    FileState = "staged"
	StateUnstaged  FileState = "unstaged"
	StateBoth      FileState = "both"
	StateUntracked FileState = "untracked"
	StateConflict  FileState = "conflict"
)

type ChangeKind string

const (
	KindModified   ChangeKind = "M"
	KindAdded      ChangeKind = "A"
	KindDeleted    ChangeKind = "D"
	KindRenamed    ChangeKind = "R"
	KindCopied     ChangeKind = "C"
	KindTypeChange ChangeKind = "T"
	KindUntracked  ChangeKind = "?"
	KindUnknown    ChangeKind = ""
)

type FileStatus struct {
	Path         string
	PreviousPath string
	State        FileState
	StagedKind   ChangeKind
	WorkKind     ChangeKind
	Adds         int
	Dels         int
	AISummary    string
	AIError      string
	AIStatus     AISummaryStatus
	FileMap      *repomap.FileMap
	RepomapError error
	TestStatus   TestStatus
	LintStatus   LintStatus
	// Problems holds the individual failing tests and lint violations for
	// this file, pulled from the last snapshot. TestStatus/LintStatus carry
	// the counts for the inline badges; Problems carries the messages the
	// trailing "Problems" section renders.
	Problems       []Problem
	ResultsStale   bool
	ConflictReason ConflictReason
	// ModifiedAt is the working-tree mtime of the file. Zero when the file
	// no longer exists on disk (e.g. deletions) or stat fails.
	ModifiedAt time.Time
}

// ConflictReason explains why a FileStatus is in StateConflict. Empty when
// the file is not conflicted.
type ConflictReason string

const (
	// ConflictReasonUnmerged means git's porcelain output flagged the file as
	// unmerged (UU/AA/DD or any pair containing 'U').
	ConflictReasonUnmerged ConflictReason = "unmerged"
	// ConflictReasonMarker means the working-tree content contains unresolved
	// conflict markers even though git no longer flags the file as unmerged
	// (e.g., the user `git add`-ed a file that still has <<<<<<< / >>>>>>>).
	ConflictReasonMarker ConflictReason = "marker"
)

type TestStatus struct {
	Passed  int
	Failed  int
	Skipped int
}

type LintStatus struct {
	Errors   int
	Warnings int
	Infos    int
}

type Result struct {
	WorkDir      string
	Branch       string
	Files        []FileStatus
	ResultsSHA   string
	CurrentSHA   string
	ResultsStale bool
	// Verbose expands the trailing "Problems" section to every failing
	// test / lint violation with full multi-line messages. When false the
	// section caps each file and prints a `gavel status -v` hint. Excluded
	// from JSON output so `--format json` stays a pure data dump.
	Verbose bool `json:"-"`
}

type Options struct {
	NoRepomap bool
	NoResults bool
	// FolderFilter, when non-empty, restricts the result to files at or
	// below this slash-separated path relative to workDir. The filter is
	// applied before line-count, repomap, and snapshot enrichment so we
	// don't waste work on files that will be dropped.
	FolderFilter string
	// Verbose is threaded onto Result.Verbose so the renderer can expand the
	// Problems section. Set from the global `-v` count by the CLI.
	Verbose bool
	Context context.Context `json:"-"`
	Summary *SummaryOptions `json:"-"`
}

// fetchFileMapFunc is the indirection point for repomap lookups so tests can
// stub enrichment without touching the real repomap configuration.
var fetchFileMapFunc = repomap.GetFileMap

func Gather(workDir string, opts Options) (*Result, error) {
	result, err := GatherBase(workDir, opts)
	if err != nil {
		return nil, err
	}
	if opts.Summary == nil {
		return result, nil
	}

	prompting.Prepare()
	result.PrepareAISummaries()
	summary := *opts.Summary
	summary.WorkDir, summary.Files = workDir, result.Files
	for update := range StreamAISummaries(opts.Context, summary) {
		result.ApplyAISummaryUpdate(update)
	}

	return result, nil
}

func GatherBase(workDir string, opts Options) (*Result, error) {
	if workDir == "" {
		return nil, errors.New("status.GatherBase: workDir is required")
	}
	ctx := opts.Context
	if ctx == nil {
		ctx = context.Background()
	}

	branch, err := currentBranch(ctx, workDir)
	if err != nil {
		return nil, err
	}

	raw, err := runGitStatus(ctx, workDir)
	if err != nil {
		return nil, err
	}

	files, err := parseStatusPorcelain(raw)
	if err != nil {
		return nil, fmt.Errorf("parse git status: %w", err)
	}
	files = filterGavelCache(files)
	files = filterByFolder(files, opts.FolderFilter)
	stopFile := rpchttp.Track(ctx, "file")
	files = filterGitIgnored(files, workDir)
	stopFile()

	if err := enrichWithLineCounts(ctx, workDir, files); err != nil {
		return nil, err
	}

	stopFile = rpchttp.Track(ctx, "file")
	enrichWithModTime(workDir, files)
	enrichWithConflictMarkers(workDir, files)
	stopFile()

	if !opts.NoRepomap {
		for i := range files {
			enrichWithRepomap(&files[i], workDir)
		}
	}

	result := &Result{
		WorkDir: workDir,
		Branch:  branch,
		Files:   files,
		Verbose: opts.Verbose,
	}

	if !opts.NoResults {
		if err := enrichWithSnapshot(ctx, workDir, result); err != nil {
			return nil, err
		}
	}

	return result, nil
}
