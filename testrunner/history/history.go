package history

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/gavel/snapshots"
	"github.com/flanksource/gavel/testrunner/parsers"
	testui "github.com/flanksource/gavel/testrunner/ui"
)

var ErrNoHistory = errors.New("no test history found in .gavel; run gavel test once")

type Options struct {
	WorkDir string
	Paths   []string
}

type Report struct {
	WorkDir    string    `json:"work_dir,omitempty"`
	RunCount   int       `json:"run_count"`
	FirstRunAt time.Time `json:"first_run_at,omitempty"`
	LastRunAt  time.Time `json:"last_run_at,omitempty"`
	Tests      []Entry   `json:"tests"`
}

type Entry struct {
	Framework      parsers.Framework `json:"framework,omitempty"`
	WorkDir        string            `json:"work_dir,omitempty"`
	PackagePath    string            `json:"package_path,omitempty"`
	File           string            `json:"file,omitempty"`
	Line           int               `json:"line,omitempty"`
	Suite          []string          `json:"suite,omitempty"`
	Name           string            `json:"name,omitempty"`
	ExecutionCount int               `json:"execution_count"`
	PassCount      int               `json:"pass_count"`
	FailCount      int               `json:"fail_count"`
	PassRate       float64           `json:"pass_rate"`
	MinDuration    time.Duration     `json:"min_duration"`
	AvgDuration    time.Duration     `json:"avg_duration"`
	MaxDuration    time.Duration     `json:"max_duration"`
	AddedAt        time.Time         `json:"added_at,omitempty"`
	LastPassedAt   time.Time         `json:"last_passed_at,omitempty"`
	LastFailedAt   time.Time         `json:"last_failed_at,omitempty"`

	durationTotal time.Duration
}

type runSnapshot struct {
	path    string
	started time.Time
	snap    testui.Snapshot
}

type entryKey struct {
	framework   parsers.Framework
	workDir     string
	packagePath string
	file        string
	suite       string
	name        string
}

func Load(opts Options) (*Report, error) {
	if opts.WorkDir == "" {
		return nil, errors.New("history.Load: workDir is required")
	}
	workDir, err := filepath.Abs(opts.WorkDir)
	if err != nil {
		return nil, fmt.Errorf("resolve workdir: %w", err)
	}

	runs, err := loadRunSnapshots(workDir)
	if err != nil {
		return nil, err
	}
	if len(runs) == 0 {
		return nil, ErrNoHistory
	}

	filters := normalizeFilters(opts.Paths)
	byKey := map[entryKey]*Entry{}
	for _, run := range runs {
		runRoot := workDir
		if run.snap.Git != nil && run.snap.Git.Root != "" {
			runRoot = run.snap.Git.Root
		}
		for _, test := range leafTests(run.snap.Tests) {
			if !isExecuted(test) {
				continue
			}
			testWorkDir := test.WorkDir
			if testWorkDir == "" {
				testWorkDir = runRoot
			}
			if !matchesFilters(test, testWorkDir, filters) {
				continue
			}
			key := makeKey(test, testWorkDir)
			entry := byKey[key]
			if entry == nil {
				entry = &Entry{
					Framework:   key.framework,
					WorkDir:     key.workDir,
					PackagePath: key.packagePath,
					File:        key.file,
					Suite:       append([]string(nil), test.Suite...),
					Name:        test.Name,
					Line:        test.Line,
					AddedAt:     run.started,
				}
				byKey[key] = entry
			}
			entry.record(test, run.started)
		}
	}

	report := &Report{
		WorkDir:    workDir,
		RunCount:   len(runs),
		FirstRunAt: runs[0].started,
		LastRunAt:  runs[len(runs)-1].started,
	}
	for _, entry := range byKey {
		entry.finish()
		report.Tests = append(report.Tests, *entry)
	}
	sortEntries(report.Tests)
	return report, nil
}

