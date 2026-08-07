package prwatch

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/report"
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
