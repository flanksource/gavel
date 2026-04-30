package git

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/models"
)

const (
	numstatRecordSep = "\x1e"
	numstatUnitSep   = "\x1f"
	numstatSentinel  = "CMT"
	// groupKeySep separates per-dimension key parts in the internal aggregation
	// map. Using \x1f (unit separator) avoids collision with author names or
	// commit type strings, which never contain control characters.
	groupKeySep = "\x1f"
)

type GroupBy string

const (
	GroupByType        GroupBy = "type"
	GroupByAuthor      GroupBy = "author"
	GroupByCommitYear  GroupBy = "year"
	GroupByCommitMonth GroupBy = "month"
	GroupByCommitWeek  GroupBy = "week"
	GroupByCommitDay   GroupBy = "day"
	GroupByRepo        GroupBy = "repo"
)

const unknownLabel = "unknown"

type SummaryByTypeOptions struct {
	HistoryOptions `json:",inline"`
	GroupBy        []string `json:"group_by" flag:"group-by" help:"Comma-separated grouping dimensions: 'type' (Conventional Commit type), 'author', 'year', 'month', 'week', 'day', 'repo'. Combine for multi-key grouping (e.g. --group-by month,author,type produces one row per (month, author, type) tuple)." default:"type"`
	IgnoreOutliers bool     `json:"ignore_outliers" flag:"ignore-outliers" help:"Skip dependency-lock and generated files (go.sum, package-lock.json, yarn.lock, etc.) when summing additions and deletions"`
	IncludeMerges  bool     `json:"include_merges" flag:"include-merges" help:"Include merge commits in the summary" default:"false"`
	ProgressEvery  int      `json:"progress_every" flag:"progress-every" help:"Log progress every N commits (0 disables)" default:"1000"`

	// RepoPaths is populated from positional Args when those args resolve to
	// existing git work trees. When non-empty it overrides Path; the summary is
	// run per repo and the results are merged with a "repo" group dimension.
	RepoPaths []string `flag:"-" json:"-"`
}

type GroupSummary struct {
	// Group is the per-dimension key parts in the order specified by GroupSummaries.By.
	Group []string `json:"group"`
	Count `json:",inline"`
}

type GroupSummaries struct {
	By      []GroupBy      `json:"by"`
	Groups  []GroupSummary `json:"groups"`
	Skipped int            `json:"skipped_outlier_files,omitempty"`
}

type commitMeta struct {
	hash       string
	date       time.Time
	authorName string
	subject    string
	repo       string
}

type numstatRow struct {
	adds     int
	dels     int
	file     string
	isBinary bool
}

type groupAccumulator struct {
	keyParts []string
	adds     int
	dels     int
	commits  int
	files    map[string]struct{}
}

// outlierBaseNames is the hardcoded set of common dependency-lock and
// machine-generated files. Matched on basename to catch nested copies (e.g.
// frontend/package-lock.json, ui/yarn.lock).
var outlierBaseNames = map[string]struct{}{
	"go.sum":              {},
	"package-lock.json":   {},
	"yarn.lock":           {},
	"pnpm-lock.yaml":      {},
	"npm-shrinkwrap.json": {},
	"bun.lockb":           {},
	"bun.lock":            {},
	"poetry.lock":         {},
	"Pipfile.lock":        {},
	"Cargo.lock":          {},
	"composer.lock":       {},
	"Gemfile.lock":        {},
	"mix.lock":            {},
	"flake.lock":          {},
	"pubspec.lock":        {},
	"Podfile.lock":        {},
}

func isOutlierFile(path string) bool {
	_, ok := outlierBaseNames[filepath.Base(path)]
	return ok
}

