package ui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/testrunner/bench"
	"github.com/flanksource/gavel/testrunner/parsers"
	testui "github.com/flanksource/gavel/testrunner/ui"
)

type GavelResultsSummary struct {
	ArtifactID       int64  `json:"artifactId"`
	ArtifactURL      string `json:"artifactUrl"`
	TestsPassed      int    `json:"testsPassed"`
	TestsFailed      int    `json:"testsFailed"`
	TestsSkipped     int    `json:"testsSkipped"`
	TestsTotal       int    `json:"testsTotal"`
	LintViolations   int    `json:"lintViolations"`
	LintLinters      int    `json:"lintLinters"`
	HasBench         bool   `json:"hasBench"`
	BenchRegressions int    `json:"benchRegressions,omitempty"`
	Error            string `json:"error,omitempty"`
	// TopFailures lists the first 5 failing tests for at-a-glance triage.
	// Populated in walk order (stable) so the same artifact always yields
	// the same head items.
	TopFailures []TestFailure `json:"topFailures,omitempty"`
	// TopLintViolations lists the first 5 lint findings across all linters.
	TopLintViolations []LintViolation `json:"topLintViolations,omitempty"`
	// Jobs is the per-job breakdown when the PR has more than one gavel
	// artifact. Empty when only a single artifact was discovered. The
	// outer counts/TopFailures fields are the sum across Jobs so PR
	// sidebar badges and existing single-summary consumers keep working.
	Jobs []GavelJobSummary `json:"jobs,omitempty"`
}

// GavelJobSummary holds the merged results from one workflow run / job.
// When a single job uploads N gavel-* artifacts (matrix shards, separate
// bench json, etc.), all of them are downloaded and merged into one entry.
type GavelJobSummary struct {
	// JobName is a human label — typically the artifact name (or a comma list
	// when several artifacts were merged into the same run).
	JobName string `json:"jobName"`
	// RunID is the workflow run that produced these artifacts.
	RunID int64 `json:"runId,omitempty"`
	// ArtifactIDs lists every artifact ID merged into this entry. The first
	// ID is treated as the canonical "open in UI" target.
	ArtifactIDs       []int64         `json:"artifactIds"`
	ArtifactURL       string          `json:"artifactUrl,omitempty"`
	TestsPassed       int             `json:"testsPassed"`
	TestsFailed       int             `json:"testsFailed"`
	TestsSkipped      int             `json:"testsSkipped"`
	TestsTotal        int             `json:"testsTotal"`
	LintViolations    int             `json:"lintViolations"`
	LintLinters       int             `json:"lintLinters"`
	HasBench          bool            `json:"hasBench"`
	BenchRegressions  int             `json:"benchRegressions,omitempty"`
	Error             string          `json:"error,omitempty"`
	TopFailures       []TestFailure   `json:"topFailures,omitempty"`
	TopLintViolations []LintViolation `json:"topLintViolations,omitempty"`
}

type TestFailure struct {
	Name    string `json:"name"`
	Suite   string `json:"suite,omitempty"`
	File    string `json:"file,omitempty"`
	Line    int    `json:"line,omitempty"`
	Message string `json:"message,omitempty"`
	Details string `json:"details,omitempty"`
}

type LintViolation struct {
	Linter  string `json:"linter"`
	File    string `json:"file,omitempty"`
	Line    int    `json:"line,omitempty"`
	Rule    string `json:"rule,omitempty"`
	Message string `json:"message,omitempty"`
}

// gavelResultJSON mirrors the dual-format JSON that gavel emits:
//   - test only: a plain JSON array of parsers.Test
//   - test --lint: an object with "tests" and "lint" keys
type gavelResultJSON struct {
	Tests []parsers.Test          `json:"tests"`
	Lint  []*linters.LinterResult `json:"lint"`
	Bench *bench.BenchComparison  `json:"bench"`
}

func (g *gavelResultJSON) UnmarshalJSON(data []byte) error {
	if strings.HasPrefix(strings.TrimSpace(string(data)), "[") {
		var tests []parsers.Test
		if err := json.Unmarshal(data, &tests); err != nil {
			return err
		}
		g.Tests = tests
		return nil
	}
	type alias gavelResultJSON
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	*g = gavelResultJSON(a)
	return nil
}

