package prwatch

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/report"
	"github.com/flanksource/gavel/testrunner/parsers"
)

// maxDetailItems bounds how many failing tests / lint violations a summary
// carries. The counts stay exact; only the detail lists are capped so a shard
// summary stays small enough to stream in an SSE frame. The full run is one
// click away behind ArtifactURL.
const maxDetailItems = 5

// GavelResultsSummary is one gavel CI artifact reduced to what a PR view needs:
// exact roll-up counts plus a bounded slice of the run itself. Failures and
// Lint keep gavel's native shapes so every consumer renders them with the same
// renderers `gavel test` / `gavel lint` use — parsers.Test.Pretty and
// linters.SummaryView in the terminal, TestNode / LintView in the dashboard.
type GavelResultsSummary struct {
	StickyID         string        `json:"stickyId,omitempty"`
	RunID            int64         `json:"-"`
	CommentID        int64         `json:"-"`
	ArtifactID       int64         `json:"artifactId"`
	ArtifactURL      string        `json:"artifactUrl"`
	TestsPassed      int           `json:"testsPassed"`
	TestsFailed      int           `json:"testsFailed"`
	TestsSkipped     int           `json:"testsSkipped"`
	TestsTotal       int           `json:"testsTotal"`
	LintViolations   int           `json:"lintViolations"`
	LintLinters      int           `json:"lintLinters"`
	HasBench         bool          `json:"hasBench"`
	BenchRegressions int           `json:"benchRegressions,omitempty"`
	Error            string        `json:"error,omitempty"`
	ExitCode         *int          `json:"exitCode,omitempty"`
	LogTail          string        `json:"logTail,omitempty"`
	Duration         time.Duration `json:"duration,omitempty"`

	// Failures holds up to maxDetailItems failing leaf tests, with their
	// stdout/stderr truncated to the compact-report budget.
	Failures []parsers.Test `json:"failures,omitempty"`
	// Lint holds the linters that found violations or errored, with their
	// violations trimmed to maxDetailItems in total across linters.
	Lint []*linters.LinterResult `json:"lint,omitempty"`

	// Commands are the local `gavel` invocations that re-run this shard's
	// failures. They are PR-scoped rather than shard-scoped: `--pr` merges
	// every shard's artifact into one snapshot before narrowing.
	Commands []string `json:"commands,omitempty"`
}

// HasFailure reports whether this artifact should fail the caller's exit code.
// It is deliberately broader than statusIcon's Fail branch: a shard whose
// results could not be read (Error) or whose linter died (hasLintFailure)
// renders amber, because a warning must never flip a parent node red in the
// tree — but it is still an unverified shard, and an unverified shard must not
// exit 0. Lint violations alone are excluded: they are Warned in gavel's result
// model, and Warned never fails.
func (s GavelResultsSummary) HasFailure() bool {
	return s.Error != "" || s.TestsFailed > 0 || s.BenchRegressions > 0 || s.hasLintFailure()
}

func ComputeGavelSummary(jsonBytes []byte, artifact github.GavelArtifact) *GavelResultsSummary {
	summary := newGavelSummary(artifact)
	var data report.ResultFile
	if err := json.Unmarshal(jsonBytes, &data); err != nil {
		summary.Error = fmt.Sprintf("parse artifact: %v", err)
		return summary
	}
	if data.IsCrash() {
		summary.Error = data.Error
		summary.ExitCode = data.ExitCodeValue()
		summary.LogTail = data.LogTail
		return summary
	}
	for _, root := range data.Tests {
		walkTestCounts(root, summary)
	}
	appendLintSummary(data, summary)
	if data.Bench != nil {
		summary.HasBench = true
		for _, delta := range data.Bench.Deltas {
			if delta.IsRegression(data.Bench.Threshold) {
				summary.BenchRegressions++
			}
		}
	}
	return summary
}

func newGavelSummary(artifact github.GavelArtifact) *GavelResultsSummary {
	return &GavelResultsSummary{
		StickyID: artifact.StickyID, RunID: artifact.RunID, CommentID: artifact.CommentID,
		ArtifactID: artifact.ArtifactID, ArtifactURL: artifact.ArtifactURL,
	}
}

// appendLintSummary counts every violation but keeps only the linters worth
// showing: those with violations, and those that failed outright (an eslint
// config error is the most useful thing in the artifact when it happens).
func appendLintSummary(data report.ResultFile, summary *GavelResultsSummary) {
	remaining := maxDetailItems
	for _, result := range data.Lint {
		if result == nil || result.Skipped {
			continue
		}
		summary.LintLinters++
		summary.LintViolations += len(result.Violations)
		if len(result.Violations) == 0 && result.Error == "" {
			continue
		}
		kept := min(remaining, len(result.Violations))
		if kept == 0 && result.Error == "" {
			continue
		}
		trimmed := *result
		trimmed.RawOutput = ""
		trimmed.Violations = append([]models.Violation(nil), result.Violations[:kept]...)
		remaining -= kept
		summary.Lint = append(summary.Lint, &trimmed)
	}
}

func walkTestCounts(test parsers.Test, summary *GavelResultsSummary) {
	for _, child := range test.Children {
		walkTestCounts(child, summary)
	}
	if len(test.Children) > 0 || test.IsFolder() {
		return
	}
	summary.TestsTotal++
	summary.Duration += test.Duration
	switch {
	case test.Failed:
		summary.TestsFailed++
		if len(summary.Failures) < maxDetailItems {
			summary.Failures = append(summary.Failures, detailTest(test))
		}
	case test.Skipped, test.Pending:
		summary.TestsSkipped++
	case test.Passed:
		summary.TestsPassed++
	}
}

