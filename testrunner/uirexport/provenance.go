package uirexport

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/utils"
)

// Provenance is the code a run's tests ran on: the module's HEAD commit, whether its worktree had
// changes HEAD does not hold, and for a dirty worktree the Git tree of the whole working tree. It is
// captured before the tests start, so files they write do not change it.
type Provenance struct {
	GitCommit string
	Dirty     bool
	GitTree   string
}

// CaptureProvenance reads the HEAD commit and worktree state of the Go module holding workDir, and
// when it is dirty the tree hash of its repository's working tree (workingTree). output is the
// export file, which is gavel's own output and never makes the worktree dirty or enters the tree.
//
// Gavel's output directories ignore themselves (utils.MkdirOutput); ones an older gavel left without
// their .gitignore get it here first, so they count neither as changes nor as part of the tree.
func CaptureProvenance(workDir, output string) (Provenance, error) {
	m, err := findModule(workDir)
	if err != nil {
		return Provenance{}, err
	}
	top, err := git(m.root, nil, "rev-parse", "--show-toplevel")
	if err != nil {
		return Provenance{}, fmt.Errorf("the Git checkout of %s: %w", m.root, err)
	}
	for _, dir := range []string{top, m.root, workDir} {
		for _, name := range []string{".gavel", ".tmp"} {
			if err := utils.IgnoreOutputDir(filepath.Join(dir, name)); err != nil {
				return Provenance{}, err
			}
		}
	}
	// The module's directory and the export file, relative to the checkout's top as Git reports paths.
	prefix, err := git(m.root, nil, "rev-parse", "--show-prefix")
	if err != nil {
		return Provenance{}, fmt.Errorf("the Git checkout of %s: %w", m.root, err)
	}
	exportFile := ""
	if output != "" {
		exportFile = topRelative(top, output)
	}
	var provenance Provenance
	if provenance.GitCommit, err = git(m.root, nil, "rev-parse", "--verify", "HEAD"); err != nil {
		return Provenance{}, fmt.Errorf("the HEAD commit of %s: %w", m.root, err)
	}
	if provenance.Dirty, err = dirty(m, prefix, exportFile); err != nil {
		return Provenance{}, err
	}
	if provenance.Dirty {
		if provenance.GitTree, err = workingTree(top, exportFile); err != nil {
			return Provenance{}, err
		}
	}
	logger.V(1).Infof("uir export: commit %s, dirty %t, tree %q", provenance.GitCommit, provenance.Dirty, provenance.GitTree)
	return provenance, nil
}

// dirty reports whether the module's worktree differs from HEAD in a way uir would index: a tracked
// file changed, staged or not, or an untracked, non-ignored Go file uir would index (one outside
// directories starting with "." or "_", vendor, testdata and nested modules). The export file does
// not count; neither do nested modules, which the export leaves out. prefix is the module's directory
// and exportFile the export file, both relative to the checkout's top.
func dirty(m *module, prefix, exportFile string) (bool, error) {
	// Porcelain paths are relative to the checkout's top whatever the working directory.
	status, err := git(m.root, nil, "--no-optional-locks", "status", "--porcelain=v1", "-z", "--untracked-files=all", "--", ".")
	if err != nil {
		return false, fmt.Errorf("the Git worktree state of %s: %w", m.root, err)
	}
	entries := strings.Split(status, "\x00")
	for i := 0; i < len(entries); i++ {
		entry := entries[i]
		if len(entry) < 4 {
			continue
		}
		code, path := entry[:2], entry[3:]
		if code[0] == 'R' || code[0] == 'C' {
			// A rename or copy is followed by the path it came from, which changed too.
			i++
		}
		inModule, ok := strings.CutPrefix(path, prefix)
		if !ok || path == exportFile {
			continue
		}
		rel := m.relative(m.root, filepath.Join(m.root, filepath.FromSlash(inModule)))
		if rel == "" || code == "??" && !indexedSource(rel) {
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

// workingTree is the Git tree hash of the checkout at top as its working tree stands: HEAD with every
// tracked change and untracked, non-ignored file added, less exportFile. It is written through a
// temporary index, so the checkout's own index and worktree are untouched; uir computes the same
// tree at import time to check the code has not changed since the run. exportFile is relative to top.
func workingTree(top, exportFile string) (string, error) {
	scratch, err := os.MkdirTemp("", "gavel-uir-index-")
	if err != nil {
		return "", err
	}
	defer func() {
		if err := os.RemoveAll(scratch); err != nil {
			logger.Warnf("remove %s: %v", scratch, err)
		}
	}()
	env := append(os.Environ(), "GIT_INDEX_FILE="+filepath.Join(scratch, "index"))
	add := []string{"add", "-A", "--", "."}
	if exportFile != "" {
		add = append(add, ":(exclude,literal)"+exportFile)
	}
	for _, args := range [][]string{{"read-tree", "HEAD"}, add} {
		if _, err := git(top, env, args...); err != nil {
			return "", fmt.Errorf("the working tree of %s: %w", top, err)
		}
	}
	tree, err := git(top, env, "write-tree")
	if err != nil {
		return "", fmt.Errorf("the working tree of %s: %w", top, err)
	}
	return tree, nil
}

// topRelative is path relative to the checkout's top, with forward slashes, or "" outside it. Git
// reports top with symlinks resolved, so path's directory is resolved too.
func topRelative(top, path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		return ""
	}
	dir := filepath.Dir(abs)
	if resolved, err := filepath.EvalSymlinks(dir); err == nil {
		dir = resolved
	}
	rel, err := filepath.Rel(top, filepath.Join(dir, filepath.Base(abs)))
	if err != nil || !filepath.IsLocal(rel) {
		return ""
	}
	return filepath.ToSlash(rel)
}

// git runs git in dir, with env when it is set, and returns its output less trailing newlines.
func git(dir string, env []string, args ...string) (string, error) {
	command := exec.Command("git", args...)
	command.Dir, command.Env = dir, env
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