// mergeGavelResults concatenates tests/lint/bench from multiple JSON payloads
// belonging to the same job. The first non-nil bench wins (bench comparisons
// are aggregate-level; we don't try to merge two of them).
func mergeGavelResults(payloads ...gavelResultJSON) gavelResultJSON {
	var merged gavelResultJSON
	for _, p := range payloads {
		merged.Tests = append(merged.Tests, p.Tests...)
		merged.Lint = append(merged.Lint, p.Lint...)
		if merged.Bench == nil {
			merged.Bench = p.Bench
		}
	}
	return merged
}

func computeGavelSummary(jsonBytes []byte, artifactID int64, artifactURL string) *GavelResultsSummary {
	var data gavelResultJSON
	if err := json.Unmarshal(jsonBytes, &data); err != nil {
		return &GavelResultsSummary{
			ArtifactID:  artifactID,
			ArtifactURL: artifactURL,
			Error:       fmt.Sprintf("parse artifact: %v", err),
		}
	}

	summary := &GavelResultsSummary{
		ArtifactID:  artifactID,
		ArtifactURL: artifactURL,
	}
	applyResultsToSummary(data, summary)
	return summary
}

// applyResultsToSummary populates the count/top fields of a GavelResultsSummary
// from a (possibly-merged) gavelResultJSON payload. Used both by the
// single-artifact path and by the multi-artifact merge path.
func applyResultsToSummary(data gavelResultJSON, summary *GavelResultsSummary) {
	for _, root := range data.Tests {
		walkTestCounts(root, summary)
	}

	for _, lr := range data.Lint {
		if lr.Skipped {
			continue
		}
		summary.LintLinters++
		summary.LintViolations += len(lr.Violations)
		for _, v := range lr.Violations {
			if len(summary.TopLintViolations) >= 5 {
				break
			}
			summary.TopLintViolations = append(summary.TopLintViolations, LintViolation{
				Linter:  lr.Linter,
				File:    v.File,
				Line:    v.Line,
				Rule:    violationRule(v),
				Message: derefString(v.Message),
			})
		}
	}

	if data.Bench != nil {
		summary.HasBench = true
		for _, d := range data.Bench.Deltas {
			if d.IsRegression(data.Bench.Threshold) {
				summary.BenchRegressions++
			}
		}
	}
}

// computeGavelJobSummary downloads every artifact in `arts`, merges all of
// their .json payloads into one logical result set, and returns a single
// GavelJobSummary representing the job.
//
// `arts` MUST share a workflow run — typically the slice produced by grouping
// FindGavelArtifacts() output by RunID. Download errors on individual
// artifacts are recorded in the Error field rather than aborting the merge,
// so a partially-broken job still surfaces its successful files.
func computeGavelJobSummary(opts github.Options, arts []github.GavelArtifact) GavelJobSummary {
	if len(arts) == 0 {
		return GavelJobSummary{Error: "no artifacts"}
	}

	job := GavelJobSummary{
		RunID:       arts[0].RunID,
		JobName:     jobNameFromArtifacts(arts),
		ArtifactURL: arts[0].URL,
	}
	job.ArtifactIDs = make([]int64, 0, len(arts))

	var payloads []gavelResultJSON
	var partialErrors []string
	for _, a := range arts {
		job.ArtifactIDs = append(job.ArtifactIDs, a.ID)
		if a.Expired {
			partialErrors = append(partialErrors, fmt.Sprintf("artifact %s (%d) expired", a.Name, a.ID))
			continue
		}
		files, err := github.DownloadArtifactFiles(opts, a.ID)
		if err != nil {
			partialErrors = append(partialErrors, fmt.Sprintf("download %s (%d): %v", a.Name, a.ID, err))
			continue
		}
		for _, f := range files {
			var p gavelResultJSON
			if err := json.Unmarshal(f.Body, &p); err != nil {
				partialErrors = append(partialErrors, fmt.Sprintf("parse %s/%s: %v", a.Name, f.Name, err))
				continue
			}
			payloads = append(payloads, p)
		}
	}

	merged := mergeGavelResults(payloads...)

	// Reuse the summary-population logic via a transient summary, then copy
	// the count fields onto the job struct. Keeps a single source of truth
	// for "how do we count tests".
	tmp := &GavelResultsSummary{}
	applyResultsToSummary(merged, tmp)

	job.TestsPassed = tmp.TestsPassed
	job.TestsFailed = tmp.TestsFailed
	job.TestsSkipped = tmp.TestsSkipped
	job.TestsTotal = tmp.TestsTotal
	job.LintViolations = tmp.LintViolations
	job.LintLinters = tmp.LintLinters
	job.HasBench = tmp.HasBench
	job.BenchRegressions = tmp.BenchRegressions
	job.TopFailures = tmp.TopFailures
	job.TopLintViolations = tmp.TopLintViolations

	if len(partialErrors) > 0 && len(payloads) == 0 {
		// Total failure: surface the first error verbatim so the UI shows
		// it instead of a misleading "0 tests" panel.
		job.Error = partialErrors[0]
	} else if len(partialErrors) > 0 {
		job.Error = fmt.Sprintf("%d of %d artifact(s) failed to load: %s",
			len(partialErrors), len(arts), strings.Join(partialErrors, "; "))
	}
	return job
}