// parseNumstatLine parses a single line of `git log --numstat` output.
//
// Returns ok=false for blank lines or malformed rows. For binary files git
// emits "-\t-\tpath"; we report the file but mark isBinary so callers can
// include it in file counts without affecting line totals.
func parseNumstatLine(line string) (adds int, dels int, file string, isBinary bool, ok bool) {
	if line == "" {
		return 0, 0, "", false, false
	}
	fields := strings.SplitN(line, "\t", 3)
	if len(fields) != 3 {
		return 0, 0, "", false, false
	}
	addsField, delsField, path := fields[0], fields[1], fields[2]
	if path == "" {
		return 0, 0, "", false, false
	}
	if addsField == "-" && delsField == "-" {
		return 0, 0, path, true, true
	}
	a, errA := strconv.Atoi(addsField)
	d, errD := strconv.Atoi(delsField)
	if errA != nil || errD != nil {
		return 0, 0, "", false, false
	}
	return a, d, path, false, true
}

// resolveDimensionKey returns the bucket key for a single grouping dimension.
// Empty/missing values fall back to "unknown".
func resolveDimensionKey(meta commitMeta, by GroupBy) string {
	switch by {
	case GroupByAuthor:
		name := strings.TrimSpace(meta.authorName)
		if name == "" {
			return unknownLabel
		}
		return name
	case GroupByCommitYear, GroupByCommitMonth, GroupByCommitWeek, GroupByCommitDay:
		return formatCommitDateBucket(meta.date, by)
	case GroupByRepo:
		if meta.repo == "" {
			return unknownLabel
		}
		return meta.repo
	default: // GroupByType
		commitType, _, _ := parseCommitTypeAndScope(meta.subject)
		if commitType == models.CommitTypeUnknown {
			return unknownLabel
		}
		return string(commitType)
	}
}

// formatCommitDateBucket formats a commit timestamp as the canonical bucket
// label for the given time-based dimension. Returns "unknown" for zero times
// so callers can keep the "missing → unknown" sort behaviour.
func formatCommitDateBucket(t time.Time, by GroupBy) string {
	if t.IsZero() {
		return unknownLabel
	}
	switch by {
	case GroupByCommitYear:
		return t.Format("2006")
	case GroupByCommitMonth:
		return t.Format("2006-01")
	case GroupByCommitWeek:
		year, week := t.ISOWeek()
		return fmt.Sprintf("%04d-W%02d", year, week)
	case GroupByCommitDay:
		return t.Format("2006-01-02")
	}
	return unknownLabel
}

// resolveGroupKey returns the per-dimension key parts for a commit, in the
// order requested by `by`.
func resolveGroupKey(meta commitMeta, by []GroupBy) []string {
	parts := make([]string, len(by))
	for i, dim := range by {
		parts[i] = resolveDimensionKey(meta, dim)
	}
	return parts
}

// aggregate folds one commit's metadata and numstat rows into the per-group
// accumulator map. Files are tracked uniquely per group, so the same path
// touched by two commits in the same group counts as one file. When
// ignoreOutliers is set, dependency-lock files are dropped before aggregating
// adds/dels and are excluded from file counts; the number of dropped rows is
// returned.
func aggregate(meta commitMeta, rows []numstatRow, agg map[string]*groupAccumulator, by []GroupBy, ignoreOutliers bool) (skipped int) {
	parts := resolveGroupKey(meta, by)
	mapKey := strings.Join(parts, groupKeySep)
	acc := agg[mapKey]
	if acc == nil {
		acc = &groupAccumulator{keyParts: parts, files: make(map[string]struct{})}
		agg[mapKey] = acc
	}
	acc.commits++
	for _, r := range rows {
		if ignoreOutliers && isOutlierFile(r.file) {
			skipped++
			continue
		}
		if !r.isBinary {
			acc.adds += r.adds
			acc.dels += r.dels
		}
		if r.file != "" {
			acc.files[r.file] = struct{}{}
		}
	}
	return skipped
}

// parseNumstatStream reads `git log --numstat` output line-by-line and invokes
// onCommit once per commit. Memory is O(rows-in-current-commit) regardless of
// total history size.
func parseNumstatStream(r io.Reader, onCommit func(commitMeta, []numstatRow) error) error {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)

	var (
		current commitMeta
		rows    []numstatRow
		started bool
	)

	flush := func() error {
		if !started {
			return nil
		}
		return onCommit(current, rows)
	}

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		if header, ok := parseCommitHeader(line); ok {
			if err := flush(); err != nil {
				return err
			}
			current = header
			rows = rows[:0]
			started = true
			continue
		}
		if !started {
			continue
		}
		adds, dels, file, isBinary, ok := parseNumstatLine(line)
		if !ok {
			continue
		}
		rows = append(rows, numstatRow{adds: adds, dels: dels, file: file, isBinary: isBinary})
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scan git log output: %w", err)
	}
	return flush()
}

