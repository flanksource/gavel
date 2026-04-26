package commit

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/verify"
	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
)

const (
	maxExecutableBytes int64 = 10 * 1024   // executable files > 10 KB
	maxFileBytes       int64 = 1024 * 1024 // any file > 1 MB
)

const (
	gitModeExecutable = "100755"
	gitModeSymlink    = "120000"
	gitModeSubmodule  = "160000"
)

var ErrFileSizeCancelled = errors.New("commit cancelled: staged file exceeds size limit")

type FileSizeReason int

const (
	ReasonOversizeFile FileSizeReason = iota
	ReasonOversizeExecutable
)

func (r FileSizeReason) String() string {
	switch r {
	case ReasonOversizeFile:
		return "file"
	case ReasonOversizeExecutable:
		return "executable"
	default:
		return "unknown"
	}
}

type FileSizeViolation struct {
	File         string
	Size         int64
	IsExecutable bool
	Reason       FileSizeReason
	Limit        int64
	// folderSiblings is populated by the prompt loop so the "folder" option can
	// report how many files it would unstage. Zero outside the prompt loop.
	folderSiblings int
}

type FileSizeDecision int

const (
	FileSizeDecisionCancel FileSizeDecision = iota
	FileSizeDecisionGitIgnoreFile
	FileSizeDecisionGitIgnoreFolder
	FileSizeDecisionAllow
	FileSizeDecisionAllowFolder
)

type FileSizeDecider func(ctx context.Context, v FileSizeViolation) (FileSizeDecision, error)

type FileSizeParams struct {
	WorkDir     string
	GitRoot     string
	StagedFiles []string
	Changes     []stagedChange
	Config      verify.CommitConfig
	Decider     FileSizeDecider
	SaveDir     string
	Mode        string
}

var interactiveFileSizeDecider = runPromptFileSizeDecider

// stagedFileStat captures the index-side metadata for a single staged path.
// The evaluator takes a slice of these so unit tests do not need a git repo.
type stagedFileStat struct {
	Path         string
	Size         int64
	IsExecutable bool
	IsSymlink    bool
	IsSubmodule  bool
}

// EvaluateFileSizeViolations returns the subset of staged stats that breach a
// size limit, skipping any path matched by allow. head is the same-path view
// from HEAD; when a path already had an equal-or-worse violation in HEAD with
// the same exec bit, we suppress it so unrelated edits don't re-prompt on
// pre-existing oversized blobs. head may be nil for a fresh repo.
func EvaluateFileSizeViolations(staged, head []stagedFileStat, allow []string) ([]FileSizeViolation, error) {
	if len(staged) == 0 {
		return nil, nil
	}

	allowMatchers, err := parsePatterns(allow, "commit.allow")
	if err != nil {
		return nil, err
	}
	allowMatcher := gitignore.NewMatcher(allowMatchers)

	headByPath := make(map[string]stagedFileStat, len(head))
	for _, h := range head {
		headByPath[h.Path] = h
	}

	var out []FileSizeViolation
	for _, s := range staged {
		if s.IsSymlink || s.IsSubmodule {
			continue
		}
		if allowMatcher.Match(splitGitPath(s.Path), false) {
			continue
		}
		reason, limit, ok := classifyFileSize(s)
		if !ok {
			continue
		}
		if h, had := headByPath[s.Path]; had && suppressesViolation(reason, s, h) {
			continue
		}
		out = append(out, FileSizeViolation{
			File:         s.Path,
			Size:         s.Size,
			IsExecutable: s.IsExecutable,
			Reason:       reason,
			Limit:        limit,
		})
	}
	return out, nil
}

func classifyFileSize(s stagedFileStat) (FileSizeReason, int64, bool) {
	if s.Size > maxFileBytes {
		return ReasonOversizeFile, maxFileBytes, true
	}
	if s.IsExecutable && s.Size > maxExecutableBytes {
		return ReasonOversizeExecutable, maxExecutableBytes, true
	}
	return 0, 0, false
}