func loadRunSnapshots(workDir string) ([]runSnapshot, error) {
	dir := filepath.Join(workDir, snapshots.Dir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("read %s: %w", dir, err)
	}

	var runs []runSnapshot
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasPrefix(name, snapshots.PerRunPrefix) || !strings.HasSuffix(name, ".json") {
			continue
		}
		path := filepath.Join(dir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read history snapshot %s: %w", path, err)
		}
		var snap testui.Snapshot
		if err := json.Unmarshal(data, &snap); err != nil {
			return nil, fmt.Errorf("decode history snapshot %s: %w", path, err)
		}
		started := snapshotTime(snap, name, path, entry)
		runs = append(runs, runSnapshot{path: path, started: started, snap: snap})
	}

	sort.Slice(runs, func(i, j int) bool {
		if runs[i].started.Equal(runs[j].started) {
			return runs[i].path < runs[j].path
		}
		return runs[i].started.Before(runs[j].started)
	})
	return runs, nil
}

func snapshotTime(snap testui.Snapshot, name, path string, entry os.DirEntry) time.Time {
	if snap.Metadata != nil && !snap.Metadata.Started.IsZero() {
		return snap.Metadata.Started.UTC()
	}
	stem := strings.TrimSuffix(strings.TrimPrefix(name, snapshots.PerRunPrefix), ".json")
	if ts, err := time.Parse(snapshots.PerRunTimestampLayout, stem); err == nil {
		return ts.UTC()
	}
	if info, err := entry.Info(); err == nil {
		return info.ModTime().UTC()
	}
	if info, err := os.Stat(path); err == nil {
		return info.ModTime().UTC()
	}
	return time.Time{}
}

func leafTests(tests []parsers.Test) []parsers.Test {
	var leaves []parsers.Test
	var walk func(parsers.Test)
	walk = func(test parsers.Test) {
		if len(test.Children) == 0 {
			leaves = append(leaves, test)
			return
		}
		for _, child := range test.Children {
			walk(child)
		}
	}
	for _, test := range tests {
		walk(test)
	}
	return leaves
}

func isExecuted(test parsers.Test) bool {
	if test.Skipped || test.Pending || test.Running {
		return false
	}
	return test.Passed || test.Failed || test.TimedOut
}

func makeKey(test parsers.Test, workDir string) entryKey {
	file := normalizeFile(test.File, workDir)
	return entryKey{
		framework:   test.Framework,
		workDir:     cleanWorkDir(workDir),
		packagePath: cleanPath(test.PackagePath),
		file:        file,
		suite:       strings.Join(test.Suite, "\x00"),
		name:        test.Name,
	}
}

func (e *Entry) record(test parsers.Test, ranAt time.Time) {
	e.ExecutionCount++
	if e.Line == 0 && test.Line > 0 {
		e.Line = test.Line
	}
	if test.TimedOut || test.Failed {
		e.FailCount++
		if ranAt.After(e.LastFailedAt) || e.LastFailedAt.IsZero() {
			e.LastFailedAt = ranAt
		}
	} else if test.Passed {
		e.PassCount++
		if ranAt.After(e.LastPassedAt) || e.LastPassedAt.IsZero() {
			e.LastPassedAt = ranAt
		}
	}
	if e.ExecutionCount == 1 || test.Duration < e.MinDuration {
		e.MinDuration = test.Duration
	}
	if test.Duration > e.MaxDuration {
		e.MaxDuration = test.Duration
	}
	e.durationTotal += test.Duration
}

func (e *Entry) finish() {
	if e.ExecutionCount > 0 {
		e.AvgDuration = time.Duration(int64(e.durationTotal) / int64(e.ExecutionCount))
		e.PassRate = float64(e.PassCount) / float64(e.ExecutionCount)
	}
}

func normalizeFilters(paths []string) []string {
	filters := make([]string, 0, len(paths))
	for _, path := range paths {
		if filter := cleanPath(path); filter != "" && filter != "." {
			filters = append(filters, filter)
		}
	}
	return filters
}

