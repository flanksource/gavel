package ui

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/flanksource/gavel/github"
	prlifecycle "github.com/flanksource/gavel/pr/lifecycle"
	"github.com/flanksource/gavel/pr/model"
	"github.com/flanksource/gavel/prwatch"
)

func (s *Server) prLifecycleFailures(host prlifecycle.Host, snapshot prlifecycle.Snapshot) ([]prlifecycle.Failure, error) {
	comments := make([]model.PRComment, 0, len(snapshot.Activity))
	for _, activity := range snapshot.Activity {
		comments = append(comments, model.PRComment{Body: activity.Body})
	}
	artifacts := github.FindGavelArtifacts(comments)
	options := s.ghOpts
	options.Repo = host.Repo.FullName()
	result := []prlifecycle.Failure{}
	for _, artifact := range artifacts {
		run, err := github.FetchRunJobs(options, artifact.RunID, github.RunLogOptions{})
		if err != nil {
			return nil, fmt.Errorf("PR CI run %d: %w", artifact.RunID, err)
		}
		if run.HeadSHA != snapshot.PR.HeadSHA {
			continue
		}
		var completedAt time.Time
		for _, job := range run.Jobs {
			if job.CompletedAt.After(completedAt) {
				completedAt = job.CompletedAt
			}
		}
		data, err := github.DownloadArtifact(options, artifact.ArtifactID)
		if err != nil {
			return nil, fmt.Errorf("PR CI artifact %d: %w", artifact.ArtifactID, err)
		}
		summary := prwatch.ComputeGavelSummary(data, artifact)
		if summary.Error != "" {
			return nil, fmt.Errorf("PR CI artifact %d: %s", artifact.ArtifactID, summary.Error)
		}
		encoded, err := json.Marshal(summary)
		if err != nil {
			return nil, err
		}
		for _, classified := range []struct {
			kind   string
			failed bool
		}{{"test", summary.TestsFailed > 0}, {"lint", summary.LintViolations > 0}} {
			if !classified.failed {
				continue
			}
			result = append(result, prlifecycle.Failure{Kind: classified.kind, RunID: strconv.FormatInt(artifact.RunID, 10) + ":" + strconv.FormatInt(artifact.ArtifactID, 10), HeadSHA: run.HeadSHA, Summary: string(encoded), URL: artifact.ArtifactURL, At: completedAt})
		}
	}
	return result, nil
}