// suppressesViolation reports whether the HEAD state already contained an
// equivalent or worse violation at the same path, so this staged change did
// not introduce or grow the problem.
func suppressesViolation(reason FileSizeReason, staged, head stagedFileStat) bool {
	if head.IsSymlink || head.IsSubmodule {
		return false
	}
	switch reason {
	case ReasonOversizeFile:
		return head.Size >= staged.Size
	case ReasonOversizeExecutable:
		return head.IsExecutable && head.Size >= staged.Size
	default:
		return false
	}
}

// RunFileSizeCheck evaluates staged files for size/executable violations and
// prompts the user for each new or grown violation. Follows the same idiom as
// RunGitIgnoreCheck: "prompt" (default), "fail", or "skip"; non-TTY prompts
// without an injected Decider escalate to "fail".
func RunFileSizeCheck(ctx context.Context, p FileSizeParams) (CheckOutcome, error) {
	mode, err := normalizeCheckMode(p.Mode, "--precommit")
	if err != nil {
		return CheckOutcome{}, err
	}
	if mode == CheckModeSkip {
		return CheckOutcome{}, nil
	}

	candidates := filterDeletedPaths(p.StagedFiles, p.Changes)
	if len(candidates) == 0 {
		return CheckOutcome{}, nil
	}

	stagedStats, err := readStagedStats(p.WorkDir, candidates)
	if err != nil {
		return CheckOutcome{}, err
	}
	headStats, err := readHeadStats(p.WorkDir, candidates)
	if err != nil {
		return CheckOutcome{}, err
	}
	violations, err := EvaluateFileSizeViolations(stagedStats, headStats, p.Config.Allow)
	if err != nil {
		return CheckOutcome{}, err
	}
	if len(violations) == 0 {
		return CheckOutcome{}, nil
	}

	if mode == CheckModePrompt && p.Decider == nil && !stdinIsTerminal() {
		logger.Warnf("file-size check: stdin is not a terminal; escalating to --precommit=fail")
		mode = CheckModeFail
	}

	switch mode {
	case CheckModeFail:
		return CheckOutcome{}, formatFileSizeError(violations)
	case CheckModePrompt:
	default:
		return CheckOutcome{}, fmt.Errorf("unknown --precommit mode: %q", mode)
	}

	decider := p.Decider
	if decider == nil {
		decider = interactiveFileSizeDecider
	}

	plan := fileSizePlan{}
	decided := make(map[string]struct{}, len(violations))
	for i, v := range violations {
		if _, ok := decided[v.File]; ok {
			continue
		}
		v.withFolderSiblings(violations[i+1:], decided)
		d, err := decider(ctx, v)
		if err != nil {
			return CheckOutcome{}, fmt.Errorf("file-size prompt: %w", err)
		}
		switch d {
		case FileSizeDecisionCancel:
			return CheckOutcome{Cancelled: true}, nil
		case FileSizeDecisionGitIgnoreFile:
			plan.gitIgnoreEntries = appendUnique(plan.gitIgnoreEntries, v.File)
			plan.unstage = appendUnique(plan.unstage, v.File)
			decided[v.File] = struct{}{}
		case FileSizeDecisionGitIgnoreFolder:
			entry := gitIgnoreFolderEntry(v.File)
			if entry == "" {
				entry = v.File
			}
			plan.gitIgnoreEntries = appendUnique(plan.gitIgnoreEntries, entry)
			plan.unstage = appendUnique(plan.unstage, v.File)
			decided[v.File] = struct{}{}
			for _, o := range violations[i+1:] {
				if entry != v.File && strings.HasPrefix(filepath.ToSlash(o.File)+"/", entry) {
					plan.unstage = appendUnique(plan.unstage, o.File)
					decided[o.File] = struct{}{}
				}
			}
		case FileSizeDecisionAllow:
			plan.allowEntries = appendUnique(plan.allowEntries, v.File)
			plan.allowGitIgnoreLines = appendUnique(plan.allowGitIgnoreLines, "!"+v.File)
			plan.allowed = appendUnique(plan.allowed, v.File)
			decided[v.File] = struct{}{}
		case FileSizeDecisionAllowFolder:
			folder := gitIgnoreFolderEntry(v.File)
			if folder == "" {
				plan.allowEntries = appendUnique(plan.allowEntries, v.File)
				plan.allowGitIgnoreLines = appendUnique(plan.allowGitIgnoreLines, "!"+v.File)
				plan.allowed = appendUnique(plan.allowed, v.File)
				decided[v.File] = struct{}{}
				break
			}
			plan.allowEntries = appendUnique(plan.allowEntries, folder+"**")
			plan.allowGitIgnoreLines = appendUnique(plan.allowGitIgnoreLines, "!"+folder)
			plan.allowed = appendUnique(plan.allowed, v.File)
			decided[v.File] = struct{}{}
			for _, o := range violations[i+1:] {
				if strings.HasPrefix(filepath.ToSlash(o.File)+"/", folder) {
					plan.allowed = appendUnique(plan.allowed, o.File)
					decided[o.File] = struct{}{}
				}
			}
		default:
			return CheckOutcome{}, fmt.Errorf("unknown file-size decision %v for %q", d, v.File)
		}
	}

	return applyFileSizePlan(p, plan)
}

