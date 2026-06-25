package status

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	clickyai "github.com/flanksource/clicky/ai"
	"github.com/flanksource/gavel/internal/prompting"
	"github.com/flanksource/gavel/utils"
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
	Path           string
	PreviousPath   string
	State          FileState
	StagedKind     ChangeKind
	WorkKind       ChangeKind
	Adds           int
	Dels           int
	AISummary      string
	AIError        string
	AIStatus       AISummaryStatus
	FileMap        *repomap.FileMap
	RepomapError   error
	TestStatus     TestStatus
	LintStatus     LintStatus
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
}

type Options struct {
	NoRepomap bool
	// FolderFilter, when non-empty, restricts the result to files at or
	// below this slash-separated path relative to workDir. The filter is
	// applied before line-count, repomap, and snapshot enrichment so we
	// don't waste work on files that will be dropped.
	FolderFilter string
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
	files = filterByFolder(files, opts.FolderFilter)
	files = filterGitIgnored(files, workDir)

	if err := enrichWithLineCounts(workDir, files); err != nil {
		return nil, err
	}

	enrichWithModTime(workDir, files)
	enrichWithConflictMarkers(workDir, files)

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
		if fs.State == StateConflict {
			fs.ConflictReason = ConflictReasonUnmerged
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

// filterByFolder keeps only files at or under prefix (slash-separated, relative
// to workDir). An empty prefix is a no-op. Prefix "." is treated as empty so
// callers can pass cwd-relative paths through filepath.Rel without a special case.
func filterByFolder(files []FileStatus, prefix string) []FileStatus {
	prefix = strings.Trim(filepath.ToSlash(prefix), "/")
	if prefix == "" || prefix == "." {
		return files
	}
	out := files[:0]
	for _, f := range files {
		if f.Path == prefix || strings.HasPrefix(f.Path, prefix+"/") {
			out = append(out, f)
		}
	}
	return out
}

// filterGitIgnored drops files matching the repo's .gitignore, so force-tracked
// but ignored bundles (e.g. testrunner/ui/dist/testui.js) don't surface as
// modifications. Staged files are kept regardless: what status shows then
// matches what `gavel commit` will commit, and a manually `git add`-ed ignored
// file stays visible. A !-negation in .gitignore re-includes a path.
func filterGitIgnored(files []FileStatus, workDir string) []FileStatus {
	paths := make([]string, len(files))
	for i, f := range files {
		paths[i] = filepath.Join(workDir, f.Path)
	}
	_, ignored := utils.PartitionGitIgnored(paths, workDir)
	if len(ignored) == 0 {
		return files
	}
	ignoredSet := make(map[string]struct{}, len(ignored))
	for _, p := range ignored {
		ignoredSet[p] = struct{}{}
	}
	out := files[:0]
	for i, f := range files {
		_, isIgnored := ignoredSet[paths[i]]
		if !isIgnored || f.State == StateStaged || f.State == StateBoth || f.State == StateConflict {
			out = append(out, f)
		}
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

// enrichWithModTime stats each file's working-tree path and records the
// resulting mtime on FileStatus.ModifiedAt. Best-effort: stat failures (file
// is deleted, broken symlink, permission error) leave ModifiedAt zero.
func enrichWithModTime(workDir string, files []FileStatus) {
	for i := range files {
		f := &files[i]
		if f.WorkKind == KindDeleted || f.StagedKind == KindDeleted {
			continue
		}
		info, err := os.Stat(filepath.Join(workDir, f.Path))
		if err != nil {
			continue
		}
		f.ModifiedAt = info.ModTime()
	}
}

// HumanAge returns a Starship-style relative age ("3m", "5h", "2d") for the
// given duration. Returns "" for zero or negative durations so callers can
// skip rendering when ModifiedAt is unavailable.
func HumanAge(d time.Duration) string {
	if d <= 0 {
		return ""
	}
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	}
}

// Conflict-marker scanning constants.
const (
	conflictMaxFileBytes  = 1 << 20 // 1 MiB; bigger files are skipped (binary or generated)
	conflictBinarySniffer = 8192    // bytes inspected for NUL byte to skip binaries
	conflictMaxLineBytes  = 1 << 20 // 1 MiB single-line cap for the scanner
)

// enrichWithConflictMarkers promotes any file whose working-tree content
// contains unresolved git conflict markers to StateConflict, even when git's
// porcelain output no longer flags the file as unmerged. This catches the
// "git add of a partially-resolved file" case where the staged content still
// has <<<<<<< / ======= / >>>>>>> lines.
//
// Best-effort: I/O errors, binary files, and oversized files are silently
// skipped so a single unreadable file never breaks the whole status command.
func enrichWithConflictMarkers(workDir string, files []FileStatus) {
	for i := range files {
		f := &files[i]
		if f.State == StateConflict {
			continue
		}
		if f.WorkKind == KindDeleted || f.StagedKind == KindDeleted {
			continue
		}
		if fileHasConflictMarkers(filepath.Join(workDir, f.Path)) {
			f.State = StateConflict
			f.ConflictReason = ConflictReasonMarker
		}
	}
}

// fileHasConflictMarkers returns true iff the file at absPath contains all
// three of the standard merge-conflict markers — a line starting with
// "<<<<<<< ", a line that is exactly "=======", and a line starting with
// ">>>>>>> ". Requiring all three together avoids false positives on
// Markdown rule lines or code that happens to contain `=======` alone.
func fileHasConflictMarkers(absPath string) bool {
	info, err := os.Stat(absPath)
	if err != nil || !info.Mode().IsRegular() || info.Size() == 0 || info.Size() > conflictMaxFileBytes {
		return false
	}

	file, err := os.Open(absPath)
	if err != nil {
		return false
	}
	defer file.Close()

	if isLikelyBinary(file) {
		return false
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return false
	}

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 64*1024), conflictMaxLineBytes)

	var hasStart, hasMid, hasEnd bool
	for scanner.Scan() {
		line := scanner.Bytes()
		switch {
		case !hasStart && bytes.HasPrefix(line, []byte("<<<<<<< ")):
			hasStart = true
		case !hasMid && (bytes.Equal(line, []byte("=======")) || bytes.HasPrefix(line, []byte("======= "))):
			hasMid = true
		case !hasEnd && bytes.HasPrefix(line, []byte(">>>>>>> ")):
			hasEnd = true
		}
		if hasStart && hasMid && hasEnd {
			return true
		}
	}
	return false
}

// isLikelyBinary samples the first few KB of the reader and returns true if
// it contains a NUL byte, matching git's own binary-detection heuristic. The
// reader is left at an arbitrary offset; callers should Seek back if needed.
func isLikelyBinary(r io.Reader) bool {
	buf := make([]byte, conflictBinarySniffer)
	n, err := io.ReadFull(r, buf)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return true
	}
	return bytes.IndexByte(buf[:n], 0) >= 0
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
