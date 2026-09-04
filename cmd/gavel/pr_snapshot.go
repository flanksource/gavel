package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/baseline"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/prwatch"
	"github.com/flanksource/gavel/snapshots"
	testui "github.com/flanksource/gavel/testrunner/ui"
)

// parsePRRef parses the PR references every gavel command accepts: a bare
// number, "#12", "owner/repo#12", or a PR URL. An empty repo means "resolve it
// from the working directory's git remote".
func parsePRRef(ref string) (repo string, number int, err error) {
	trimmed := strings.TrimSpace(ref)
	if trimmed == "" {
		return "", 0, fmt.Errorf("empty PR reference")
	}

	if strings.HasPrefix(trimmed, "http://") || strings.HasPrefix(trimmed, "https://") || strings.HasPrefix(trimmed, "github.com/") {
		repo, number, err = github.ParsePRURL(trimmed)
		if err != nil {
			return "", 0, fmt.Errorf("invalid PR URL %q: %w", ref, err)
		}
		return repo, number, nil
	}

	numberPart := strings.TrimPrefix(trimmed, "#")
	if idx := strings.Index(trimmed, "#"); idx > 0 {
		repo, err = resolveRepoArg(trimmed[:idx])
		if err != nil {
			return "", 0, err
		}
		numberPart = trimmed[idx+1:]
	}

	number, err = strconv.Atoi(numberPart)
	if err != nil || number <= 0 {
		return "", 0, fmt.Errorf("expected a PR number, #number, owner/repo#number, or PR URL, got %q", ref)
	}
	return repo, number, nil
}

// prNarrowTarget selects which half of a downloaded PR snapshot must be
// non-empty for the run to have anything to do.
type prNarrowTarget string

const (
	prNarrowTests prNarrowTarget = "tests"
	prNarrowLint  prNarrowTarget = "lint"
)

// resolvePRFailed turns --pr into the --failed snapshot path the runner and
// the linter already know how to narrow with. It fails loudly rather than
// running everything when the PR has nothing to re-run, and rejects the
// combination with the other narrowing flags instead of silently picking one.
func resolvePRFailed(ref, workDir, failedPath, baselinePath string, target prNarrowTarget) (string, error) {
	if failedPath != "" {
		return "", fmt.Errorf("--pr and --failed both narrow the run; pass only one")
	}
	if baselinePath != "" {
		return "", fmt.Errorf("--pr and --baseline both read a previous run; pass only one")
	}

	path, snap, err := resolvePRSnapshot(ref, workDir)
	if err != nil {
		return "", fmt.Errorf("--pr %s: %w", ref, err)
	}

	switch target {
	case prNarrowTests:
		if len(baseline.ExtractFailedTestPackages(snap.Tests)) == 0 {
			return "", fmt.Errorf("--pr %s: no failing tests in the PR's gavel results (%s)", ref, path)
		}
	case prNarrowLint:
		if names, _ := baseline.ExtractFailedLintTargets(snap.Lint); len(names) == 0 {
			return "", fmt.Errorf("--pr %s: no lint violations in the PR's gavel results (%s)", ref, path)
		}
	}
	return path, nil
}

// resolvePRSnapshot downloads every gavel artifact published on the referenced
// PR, merges them, and caches the result at .gavel/pr-<n>.json so `--pr` can
// hand a plain path to the same `--failed` narrowing a local re-run uses.
func resolvePRSnapshot(ref, workDir string) (path string, snap *testui.Snapshot, err error) {
	repo, number, err := parsePRRef(ref)
	if err != nil {
		return "", nil, err
	}

	if workDir == "" {
		workDir, err = getWorkingDir()
		if err != nil {
			return "", nil, fmt.Errorf("failed to get working directory: %w", err)
		}
	}

	ghOpts := github.Options{Repo: repo}
	if repo == "" {
		ghOpts.WorkDir = workDir
	}

	snap, err = prwatch.DownloadPRSnapshot(ghOpts, number)
	if err != nil {
		return "", nil, err
	}

	dir := filepath.Join(workDir, snapshots.Dir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", nil, fmt.Errorf("create %s: %w", dir, err)
	}
	path = filepath.Join(dir, fmt.Sprintf("pr-%d.json", number))
	if err := snapshots.WriteSnapshot(path, snap); err != nil {
		return "", nil, err
	}
	logger.Infof("--pr %d: downloaded gavel results to %s", number, path)
	return path, snap, nil
}
