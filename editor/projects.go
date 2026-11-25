package editor

import (
	"fmt"
	"net/http"
	"path/filepath"
	"sort"

	"github.com/flanksource/gavel/projects"
	"github.com/flanksource/gavel/verify"
	"github.com/flanksource/gavel/workspaceaccess"
)

// EditProject is a registered project the editor can switch to.
type EditProject struct {
	Name string `json:"name"`
	Dir  string `json:"dir"`
}

// EditListing is which files the open folders list, where it came from, and
// the folder whose .gavel.yaml a save writes.
type EditListing struct {
	verify.EditConfig
	// Configured is the edit section of the folder's merged .gavel.yaml.
	Configured verify.EditConfig `json:"configured"`
	// Flags are the gavel edit flags, which override Configured key by key.
	Flags verify.EditConfig `json:"flags"`
	Dir   string            `json:"dir"`
	// Defaults are the extensions listed when none are configured.
	Defaults []string `json:"defaults"`
}

// ListingFor is the listing of the folder dir: its .gavel.yaml edit section,
// with each key the gavel edit flags set overriding it.
func (s *Server) ListingFor(dir string) (verify.EditConfig, error) {
	cfg, err := verify.LoadGavelConfig(dir)
	if err != nil {
		return verify.EditConfig{}, err
	}
	listing := cfg.Edit
	if len(s.ListingFlags.Extensions) > 0 {
		listing.Extensions = s.ListingFlags.Extensions
	}
	if len(s.ListingFlags.IncludeIgnored) > 0 {
		listing.IncludeIgnored = s.ListingFlags.IncludeIgnored
	}
	return listing, nil
}

// ListingDir is the folder whose .gavel.yaml configures the listing: the first edited folder.
func (s *Server) ListingDir() (string, error) {
	roots := s.Store.Roots()
	if len(roots) == 0 {
		return "", fmt.Errorf("no folder is open")
	}
	return roots[0], nil
}

// editProjects serves GET /api/edit/projects: the registered local projects, by name.
func (s *Server) editProjects(w http.ResponseWriter, r *http.Request) {
	catalog, err := projects.Shared(r.Context())
	if err != nil {
		respond(w, nil, err)
		return
	}
	listed, err := catalog.List(r.Context())
	if err != nil {
		respond(w, nil, err)
		return
	}
	out := []EditProject{}
	for _, project := range listed {
		if project.Kind == projects.KindDir {
			out = append(out, EditProject{Name: project.Name, Dir: filepath.Clean(project.ResolvedDir())})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	respond(w, map[string]any{"projects": out, "roots": s.Store.Roots()}, nil)
}

// switchProject serves POST /api/edit/project {dir}: every edited file is
// replaced by the folder dir, listed by its own .gavel.yaml and the flags.
func (s *Server) switchProject(w http.ResponseWriter, r *http.Request) {
	if !workspaceaccess.Trusted(r) {
		http.Error(w, "switching projects requires a local same-origin caller", http.StatusForbidden)
		return
	}
	var input struct {
		Dir string `json:"dir"`
	}
	if !decode(w, r, &input) {
		return
	}
	if !filepath.IsAbs(input.Dir) {
		http.Error(w, fmt.Sprintf("project dir must be absolute: %q", input.Dir), http.StatusBadRequest)
		return
	}
	listing, err := s.ListingFor(input.Dir)
	if err == nil {
		err = s.Store.Switch(r.Context(), input.Dir, listing)
	}
	if err == nil {
		err = s.Reload()
	}
	if err != nil {
		workspaceaccess.Respond(w, nil, err)
		return
	}
	respond(w, map[string]any{"roots": s.Store.Roots(), "files": s.Store.Files()}, nil)
}

// editListing serves GET /api/edit/options.
func (s *Server) editListing(w http.ResponseWriter, r *http.Request) {
	listing, err := s.describeListing()
	respond(w, listing, err)
}

func (s *Server) describeListing() (EditListing, error) {
	dir, err := s.ListingDir()
	if err != nil {
		return EditListing{}, err
	}
	cfg, err := verify.LoadGavelConfig(dir)
	if err != nil {
		return EditListing{}, err
	}
	return EditListing{EditConfig: s.Store.Listing(), Configured: cfg.Edit, Flags: s.ListingFlags, Dir: dir, Defaults: builtinExtensions()}, nil
}

// saveListing serves PUT /api/edit/options {extensions, includeIgnored, save}:
// the open folders are listed again by it, and with save it becomes the edit
// section of the first folder's .gavel.yaml.
func (s *Server) saveListing(w http.ResponseWriter, r *http.Request) {
	if !workspaceaccess.Trusted(r) {
		http.Error(w, "changing the listing requires a local same-origin caller", http.StatusForbidden)
		return
	}
	var input struct {
		verify.EditConfig
		Save bool `json:"save"`
	}
	if !decode(w, r, &input) {
		return
	}
	listing, err := input.EditConfig.Normalize()
	if err != nil {
		respond(w, nil, err)
		return
	}
	if input.Save {
		dir, err := s.ListingDir()
		if err == nil && s.Store.Access != nil {
			_, err = s.Store.Access.Authorize(r.Context(), dir)
		}
		if err == nil {
			err = verify.SaveEditConfig(dir, listing)
		}
		if err != nil {
			respond(w, nil, err)
			return
		}
	}
	if err := s.Store.Relist(listing); err != nil {
		respond(w, nil, err)
		return
	}
	if err := s.Reload(); err != nil {
		respond(w, nil, err)
		return
	}
	described, err := s.describeListing()
	respond(w, map[string]any{"listing": described, "roots": s.Store.Roots(), "files": s.Store.Files()}, err)
}
