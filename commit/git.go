package commit

import (
	"fmt"
	"os/exec"
	"strings"

	gavelgit "github.com/flanksource/gavel/git"
	"github.com/flanksource/gavel/models"
)

type stagedChange struct {
	Path         string
	PreviousPath string
	Status       string
	Adds         int
	Dels         int
	Patch        string
}

type stagedSource struct {
	Files   []string
	Diff    string
	Changes []stagedChange
	// PendingRestores carries deferred working-tree edits that must be applied
	// AFTER the commit subprocess succeeds — currently only the linked-deps
	// "upgrade" path uses it to re-add `replace` directives that were dropped
	// from the committed snapshot.
	PendingRestores []pendingRestore
}

func (s stagedSource) GitPaths() []string {
	var paths []string
	seen := make(map[string]struct{}, len(s.Changes)*2)
	for _, change := range s.Changes {
		for _, path := range change.GitPaths() {
			if _, ok := seen[path]; ok {
				continue
			}
			seen[path] = struct{}{}
			paths = append(paths, path)
		}
	}
	return paths
}

func (c stagedChange) GitPaths() []string {
	paths := make([]string, 0, 2)
	if c.PreviousPath != "" {
		paths = append(paths, c.PreviousPath)
	}
	if c.Path != "" && c.Path != c.PreviousPath {
		paths = append(paths, c.Path)
	}
	if len(paths) == 0 && c.Path != "" {
		paths = append(paths, c.Path)
	}
	return paths
}

func readStagedSource(workDir string) (stagedSource, error) {
	files, err := stagedFiles(workDir)
	if err != nil {
		return stagedSource{}, fmt.Errorf("list staged files: %w", err)
	}
	if len(files) == 0 {
		return stagedSource{}, nil
	}

	diff, err := stagedDiff(workDir)
	if err != nil {
		return stagedSource{}, fmt.Errorf("read staged diff: %w", err)
	}
	if strings.TrimSpace(diff) == "" {
		return stagedSource{}, fmt.Errorf("staged file list was non-empty but diff is empty: %v", files)
	}

	changes, err := parseStagedChanges(diff)
	if err != nil {
		return stagedSource{}, fmt.Errorf("parse staged diff: %w", err)
	}
	return stagedSource{
		Files:   files,
		Diff:    diff,
		Changes: changes,
	}, nil
}

func stagedDiff(workDir string) (string, error) {
	cmd := exec.Command("git", "diff", "--cached", "--find-renames")
	cmd.Dir = workDir
	var stderr strings.Builder
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git diff --cached: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	return string(out), nil
}

func stagedFiles(workDir string) ([]string, error) {
	cmd := exec.Command("git", "diff", "--cached", "--name-only", "--find-renames")
	cmd.Dir = workDir
	var stderr strings.Builder
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git diff --cached --name-only: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	var files []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			files = append(files, line)
		}
	}
	return files, nil
}