// detailTest bounds a failing leaf for transport: it is already childless, and
// its captured output is clamped to the compact-report budget so one runaway
// log cannot dominate the payload.
func detailTest(test parsers.Test) parsers.Test {
	trimmed := test
	trimmed.Children = nil
	trimmed.Stdout = truncateOutput(test.Stdout)
	trimmed.Stderr = truncateOutput(test.Stderr)
	return trimmed
}

func truncateOutput(body string) string {
	if strings.TrimSpace(body) == "" {
		return ""
	}
	return report.TruncateBlock(body, report.DefaultBudget.MaxLinesPerFailure, report.DefaultBudget.MaxCharsPerLine)
}

func AggregateGavelSummaries(shards []*GavelResultsSummary) *GavelResultsSummary {
	var only *GavelResultsSummary
	aggregate := &GavelResultsSummary{}
	count := 0
	for _, shard := range shards {
		if shard == nil {
			continue
		}
		count++
		only = shard
		aggregateGavelSummary(aggregate, shard)
	}
	if count == 0 {
		return nil
	}
	if count == 1 {
		return only
	}
	return aggregate
}

func aggregateGavelSummary(aggregate, shard *GavelResultsSummary) {
	aggregate.TestsPassed += shard.TestsPassed
	aggregate.TestsFailed += shard.TestsFailed
	aggregate.TestsSkipped += shard.TestsSkipped
	aggregate.TestsTotal += shard.TestsTotal
	aggregate.LintViolations += shard.LintViolations
	aggregate.LintLinters += shard.LintLinters
	aggregate.BenchRegressions += shard.BenchRegressions
	aggregate.Duration += shard.Duration
	aggregate.HasBench = aggregate.HasBench || shard.HasBench
	for _, failure := range shard.Failures {
		if len(aggregate.Failures) < maxDetailItems {
			aggregate.Failures = append(aggregate.Failures, failure)
		}
	}
	for _, result := range shard.Lint {
		if len(aggregate.Lint) < maxDetailItems {
			aggregate.Lint = append(aggregate.Lint, result)
		}
	}
}

// annotateReproduceCommands attaches the `gavel test --pr` / `gavel lint --pr`
// invocations that re-run what the shard failed on. Only failing shards get
// them — a green shard has nothing to reproduce.
func annotateReproduceCommands(summaries []*GavelResultsSummary, repo string, prNumber int) {
	ref := strconv.Itoa(prNumber)
	if repo != "" {
		ref = fmt.Sprintf("%s#%d", repo, prNumber)
	}
	for _, summary := range summaries {
		if summary == nil {
			continue
		}
		summary.Commands = nil
		if summary.TestsFailed > 0 {
			summary.Commands = append(summary.Commands, "gavel test --pr "+ref)
		}
		if summary.LintViolations > 0 || summary.hasLintFailure() {
			summary.Commands = append(summary.Commands, "gavel lint --pr "+ref)
		}
	}
}

type artifactDownloader func(github.Options, int64) ([]byte, error)

func FetchGavelArtifacts(opts github.Options, artifacts []github.GavelArtifact) []*GavelResultsSummary {
	return fetchGavelArtifacts(opts, artifacts, github.DownloadArtifact)
}

func fetchGavelArtifacts(opts github.Options, artifacts []github.GavelArtifact, download artifactDownloader) []*GavelResultsSummary {
	const maxConcurrent = 4
	out := make([]*GavelResultsSummary, len(artifacts))
	semaphore := make(chan struct{}, maxConcurrent)
	var group sync.WaitGroup
	for i, artifact := range artifacts {
		group.Add(1)
		semaphore <- struct{}{}
		go func(i int, artifact github.GavelArtifact) {
			defer group.Done()
			defer func() { <-semaphore }()
			jsonBytes, err := download(opts, artifact.ArtifactID)
			if errors.Is(err, github.ErrArtifactResultsNotFound) {
				return
			}
			if err != nil {
				logger.Warnf("artifact %d (%s) download failed: %v", artifact.ArtifactID, artifact.StickyID, err)
				out[i] = newGavelSummary(artifact)
				out[i].Error = err.Error()
				return
			}
			out[i] = ComputeGavelSummary(jsonBytes, artifact)
		}(i, artifact)
	}
	group.Wait()
	return compactGavelSummaries(out)
}

func compactGavelSummaries(summaries []*GavelResultsSummary) []*GavelResultsSummary {
	usable := make([]*GavelResultsSummary, 0, len(summaries))
	for _, summary := range summaries {
		if summary != nil {
			usable = append(usable, summary)
		}
	}
	return usable
}

func removeRenderedArtifactComments(comments []github.PRComment, results []*GavelResultsSummary) []github.PRComment {
	commentIDs := make(map[int64]bool, len(results))
	for _, result := range results {
		if result != nil && result.CommentID != 0 {
			commentIDs[result.CommentID] = true
		}
	}
	filtered := make([]github.PRComment, 0, len(comments))
	for _, comment := range comments {
		if !commentIDs[comment.ID] {
			filtered = append(filtered, comment)
		}
	}
	return filtered
}