type fileSizePlan struct {
	gitIgnoreEntries    []string
	allowEntries        []string
	allowGitIgnoreLines []string // negation lines (`!path`, `!folder/`) added to .gitignore as overrides
	unstage             []string
	allowed             []string
}

// withFolderSiblings decorates v with sibling metadata used by the prompt so
// the "folder" option can report how many files it would unstage. Siblings
// already covered by a prior decision are skipped.
func (v *FileSizeViolation) withFolderSiblings(rest []FileSizeViolation, decided map[string]struct{}) {
	dir := gitIgnoreFolderEntry(v.File)
	if dir == "" {
		return
	}
	n := 0
	for _, o := range rest {
		if _, ok := decided[o.File]; ok {
			continue
		}
		if strings.HasPrefix(filepath.ToSlash(o.File)+"/", dir) {
			n++
		}
	}
	v.folderSiblings = n
}

func applyFileSizePlan(p FileSizeParams, plan fileSizePlan) (CheckOutcome, error) {
	var outcome CheckOutcome

	gitRoot := p.GitRoot
	if gitRoot == "" {
		gitRoot = p.WorkDir
	}

	if len(plan.unstage) > 0 {
		appended, err := appendGitIgnore(gitRoot, plan.gitIgnoreEntries)
		if err != nil {
			return CheckOutcome{}, fmt.Errorf("append .gitignore: %w", err)
		}
		if err := resetFiles(p.WorkDir, plan.unstage); err != nil {
			return CheckOutcome{}, fmt.Errorf("unstage: %w", err)
		}
		outcome.Unstaged = plan.unstage
		outcome.GitIgnored = appended
	}

	if len(plan.allowed) > 0 {
		saveDir := p.SaveDir
		if saveDir == "" {
			saveDir = gitRoot
		}
		added, err := appendAllow(saveDir, plan.allowEntries)
		if err != nil {
			return CheckOutcome{}, fmt.Errorf("update .gavel.yaml: %w", err)
		}
		outcome.Allowed = added

		negated, err := appendGitIgnoreAllow(gitRoot, plan.allowGitIgnoreLines)
		if err != nil {
			return CheckOutcome{}, fmt.Errorf("append .gitignore allow override: %w", err)
		}
		outcome.GitIgnored = appendUniqueAll(outcome.GitIgnored, negated)
	}

	return outcome, nil
}

func formatFileSizeError(violations []FileSizeViolation) error {
	lines := make([]string, 0, len(violations)+1)
	lines = append(lines, "staged files exceed commit size limits:")
	for _, v := range violations {
		lines = append(lines, "  - "+describeViolation(v))
	}
	return errors.New(strings.Join(lines, "\n"))
}

func describeViolation(v FileSizeViolation) string {
	switch v.Reason {
	case ReasonOversizeExecutable:
		return fmt.Sprintf("%s is executable and %s (exceeds %s executable limit)",
			v.File, humanBytes(v.Size), humanBytes(v.Limit))
	default:
		return fmt.Sprintf("%s is %s (exceeds %s file limit)",
			v.File, humanBytes(v.Size), humanBytes(v.Limit))
	}
}