// jobNameFromArtifacts produces a human label for a job entry. When all
// artifacts share a name we use that; when names differ we join them so the
// UI can show "gavel-results, gavel-bench".
func jobNameFromArtifacts(arts []github.GavelArtifact) string {
	if len(arts) == 0 {
		return ""
	}
	seen := make(map[string]bool, len(arts))
	var names []string
	for _, a := range arts {
		if seen[a.Name] {
			continue
		}
		seen[a.Name] = true
		names = append(names, a.Name)
	}
	return strings.Join(names, ", ")
}

// summarizeGavelArtifacts groups the discovered artifacts by workflow run,
// downloads + merges each group, then folds the per-job results into a
// single PR-level summary. The PR-level summary's Jobs field carries the
// per-job breakdown for surfaces that want detail.
func summarizeGavelArtifacts(opts github.Options, arts []github.GavelArtifact) *GavelResultsSummary {
	if len(arts) == 0 {
		return nil
	}

	byRun := make(map[int64][]github.GavelArtifact)
	var runOrder []int64
	for _, a := range arts {
		if _, ok := byRun[a.RunID]; !ok {
			runOrder = append(runOrder, a.RunID)
		}
		byRun[a.RunID] = append(byRun[a.RunID], a)
	}

	pr := &GavelResultsSummary{
		ArtifactID:  arts[0].ID,
		ArtifactURL: arts[0].URL,
	}
	for _, runID := range runOrder {
		job := computeGavelJobSummary(opts, byRun[runID])
		pr.Jobs = append(pr.Jobs, job)
		pr.TestsPassed += job.TestsPassed
		pr.TestsFailed += job.TestsFailed
		pr.TestsSkipped += job.TestsSkipped
		pr.TestsTotal += job.TestsTotal
		pr.LintViolations += job.LintViolations
		pr.LintLinters += job.LintLinters
		if job.HasBench {
			pr.HasBench = true
		}
		pr.BenchRegressions += job.BenchRegressions
		for _, f := range job.TopFailures {
			if len(pr.TopFailures) >= 5 {
				break
			}
			pr.TopFailures = append(pr.TopFailures, f)
		}
		for _, v := range job.TopLintViolations {
			if len(pr.TopLintViolations) >= 5 {
				break
			}
			pr.TopLintViolations = append(pr.TopLintViolations, v)
		}
	}
	return pr
}

func walkTestCounts(t parsers.Test, s *GavelResultsSummary) {
	for _, child := range t.Children {
		walkTestCounts(child, s)
	}
	if len(t.Children) > 0 || t.IsFolder() {
		return
	}
	s.TestsTotal++
	switch {
	case t.Failed:
		s.TestsFailed++
		if len(s.TopFailures) < 5 {
			s.TopFailures = append(s.TopFailures, toTestFailure(t))
		}
	case t.Skipped, t.Pending:
		s.TestsSkipped++
	case t.Passed:
		s.TestsPassed++
	}
}

func toTestFailure(t parsers.Test) TestFailure {
	suite := ""
	if len(t.Suite) > 0 {
		suite = strings.Join(t.Suite, " › ")
	}
	details := t.Stderr
	if details == "" {
		details = t.Stdout
	}
	return TestFailure{
		Name:    t.Name,
		Suite:   suite,
		File:    t.File,
		Line:    t.Line,
		Message: t.Message,
		Details: details,
	}
}

func violationRule(v models.Violation) string {
	if v.Code != nil && *v.Code != "" {
		return *v.Code
	}
	if v.Rule != nil && v.Rule.Pattern != "" {
		return v.Rule.Pattern
	}
	return ""
}

