package snapshots

import (
	"errors"
	"path/filepath"
	"slices"
	"time"

	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/testrunner/parsers"
	testui "github.com/flanksource/gavel/testrunner/ui"
)

// Problems are the current lint violations and failed tests of a set of files,
// as the workspace's recorded runs last saw them.
type Problems struct {
	// Lint holds, per linter and work dir, the violations of the newest run of
	// that linter over each file.
	Lint []*linters.LinterResult
	// Tests are the failed leaf tests of the newest test run located in one of
	// the files, each with an absolute WorkDir and no children.
	Tests []parsers.Test
	// LintEnded is when the newest run that linted each file ended.
	LintEnded  map[string]time.Time
	TestsEnded time.Time
}

type lintKey struct{ linter, workDir string }

type lintAggregate struct {
	files   []string
	covered map[lintKey]map[string]bool
	settled map[string]bool
	results map[lintKey]*linters.LinterResult
	order   []lintKey
	ended   map[string]time.Time
}

// CurrentProblems aggregates workDir's run-*.json snapshots into the problems
// of files (absolute paths). Runs are visited newest first: a lint run settles
// each file it covered for each linter it ran, so a violation fixed by a later
// run limited to one file disappears while other files keep older results. The
// walk ends once a whole-tree lint run has covered every file and a test run
// has been read.
func CurrentProblems(workDir string, files []string) (*Problems, error) {
	if workDir == "" {
		return nil, errors.New("snapshots.CurrentProblems: workDir is required")
	}
	var wanted []string
	for _, file := range files {
		wanted = append(wanted, canonicalPath(file))
	}
	slices.Sort(wanted)
	wanted = slices.Compact(wanted)
	store, err := OpenStore(workDir)
	if err != nil {
		return nil, err
	}
	objects, err := runObjects(store)
	if err != nil {
		return nil, err
	}
	agg := &lintAggregate{files: wanted, covered: map[lintKey]map[string]bool{}, settled: map[string]bool{},
		results: map[lintKey]*linters.LinterResult{}, ended: map[string]time.Time{}}
	problems := &Problems{}
	testsDone := false
	for _, object := range objects {
		lintDone := len(agg.settled) == len(wanted)
		if lintDone && testsDone {
			break
		}
		hasTests, hasLint, err := runContents(store, workDir, object.Key)
		if err != nil {
			return nil, err
		}
		if (!hasTests || testsDone) && (!hasLint || lintDone) {
			continue
		}
		snap, err := loadRun(store, workDir, object.Key)
		if err != nil {
			return nil, err
		}
		if snap.Status.Running {
			continue
		}
		if hasTests && !testsDone {
			problems.Tests, problems.TestsEnded, testsDone = failedTestsIn(snap, wanted), snapshotEnded(snap), true
		}
		if hasLint && !lintDone {
			agg.add(snap)
		}
	}
	for _, key := range agg.order {
		problems.Lint = append(problems.Lint, agg.results[key])
	}
	problems.LintEnded = agg.ended
	return problems, nil
}

func (a *lintAggregate) add(snap *testui.Snapshot) {
	scopes, whole := lintRunScope(snap)
	ended := snapshotEnded(snap)
	for _, result := range snap.Lint {
		if result == nil || result.Skipped {
			continue
		}
		workDir := canonicalPath(result.WorkDir)
		key := lintKey{result.Linter, workDir}
		var newly []string
		for _, file := range a.files {
			if a.settled[file] || a.covered[key][file] || !inLintScope(file, workDir, scopes) {
				continue
			}
			if a.covered[key] == nil {
				a.covered[key] = map[string]bool{}
			}
			a.covered[key][file] = true
			if _, ok := a.ended[file]; !ok {
				a.ended[file] = ended
			}
			newly = append(newly, file)
		}
		if len(newly) > 0 {
			a.collect(key, result, workDir, newly)
		}
	}
	if !whole {
		return
	}
	for _, file := range a.files {
		if inLintScope(file, "", scopes) {
			a.settled[file] = true
		}
	}
}

// collect adds result's violations located in files to the aggregate result of key.
func (a *lintAggregate) collect(key lintKey, result *linters.LinterResult, workDir string, files []string) {
	located := *result
	located.WorkDir = workDir
	located.Violations = slices.Clone(result.Violations)
	for i, violation := range located.Violations {
		if filepath.IsAbs(violation.File) {
			located.Violations[i].File = canonicalPath(violation.File)
		}
	}
	located.Ignored = nil
	linters.FilterViolationsByUserScope([]*linters.LinterResult{&located}, workDir, files)
	located.Violations = slices.DeleteFunc(located.Violations, func(v models.Violation) bool { return v.File == "" })
	merged, ok := a.results[key]
	if !ok {
		a.results[key] = &located
		a.order = append(a.order, key)
		return
	}
	merged.Violations = append(merged.Violations, located.Violations...)
}

// lintRunScope is what a lint run covered: its LintFiles, else the files of
// its change selection, else the whole tree under its git root (whole). Nil
// scopes cover every file.
func lintRunScope(snap *testui.Snapshot) (scopes []string, whole bool) {
	if snap.Metadata != nil && len(snap.Metadata.LintFiles) > 0 {
		for _, file := range snap.Metadata.LintFiles {
			scopes = append(scopes, canonicalPath(file))
		}
		return scopes, false
	}
	if snap.Metadata != nil && len(snap.Metadata.Changes) > 0 {
		for _, set := range snap.Metadata.Changes {
			for _, change := range set.Changes {
				scopes = append(scopes, canonicalPath(filepath.Join(set.Root, filepath.FromSlash(change.Path))))
			}
		}
		return scopes, false
	}
	if snap.Git != nil && snap.Git.Root != "" {
		return []string{canonicalPath(snap.Git.Root)}, true
	}
	return nil, true
}

func inLintScope(file, workDir string, scopes []string) bool {
	if workDir != "" && !linters.PathMatchesAnyScope(file, []string{workDir}) {
		return false
	}
	return scopes == nil || linters.PathMatchesAnyScope(file, scopes)
}

// failedTestsIn returns the failed tests of snap located in files, reporting a
// failed suite only when none of its descendants is.
func failedTestsIn(snap *testui.Snapshot, files []string) []parsers.Test {
	root := ""
	if snap.Git != nil {
		root = snap.Git.Root
	}
	var walk func(tests []parsers.Test) []parsers.Test
	walk = func(tests []parsers.Test) []parsers.Test {
		var failed []parsers.Test
		for _, test := range tests {
			if below := walk(test.Children); len(below) > 0 {
				failed = append(failed, below...)
				continue
			}
			if !test.Failed || test.File == "" || test.Line == 0 {
				continue
			}
			workDir := test.WorkDir
			if workDir == "" {
				workDir = root
			}
			file := test.File
			if !filepath.IsAbs(file) {
				file = filepath.Join(workDir, file)
			}
			if !slices.Contains(files, canonicalPath(file)) {
				continue
			}
			test.WorkDir = canonicalPath(workDir)
			test.File = canonicalPath(file)
			test.Children = nil
			failed = append(failed, test)
		}
		return failed
	}
	return walk(snap.Tests)
}

func snapshotEnded(snap *testui.Snapshot) time.Time {
	if snap.Metadata == nil {
		return time.Time{}
	}
	return snap.Metadata.Ended.UTC()
}

// canonicalPath resolves symlinks so paths recorded by a run compare equal to
// the files the caller names; a path that no longer exists stays as recorded.
func canonicalPath(path string) string {
	if path == "" {
		return ""
	}
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		return resolved
	}
	return filepath.Clean(path)
}