func fileSizePromptHeader(v FileSizeViolation) string {
	switch v.Reason {
	case ReasonOversizeExecutable:
		return fmt.Sprintf("Staged %q is executable and %s (exceeds %s executable limit)",
			v.File, humanBytes(v.Size), humanBytes(v.Limit))
	default:
		return fmt.Sprintf("Staged %q is %s (exceeds %s file limit)",
			v.File, humanBytes(v.Size), humanBytes(v.Limit))
	}
}

type fileSizeChoice struct {
	Text     string
	Decision FileSizeDecision
}

func fileSizeChoices(v FileSizeViolation) []fileSizeChoice {
	choices := []fileSizeChoice{
		{
			Text:     fmt.Sprintf("Unstage and add file %q to .gitignore", v.File),
			Decision: FileSizeDecisionGitIgnoreFile,
		},
	}
	if dir := gitIgnoreFolderEntry(v.File); dir != "" {
		text := fmt.Sprintf("Unstage and add folder %q to .gitignore", dir)
		if v.folderSiblings > 0 {
			text = fmt.Sprintf("Unstage and add folder %q to .gitignore (unstages %d files)", dir, v.folderSiblings+1)
		}
		choices = append(choices, fileSizeChoice{Text: text, Decision: FileSizeDecisionGitIgnoreFolder})
	}
	choices = append(choices,
		fileSizeChoice{
			Text:     fmt.Sprintf("Allow this file (%q in .gavel.yaml + !%s in .gitignore)", v.File, v.File),
			Decision: FileSizeDecisionAllow,
		},
	)
	if folder := gitIgnoreFolderEntry(v.File); folder != "" {
		choices = append(choices, fileSizeChoice{
			Text:     fmt.Sprintf("Allow folder %q (and everything under it: !%s in .gitignore + %s** in .gavel.yaml)", folder, folder, folder),
			Decision: FileSizeDecisionAllowFolder,
		})
	}
	choices = append(choices,
		fileSizeChoice{
			Text:     "Cancel commit",
			Decision: FileSizeDecisionCancel,
		},
	)
	return choices
}

func runPromptFileSizeDecider(_ context.Context, v FileSizeViolation) (FileSizeDecision, error) {
	header := fileSizePromptHeader(v)
	choices := fileSizeChoices(v)
	items := make([]string, len(choices))
	for i, c := range choices {
		items[i] = c.Text
	}
	idx, ok := promptSelectIndex(header, items)
	if !ok {
		return FileSizeDecisionCancel, nil
	}
	return choices[idx].Decision, nil
}

// applyFileSizeCheck runs the check and, if any file was unstaged, re-reads
// the staged source so the caller sees the updated file list.
func applyFileSizeCheck(ctx context.Context, opts Options, source stagedSource) (stagedSource, error) {
	if !shouldRunPrecommitChecks(opts.PrecommitMode) {
		return source, nil
	}

	outcome, err := RunFileSizeCheck(ctx, FileSizeParams{
		WorkDir:     opts.WorkDir,
		GitRoot:     opts.WorkDir,
		StagedFiles: source.Files,
		Changes:     source.Changes,
		Config:      opts.Config,
		SaveDir:     opts.WorkDir,
		Mode:        opts.PrecommitMode,
	})
	if err != nil {
		return source, err
	}
	if outcome.Cancelled {
		return source, ErrFileSizeCancelled
	}
	if len(outcome.Unstaged) == 0 {
		return source, nil
	}
	refreshed, err := readStagedSource(opts.WorkDir)
	if err != nil {
		return source, fmt.Errorf("re-read staged source after file-size unstage: %w", err)
	}
	return refreshed, nil
}

func filterDeletedPaths(files []string, changes []stagedChange) []string {
	if len(changes) == 0 {
		return files
	}
	deleted := make(map[string]struct{})
	for _, c := range changes {
		if c.Status == "deleted" {
			deleted[c.Path] = struct{}{}
		}
	}
	if len(deleted) == 0 {
		return files
	}
	out := make([]string, 0, len(files))
	for _, f := range files {
		if _, skip := deleted[f]; skip {
			continue
		}
		out = append(out, f)
	}
	return out
}

