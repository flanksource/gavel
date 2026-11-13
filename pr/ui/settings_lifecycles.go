package ui

import (
	"net/http"

	"github.com/flanksource/gavel/workflows"
)

func (s *Server) handleSettingsLifecycles(w http.ResponseWriter, r *http.Request) {
	_, dir, err := resolveSettingsDir(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	kind := r.URL.Query().Get("kind")
	var view workflows.View
	switch r.Method {
	case http.MethodGet:
		view, err = workflows.Inspect(kind, dir)
	case http.MethodPost, http.MethodPut:
		var edit workflows.Edit
		r.Body = http.MaxBytesReader(w, r.Body, 2<<20)
		if err := decodeRuntimePresetRequest(r, &edit); err != nil {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		if r.Method == http.MethodPost {
			view, err = workflows.Preview(kind, dir, &edit)
		} else {
			view, err = workflows.Save(kind, dir, edit)
		}
	default:
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, view)
}