// parseCommitHeader extracts a commitMeta from a header line of the form
// "\x1eCMT\x1f<hash>\x1f<iso-date>\x1f<author>\x1f<subject>". Returns ok=false
// otherwise.
func parseCommitHeader(line string) (commitMeta, bool) {
	if !strings.HasPrefix(line, numstatRecordSep+numstatSentinel+numstatUnitSep) {
		return commitMeta{}, false
	}
	payload := strings.TrimPrefix(line, numstatRecordSep+numstatSentinel+numstatUnitSep)
	fields := strings.Split(payload, numstatUnitSep)
	if len(fields) < 4 {
		return commitMeta{}, false
	}
	date, _ := time.Parse(time.RFC3339, fields[1])
	return commitMeta{
		hash:       fields[0],
		date:       date,
		authorName: fields[2],
		subject:    fields[3],
	}, true
}

func buildNumstatArgs(opts SummaryByTypeOptions) []string {
	format := numstatRecordSep + numstatSentinel + numstatUnitSep + "%H" + numstatUnitSep + "%aI" + numstatUnitSep + "%an" + numstatUnitSep + "%s"
	args := []string{
		"log",
		"--numstat",
		"--date=iso-strict",
		"--pretty=format:" + format,
	}
	if !opts.IncludeMerges {
		args = append(args, "--no-merges")
	}
	args = append(args, "--all")
	if !opts.Since.IsZero() {
		args = append(args, "--since="+opts.Since.Format(time.RFC3339))
	}
	if !opts.Until.IsZero() {
		args = append(args, "--until="+opts.Until.Format(time.RFC3339))
	}
	for _, author := range opts.Author {
		args = append(args, "--author="+author)
	}
	if opts.Message != "" {
		args = append(args, "--grep="+opts.Message)
	}
	if len(opts.FilePaths) > 0 {
		args = append(args, "--")
		args = append(args, opts.FilePaths...)
	}
	return args
}

// normalizeGroupBy parses raw flag input (which may be CSV-separated, repeated,
// or empty) into an ordered list of validated GroupBy values, deduplicated
// while preserving the first occurrence's position.
func normalizeGroupBy(raw []string) ([]GroupBy, error) {
	if len(raw) == 0 {
		return []GroupBy{GroupByType}, nil
	}
	var out []GroupBy
	seen := make(map[GroupBy]struct{})
	for _, entry := range raw {
		for _, part := range strings.Split(entry, ",") {
			part = strings.TrimSpace(strings.ToLower(part))
			if part == "" {
				continue
			}
			var dim GroupBy
			switch part {
			case string(GroupByType):
				dim = GroupByType
			case string(GroupByAuthor):
				dim = GroupByAuthor
			case string(GroupByCommitYear):
				dim = GroupByCommitYear
			case string(GroupByCommitMonth):
				dim = GroupByCommitMonth
			case string(GroupByCommitWeek):
				dim = GroupByCommitWeek
			case string(GroupByCommitDay):
				dim = GroupByCommitDay
			case string(GroupByRepo):
				dim = GroupByRepo
			default:
				return nil, fmt.Errorf("invalid --group-by %q: expected one of type, author, year, month, week, day, repo", part)
			}
			if _, dup := seen[dim]; dup {
				continue
			}
			seen[dim] = struct{}{}
			out = append(out, dim)
		}
	}
	if len(out) == 0 {
		return []GroupBy{GroupByType}, nil
	}
	return out, nil
}

