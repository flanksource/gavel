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
	// StickyID is the gavel sticky-comment id, e.g. "gavel-test-pg15".
	// Empty for the legacy single-artifact path or for an aggregate.
	StickyID         string `json:"stickyId,omitempty"`
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
}

// aggregateGavelSummaries rolls up per-shard results into a single summary
// suitable for the sidebar badge, where there is only room for one number
// per PR. The detail page renders the shards individually and does not use
// this. Returns nil if shards is empty.
func aggregateGavelSummaries(shards []*GavelResultsSummary) *GavelResultsSummary {
	if len(shards) == 0 {
		return nil
	}
	if len(shards) == 1 {
		return shards[0]
	}
	agg := &GavelResultsSummary{}
	for _, s := range shards {
		if s == nil {
			continue
		}
		agg.TestsPassed += s.TestsPassed
		agg.TestsFailed += s.TestsFailed
		agg.TestsSkipped += s.TestsSkipped
		agg.TestsTotal += s.TestsTotal
		agg.LintViolations += s.LintViolations
		agg.LintLinters += s.LintLinters
		agg.BenchRegressions += s.BenchRegressions
		if s.HasBench {
			agg.HasBench = true
		}
		for _, f := range s.TopFailures {
			if len(agg.TopFailures) >= 5 {
				break
			}
			agg.TopFailures = append(agg.TopFailures, f)
		}
		for _, v := range s.TopLintViolations {
			if len(agg.TopLintViolations) >= 5 {
				break
			}
			agg.TopLintViolations = append(agg.TopLintViolations, v)
		}
	}
	return agg
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

	return summary
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
	message := t.Message
	if d := t.FailureDetail; d != nil && d.Summary != "" {
		message = d.Summary
	}
	return TestFailure{
		Name:    t.Name,
		Suite:   suite,
		File:    t.File,
		Line:    t.Line,
		Message: message,
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

// fetchGavelArtifacts downloads each artifact and writes the resulting
// summary into out[i]. Concurrency is capped so a PR with many matrix
// shards doesn't fan out one goroutine per shard onto GitHub's artifacts
// API. The caller owns out and must size it to len(artifacts).
func fetchGavelArtifacts(opts github.Options, artifacts []github.GavelArtifact, out []*GavelResultsSummary) {
	const maxConcurrent = 4
	sem := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup
	for i, a := range artifacts {
		wg.Add(1)
		sem <- struct{}{}
		go func(i int, a github.GavelArtifact) {
			defer wg.Done()
			defer func() { <-sem }()
			jsonBytes, err := github.DownloadArtifact(opts, a.ArtifactID)
			if err != nil {
				logger.Warnf("artifact %d (%s) download failed: %v", a.ArtifactID, a.StickyID, err)
				out[i] = &GavelResultsSummary{
					StickyID:    a.StickyID,
					ArtifactID:  a.ArtifactID,
					ArtifactURL: a.ArtifactURL,
					Error:       err.Error(),
				}
				return
			}
			summary := computeGavelSummary(jsonBytes, a.ArtifactID, a.ArtifactURL)
			summary.StickyID = a.StickyID
			out[i] = summary
		}(i, a)
	}
	wg.Wait()
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
	jsonBytes, err := github.DownloadArtifact(opts, artifactID)
	if err != nil {
		return nil, err
	}

	summary := computeGavelSummary(jsonBytes, artifactID, "")

	srv := testui.NewServer()
	var snap testui.Snapshot
	if err := json.Unmarshal(jsonBytes, &snap); err != nil {
		logger.Warnf("artifact %d: unmarshal as snapshot: %v, trying legacy format", artifactID, err)
		var data gavelResultJSON
		if err := json.Unmarshal(jsonBytes, &data); err != nil {
			return nil, fmt.Errorf("parse artifact %d: %w", artifactID, err)
		}
		snap = testui.Snapshot{
			Tests: data.Tests,
			Lint:  data.Lint,
			Bench: data.Bench,
			Status: testui.SnapshotStatus{
				LintRun: len(data.Lint) > 0,
			},
		}
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
