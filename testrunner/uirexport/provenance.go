package uirexport

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/flanksource/commons/logger"
)

// Provenance is the code a run's tests ran on: the module's HEAD commit, whether its worktree had
// changes HEAD does not hold, and the uir snapshot indexed from it. It is captured before the tests
// start, so files they write do not change it.
type Provenance struct {
	GitCommit  string
	Dirty      bool
	SnapshotID string
}

// CaptureProvenance reads the HEAD commit and worktree state of the Go module holding workDir, then
// runs `uir reindex <module> --include-tests` for the snapshot the tests will run on. Without uir, or
// when the reindex fails, the snapshot is left empty with a warning; uir refuses a dirty run without
// one. output is the export file, which is gavel's own output and never makes the worktree dirty.
func CaptureProvenance(workDir, output string) (Provenance, error) {
	m, err := findModule(workDir)
	if err != nil {
		return Provenance{}, err
	}
	var provenance Provenance
	if provenance.GitCommit, err = git(m.root, "rev-parse", "--verify", "HEAD"); err != nil {
		return Provenance{}, fmt.Errorf("the HEAD commit of %s: %w", m.root, err)
	}
	if provenance.Dirty, err = dirty(m, output); err != nil {
		return Provenance{}, err
	}
	provenance.SnapshotID, err = reindex(m.root)
	switch {
	case err != nil && provenance.Dirty:
		logger.Warnf("uir export: %s has uncommitted changes and %v; the export has no snapshot_id, so `uir tests import` will refuse it", m.root, err)
	case err != nil:
		logger.Warnf("uir export: %v; exporting the run against commit %s without a snapshot_id", err, provenance.GitCommit)
	}
	logger.V(1).Infof("uir export: commit %s, dirty %t, snapshot %q", provenance.GitCommit, provenance.Dirty, provenance.SnapshotID)
	return provenance, nil
}

// dirty reports whether the module's worktree differs from HEAD in a way uir would index: a tracked
// file changed, staged or not, or an untracked Go file uir would index (one that is not ignored and
// lies outside directories starting with ".", vendor, testdata and nested modules). Gavel's own output,
// any path under a .gavel or .tmp directory and the export file, does not count; neither do nested
// modules, which the export leaves out.
func dirty(m *module, output string) (bool, error) {
	top, err := git(m.root, "rev-parse", "--show-toplevel")
	if err != nil {
		return false, fmt.Errorf("the Git checkout of %s: %w", m.root, err)
	}
	exportFile := ""
	if output != "" {
		if abs, err := filepath.Abs(output); err == nil {
			exportFile = abs
		}
	}
	// Porcelain paths are relative to the checkout's top whatever the working directory.
	status, err := git(m.root, "--no-optional-locks", "status", "--porcelain=v1", "-z", "--untracked-files=all", "--", ".")
	if err != nil {
		return false, fmt.Errorf("the Git worktree state of %s: %w", m.root, err)
	}
	entries := strings.Split(status, "\x00")
	for i := 0; i < len(entries); i++ {
		entry := entries[i]
		if len(entry) < 4 {
			continue
		}
		code, path := entry[:2], filepath.Join(top, filepath.FromSlash(entry[3:]))
		if code[0] == 'R' || code[0] == 'C' {
			// A rename or copy is followed by the path it came from, which changed too.
			i++
		}
		rel := m.relative(m.root, path)
		if rel == "" || path == exportFile || slices.ContainsFunc(strings.Split(rel, "/"), func(part string) bool { return part == ".gavel" || part == ".tmp" }) {
			continue
		}
		if code == "??" && !indexedSource(rel) {
			continue
		}
		logger.V(1).Infof("uir export: %s is dirty: %s %s", m.root, code, rel)
		return true, nil
	}
	return false, nil
}

// indexedSource reports whether uir indexes the module-relative file rel with test files included.
func indexedSource(rel string) bool {
	if filepath.Ext(rel) != ".go" {
		return false
	}
	dirs := strings.Split(rel, "/")
	return !slices.ContainsFunc(dirs[:len(dirs)-1], func(dir string) bool {
		return strings.HasPrefix(dir, ".") || strings.HasPrefix(dir, "_") || dir == "vendor" || dir == "testdata"
	})
}

// reindex runs `uir reindex <root> --include-tests` and returns the snapshot it reports for root.
func reindex(root string) (string, error) {
	if _, err := exec.LookPath("uir"); err != nil {
		return "", errors.New("uir is not installed")
	}
	command := exec.Command("uir", "reindex", root, "--include-tests", "--json", "--no-progress")
	command.Dir = root
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	if err := command.Run(); err != nil {
		// uir logs before it fails; its last line is the error.
		lines := strings.Split(strings.TrimSpace(stderr.String()), "\n")
		return "", fmt.Errorf("`uir reindex %s --include-tests` failed (a module uir has not added needs `uir add %s --include-tests` first): %w: %s", root, root, err, lines[len(lines)-1])
	}
	var run struct {
		Results []struct {
			Location   string `json:"location"`
			SnapshotID string `json:"snapshot_id"`
			Error      string `json:"error"`
		} `json:"results"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &run); err != nil {
		return "", fmt.Errorf("decode `uir reindex` output: %w", err)
	}
	for _, result := range run.Results {
		if !samePath(result.Location, root) {
			continue
		}
		if result.Error != "" {
			return "", fmt.Errorf("`uir reindex` failed for %s: %s", root, result.Error)
		}
		if result.SnapshotID != "" {
			return result.SnapshotID, nil
		}
	}
	return "", fmt.Errorf("`uir reindex` reported no snapshot for %s", root)
}

// samePath reports whether a and b name the same directory once symlinks are resolved.
func samePath(a, b string) bool {
	if filepath.Clean(a) == filepath.Clean(b) {
		return true
	}
	resolvedA, errA := filepath.EvalSymlinks(a)
	resolvedB, errB := filepath.EvalSymlinks(b)
	return errA == nil && errB == nil && resolvedA == resolvedB
}

// git runs git in dir and returns its trimmed output.
func git(dir string, args ...string) (string, error) {
	command := exec.Command("git", args...)
	command.Dir = dir
	output, err := command.Output()
	if err != nil {
		var exit *exec.ExitError
		if errors.As(err, &exit) {
			err = fmt.Errorf("git %s: %w: %s", args[0], err, bytes.TrimSpace(exit.Stderr))
		}
		return "", err
	}
	return strings.TrimRight(string(output), "\n"), nil
}
