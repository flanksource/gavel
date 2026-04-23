package status

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	clickyai "github.com/flanksource/clicky/ai"
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
	ResultsStale bool
}

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
}

type Options struct {
	NoRepomap    bool
	Agent        clickyai.Agent  `json:"-"`
	Context      context.Context `json:"-"`
	AIMaxWorkers int             `json:"-"`
}

// fetchFileMapFunc is the indirection point for repomap lookups so tests can
// stub enrichment without touching the real repomap configuration.
var fetchFileMapFunc = repomap.GetFileMap

func Gather(workDir string, opts Options) (*Result, error) {
	result, err := GatherBase(workDir, opts)
	if err != nil {
		return nil, err
	}
	if opts.Agent == nil {
		return result, nil
	}

	prompting.Prepare()
	result.PrepareAISummaries()
	for update := range StreamAISummaries(opts.Context, workDir, opts.Agent, result.Files, opts.AIMaxWorkers) {
		result.ApplyAISummaryUpdate(update)
	}

	return result, nil
}

func GatherBase(workDir string, opts Options) (*Result, error) {
	if workDir == "" {
		return nil, errors.New("status.GatherBase: workDir is required")
	}

	branch, err := currentBranch(workDir)
	if err != nil {
		return nil, err
	}

	raw, err := runGitStatus(workDir)
	if err != nil {
		return nil, err
	}

	files, err := parseStatusPorcelain(raw)
	if err != nil {
		return nil, fmt.Errorf("parse git status: %w", err)
	}
	files = filterGavelCache(files)

	if err := enrichWithLineCounts(workDir, files); err != nil {
		return nil, err
	}

	if !opts.NoRepomap {
		for i := range files {
			enrichWithRepomap(&files[i], workDir)
		}
	}

	result := &Result{
		WorkDir: workDir,
		Branch:  branch,
		Files:   files,
	}

	if err := enrichWithSnapshot(workDir, result); err != nil {
		return nil, err
	}

	return result, nil
}

func runGitStatus(workDir string) ([]byte, error) {
	cmd := exec.Command("git", "status", "--porcelain=v1", "-z")
	cmd.Dir = workDir
	var stderr strings.Builder
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git status: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	return out, nil
}

func currentBranch(workDir string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = workDir
	out, err := cmd.Output()
	if err != nil {
		// No HEAD yet (fresh repo) or other transient state — report blank.
		return "", nil
	}
	name := strings.TrimSpace(string(out))
	if name == "HEAD" {
		return "(detached)", nil
	}
	return name, nil
}

// parseStatusPorcelain decodes `git status --porcelain=v1 -z` output. Each
// record is: 2 status bytes, 1 space, path, NUL terminator. For R/C entries
// the record is followed by another NUL-terminated path containing the
// original (pre-rename) name, per git's v1 porcelain format.
func parseStatusPorcelain(raw []byte) ([]FileStatus, error) {
	records := splitNullDelimited(raw)
	var files []FileStatus
	for i := 0; i < len(records); i++ {
		rec := records[i]
		if len(rec) < 3 {
			continue
		}
		stagedByte, workByte := rec[0], rec[1]
		payload := string(rec[3:])

		fs := FileStatus{
			Path:       payload,
			StagedKind: mapKindByte(stagedByte),
			WorkKind:   mapKindByte(workByte),
			State:      deriveState(stagedByte, workByte),
		}

		if stagedByte == 'R' || stagedByte == 'C' || workByte == 'R' || workByte == 'C' {
			if i+1 >= len(records) {
				return nil, fmt.Errorf("rename record missing original path: %q", rec)
			}
			i++
			fs.PreviousPath = string(records[i])
		}

		files = append(files, fs)
	}
	return files, nil
}

// filterGavelCache drops .gavel/ entries from the status display since they
// are managed by gavel itself and would show up as untracked otherwise.
func filterGavelCache(files []FileStatus) []FileStatus {
	out := files[:0]
	for _, f := range files {
		if strings.HasPrefix(f.Path, ".gavel/") || f.Path == ".gavel" {
			continue
		}
		out = append(out, f)
	}
	return out
}

func splitNullDelimited(raw []byte) []string {
	if len(raw) == 0 {
		return nil
	}
	s := string(raw)
	s = strings.TrimRight(s, "\x00")
	if s == "" {
		return nil
	}
	return strings.Split(s, "\x00")
}

