package ui

import (
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/prwatch"
	testui "github.com/flanksource/gavel/testrunner/ui"
)

type GavelResultsSummary = prwatch.GavelResultsSummary

func aggregateGavelSummaries(shards []*GavelResultsSummary) *GavelResultsSummary {
	return prwatch.AggregateGavelSummaries(shards)
}

func computeGavelSummary(jsonBytes []byte, artifactID int64, artifactURL string) *GavelResultsSummary {
	return prwatch.ComputeGavelSummary(jsonBytes, github.GavelArtifact{ArtifactID: artifactID, ArtifactURL: artifactURL})
}

func fetchGavelArtifacts(opts github.Options, artifacts []github.GavelArtifact) []*GavelResultsSummary {
	return prwatch.FetchGavelArtifacts(opts, artifacts)
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
	// Cap at 50 entries — simple eviction: drop everything when full, closing
	// the trace stores the dropped runs reopened.
	if len(c.entries) >= 50 {
		for id, dropped := range c.entries {
			if err := dropped.srv.CloseTraces(); err != nil {
				logger.Errorf("close the trace stores of artifact %d: %v", id, err)
			}
		}
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

	snap, err := prwatch.DecodeArtifactSnapshot(jsonBytes)
	if err != nil {
		return nil, fmt.Errorf("artifact %d: %w", artifactID, err)
	}

	srv := testui.NewServer()
	srv.LoadSnapshot(snap)
	srv.MarkDone()
	// A CI artifact names a store on the machine that ran it, so it is usually
	// gone here and the run is served without it; a local one opens.
	if run, err := testui.TraceRunOf(snap.Metadata); err != nil {
		return nil, fmt.Errorf("artifact %d: %w", artifactID, err)
	} else if run != nil {
		if err := srv.ReopenTraces(*run); err != nil {
			return nil, fmt.Errorf("artifact %d: %w", artifactID, err)
		}
	}

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

	// API, bundle and export routes: strip prefix and delegate to testui handler
	if strings.HasPrefix(rest, "/api/") || strings.HasPrefix(rest, testui.QueryMount+"/") || strings.HasPrefix(rest, testui.AssetsMount+"/") || isExportPath(rest) {
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
    <script>window.__gavelBackTo = %s; window.__gavelBasePath = %s;</script>
</head>
<body>
    <div id="root"></div>
    <script type="module" src="%s"></script>
</body>
</html>`, jsonString(backTo), jsonString(prefix), html.EscapeString(prefix+testui.BundlePath))
}

func jsonString(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}