// readStagedStats returns index-side stats for each path in files. Reads mode
// + SHA via one `git ls-files -s -z`, then sizes via one `git cat-file
// --batch-check` piped through stdin. Blob bytes are never read.
func readStagedStats(workDir string, files []string) ([]stagedFileStat, error) {
	if len(files) == 0 {
		return nil, nil
	}
	entries, err := lsFilesStage(workDir, files)
	if err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		return nil, nil
	}
	return annotateWithSizes(workDir, entries)
}

// readHeadStats returns HEAD-side stats for each path. Missing entries (either
// no HEAD yet or path not in HEAD) are simply absent from the result. Caller
// treats absence as "newly introduced".
func readHeadStats(workDir string, files []string) ([]stagedFileStat, error) {
	if len(files) == 0 {
		return nil, nil
	}
	out := make([]stagedFileStat, 0, len(files))
	for _, path := range files {
		entry, ok, err := lsTreeHead(workDir, path)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}
		stat := entry.stagedFileStat
		if stat.IsSymlink || stat.IsSubmodule {
			out = append(out, stat)
			continue
		}
		size, err := catFileSize(workDir, entry.sha)
		if err != nil {
			return nil, err
		}
		stat.Size = size
		out = append(out, stat)
	}
	return out, nil
}

type lsEntry struct {
	stagedFileStat
	sha string
}

// lsFilesStage parses `git ls-files -s -z -- <files...>` into lsEntry records.
// Each record line is: `<mode> <sha> <stage>\t<path>` (NUL-delimited).
func lsFilesStage(workDir string, files []string) ([]lsEntry, error) {
	args := append([]string{"ls-files", "-s", "-z", "--"}, files...)
	cmd := exec.Command("git", args...)
	cmd.Dir = workDir
	var stderr strings.Builder
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git ls-files -s: %w: %s", err, strings.TrimSpace(stderr.String()))
	}

	var entries []lsEntry
	for _, rec := range bytes.Split(out, []byte{0}) {
		if len(rec) == 0 {
			continue
		}
		entry, err := parseLsFilesRecord(string(rec))
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

// parseLsFilesRecord parses one `<mode> <sha> <stage>\t<path>` record.
func parseLsFilesRecord(rec string) (lsEntry, error) {
	tab := strings.IndexByte(rec, '\t')
	if tab < 0 {
		return lsEntry{}, fmt.Errorf("ls-files: missing tab separator in %q", rec)
	}
	head, path := rec[:tab], rec[tab+1:]
	parts := strings.Fields(head)
	if len(parts) != 3 {
		return lsEntry{}, fmt.Errorf("ls-files: unexpected format %q", rec)
	}
	mode, sha := parts[0], parts[1]
	return lsEntry{
		stagedFileStat: stagedFileStat{
			Path:         path,
			IsExecutable: mode == gitModeExecutable,
			IsSymlink:    mode == gitModeSymlink,
			IsSubmodule:  mode == gitModeSubmodule,
		},
		sha: sha,
	}, nil
}

// annotateWithSizes fills in Size for non-symlink, non-submodule entries by
// piping SHAs to a single `git cat-file --batch-check=%(objectsize)` call.
func annotateWithSizes(workDir string, entries []lsEntry) ([]stagedFileStat, error) {
	batchIdx := make([]int, 0, len(entries))
	var stdinBuf bytes.Buffer
	for i, e := range entries {
		if e.IsSymlink || e.IsSubmodule {
			continue
		}
		batchIdx = append(batchIdx, i)
		stdinBuf.WriteString(e.sha)
		stdinBuf.WriteByte('\n')
	}

	out := make([]stagedFileStat, len(entries))
	for i, e := range entries {
		out[i] = e.stagedFileStat
	}

	if len(batchIdx) == 0 {
		return out, nil
	}

	cmd := exec.Command("git", "cat-file", "--batch-check=%(objectsize)")
	cmd.Dir = workDir
	cmd.Stdin = &stdinBuf
	var stderr strings.Builder
	cmd.Stderr = &stderr
	stdout, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git cat-file --batch-check: %w: %s", err, strings.TrimSpace(stderr.String()))
	}

	scanner := bufio.NewScanner(bytes.NewReader(stdout))
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	for _, idx := range batchIdx {
		if !scanner.Scan() {
			return nil, fmt.Errorf("git cat-file --batch-check: truncated output after %d entries", idx)
		}
		line := strings.TrimSpace(scanner.Text())
		size, perr := parseBatchCheckSize(line)
		if perr != nil {
			return nil, fmt.Errorf("git cat-file --batch-check: %w (line %q)", perr, line)
		}
		out[idx].Size = size
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("git cat-file --batch-check: %w", err)
	}
	return out, nil
}