func addFiles(workDir string, files []string) error {
	if len(files) == 0 {
		return nil
	}
	args := append([]string{"add", "-A", "--"}, files...)
	cmd := exec.Command("git", args...)
	cmd.Dir = workDir
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git add -A: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// gitRmCached runs `git rm --cached -- <files>` in workDir, removing the
// listed paths from the index without touching the working tree. Used by the
// interactive picker after the user adds an entry to .gitignore for a file
// that was already tracked, so the new ignore actually takes effect.
//
// `--ignore-unmatch` keeps the call idempotent when a file has already been
// untracked (e.g. by a previous picker iteration).
func gitRmCached(workDir string, files []string) error {
	if len(files) == 0 {
		return nil
	}
	args := append([]string{"rm", "--cached", "--ignore-unmatch", "--"}, files...)
	cmd := exec.Command("git", args...)
	cmd.Dir = workDir
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git rm --cached: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func resetFiles(workDir string, files []string) error {
	if len(files) == 0 {
		return nil
	}
	args := append([]string{"reset", "--"}, files...)
	cmd := exec.Command("git", args...)
	cmd.Dir = workDir
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git reset --: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func commitWithMessage(workDir, msg string) (string, error) {
	cmd := exec.Command("git", "commit", "-m", msg)
	cmd.Dir = workDir
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("git commit: %w: %s", err, strings.TrimSpace(string(out)))
	}
	hashCmd := exec.Command("git", "rev-parse", "HEAD")
	hashCmd.Dir = workDir
	out, err := hashCmd.Output()
	if err != nil {
		return "", fmt.Errorf("git rev-parse HEAD: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// loadAheadCommits returns local commits on HEAD that are not yet on the
// remote tracking branch, oldest-first. Used by --push when nothing is
// staged so we have something to seed PR title/body generation with.
//
// Resolution order for the base ref:
//  1. The branch's upstream (@{upstream}) when configured.
//  2. origin/<branch> when that ref exists locally.
//  3. defaultBase (e.g. "origin/main") when non-empty and the ref exists.
//
// Returns an empty slice when no base can be resolved or HEAD is not ahead.
func loadAheadCommits(workDir, branch, defaultBase string) ([]CommitResult, error) {
	base := resolveAheadBase(workDir, branch, defaultBase)
	if base == "" {
		return nil, nil
	}
	hashes, err := revList(workDir, base+"..HEAD")
	if err != nil {
		return nil, err
	}
	commits := make([]CommitResult, 0, len(hashes))
	for i := len(hashes) - 1; i >= 0; i-- {
		h := hashes[i]
		msg, mErr := commitMessage(workDir, h)
		if mErr != nil {
			return nil, fmt.Errorf("read commit %s: %w", h, mErr)
		}
		files, fErr := commitFiles(workDir, h)
		if fErr != nil {
			return nil, fmt.Errorf("read files for %s: %w", h, fErr)
		}
		commits = append(commits, CommitResult{
			Hash:    h,
			Message: msg,
			Files:   files,
		})
	}
	return commits, nil
}

func resolveAheadBase(workDir, branch, defaultBase string) string {
	if branch != "" {
		if validRef(workDir, branch+"@{upstream}") {
			return branch + "@{upstream}"
		}
		if validRef(workDir, "origin/"+branch) {
			return "origin/" + branch
		}
	}
	if defaultBase != "" && validRef(workDir, defaultBase) {
		return defaultBase
	}
	return ""
}

func validRef(workDir, ref string) bool {
	cmd := exec.Command("git", "rev-parse", "--verify", "--quiet", ref)
	if workDir != "" {
		cmd.Dir = workDir
	}
	return cmd.Run() == nil
}

func revList(workDir, spec string) ([]string, error) {
	cmd := exec.Command("git", "rev-list", spec)
	if workDir != "" {
		cmd.Dir = workDir
	}
	var stderr strings.Builder
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git rev-list %s: %w: %s", spec, err, strings.TrimSpace(stderr.String()))
	}
	var hashes []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			hashes = append(hashes, line)
		}
	}
	return hashes, nil
}

func commitMessage(workDir, hash string) (string, error) {
	cmd := exec.Command("git", "log", "-1", "--pretty=%B", hash)
	if workDir != "" {
		cmd.Dir = workDir
	}
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// lastTouchingCommit returns the most recent commit on base..HEAD that
// touched file. Empty string with nil error means no in-range commit
// modified that file (caller should treat the file as a "leftover").
func lastTouchingCommit(workDir, base, file string) (string, error) {
	spec := base + "..HEAD"
	cmd := exec.Command("git", "log", "-1", "--format=%H", spec, "--", file)
	if workDir != "" {
		cmd.Dir = workDir
	}
	var stderr strings.Builder
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git log %s -- %s: %w: %s", spec, file, err, strings.TrimSpace(stderr.String()))
	}
	return strings.TrimSpace(string(out)), nil
}

// commitFixup runs `git commit --fixup=<targetHash>` against the currently
// staged index and returns the new HEAD hash. The fixup message format
// (`fixup! <subject of target>`) is produced by git itself, so we do not
// build it manually.
func commitFixup(workDir, targetHash string) (string, error) {
	cmd := exec.Command("git", "commit", "--fixup="+targetHash)
	cmd.Dir = workDir
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("git commit --fixup=%s: %w: %s", targetHash, err, strings.TrimSpace(string(out)))
	}
	hashCmd := exec.Command("git", "rev-parse", "HEAD")
	hashCmd.Dir = workDir
	out, err := hashCmd.Output()
	if err != nil {
		return "", fmt.Errorf("git rev-parse HEAD: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func commitFiles(workDir, hash string) ([]string, error) {
	cmd := exec.Command("git", "show", "--name-only", "--pretty=", hash)
	if workDir != "" {
		cmd.Dir = workDir
	}
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var files []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			files = append(files, line)
		}
	}
	return files, nil
}

func parseStagedChanges(diff string) ([]stagedChange, error) {
	summaries, err := gavelgit.ParsePatch(diff)
	if err != nil {
		return nil, err
	}

	summaryByPath := make(map[string]models.CommitChange, len(summaries))
	for _, summary := range summaries {
		summaryByPath[summary.File] = summary
	}

	lines := strings.Split(diff, "\n")
	changes := make([]stagedChange, 0, len(summaries))
	var current *stagedChange
	var patchLines []string

	flush := func() {
		if current == nil {
			return
		}
		current.Patch = strings.Join(patchLines, "\n")
		if summary, ok := summaryByPath[current.Path]; ok {
			current.Adds = summary.Adds
			current.Dels = summary.Dels
			current.Status = changeStatus(summary.Type)
		}
		changes = append(changes, *current)
		current = nil
		patchLines = nil
	}

	for _, line := range lines {
		if strings.HasPrefix(line, "diff --git ") {
			flush()
			oldPath, newPath := parseDiffGitPaths(line)
			current = &stagedChange{
				Path:         newPath,
				PreviousPath: oldPath,
				Status:       "updated",
			}
			patchLines = []string{line}
			continue
		}
		if current == nil {
			continue
		}
		patchLines = append(patchLines, line)
		switch {
		case strings.HasPrefix(line, "new file mode "):
			current.Status = "inserted"
		case strings.HasPrefix(line, "deleted file mode "):
			current.Status = "deleted"
		case strings.HasPrefix(line, "rename from "):
			current.Status = "renamed"
			current.PreviousPath = strings.TrimSpace(strings.TrimPrefix(line, "rename from "))
		case strings.HasPrefix(line, "rename to "):
			current.Status = "renamed"
			current.Path = strings.TrimSpace(strings.TrimPrefix(line, "rename to "))
		}
	}
	flush()

	return changes, nil
}

func parseDiffGitPaths(line string) (string, string) {
	fields := strings.Fields(line)
	if len(fields) < 4 {
		return "", ""
	}
	return trimGitDiffPath(fields[2]), trimGitDiffPath(fields[3])
}

func trimGitDiffPath(path string) string {
	path = strings.Trim(path, `"`)
	path = strings.TrimPrefix(path, "a/")
	path = strings.TrimPrefix(path, "b/")
	return path
}

func changeStatus(changeType models.SourceChangeType) string {
	switch changeType {
	case models.SourceChangeTypeAdded:
		return "inserted"
	case models.SourceChangeTypeDeleted:
		return "deleted"
	case models.SourceChangeTypeRenamed:
		return "renamed"
	default:
		return "updated"
	}
}