func matchesFilters(test parsers.Test, workDir string, filters []string) bool {
	if len(filters) == 0 {
		return true
	}
	candidates := filterCandidates(test, workDir)
	for _, filter := range filters {
		for _, candidate := range candidates {
			if candidate == filter || strings.HasPrefix(candidate, filter+"/") {
				return true
			}
		}
	}
	return false
}

func filterCandidates(test parsers.Test, workDir string) []string {
	var candidates []string
	add := func(path string) {
		path = cleanPath(path)
		if path == "" || path == "." {
			return
		}
		candidates = append(candidates, path)
	}
	add(test.PackagePath)
	add(strings.TrimPrefix(test.PackagePath, "./"))

	file := normalizeFile(test.File, workDir)
	add(file)
	if file != "" && file != "." {
		add(filepath.ToSlash(filepath.Dir(file)))
	}
	return candidates
}

func normalizeFile(path, workDir string) string {
	if path == "" {
		return ""
	}
	if filepath.IsAbs(path) {
		if workDir != "" {
			if rel, err := filepath.Rel(workDir, path); err == nil && !strings.HasPrefix(rel, "..") {
				path = rel
			}
		}
	}
	return cleanPath(path)
}

func cleanPath(path string) string {
	path = filepath.ToSlash(strings.TrimSpace(path))
	path = strings.TrimPrefix(path, "./")
	path = strings.Trim(path, "/")
	if path == "" {
		return ""
	}
	return path
}

func cleanWorkDir(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	return filepath.ToSlash(filepath.Clean(path))
}

func sortEntries(entries []Entry) {
	sort.Slice(entries, func(i, j int) bool {
		a, b := entries[i], entries[j]
		if a.PackagePath != b.PackagePath {
			return a.PackagePath < b.PackagePath
		}
		if a.File != b.File {
			return a.File < b.File
		}
		if strings.Join(a.Suite, "\x00") != strings.Join(b.Suite, "\x00") {
			return strings.Join(a.Suite, "\x00") < strings.Join(b.Suite, "\x00")
		}
		if a.Name != b.Name {
			return a.Name < b.Name
		}
		return a.Framework < b.Framework
	})
}

func (r Report) Pretty() api.Text {
	label := fmt.Sprintf("Test history: %d tests across %d runs", len(r.Tests), r.RunCount)
	t := clicky.Text(label, "bold text-blue-500")
	if !r.FirstRunAt.IsZero() && !r.LastRunAt.IsZero() {
		t = t.Space().Append(fmt.Sprintf("(%s to %s)", formatDate(r.FirstRunAt), formatDate(r.LastRunAt)), "text-muted")
	}
	return t
}

func (r Report) GetChildren() []api.TreeNode {
	type pkgBucket struct {
		name    string
		entries []Entry
	}
	byPkg := map[string]*pkgBucket{}
	var order []string
	for _, entry := range r.Tests {
		pkg := entry.PackagePath
		if pkg == "" {
			pkg = "(unknown package)"
		}
		bucket := byPkg[pkg]
		if bucket == nil {
			bucket = &pkgBucket{name: pkg}
			byPkg[pkg] = bucket
			order = append(order, pkg)
		}
		bucket.entries = append(bucket.entries, entry)
	}
	sort.Strings(order)

	children := make([]api.TreeNode, 0, len(order))
	for _, pkg := range order {
		bucket := byPkg[pkg]
		children = append(children, buildPackageNode(bucket.name, bucket.entries))
	}
	return children
}

type groupNode struct {
	label    string
	style    string
	entries  []Entry
	children []api.TreeNode
}

func (n *groupNode) Pretty() api.Text {
	t := clicky.Text(n.label, n.style)
	if len(n.entries) > 0 {
		t = t.Space().Append(formatGroupStats(n.entries), "text-muted")
	}
	return t
}

func (n *groupNode) GetChildren() []api.TreeNode {
	return n.children
}

type testNode struct {
	entry Entry
}

