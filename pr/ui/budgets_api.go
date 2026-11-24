// Serves /api/budgets: a hosted organization's budgets for its substrate runs, with how much of each
// is used, to every member; only its admins set and remove them.
package ui

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/flanksource/gavel/budget"
	"github.com/flanksource/gavel/tenant"
)

// BudgetsResponse is the organization's budgets with their use; CanEdit says whether the caller
// may change them.
type BudgetsResponse struct {
	Budgets []budget.Status `json:"budgets"`
	CanEdit bool            `json:"canEdit"`
}

// budgetsFor answers why there are no budgets on this dashboard, and, when admin is set, why the
// caller may not change them.
func (s *Server) budgetsFor(w http.ResponseWriter, r *http.Request, admin bool) (tenant.Scope, budget.Store, bool) {
	if !s.sandboxRunsAvailable(w, r) {
		return tenant.Scope{}, budget.Store{}, false
	}
	scope := tenant.FromContext(r.Context())
	if admin && !scope.IsAdmin() {
		writeJSONError(w, http.StatusForbidden, errors.New("only an organization admin can change its budgets"))
		return scope, budget.Store{}, false
	}
	return scope, budget.Store{DB: s.agentRuns.Dispatcher.DB}, true
}

func (s *Server) handleBudgetList(w http.ResponseWriter, r *http.Request) {
	scope, store, ok := s.budgetsFor(w, r, false)
	if !ok {
		return
	}
	statuses, err := store.Status(r.Context(), s.orgID(), s.agentRuns.Dispatcher.RunLifetime())
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, BudgetsResponse{Budgets: statuses, CanEdit: scope.IsAdmin()})
}

func (s *Server) handleBudgetSet(w http.ResponseWriter, r *http.Request) {
	scope, store, ok := s.budgetsFor(w, r, true)
	if !ok {
		return
	}
	var body struct {
		Amount float64 `json:"amount"`
		Window string  `json:"window"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4<<10)).Decode(&body); err != nil {
		writeJSONError(w, http.StatusBadRequest, fmt.Errorf(`a budget is JSON like {"amount": 40, "window": "now/M"}: %w`, err))
		return
	}
	set, err := store.Set(r.Context(), s.orgID(), scope.UserID, budget.Budget{
		Resource: budget.Resource(r.PathValue("resource")), Kind: budget.Kind(r.PathValue("kind")), Amount: body.Amount, Window: body.Window,
	})
	if errors.Is(err, budget.ErrInvalid) {
		writeJSONError(w, http.StatusBadRequest, err)
		return
	}
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, set)
}

func (s *Server) handleBudgetRemove(w http.ResponseWriter, r *http.Request) {
	_, store, ok := s.budgetsFor(w, r, true)
	if !ok {
		return
	}
	err := store.Remove(r.Context(), s.orgID(), budget.Resource(r.PathValue("resource")), budget.Kind(r.PathValue("kind")))
	if errors.Is(err, budget.ErrNotFound) {
		writeJSONError(w, http.StatusNotFound, err)
		return
	}
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, map[string]string{"removed": r.PathValue("kind") + " " + r.PathValue("resource")})
}