func parseBatchCheckSize(line string) (int64, error) {
	if line == "" {
		return 0, errors.New("empty line")
	}
	var n int64
	for _, ch := range line {
		if ch < '0' || ch > '9' {
			return 0, fmt.Errorf("non-numeric size %q", line)
		}
		n = n*10 + int64(ch-'0')
	}
	return n, nil
}

// lsTreeHead returns the HEAD entry for path, if any. Missing HEAD or missing
// entry yield (_, false, nil).
func lsTreeHead(workDir, path string) (lsEntry, bool, error) {
	cmd := exec.Command("git", "ls-tree", "-z", "HEAD", "--", path)
	cmd.Dir = workDir
	var stderr strings.Builder
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		msg := strings.TrimSpace(stderr.String())
		if missingHeadBlob(msg) {
			return lsEntry{}, false, nil
		}
		return lsEntry{}, false, fmt.Errorf("git ls-tree HEAD: %w: %s", err, msg)
	}
	rec := strings.TrimRight(string(out), "\x00")
	if rec == "" {
		return lsEntry{}, false, nil
	}
	entry, err := parseLsTreeRecord(rec)
	if err != nil {
		return lsEntry{}, false, err
	}
	return entry, true, nil
}

// parseLsTreeRecord parses one ls-tree record: `<mode> <type> <sha>\t<path>`.
func parseLsTreeRecord(rec string) (lsEntry, error) {
	tab := strings.IndexByte(rec, '\t')
	if tab < 0 {
		return lsEntry{}, fmt.Errorf("ls-tree: missing tab separator in %q", rec)
	}
	head, path := rec[:tab], rec[tab+1:]
	parts := strings.Fields(head)
	if len(parts) != 3 {
		return lsEntry{}, fmt.Errorf("ls-tree: unexpected format %q", rec)
	}
	mode, _, sha := parts[0], parts[1], parts[2]
	return lsEntry{
		stagedFileStat: stagedFileStat{
			Path:         path,
			IsExecutable: mode == gitModeExecutable,
			IsSymlink:    mode == gitModeSymlink,
			IsSubmodule:  mode == gitModeSubmodule,
		},
		sha: sha,
	}, nil
}

func catFileSize(workDir, sha string) (int64, error) {
	cmd := exec.Command("git", "cat-file", "-s", sha)
	cmd.Dir = workDir
	var stderr strings.Builder
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("git cat-file -s %s: %w: %s", sha, err, strings.TrimSpace(stderr.String()))
	}
	return parseBatchCheckSize(strings.TrimSpace(string(out)))
}

// humanBytes renders n using binary units with one decimal place. Integers are
// printed without a decimal ("1 KB" rather than "1.0 KB").
func humanBytes(n int64) string {
	const (
		kib = 1024
		mib = kib * 1024
		gib = mib * 1024
	)
	switch {
	case n >= gib:
		return formatUnits(n, gib, "GB")
	case n >= mib:
		return formatUnits(n, mib, "MB")
	case n >= kib:
		return formatUnits(n, kib, "KB")
	default:
		return fmt.Sprintf("%d B", n)
	}
}

func formatUnits(n, unit int64, suffix string) string {
	whole := n / unit
	frac := (n % unit) * 10 / unit
	if frac == 0 {
		return fmt.Sprintf("%d %s", whole, suffix)
	}
	return fmt.Sprintf("%d.%d %s", whole, frac, suffix)
}
