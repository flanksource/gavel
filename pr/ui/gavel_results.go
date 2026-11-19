package ui

import (
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/prwatch"
	testui "github.com/flanksource/gavel/testrunner/ui"
	"github.com/flanksource/gavel/testrunner/ui/queryapp"
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

// artifactCache caches downloaded artifact servers keyed by their artifact ID
// list ("123" for one shard, "123,456" for the merged view of several).
type artifactCache struct {
	mu      sync.RWMutex
	entries map[string]*artifactEntry
}

type artifactEntry struct {
	srv     *testui.Server
	handler http.Handler
}

var globalArtifactCache = &artifactCache{
	entries: make(map[string]*artifactEntry),
}

func (c *artifactCache) get(id string) (*artifactEntry, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	e, ok := c.entries[id]
	return e, ok
}

func (c *artifactCache) put(id string, e *artifactEntry) {
	c.mu.Lock()
	defer c.mu.Unlock()
	// Cap at 50 entries — simple eviction: drop everything when full, closing
	// the trace stores the dropped runs reopened.
	if len(c.entries) >= 50 {
		for id, dropped := range c.entries {
			if err := dropped.srv.CloseTraces(); err != nil {
				logger.Errorf("close the trace stores of artifacts %s: %v", id, err)
			}
		}
		c.entries = make(map[string]*artifactEntry)
	}
	c.entries[id] = e
}

// getOrCreateArtifacts returns the cached server for the artifact list, or
// downloads and caches one. Several artifacts are merged into one run — the
// unified view across a PR's shards.
func (s *Server) getOrCreateArtifacts(artifactIDs []int64, repo string) (*artifactEntry, error) {
	key := joinArtifactIDs(artifactIDs)
	if e, ok := globalArtifactCache.get(key); ok {
		return e, nil
	}
	opts := s.ghOpts
	opts.Repo = repo
	if len(artifactIDs) > 1 {
		artifacts := make([]github.GavelArtifact, len(artifactIDs))
		for i, id := range artifactIDs {
			artifacts[i] = github.GavelArtifact{ArtifactID: id, StickyID: strconv.FormatInt(id, 10)}
		}
		merged, err := prwatch.MergeArtifactSnapshots(opts, artifacts)
		if err != nil {
			return nil, fmt.Errorf("artifacts %s: %w", key, err)
		}
		srv := testui.NewServer()
		srv.LoadSnapshot(*merged)
		srv.MarkDone()
		entry := &artifactEntry{srv: srv, handler: srv.Handler()}
		globalArtifactCache.put(key, entry)
		return entry, nil
	}

	artifactID := artifactIDs[0]
	jsonBytes, err := github.DownloadArtifact(opts, artifactID)
	if err != nil {
		return nil, err
	}

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
		if err := queryapp.Serve(srv); err != nil {
			return nil, fmt.Errorf("artifact %d: %w", artifactID, err)
		}
		if err := srv.ReopenTraces(*run); err != nil {
			return nil, fmt.Errorf("artifact %d: %w", artifactID, err)
		}
	}

	entry := &artifactEntry{srv: srv, handler: srv.Handler()}
	globalArtifactCache.put(joinArtifactIDs(artifactIDs), entry)
	return entry, nil
}

// resultsPathPattern matches /results/{owner}/{repo}/{artifactId[,artifactId...]}[/rest...]
var resultsPathPattern = regexp.MustCompile(`^/results/([^/]+/[^/]+)/(\d+(?:,\d+)*)(/.*)?$`)

// resultsShard is one shard the results page can filter to. The PR dashboard
// lists every shard of the PR as ?shard={artifactId}:{label}; the path's
// artifact IDs are the ones currently shown.
type resultsShard struct {
	ArtifactID int64  `json:"artifactId"`
	Label      string `json:"label"`
	Selected   bool   `json:"selected"`
}

func parseArtifactIDs(list string) ([]int64, error) {
	parts := strings.Split(list, ",")
	ids := make([]int64, len(parts))
	for i, part := range parts {
		id, err := strconv.ParseInt(part, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("artifact id %q: %w", part, err)
		}
		ids[i] = id
	}
	return ids, nil
}

func joinArtifactIDs(ids []int64) string {
	parts := make([]string, len(ids))
	for i, id := range ids {
		parts[i] = strconv.FormatInt(id, 10)
	}
	return strings.Join(parts, ",")
}

func parseResultsShards(values []string, selected []int64) ([]resultsShard, error) {
	shards := make([]resultsShard, 0, len(values))
	for _, value := range values {
		idText, label, ok := strings.Cut(value, ":")
		if !ok || label == "" {
			return nil, fmt.Errorf("shard %q: want {artifactId}:{label}", value)
		}
		id, err := strconv.ParseInt(idText, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("shard %q: %w", value, err)
		}
		shards = append(shards, resultsShard{ArtifactID: id, Label: label, Selected: slices.Contains(selected, id)})
	}
	return shards, nil
}

func (s *Server) handleGavelResults(w http.ResponseWriter, r *http.Request) {
	m := resultsPathPattern.FindStringSubmatch(r.URL.Path)
	if len(m) < 3 {
		http.NotFound(w, r)
		return
	}
	repo := m[1]
	artifactIDs, err := parseArtifactIDs(m[2])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	rest := m[3] // e.g. "/api/tests" or "/tests/..." or "" or "/"

	entry, err := s.getOrCreateArtifacts(artifactIDs, repo)
	if err != nil {
		logger.Warnf("artifacts %s: %v", m[2], err)
		http.Error(w, fmt.Sprintf("failed to load artifact: %v", err), http.StatusBadGateway)
		return
	}

	prefix := fmt.Sprintf("/results/%s/%s", repo, m[2])

	// API, bundle and export routes: strip prefix and delegate to testui handler
	if strings.HasPrefix(rest, "/api/") || strings.HasPrefix(rest, testui.QueryMount+"/") || strings.HasPrefix(rest, testui.AssetsMount+"/") || isExportPath(rest) {
		http.StripPrefix(prefix, entry.handler).ServeHTTP(w, r)
		return
	}

	// HTML page: serve testrunner UI with back button
	shards, err := parseResultsShards(r.URL.Query()["shard"], artifactIDs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	backTo := r.URL.Query().Get("backTo")
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, resultsPageHTML(resultsPage{BackTo: backTo, Prefix: prefix, ShardBase: "/results/" + repo, Shards: shards}))
}

func isExportPath(path string) bool {
	return strings.HasSuffix(path, ".json") || strings.HasSuffix(path, ".md")
}

type resultsPage struct {
	BackTo string
	Prefix string
	// ShardBase is the /results/{repo} path a shard selection's artifact IDs
	// are appended to.
	ShardBase string
	Shards    []resultsShard
}

func resultsPageHTML(page resultsPage) string {
	shards := "null"
	if len(page.Shards) > 0 {
		shards = jsonString(map[string]any{"base": page.ShardBase, "shards": page.Shards})
	}
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Gavel Results</title>
    <script>window.__gavelBackTo = %s; window.__gavelBasePath = %s; window.__gavelShards = %s;</script>
</head>
<body>
    <div id="root"></div>
    <script type="module" src="%s"></script>
</body>
</html>`, jsonString(page.BackTo), jsonString(page.Prefix), shards, html.EscapeString(page.Prefix+testui.BundlePath))
}

func jsonString(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}
