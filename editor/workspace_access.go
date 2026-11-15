package editor

import (
	"context"
	"github.com/flanksource/gavel/workspaceaccess"
	"net/http"
)

func (s *Store) EnableAccess() {
	roots := []string{}
	for _, r := range s.roots {
		roots = append(roots, r.canonical)
	}
	s.Access = workspaceaccess.New(roots)
}
func (s *Store) OpenFolder(ctx context.Context, path string) error {
	if _, err := s.Access.Authorize(ctx, path); err != nil {
		return err
	}
	added, err := newStore([]string{path}, true)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, root := range added.roots {
		if err = s.addRoot(root.display); err != nil {
			return err
		}
	}
	for _, e := range added.files {
		s.add(e)
	}
	return nil
}
func (s *Server) opened(w http.ResponseWriter, r *http.Request) {
	if !workspaceaccess.Trusted(r) {
		http.Error(w, "opening folders requires a local same-origin caller", 403)
		return
	}
	if r.Method == "POST" {
		var input struct {
			Path string `json:"path"`
		}
		if !decode(w, r, &input) {
			return
		}
		if err := s.Store.OpenFolder(r.Context(), input.Path); err != nil {
			workspaceaccess.Respond(w, nil, err)
			return
		}
		if err := s.Reload(); err != nil {
			workspaceaccess.Respond(w, nil, err)
			return
		}
	}
	s.Store.mu.Lock()
	roots := []string{}
	for _, root := range s.Store.roots {
		roots = append(roots, root.display)
	}
	s.Store.mu.Unlock()
	workspaceaccess.Respond(w, map[string]any{"roots": roots, "files": s.Store.Files()}, nil)
}