// GetCommitGroupSummaries streams `git log --numstat`, aggregates additions,
// deletions, commits and unique files per group key tuple (the cross product
// of the requested grouping dimensions), and returns the result sorted by
// commit count descending. Tuples containing "unknown" sort last.
//
// When opts.RepoPaths has more than one entry, the summary runs once per repo
// and the rows are merged into a single result. The caller controls how
// repos appear in the output: include "repo" in --group-by to break results
// out per repo; omit it to merge across repos under the requested dimensions.
func GetCommitGroupSummaries(opts SummaryByTypeOptions) (GroupSummaries, error) {
	by, err := normalizeGroupBy(opts.GroupBy)
	if err != nil {
		return GroupSummaries{}, err
	}
	if err := opts.ParseArgs(); err != nil {
		return GroupSummaries{}, err
	}

	repos := opts.RepoPaths
	if len(repos) == 0 {
		path := opts.Path
		if path == "" {
			wd, _ := os.Getwd()
			path = wd
		}
		repos = []string{path}
	}

	agg := make(map[string]*groupAccumulator)
	totalProcessed := 0
	totalSkipped := 0
	start := time.Now()
	for _, repo := range repos {
		label := repoLabel(repo)
		processed, skipped, err := streamRepoIntoAgg(repo, label, opts, by, agg)
		if err != nil {
			return GroupSummaries{}, err
		}
		totalProcessed += processed
		totalSkipped += skipped
	}

	groups := finalizeGroupSummaries(agg)
	clicky.Infof("summarized %d commits across %d groups (%d repos) in %v (skipped %d outlier rows)",
		totalProcessed, len(groups), len(repos), time.Since(start), totalSkipped)
	return GroupSummaries{By: by, Groups: groups, Skipped: totalSkipped}, nil
}

// streamRepoIntoAgg runs `git log --numstat` for one repo and folds each
// commit into the shared accumulator under the supplied repo label.
func streamRepoIntoAgg(repo, label string, opts SummaryByTypeOptions, by []GroupBy, agg map[string]*groupAccumulator) (int, int, error) {
	args := buildNumstatArgs(opts)
	logger.Tracef("git %s (cwd=%s)", strings.Join(args, " "), repo)
	clicky.Infof("git log --numstat %s repo=%s group-by=%s", opts.HistoryOptions.Pretty().ANSI(), label, groupByString(by))

	cmd := exec.Command("git", args...)
	cmd.Dir = repo
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return 0, 0, fmt.Errorf("stdout pipe for %s: %w", repo, err)
	}
	stderr := &strings.Builder{}
	cmd.Stderr = stderr
	if err := cmd.Start(); err != nil {
		return 0, 0, fmt.Errorf("start git log in %s: %w", repo, err)
	}

	processed := 0
	skipped := 0
	progressEvery := opts.ProgressEvery
	parseErr := parseNumstatStream(stdout, func(meta commitMeta, rows []numstatRow) error {
		meta.repo = label
		skipped += aggregate(meta, rows, agg, by, opts.IgnoreOutliers)
		processed++
		if progressEvery > 0 && processed%progressEvery == 0 {
			clicky.Infof("[%s] processed %d commits", label, processed)
		}
		return nil
	})
	waitErr := cmd.Wait()
	if parseErr != nil {
		return processed, skipped, parseErr
	}
	if waitErr != nil {
		return processed, skipped, fmt.Errorf("git log in %s exited %v: %s", repo, waitErr, strings.TrimSpace(stderr.String()))
	}
	return processed, skipped, nil
}

// SplitRepoPathArgs partitions positional args into git work-tree directories
// and the remaining args (commit SHAs, ranges, file paths). An arg is treated
// as a repo path when it points to an existing directory containing a .git
// entry (file or directory, to support worktrees and submodules). Returned
// repo paths are kept in the order the user supplied them. The remaining args
// are passed through unchanged for HistoryOptions.ParseArgs to classify.
func SplitRepoPathArgs(args []string) (repos []string, rest []string, err error) {
	for _, arg := range args {
		if isGitRepoDir(arg) {
			repos = append(repos, arg)
			continue
		}
		rest = append(rest, arg)
	}
	return repos, rest, nil
}

func isGitRepoDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		return false
	}
	if _, err := os.Stat(filepath.Join(path, ".git")); err == nil {
		return true
	}
	return false
}

