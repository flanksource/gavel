package ui

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/flanksource/clicky/sse"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/github/activity"
	ghcache "github.com/flanksource/gavel/github/cache"
)

// activitySnapshot is the JSON shape returned by /api/activity and pushed
// over /api/activity/stream.
type activitySnapshot struct {
	Entries []activity.Entry `json:"entries"`
	Stats   activity.Stats   `json:"stats"`
}

func snapshotActivity(limit int) activitySnapshot {
	entries, stats := activity.Shared().Snapshot(limit)
	if entries == nil {
		entries = []activity.Entry{}
	}
	return activitySnapshot{Entries: entries, Stats: stats}
}

// parseLimit returns a sane ?limit value bounded to [1, maxLimit].
func parseLimit(raw string, def, maxLimit int) int {
	if raw == "" {
		return def
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		return def
	}
	if n > maxLimit {
		return maxLimit
	}
	return n
}

func (s *Server) handleActivity(w http.ResponseWriter, r *http.Request) {
	limit := parseLimit(r.URL.Query().Get("limit"), 100, 500)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(snapshotActivity(limit)) //nolint:errcheck
}

// activityStreamInterval is how often the activity stream re-reads the
// recorder; a frame is sent only when the snapshot changed.
var activityStreamInterval = time.Second

func (s *Server) handleActivityStream(w http.ResponseWriter, r *http.Request) {
	limit := parseLimit(r.URL.Query().Get("limit"), 100, 500)
	err := sse.ServeSnapshot(w, r, sse.SnapshotOptions{
		Interval: activityStreamInterval,
		Load:     func(context.Context) (any, error) { return snapshotActivity(limit), nil },
	})
	if err != nil {
		logger.Warnf("activity stream: %v", err)
	}
}

func (s *Server) handleActivityReset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	activity.Shared().Reset()
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, `{"status":"reset"}`)
}

func (s *Server) handleActivityCache(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ghcache.Shared().Status()) //nolint:errcheck
}
