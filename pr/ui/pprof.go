package ui

import (
	"encoding/json"
	"net"
	"net/http"
	"net/http/pprof"

	"github.com/flanksource/captain/pkg/monitor"
	"github.com/flanksource/commons/logger"
)

// registerPprof exposes the Go runtime profiler on the serve mux. `gavel serve`
// is a long-lived process that embeds the session monitor, so attributing its
// CPU and heap needs a live profile rather than a guess.
//
// The handlers are gated to loopback: profiles carry goroutine stacks, the
// process command line, and enough timing detail to fingerprint the workload,
// and the serve port is reachable from the rest of the network.
func registerPprof(mux *http.ServeMux) {
	// pprof.Index serves the named profiles (heap, goroutine, allocs, block,
	// mutex) for any path under the prefix that is not handled explicitly.
	for path, handler := range map[string]http.HandlerFunc{
		"/debug/pprof/":        pprof.Index,
		"/debug/pprof/cmdline": pprof.Cmdline,
		"/debug/pprof/profile": pprof.Profile,
		"/debug/pprof/symbol":  pprof.Symbol,
		"/debug/pprof/trace":   pprof.Trace,
	} {
		mux.Handle(path, loopbackOnly(handler))
	}
}

// registerIngestStats serves the embedded session monitor's counters next to
// the profiler. offerRatio is the one to read: it is what parsed transcript
// lines actually reached the database, and a value that stays at 1 means every
// append is rewriting a whole transcript again.
func registerIngestStats(mux *http.ServeMux, read func() (monitor.IngestStats, bool)) {
	mux.Handle("/debug/ingest", loopbackOnly(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		stats, running := read()
		if !running {
			http.Error(w, "session monitor is not running", http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(struct {
			monitor.IngestStats
			OfferRatio float64 `json:"offerRatio"`
		}{IngestStats: stats, OfferRatio: stats.OfferRatio()}); err != nil {
			logger.Warnf("write ingest stats: %v", err)
		}
	})))
}

// loopbackOnly answers 404 rather than 403 for off-host callers so the profiler
// is not advertised to anyone who cannot already use it.
func loopbackOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !isLoopbackRequest(r) {
			http.NotFound(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func isLoopbackRequest(r *http.Request) bool {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}
