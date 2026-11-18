package editor

import (
	"context"
	"net/http"

	"github.com/flanksource/gavel/verify"
	"github.com/flanksource/gavel/workspaceaccess"
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
	s.mu.Lock()
	listing := s.listing
	s.mu.Unlock()
	added, err := newStore(StoreOptions{Paths: []string{path}, Listing: listing}, true)
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
	s.paths = append(s.paths, added.paths...)
	return nil
}

// Switch replaces every edited file and folder with the folder dir, listed by listing.
func (s *Store) Switch(ctx context.Context, dir string, listing verify.EditConfig) error {
	if s.Access != nil {
		if _, err := s.Access.Authorize(ctx, dir); err != nil {
			return err
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.replaceLocked(StoreOptions{Paths: []string{dir}, Listing: listing})
}

// Relist lists the edited files and folders again by listing.
func (s *Store) Relist(listing verify.EditConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.replaceLocked(StoreOptions{Paths: s.paths, Listing: listing})
}

func (s *Store) replaceLocked(opts StoreOptions) error {
	fresh, err := newStore(opts, true)
	if err != nil {
		return err
	}
	s.files, s.roots, s.paths, s.listing = fresh.files, fresh.roots, fresh.paths, fresh.listing
	return nil
}

func (s *Store) refresh(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.Access != nil {
		for _, root := range s.roots {
			if _, err := s.Access.Authorize(ctx, root.display); err != nil {
				return err
			}
		}
	}
	return s.replaceLocked(StoreOptions{Paths: s.paths, Listing: s.listing})
}

func (s *Server) refreshFiles(w http.ResponseWriter, r *http.Request) {
	if !workspaceaccess.Trusted(r) {
		http.Error(w, "refreshing files requires a local same-origin caller", http.StatusForbidden)
		return
	}
	if err := s.Store.refresh(r.Context()); err != nil {
		workspaceaccess.Respond(w, nil, err)
		return
	}
	if err := s.Reload(); err != nil {
		respond(w, nil, err)
		return
	}
	respond(w, map[string]any{"roots": s.Store.Roots(), "files": s.Store.Files()}, nil)
}

// Listing is which files the folders list.
func (s *Store) Listing() verify.EditConfig {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.listing
}

// Roots are the edited folders, as selected.
func (s *Store) Roots() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	roots := []string{}
	for _, root := range s.roots {
		roots = append(roots, root.display)
	}
	return roots
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
	workspaceaccess.Respond(w, map[string]any{"roots": s.Store.Roots(), "files": s.Store.Files()}, nil)
}
