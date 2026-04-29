package git

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
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
)

type SummaryByTypeOptions struct {
	HistoryOptions `json:",inline"`
	ByType         bool `json:"by_type" flag:"by-type" help:"Group commits by Conventional Commit type" default:"true"`
	IncludeMerges  bool `json:"include_merges" flag:"include-merges" help:"Include merge commits in the summary" default:"false"`
	ProgressEvery  int  `json:"progress_every" flag:"progress-every" help:"Log progress every N commits (0 disables)" default:"1000"`
}

type TypeSummary struct {
	Type  models.CommitType `json:"type"`
	Count `json:",inline"`
}

type TypeSummaries []TypeSummary

type commitMeta struct {
	hash    string
	date    time.Time
	subject string
}

type numstatRow struct {
	adds     int
	dels     int
	file     string
	isBinary bool
}

type typeAccumulator struct {
	adds    int
	dels    int
	commits int
	files   map[string]struct{}
}

// parseNumstatLine parses a single line of `git log --numstat` output.
//
// Returns ok=false for blank lines or malformed rows. For binary files, git emits
// "-\t-\tpath"; we report the file but mark isBinary so callers can include it in
// file counts without affecting line totals.
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

// aggregateByType folds one commit's metadata and numstat rows into the
// per-type accumulator map. Files are tracked uniquely per type so the same path
// touched by two `feat` commits counts as one file under `feat`.
func aggregateByType(meta commitMeta, rows []numstatRow, agg map[models.CommitType]*typeAccumulator) {
	commitType, _, _ := parseCommitTypeAndScope(meta.subject)
	acc := agg[commitType]
	if acc == nil {
		acc = &typeAccumulator{files: make(map[string]struct{})}
		agg[commitType] = acc
	}
	acc.commits++
	for _, r := range rows {
		if !r.isBinary {
			acc.adds += r.adds
			acc.dels += r.dels
		}
		if r.file != "" {
			acc.files[r.file] = struct{}{}
		}
	}
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
// "\x1eCMT\x1f<hash>\x1f<iso-date>\x1f<subject>". Returns ok=false otherwise.
func parseCommitHeader(line string) (commitMeta, bool) {
	if !strings.HasPrefix(line, numstatRecordSep+numstatSentinel+numstatUnitSep) {
		return commitMeta{}, false
	}
	payload := strings.TrimPrefix(line, numstatRecordSep+numstatSentinel+numstatUnitSep)
	fields := strings.Split(payload, numstatUnitSep)
	if len(fields) < 3 {
		return commitMeta{}, false
	}
	date, _ := time.Parse(time.RFC3339, fields[1])
	return commitMeta{hash: fields[0], date: date, subject: fields[2]}, true
}

func buildNumstatArgs(opts SummaryByTypeOptions) []string {
	format := numstatRecordSep + numstatSentinel + numstatUnitSep + "%H" + numstatUnitSep + "%aI" + numstatUnitSep + "%s"
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

// GetCommitStatsByType streams `git log --numstat`, aggregates additions,
// deletions, commits and unique files per Conventional Commit type, and returns
// the result sorted by commit count descending. CommitTypeUnknown sorts last.
func GetCommitStatsByType(opts SummaryByTypeOptions) (TypeSummaries, error) {
	if opts.Path == "" {
		wd, _ := os.Getwd()
		opts.Path = wd
	}
	if err := opts.ParseArgs(); err != nil {
		return nil, err
	}

	args := buildNumstatArgs(opts)
	logger.Tracef("git %s", strings.Join(args, " "))
	clicky.Infof("git log --numstat %s", opts.HistoryOptions.Pretty().ANSI())
	start := time.Now()

	cmd := exec.Command("git", args...)
	cmd.Dir = opts.Path
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("stdout pipe: %w", err)
	}
	stderr := &strings.Builder{}
	cmd.Stderr = stderr
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start git log: %w", err)
	}

	agg := make(map[models.CommitType]*typeAccumulator)
	processed := 0
	progressEvery := opts.ProgressEvery
	parseErr := parseNumstatStream(stdout, func(meta commitMeta, rows []numstatRow) error {
		aggregateByType(meta, rows, agg)
		processed++
		if progressEvery > 0 && processed%progressEvery == 0 {
			clicky.Infof("processed %d commits", processed)
		}
		return nil
	})
	waitErr := cmd.Wait()
	if parseErr != nil {
		return nil, parseErr
	}
	if waitErr != nil {
		return nil, fmt.Errorf("git log exited %v: %s", waitErr, strings.TrimSpace(stderr.String()))
	}

	result := finalizeTypeSummaries(agg)
	clicky.Infof("summarized %d commits across %d types in %v", processed, len(result), time.Since(start))
	return result, nil
}

func finalizeTypeSummaries(agg map[models.CommitType]*typeAccumulator) TypeSummaries {
	out := make(TypeSummaries, 0, len(agg))
	for ct, acc := range agg {
		out = append(out, TypeSummary{
			Type: ct,
			Count: Count{
				Adds:    acc.adds,
				Dels:    acc.dels,
				Commits: acc.commits,
				Files:   len(acc.files),
			},
		})
	}
	sort.SliceStable(out, func(i, j int) bool {
		if (out[i].Type == models.CommitTypeUnknown) != (out[j].Type == models.CommitTypeUnknown) {
			return out[j].Type == models.CommitTypeUnknown
		}
		if out[i].Commits != out[j].Commits {
			return out[i].Commits > out[j].Commits
		}
		return out[i].Type < out[j].Type
	})
	return out
}

// Total returns the grand total Count across all types.
func (ts TypeSummaries) Total() Count {
	var total Count
	for _, t := range ts {
		total.Adds += t.Adds
		total.Dels += t.Dels
		total.Commits += t.Commits
		total.Files += t.Files
	}
	return total
}

func (ts TypeSummaries) Pretty() api.Text {
	t := clicky.Text("Commits by type", "font-bold").NewLine()
	maxTypeWidth := len("unknown")
	for _, row := range ts {
		name := string(row.Type)
		if name == "" {
			name = "unknown"
		}
		if len(name) > maxTypeWidth {
			maxTypeWidth = len(name)
		}
	}
	for _, row := range ts {
		typeName := string(row.Type)
		if typeName == "" {
			typeName = "unknown"
		}
		t = t.Append("  ").
			Append(fmt.Sprintf("%-*s", maxTypeWidth, typeName), "font-mono").
			Append(fmt.Sprintf("  %7d commits", row.Commits), "text-muted").
			Append(fmt.Sprintf("  +%-7d", row.Adds), "text-green-600").
			Append(fmt.Sprintf("-%-7d", row.Dels), "text-red-600").
			Append(fmt.Sprintf("  %5d files", row.Files), "text-muted").
			NewLine()
	}
	total := ts.Total()
	t = t.Append("  ").
		Append(fmt.Sprintf("%-*s", maxTypeWidth, "TOTAL"), "font-bold").
		Append(fmt.Sprintf("  %7d commits", total.Commits), "text-muted").
		Append(fmt.Sprintf("  +%-7d", total.Adds), "text-green-600").
		Append(fmt.Sprintf("-%-7d", total.Dels), "text-red-600").
		Append(fmt.Sprintf("  %5d files", total.Files), "text-muted").
		NewLine()
	return t
}
