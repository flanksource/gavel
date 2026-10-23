// Package runs serves the run history recorded in .gavel/run-*.json next to a
// testrunner UI: a paged listing at /api/runs and each stored run's own view
// mounted at /runs/{runID}/.
package runs

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/flanksource/gavel/snapshots"
	testui "github.com/flanksource/gavel/testrunner/ui"
)

const (
	// Mount is the URL prefix stored runs are served under.
	Mount = "/runs"
	// ListPath answers the listing at the root and under every run's prefix,
	// so the page can always reach it relative to its own base path.
	ListPath = "/api/runs"

	defaultPageSize = 50
	maxPageSize     = 200
)

var runIDPattern = regexp.MustCompile(`^` + regexp.QuoteMeta(snapshots.PerRunPrefix) + `[A-Za-z0-9-]+$`)

// Options configures Handler.
type Options struct {
	// WorkDir holds the .gavel directory whose run-*.json files are listed.
	WorkDir string
	// Latest serves every path outside Mount and ListPath: the default view.
	Latest http.Handler
	// LatestRunIDs are the runs merged into the default view.
	LatestRunIDs []string
}

// Run is one listed run and the URL of its view.
type Run struct {
	snapshots.RunInfo
	Href string `json:"href"`
}

// Page is one page of the run listing.
type Page struct {
	// Current is the run being viewed; empty for the default view.
	Current string `json:"current,omitempty"`
	// Latest are the runs merged into the default view at LatestHref.
	Latest     []string `json:"latest,omitempty"`
	LatestHref string   `json:"latestHref"`
	Runs       []Run    `json:"runs"`
	// Next is the cursor for the following page; empty when none follows.
	Next string `json:"next,omitempty"`
}

type handler struct {
	opts    Options
	servers *testui.MultiServer
	loadMu  sync.Mutex
}

// Handler serves the run listing and the stored runs, delegating everything
// else to opts.Latest.
func Handler(opts Options) http.Handler {
	return &handler{opts: opts, servers: testui.NewMultiServer()}
}

// Close releases the trace stores of every stored run that was opened.
func Close(h http.Handler) error {
	if rh, ok := h.(*handler); ok {
		return rh.servers.CloseTraces()
	}
	return nil
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == ListPath {
		h.serveList(w, r, "")
		return
	}
	rest, ok := strings.CutPrefix(r.URL.Path, Mount+"/")
	if !ok {
		h.opts.Latest.ServeHTTP(w, r)
		return
	}
	runID, rest, _ := strings.Cut(rest, "/")
	if !runIDPattern.MatchString(runID) {
		http.NotFound(w, r)
		return
	}
	prefix := Mount + "/" + runID
	if r.URL.Path == prefix {
		http.Redirect(w, r, prefix+"/", http.StatusMovedPermanently)
		return
	}
	if "/"+rest == ListPath {
		h.serveList(w, r, runID)
		return
	}
	srv, err := h.server(runID, prefix)
	if errors.Is(err, os.ErrNotExist) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.StripPrefix(prefix, srv.Handler()).ServeHTTP(w, r)
}

// server returns the replay server for runID, loading its snapshot on first
// use. Loading is serialized so concurrent first requests share one server.
func (h *handler) server(runID, prefix string) (*testui.Server, error) {
	h.loadMu.Lock()
	defer h.loadMu.Unlock()
	if srv, ok := h.servers.Get(runID); ok {
		return srv, nil
	}
	snap, err := snapshots.LoadRun(h.opts.WorkDir, runID)
	if err != nil {
		return nil, err
	}
	srv := testui.NewServer()
	srv.SetBasePath(prefix)
	srv.LoadSnapshot(*snap)
	run, err := testui.TraceRunOf(snap.Metadata)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", runID, err)
	}
	if run != nil {
		if err := srv.ReopenTraces(*run); err != nil {
			return nil, fmt.Errorf("%s: %w", runID, err)
		}
	}
	h.servers.Put(runID, srv)
	return srv, nil
}

func (h *handler) serveList(w http.ResponseWriter, r *http.Request, current string) {
	limit, err := pageSize(r.URL.Query().Get("limit"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	infos, err := snapshots.ListRuns(h.opts.WorkDir, snapshots.ListRunsOptions{Limit: limit, Before: r.URL.Query().Get("before")})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	page := Page{Current: current, Latest: h.opts.LatestRunIDs, LatestHref: "/", Runs: make([]Run, 0, len(infos))}
	for _, info := range infos {
		page.Runs = append(page.Runs, Run{RunInfo: info, Href: Mount + "/" + info.RunID + "/"})
		// ListRuns pages by RunID, so the cursor is the oldest RunID returned
		// even where start-time order differs.
		if page.Next == "" || info.RunID < page.Next {
			page.Next = info.RunID
		}
	}
	if len(infos) < limit {
		page.Next = ""
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(page) //nolint:errcheck
}

func pageSize(raw string) (int, error) {
	if raw == "" {
		return defaultPageSize, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 1 || n > maxPageSize {
		return 0, fmt.Errorf("limit must be between 1 and %d, got %q", maxPageSize, raw)
	}
	return n, nil
}