func derefString(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// artifactCache caches downloaded artifact servers keyed by artifact ID.
type artifactCache struct {
	mu      sync.RWMutex
	entries map[int64]*artifactEntry
}

type artifactEntry struct {
	srv     *testui.Server
	handler http.Handler
	summary *GavelResultsSummary
}

var globalArtifactCache = &artifactCache{
	entries: make(map[int64]*artifactEntry),
}

func (c *artifactCache) get(id int64) (*artifactEntry, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	e, ok := c.entries[id]
	return e, ok
}

func (c *artifactCache) put(id int64, e *artifactEntry) {
	c.mu.Lock()
	defer c.mu.Unlock()
	// Cap at 50 entries — simple eviction: drop everything when full
	if len(c.entries) >= 50 {
		c.entries = make(map[int64]*artifactEntry)
	}
	c.entries[id] = e
}

// getOrCreate returns a cached artifact entry or downloads and caches one.
func (s *Server) getOrCreateArtifact(artifactID int64, repo string) (*artifactEntry, error) {
	if e, ok := globalArtifactCache.get(artifactID); ok {
		return e, nil
	}

	opts := s.ghOpts
	opts.Repo = repo
	files, err := github.DownloadArtifactFiles(opts, artifactID)
	if err != nil {
		return nil, err
	}

	// Merge every JSON entry in the zip into a single payload before
	// computing summary / loading the snapshot. Single-file artifacts hit
	// the same path with one payload, so behaviour is unchanged for them.
	var payloads []gavelResultJSON
	for _, f := range files {
		var p gavelResultJSON
		if err := json.Unmarshal(f.Body, &p); err != nil {
			logger.Warnf("artifact %d: parse %s: %v", artifactID, f.Name, err)
			continue
		}
		payloads = append(payloads, p)
	}
	if len(payloads) == 0 {
		return nil, fmt.Errorf("artifact %d: no parseable .json file in zip", artifactID)
	}
	merged := mergeGavelResults(payloads...)

	summary := &GavelResultsSummary{ArtifactID: artifactID}
	applyResultsToSummary(merged, summary)

	srv := testui.NewServer()
	snap := testui.Snapshot{
		Tests: merged.Tests,
		Lint:  merged.Lint,
		Bench: merged.Bench,
		Status: testui.SnapshotStatus{
			LintRun: len(merged.Lint) > 0,
		},
	}
	srv.LoadSnapshot(snap)
	srv.MarkDone()

	entry := &artifactEntry{
		srv:     srv,
		handler: srv.Handler(),
		summary: summary,
	}
	globalArtifactCache.put(artifactID, entry)
	return entry, nil
}

// resultsPathPattern matches /results/{owner}/{repo}/{artifactId}[/rest...]
var resultsPathPattern = regexp.MustCompile(`^/results/([^/]+/[^/]+)/(\d+)(/.*)?$`)

func (s *Server) handleGavelResults(w http.ResponseWriter, r *http.Request) {
	m := resultsPathPattern.FindStringSubmatch(r.URL.Path)
	if len(m) < 3 {
		http.NotFound(w, r)
		return
	}
	repo := m[1]
	artifactID, _ := strconv.ParseInt(m[2], 10, 64)
	rest := m[3] // e.g. "/api/tests" or "/tests/..." or "" or "/"

	entry, err := s.getOrCreateArtifact(artifactID, repo)
	if err != nil {
		logger.Warnf("artifact %d: %v", artifactID, err)
		http.Error(w, fmt.Sprintf("failed to load artifact: %v", err), http.StatusBadGateway)
		return
	}

	prefix := fmt.Sprintf("/results/%s/%d", repo, artifactID)

	// API and export routes: strip prefix and delegate to testui handler
	if strings.HasPrefix(rest, "/api/") || isExportPath(rest) {
		http.StripPrefix(prefix, entry.handler).ServeHTTP(w, r)
		return
	}

	// HTML page: serve testrunner UI with back button
	backTo := r.URL.Query().Get("backTo")
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, resultsPageHTML(backTo, prefix))
}

func isExportPath(path string) bool {
	return strings.HasSuffix(path, ".json") || strings.HasSuffix(path, ".md")
}

func resultsPageHTML(backTo, prefix string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Gavel Results</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <script src="https://code.iconify.design/iconify-icon/2.0.0/iconify-icon.min.js"></script>
    <script>window.__gavelBackTo = %s; window.__gavelBasePath = %s;</script>
</head>
<body>
    <div id="root"></div>
    <script>%s</script>
</body>
</html>`, jsonString(backTo), jsonString(prefix), testui.BundleJS())
}

func jsonString(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}
