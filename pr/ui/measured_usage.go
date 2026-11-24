// Serves an organization's measured CPU and memory, sampled from substrate's Prometheus through a
// Mission Control View given the organization's ID: /api/usage/measured to its admins, and
// /api/platform/orgs/{id}/measured-usage to platform admins.
package ui

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/flanksource/gavel/missioncontrol"
	"github.com/flanksource/gavel/tenant"
	"github.com/google/uuid"
)

// errNoOrgView answers the measured use routes when serve runs without its View.
var errNoOrgView = errors.New("no View measures an organization's CPU and memory: start gavel serve with --mission-control-org-view namespace/name")

// MeasuredUsage reads an organization's measured CPU and memory from View, whose queries read the
// organization's ID as var.org.
type MeasuredUsage struct {
	// MissionControl is nil when serve runs without one.
	MissionControl *missioncontrol.Source
	// View is the View's namespace/name; empty when serve runs without one.
	View string
}

// serve answers org's measured use, or why there is none.
func (m MeasuredUsage) serve(w http.ResponseWriter, r *http.Request, org uuid.UUID) {
	if m.MissionControl == nil {
		writeJSONError(w, http.StatusNotFound, errNoMissionControl)
		return
	}
	if m.View == "" {
		writeJSONError(w, http.StatusNotFound, errNoOrgView)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), platformSourceTimeout)
	defer cancel()
	view, err := m.MissionControl.View(ctx, m.View, map[string]string{"org": org.String()})
	if err != nil {
		writeJSONError(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, view)
}

// SetMeasuredUsage sets where the organization's measured CPU and memory are read from.
func (s *Server) SetMeasuredUsage(measured MeasuredUsage) {
	s.measuredUsage = measured
}

func (s *Server) handleMeasuredUsage(w http.ResponseWriter, r *http.Request) {
	if !tenant.FromContext(r.Context()).IsAdmin() {
		writeJSONError(w, http.StatusForbidden, errors.New("only an organization admin sees its measured CPU and memory"))
		return
	}
	s.measuredUsage.serve(w, r, s.orgID())
}

func (p *PlatformAPI) orgMeasuredUsage(w http.ResponseWriter, r *http.Request) {
	org, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, fmt.Errorf("organization ID %q is not a UUID", r.PathValue("id")))
		return
	}
	MeasuredUsage{MissionControl: p.MissionControl, View: p.OrgView}.serve(w, r, org)
}
