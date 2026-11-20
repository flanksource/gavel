// Serves /api/usage: what a hosted organization used in a window, to every member: the CPU and
// memory its sandbox runs were sized at for as long as they ran, and its model calls' tokens and cost.
package ui

import (
	"fmt"
	"net/http"
	"time"

	"github.com/flanksource/captain/pkg/budgets"
	"github.com/flanksource/gavel/platform"
)

// defaultUsageWindow is the window usage covers unless the request names one: this month.
const defaultUsageWindow = "now/M"

// UsageResponse is an organization's usage in Window, datemath relative to now.
type UsageResponse struct {
	Window string            `json:"window"`
	Usage  platform.OrgUsage `json:"usage"`
}

func (s *Server) handleUsage(w http.ResponseWriter, r *http.Request) {
	if !s.sandboxRunsAvailable(w, r) {
		return
	}
	window := r.URL.Query().Get("window")
	if window == "" {
		window = defaultUsageWindow
	}
	since, err := budgets.WindowStart(window, time.Now())
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err)
		return
	}
	dispatcher := s.agentRuns.Dispatcher
	usage, err := platform.Store{DB: dispatcher.DB}.OrgUsage(r.Context(), s.orgID(), since, dispatcher.RunLifetime())
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, fmt.Errorf("read the organization's usage: %w", err))
		return
	}
	writeJSON(w, UsageResponse{Window: window, Usage: usage})
}