// repoLabel returns a stable, short label for a repo path suitable for use as
// a group key. Uses the basename, falling back to the absolute path if the
// basename is empty or "."  (i.e., the cwd or root).
func repoLabel(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}
	base := filepath.Base(abs)
	if base == "" || base == "." || base == string(filepath.Separator) {
		return abs
	}
	return base
}

func groupByString(by []GroupBy) string {
	parts := make([]string, len(by))
	for i, b := range by {
		parts[i] = string(b)
	}
	return strings.Join(parts, ",")
}

func hasUnknownPart(parts []string) bool {
	for _, p := range parts {
		if p == unknownLabel {
			return true
		}
	}
	return false
}

func finalizeGroupSummaries(agg map[string]*groupAccumulator) []GroupSummary {
	out := make([]GroupSummary, 0, len(agg))
	for _, acc := range agg {
		out = append(out, GroupSummary{
			Group: acc.keyParts,
			Count: Count{
				Adds:    acc.adds,
				Dels:    acc.dels,
				Commits: acc.commits,
				Files:   len(acc.files),
			},
		})
	}
	sort.SliceStable(out, func(i, j int) bool {
		return lessGroupSummary(out[i], out[j])
	})
	return out
}

// lessGroupSummary defines the canonical sort: rows containing any "unknown"
// dimension sink to the bottom; otherwise commits-desc with a per-dimension
// lexical fallback for stability.
func lessGroupSummary(a, b GroupSummary) bool {
	aUnknown := hasUnknownPart(a.Group)
	bUnknown := hasUnknownPart(b.Group)
	if aUnknown != bUnknown {
		return bUnknown
	}
	if a.Commits != b.Commits {
		return a.Commits > b.Commits
	}
	for k := 0; k < len(a.Group) && k < len(b.Group); k++ {
		if a.Group[k] != b.Group[k] {
			return a.Group[k] < b.Group[k]
		}
	}
	return false
}

// Total returns the grand total Count across all groups.
func (gs GroupSummaries) Total() Count {
	var total Count
	for _, g := range gs.Groups {
		total.Adds += g.Adds
		total.Dels += g.Dels
		total.Commits += g.Commits
		total.Files += g.Files
	}
	return total
}

func (gs GroupSummaries) Pretty() api.Text {
	header := "Commits by " + groupByString(gs.By)
	t := clicky.Text(header, "font-bold").NewLine()

	keyWidth := groupKeyWidth(gs.Groups)
	total := gs.Total()
	roots := gs.tree()
	for i, node := range roots {
		t = renderGroupNode(t, node, "", i == len(roots)-1, keyWidth, total.Commits)
	}

	totalLabel := fmt.Sprintf("%-*s", keyWidth, "TOTAL")
	t = t.Append(totalLabel, "font-bold")
	t = appendCountLine(t, total, true, 0)
	if gs.Skipped > 0 {
		t = t.Append(fmt.Sprintf("  (skipped %d outlier-file rows)", gs.Skipped), "text-muted").NewLine()
	}
	return t
}

// groupNode is one node in the rendered tree of group summaries. Leaf nodes
// (Children empty) correspond to a single GroupSummary row; branch nodes
// represent a partial key prefix and carry the rolled-up Count.
type groupNode struct {
	Label    string
	Count    Count
	Children []*groupNode
}

// tree folds the flat sorted Groups slice into a per-prefix nested structure.
// For len(By)==1 every row becomes a top-level leaf (no nesting). For 2+
// dimensions, rows sharing a leading key part are grouped together with a
// branch node whose Count is the sum of its children.
func (gs GroupSummaries) tree() []*groupNode {
	root := &groupNode{}
	for _, row := range gs.Groups {
		insertGroupRow(root, row.Group, row.Count)
	}
	sortGroupNodesByDim(root, gs.By, 0)
	return root.Children
}

func insertGroupRow(parent *groupNode, key []string, count Count) {
	if len(key) == 0 {
		return
	}
	head := key[0]
	var node *groupNode
	for _, child := range parent.Children {
		if child.Label == head {
			node = child
			break
		}
	}
	if node == nil {
		node = &groupNode{Label: head}
		parent.Children = append(parent.Children, node)
	}
	node.Count.Add(count)
	if len(key) > 1 {
		insertGroupRow(node, key[1:], count)
	}
}

