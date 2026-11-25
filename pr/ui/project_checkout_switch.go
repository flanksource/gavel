package ui

import (
	"errors"
	"net/http"
	"strings"

	"github.com/flanksource/gavel/git/checkoutswitch"
)

type projectCheckoutSwitchRequest struct {
	Branch       string              `json:"branch"`
	ExpectedHead string              `json:"expectedHead"`
	Sync         checkoutswitch.Sync `json:"sync"`
}

// handleProjectCheckoutSwitchPlan reports what switching the project's primary
// checkout to ?branch= would do, against the last-fetched upstream.
func (s *Server) handleProjectCheckoutSwitchPlan(w http.ResponseWriter, r *http.Request) {
	project, err := s.visibleProject(r.Context(), r.PathValue("name"))
	if err != nil {
		respondError(w, statusForProjectErr(err), err.Error())
		return
	}
	plan, err := checkoutswitch.Inspect(r.Context(), project.ResolvedDir(), strings.TrimSpace(r.URL.Query().Get("branch")))
	if err != nil {
		respondCheckoutSwitchError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, plan)
}

// handleProjectCheckoutSwitch switches the project's primary checkout to a
// local branch, first fast-forwarding or resetting it to its fetched upstream
// when asked.
func (s *Server) handleProjectCheckoutSwitch(w http.ResponseWriter, r *http.Request) {
	var request projectCheckoutSwitchRequest
	if err := decodeStrict(r, &request); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	project, err := s.visibleProject(r.Context(), r.PathValue("name"))
	if err != nil {
		respondError(w, statusForProjectErr(err), err.Error())
		return
	}
	dir := project.ResolvedDir()
	result, err := checkoutswitch.Switch(r.Context(), checkoutswitch.Options{
		Repo: dir, Branch: strings.TrimSpace(request.Branch), ExpectedHead: strings.TrimSpace(request.ExpectedHead), Sync: request.Sync,
	})
	s.touchGit(dir)
	if err != nil {
		respondCheckoutSwitchError(w, err)
		return
	}
	s.notify()
	respondJSON(w, http.StatusOK, result)
}

func respondCheckoutSwitchError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, checkoutswitch.ErrInvalidOptions):
		respondError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, checkoutswitch.ErrRefused):
		respondError(w, http.StatusConflict, err.Error())
	default:
		respondError(w, http.StatusInternalServerError, err.Error())
	}
}
