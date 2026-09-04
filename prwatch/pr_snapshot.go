package prwatch

import (
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"strings"

	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/report"
	"github.com/flanksource/gavel/testrunner/parsers"
	testui "github.com/flanksource/gavel/testrunner/ui"
)

// DecodeArtifactSnapshot parses a gavel results artifact into a snapshot.
// Gavel uploads a testui.Snapshot, but the composite action's stub and older
// runs upload a report.ResultFile — which additionally accepts a bare JSON
// array of tests. Both are normalized here so every consumer of an artifact
// (the PR dashboard, `--pr` narrowing) reads the same shape.
func DecodeArtifactSnapshot(jsonBytes []byte) (testui.Snapshot, error) {
	var snap testui.Snapshot
	if err := json.Unmarshal(jsonBytes, &snap); err == nil {
		return snap, nil
	}
	var data report.ResultFile
	if err := json.Unmarshal(jsonBytes, &data); err != nil {
		return testui.Snapshot{}, fmt.Errorf("parse gavel results: %w", err)
	}
	return testui.Snapshot{
		Tests:   data.Tests,
		Lint:    data.Lint,
		Bench:   data.Bench,
		Error:   data.Error,
		LogTail: data.LogTail,
		Status:  testui.SnapshotStatus{LintRun: len(data.Lint) > 0},
	}, nil
}

type prFetcher func(github.Options, int) (*github.PRInfo, error)

// DownloadPRSnapshot merges every gavel artifact published on a PR into a
// single snapshot, so `gavel test --pr` / `gavel lint --pr` can narrow to the
// PR's failures the same way `--failed` narrows to a local run's. Matrix shards
// upload one artifact each; their trees are concatenated because a failure in
// any shard is a failure to re-run.
func DownloadPRSnapshot(opts github.Options, prNumber int) (*testui.Snapshot, error) {
	return downloadPRSnapshot(opts, prNumber, github.FetchPR, github.DownloadArtifact)
}

func downloadPRSnapshot(opts github.Options, prNumber int, fetchPR prFetcher, download artifactDownloader) (*testui.Snapshot, error) {
	pr, err := fetchPR(opts, prNumber)
	if err != nil {
		return nil, fmt.Errorf("fetch PR #%d: %w", prNumber, err)
	}

	comments := append(append([]github.PRComment{}, pr.Comments...), pr.ReviewThreads...)
	artifacts := github.FindGavelArtifacts(comments)
	if len(artifacts) == 0 {
		return nil, fmt.Errorf("PR #%d has no gavel result artifacts — no gavel sticky comment found on the PR", pr.Number)
	}

	merged := &testui.Snapshot{}
	found := 0
	for _, artifact := range artifacts {
		jsonBytes, err := download(opts, artifact.ArtifactID)
		if errors.Is(err, github.ErrArtifactResultsNotFound) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("download artifact %d (%s): %w", artifact.ArtifactID, artifact.StickyID, err)
		}
		snap, err := DecodeArtifactSnapshot(jsonBytes)
		if err != nil {
			return nil, fmt.Errorf("artifact %d (%s): %w", artifact.ArtifactID, artifact.StickyID, err)
		}
		if err := normalizeArtifactSnapshot(&snap); err != nil {
			return nil, fmt.Errorf("artifact %d (%s): %w", artifact.ArtifactID, artifact.StickyID, err)
		}
		if err := mergeSnapshotGit(merged, snap.Git); err != nil {
			return nil, fmt.Errorf("artifact %d (%s): %w", artifact.ArtifactID, artifact.StickyID, err)
		}
		found++
		merged.Tests = append(merged.Tests, snap.Tests...)
		merged.Lint = append(merged.Lint, snap.Lint...)
		merged.Status.LintRun = merged.Status.LintRun || len(snap.Lint) > 0
	}
	if found == 0 {
		return nil, fmt.Errorf("PR #%d: none of the %d gavel artifacts contained a results payload", pr.Number, len(artifacts))
	}
	return merged, nil
}

func normalizeArtifactSnapshot(snapshot *testui.Snapshot) error {
	if snapshot.Git == nil || strings.TrimSpace(snapshot.Git.Root) == "" {
		return nil
	}
	return normalizeArtifactTests(snapshot.Tests, snapshot.Git.Root)
}

func normalizeArtifactTests(tests []parsers.Test, gitRoot string) error {
	for i := range tests {
		test := &tests[i]
		workDir := test.WorkDir
		if workDir != "" {
			if test.PackagePath != "" {
				packagePath, err := artifactRelativePath(gitRoot, workDir, test.PackagePath)
				if err != nil {
					return fmt.Errorf("normalize package for test %q: %w", test.Name, err)
				}
				test.PackagePath = "./" + packagePath
			}
			if test.File != "" {
				file, err := artifactRelativePath(gitRoot, workDir, test.File)
				if err != nil {
					return fmt.Errorf("normalize file for test %q: %w", test.Name, err)
				}
				test.File = file
			}
			test.WorkDir = gitRoot
		}
		if err := normalizeArtifactTests(test.Children, gitRoot); err != nil {
			return err
		}
	}
	return nil
}

func artifactRelativePath(gitRoot, workDir, value string) (string, error) {
	root := portableArtifactPath(gitRoot)
	if !portableArtifactPathIsAbs(root) {
		return "", fmt.Errorf("git root %q is not absolute", gitRoot)
	}
	workDir = portableArtifactPath(workDir)
	workDirRel, err := portableArtifactPathRel(root, workDir)
	if err != nil {
		return "", err
	}

	value = portableArtifactPath(value)
	var relative string
	if portableArtifactPathIsAbs(value) {
		relative, err = portableArtifactPathRel(root, value)
		if err != nil {
			return "", err
		}
	} else {
		relative = path.Clean(path.Join(workDirRel, value))
		if relative == ".." || strings.HasPrefix(relative, "../") {
			return "", fmt.Errorf("path %q escapes git root %q", value, gitRoot)
		}
	}
	return relative, nil
}

func portableArtifactPath(value string) string {
	return path.Clean(strings.ReplaceAll(strings.TrimSpace(value), `\`, "/"))
}

func portableArtifactPathIsAbs(value string) bool {
	return strings.HasPrefix(value, "/") || (len(value) >= 3 && value[1] == ':' && value[2] == '/')
}

func portableArtifactPathRel(root, value string) (string, error) {
	if value == root {
		return ".", nil
	}
	prefix := strings.TrimSuffix(root, "/") + "/"
	if strings.HasPrefix(value, prefix) {
		return strings.TrimPrefix(value, prefix), nil
	}
	return "", fmt.Errorf("path %q is outside git root %q", value, root)
}

func mergeSnapshotGit(merged *testui.Snapshot, git *testui.SnapshotGit) error {
	if git == nil {
		return nil
	}
	if merged.Git == nil {
		gitCopy := *git
		merged.Git = &gitCopy
		return nil
	}
	if merged.Git.Repo != "" && git.Repo != "" && merged.Git.Repo != git.Repo {
		return fmt.Errorf("artifacts reference different repositories: %q and %q", merged.Git.Repo, git.Repo)
	}
	if merged.Git.SHA != "" && git.SHA != "" && merged.Git.SHA != git.SHA {
		merged.Git.SHA = ""
	}
	return nil
}