func mapKindByte(b byte) ChangeKind {
	switch b {
	case 'M':
		return KindModified
	case 'A':
		return KindAdded
	case 'D':
		return KindDeleted
	case 'R':
		return KindRenamed
	case 'C':
		return KindCopied
	case 'T':
		return KindTypeChange
	case '?':
		return KindUntracked
	case ' ':
		return KindUnknown
	case 'U':
		return KindModified
	default:
		return KindUnknown
	}
}

func deriveState(stagedByte, workByte byte) FileState {
	switch {
	case stagedByte == '?' && workByte == '?':
		return StateUntracked
	case isConflictPair(stagedByte, workByte):
		return StateConflict
	case stagedByte != ' ' && workByte != ' ':
		return StateBoth
	case stagedByte != ' ':
		return StateStaged
	default:
		return StateUnstaged
	}
}

// isConflictPair returns true for the unmerged-path combinations git uses in
// porcelain output: any pair with a 'U', plus the special cases AA and DD.
func isConflictPair(staged, work byte) bool {
	if staged == 'U' || work == 'U' {
		return true
	}
	if staged == 'A' && work == 'A' {
		return true
	}
	if staged == 'D' && work == 'D' {
		return true
	}
	return false
}

// enrichWithLineCounts fills in Adds/Dels for each FileStatus by combining
// staged and unstaged numstat output, and by counting lines of any untracked
// file directly (git numstat does not report untracked content).
func enrichWithLineCounts(workDir string, files []FileStatus) error {
	staged, err := numstat(workDir, true)
	if err != nil {
		return err
	}
	unstaged, err := numstat(workDir, false)
	if err != nil {
		return err
	}

	for i := range files {
		f := &files[i]
		key := f.Path
		if s, ok := staged[key]; ok {
			f.Adds += s.adds
			f.Dels += s.dels
		}
		if u, ok := unstaged[key]; ok {
			f.Adds += u.adds
			f.Dels += u.dels
		}
		if f.State == StateUntracked && f.Adds == 0 && f.Dels == 0 {
			f.Adds = countFileLines(filepath.Join(workDir, f.Path))
		}
	}
	return nil
}

type numstatEntry struct {
	adds int
	dels int
}

func numstat(workDir string, cached bool) (map[string]numstatEntry, error) {
	args := []string{"diff", "--numstat", "-z", "--find-renames"}
	if cached {
		args = append(args, "--cached")
	}
	cmd := exec.Command("git", args...)
	cmd.Dir = workDir
	var stderr strings.Builder
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git diff --numstat (cached=%v): %w: %s", cached, err, strings.TrimSpace(stderr.String()))
	}

	result := map[string]numstatEntry{}
	records := splitNullDelimited(out)
	for i := 0; i < len(records); i++ {
		rec := records[i]
		if rec == "" {
			continue
		}
		fields := strings.Split(rec, "\t")
		if len(fields) < 3 {
			continue
		}
		adds, dels := parseNumstatCount(fields[0]), parseNumstatCount(fields[1])
		path := fields[2]
		if path == "" && i+2 < len(records) {
			// Rename entry: "adds\tdels\t" then two NUL-separated paths.
			i++
			i++
			path = records[i]
		}
		result[path] = numstatEntry{adds: adds, dels: dels}
	}
	return result, nil
}

func parseNumstatCount(s string) int {
	if s == "-" {
		return 0
	}
	n := 0
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0
		}
		n = n*10 + int(r-'0')
	}
	return n
}

func countFileLines(path string) int {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	if len(data) == 0 {
		return 0
	}
	n := strings.Count(string(data), "\n")
	if !strings.HasSuffix(string(data), "\n") {
		n++
	}
	return n
}

func enrichWithRepomap(fs *FileStatus, workDir string) {
	// Untracked files usually aren't in the arch index, but repomap can still
	// classify by extension/path — try anyway, ignore empty results.
	abs := filepath.Join(workDir, fs.Path)
	fm, err := fetchFileMapFunc(abs, "")
	if err != nil {
		fs.RepomapError = err
		return
	}
	fs.FileMap = fm
}

// Counts breaks down Files by state for header rendering.
type Counts struct {
	Staged    int
	Unstaged  int
	Both      int
	Untracked int
	Conflict  int
	Adds      int
	Dels      int
}

func (r *Result) Counts() Counts {
	var c Counts
	for _, f := range r.Files {
		switch f.State {
		case StateStaged:
			c.Staged++
		case StateUnstaged:
			c.Unstaged++
		case StateBoth:
			c.Both++
		case StateUntracked:
			c.Untracked++
		case StateConflict:
			c.Conflict++
		}
		c.Adds += f.Adds
		c.Dels += f.Dels
	}
	return c
}