func (n *testNode) Pretty() api.Text {
	e := n.entry
	style := "text-green-600"
	prefix := "PASS "
	if e.lastFailedAfterPass() {
		style = "text-red-600"
		prefix = "FAIL "
	}
	t := clicky.Text(prefix, style).Append(e.Name, "bold wrap-space")
	t = t.Space().Append(fmt.Sprintf("exec %d", e.ExecutionCount), "text-muted")
	t = t.Space().Append(fmt.Sprintf("pass %.0f%%", e.PassRate*100), passRateStyle(e.PassRate))
	t = t.Space().Append(fmt.Sprintf("min %s", e.MinDuration), "text-muted")
	t = t.Space().Append(fmt.Sprintf("avg %s", e.AvgDuration), "text-muted")
	t = t.Space().Append(fmt.Sprintf("max %s", e.MaxDuration), "text-muted")
	t = t.Space().Append("added "+formatDate(e.AddedAt), "text-muted")
	t = t.Space().Append("last pass "+formatOptionalDate(e.LastPassedAt), "text-green-600")
	t = t.Space().Append("last fail "+formatOptionalDate(e.LastFailedAt), "text-red-600")
	return t
}

func (n *testNode) GetChildren() []api.TreeNode { return nil }

func buildPackageNode(pkg string, entries []Entry) api.TreeNode {
	files := map[string][]Entry{}
	var order []string
	for _, entry := range entries {
		file := entry.File
		if file == "" {
			file = "(no file)"
		}
		if _, ok := files[file]; !ok {
			order = append(order, file)
		}
		files[file] = append(files[file], entry)
	}
	sort.Strings(order)

	children := make([]api.TreeNode, 0, len(order))
	for _, file := range order {
		children = append(children, buildFileNode(file, files[file]))
	}
	return &groupNode{label: pkg, style: "bold", entries: entries, children: children}
}

func buildFileNode(file string, entries []Entry) api.TreeNode {
	suiteGroups := map[string][]Entry{}
	var order []string
	for _, entry := range entries {
		suite := strings.Join(entry.Suite, " > ")
		if suite == "" {
			suite = "(no suite)"
		}
		if _, ok := suiteGroups[suite]; !ok {
			order = append(order, suite)
		}
		suiteGroups[suite] = append(suiteGroups[suite], entry)
	}
	sort.Strings(order)

	var children []api.TreeNode
	for _, suite := range order {
		groupEntries := suiteGroups[suite]
		if suite == "(no suite)" {
			for _, entry := range groupEntries {
				children = append(children, &testNode{entry: entry})
			}
			continue
		}
		var suiteChildren []api.TreeNode
		for _, entry := range groupEntries {
			suiteChildren = append(suiteChildren, &testNode{entry: entry})
		}
		children = append(children, &groupNode{label: suite, style: "text-muted", entries: groupEntries, children: suiteChildren})
	}
	return &groupNode{label: file, style: "text-muted", entries: entries, children: children}
}

func (e Entry) lastFailedAfterPass() bool {
	if e.LastFailedAt.IsZero() {
		return false
	}
	if e.LastPassedAt.IsZero() {
		return true
	}
	return e.LastFailedAt.After(e.LastPassedAt)
}

func formatGroupStats(entries []Entry) string {
	var execs, passes int
	for _, entry := range entries {
		execs += entry.ExecutionCount
		passes += entry.PassCount
	}
	rate := 0.0
	if execs > 0 {
		rate = float64(passes) / float64(execs)
	}
	return fmt.Sprintf("(%d tests, %d execs, %.0f%% pass)", len(entries), execs, rate*100)
}

func passRateStyle(rate float64) string {
	switch {
	case rate >= 1:
		return "text-green-600"
	case rate <= 0:
		return "text-red-600"
	default:
		return "text-yellow-600"
	}
}

func formatDate(ts time.Time) string {
	if ts.IsZero() {
		return "never"
	}
	return ts.UTC().Format("2006-01-02")
}

func formatOptionalDate(ts time.Time) string {
	if ts.IsZero() {
		return "never"
	}
	return formatDate(ts)
}
