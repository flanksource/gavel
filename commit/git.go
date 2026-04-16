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