// sortGroupNodesByDim sorts each level using the strategy appropriate for the
// dimension at that depth. Time-based dimensions sort chronologically
// descending (most recent first); all others sort by commit count desc, with
// "unknown" sinking to the bottom and a lexical tiebreaker for stability.
func sortGroupNodesByDim(node *groupNode, by []GroupBy, depth int) {
	timeDim := depth < len(by) && isTimeGroupBy(by[depth])
	sort.SliceStable(node.Children, func(i, j int) bool {
		a, b := node.Children[i], node.Children[j]
		aUnknown := a.Label == unknownLabel
		bUnknown := b.Label == unknownLabel
		if aUnknown != bUnknown {
			return bUnknown
		}
		if timeDim && !aUnknown && !bUnknown {
			// Bucket labels are zero-padded so lexical desc == chronological desc.
			return a.Label > b.Label
		}
		if a.Count.Commits != b.Count.Commits {
			return a.Count.Commits > b.Count.Commits
		}
		return a.Label < b.Label
	})
	for _, child := range node.Children {
		sortGroupNodesByDim(child, by, depth+1)
	}
}

func isTimeGroupBy(by GroupBy) bool {
	switch by {
	case GroupByCommitYear, GroupByCommitMonth, GroupByCommitWeek, GroupByCommitDay:
		return true
	}
	return false
}

// groupKeyWidth returns the longest key part across every dimension of every
// row, with a floor of len("TOTAL") so the totals row aligns.
func groupKeyWidth(rows []GroupSummary) int {
	w := len("TOTAL")
	for _, row := range rows {
		for _, part := range row.Group {
			if len(part) > w {
				w = len(part)
			}
		}
	}
	return w
}

// renderGroupNode renders one tree row. The percentage shown for the row is
// relative to parentCommits, so a child reads as a share of its branch rather
// than of the grand total. Top-level rows are rendered with parentCommits set
// to the grand total.
func renderGroupNode(t api.Text, node *groupNode, prefix string, last bool, keyWidth, parentCommits int) api.Text {
	connector := "├── "
	if last {
		connector = "└── "
	}
	style := "font-mono"
	if len(node.Children) > 0 {
		style = "font-mono font-bold"
	}
	t = t.Append(prefix, "text-muted").
		Append(connector, "text-muted").
		Append(fmt.Sprintf("%-*s", keyWidth, node.Label), style)
	// Files counts are unique per leaf group; summing them at branch nodes
	// would double-count files touched under multiple sub-keys, so suppress.
	t = appendCountLine(t, node.Count, len(node.Children) == 0, parentCommits)

	childPrefix := prefix
	if last {
		childPrefix += "    "
	} else {
		childPrefix += "│   "
	}
	for i, child := range node.Children {
		t = renderGroupNode(t, child, childPrefix, i == len(node.Children)-1, keyWidth, node.Count.Commits)
	}
	return t
}

// appendCountLine writes the metric columns for one tree row. parentCommits
// is the divisor for the percentage column (parent branch's commits, or grand
// total for top-level rows); pass 0 to suppress the percentage (used for the
// TOTAL row, which is always 100%).
func appendCountLine(t api.Text, c Count, showFiles bool, parentCommits int) api.Text {
	t = t.Append(fmt.Sprintf("  %7d commits", c.Commits), "text-muted")
	if parentCommits > 0 {
		pct := 100 * float64(c.Commits) / float64(parentCommits)
		t = t.Append(fmt.Sprintf(" (%5.1f%%)", pct), "text-muted")
	} else {
		t = t.Append(strings.Repeat(" ", 9), "text-muted")
	}
	t = t.Append(fmt.Sprintf("  +%-7d", c.Adds), "text-green-600").
		Append(fmt.Sprintf("-%-7d", c.Dels), "text-red-600")
	if showFiles {
		t = t.Append(fmt.Sprintf("  %5d files", c.Files), "text-muted")
	}
	return t.NewLine()
}
