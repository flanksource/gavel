package editor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"slices"

	"github.com/flanksource/gavel/comments"
	"github.com/flanksource/gavel/comments/agents"
	"github.com/flanksource/gavel/comments/postgres"
	"github.com/flanksource/gavel/internal/database"
	ui "github.com/flanksource/gavel/pr/ui"
	"github.com/flanksource/gavel/projects"
	"github.com/flanksource/gavel/settings"
)

type CommentSelector struct {
	Kind     string                `json:"kind"`
	Version  int                   `json:"version"`
	Anchor   string                `json:"anchor,omitempty"`
	View     string                `json:"view"`
	Patch    *comments.PatchAnchor `json:"patch,omitempty"`
	Locators json.RawMessage       `json:"locators,omitempty"`
	HTML     string                `json:"html,omitempty"`
	// Quote is the exact selected text, inside Patch's lines when both are set.
	// It is source text, as the file holds it (Markdown, not the rendered
	// prose), and its exact text is never empty.
	Quote *comments.TextQuote `json:"quote,omitempty"`
	// Cell is the markdown table cell a code-range was made in; Patch is then
	// its row's one line.
	Cell *comments.TableCell `json:"cell,omitempty"`
	// Region is the box a region selector marks on an image.
	Region *comments.Region `json:"region,omitempty"`
}

var commentViews = []string{"preview", "source", "rich", "form"}

// selectorKinds are the selector kinds each file kind takes: an image has no
// text to anchor, only itself or a box on it.
var selectorKinds = map[string][]string{KindDocument: {"document", "dom", "code-range"}, KindImage: {"document", "region"}}

// validate checks the selector against the file it is on, as of doc.
func (s CommentSelector) validate(doc Document) error {
	if s.Cell != nil {
		if s.Kind != "code-range" || s.Patch == nil || s.Patch.StartLine != s.Patch.EndLine {
			return fmt.Errorf("%w: a table cell sits beside a code-range on its row's one line", comments.ErrInvalid)
		}
		if err := s.Cell.Validate(); err != nil {
			return err
		}
	}
	switch {
	case s.Version != 1 || !slices.Contains(commentViews, s.View) || len(s.HTML) > 8192:
		return fmt.Errorf("%w: selector version, view %q or html size", comments.ErrInvalid, s.View)
	case !slices.Contains(selectorKinds[doc.Kind], s.Kind):
		return fmt.Errorf("%w: a %s selector cannot anchor on a %s file", comments.ErrInvalid, s.Kind, doc.Kind)
	case (s.Region != nil) != (s.Kind == "region"):
		return fmt.Errorf("%w: a region selector, and only one, carries a region", comments.ErrInvalid)
	case s.Quote != nil && s.Kind != "dom" && s.Kind != "code-range":
		return fmt.Errorf("%w: a %s selector cannot carry a quote", comments.ErrInvalid, s.Kind)
	case s.Quote != nil && s.Patch != nil:
		if err := s.Quote.Within(*s.Patch); err != nil {
			return err
		}
	case s.Quote != nil:
		if err := s.Quote.Validate(); err != nil {
			return err
		}
	}
	switch s.Kind {
	case "dom":
		if s.Anchor == "" {
			return fmt.Errorf("%w: dom selector needs an anchor", comments.ErrInvalid)
		}
	case "region":
		return s.Region.Validate()
	case "code-range":
		return s.validatePatch(doc.Source)
	}
	return nil
}

// validatePatch requires the code range's patch to be exactly the capture of
// its lines from source, from a view that shows source lines.
func (s CommentSelector) validatePatch(source string) error {
	if s.Patch == nil || s.View == "form" {
		return fmt.Errorf("%w: code-range needs source lines from the source, preview or rich view", comments.ErrInvalid)
	}
	capture, err := comments.CapturePatch(source, s.Patch.StartLine, s.Patch.EndLine)
	if err != nil {
		return err
	}
	expected, err := json.Marshal(capture)
	if err != nil {
		return err
	}
	actual, err := json.Marshal(s.Patch)
	if err != nil {
		return err
	}
	if string(expected) != string(actual) {
		return comments.ErrConflict
	}
	return nil
}

// commentHost is pr/ui's project comment host (local tenant, asserted or OS
// user, project scopes) plus what only the editor can check: a new or
// retargeted thread must name an edited file, at its current revision.
type commentHost struct {
	ui.CommentAccess
	store *Store
}

func (h commentHost) Resolve(ctx context.Context, p comments.Principal, t comments.Target) (comments.Target, error) {
	if err := h.Authorize(ctx, p, "create", t); err != nil {
		return t, err
	}
	doc, err := h.store.editedFile(ctx, t.ScopeURI, t.URI)
	if errors.Is(err, errInvalidTarget) {
		return t, fmt.Errorf("%w: %v", comments.ErrInvalid, err)
	}
	if err != nil {
		return t, err
	}
	if t.Revision != doc.Revision {
		return t, comments.ErrConflict
	}
	var selector CommentSelector
	if err := json.Unmarshal(t.Selector, &selector); err != nil {
		return t, fmt.Errorf("%w: selector: %v", comments.ErrInvalid, err)
	}
	if t.Kind != selector.Kind {
		return t, fmt.Errorf("%w: target kind %q differs from selector kind %q", comments.ErrInvalid, t.Kind, selector.Kind)
	}
	if err := selector.validate(doc); err != nil {
		return t, err
	}
	t.Label = filepath.Base(doc.Path)
	return t, nil
}
func (s *Server) OpenComments(ctx context.Context) (*comments.Service, error) {
	if s.CommentService != nil {
		return s.CommentService(ctx)
	}
	db, err := database.Require(ctx, "Editor comments (run gavel serve once to migrate the database)")
	if err != nil {
		return nil, err
	}
	store, err := postgres.New(db)
	if err != nil {
		return nil, err
	}
	if s.Agents == nil {
		return nil, fmt.Errorf("editor comments need an agents bridge")
	}
	delivery, err := agents.Delivery(agents.Options{Bridge: s.Agents, Groups: store, Workdir: ui.CommentWorkdir, Amp: agents.AmpFromEnvironment(), Repositories: ui.CommentRepositories})
	if err != nil {
		return nil, err
	}
	catalog, err := projects.Shared(ctx)
	if err != nil {
		return nil, err
	}
	configured, err := settings.Open(ctx, catalog)
	if err != nil {
		return nil, err
	}
	host := commentHost{store: s.Store}
	locate := func(ctx context.Context, path string) (comments.Location, error) {
		return ui.LocateCommentPath(ctx, catalog, path)
	}
	routes := agents.ProjectRoutes(catalog, configured)
	routes.Session = s.Session
	return comments.New(comments.Options{Store: store, Access: host, Targets: host, Workflows: []comments.Workflow{comments.HumanReview()}, Blobs: store, AttachmentLimit: 10 << 20, Routes: routes, Delivery: delivery, Proposals: ui.CommentProposals{}, Locate: locate})
}

// me serves GET /api/me: the actor this editor authors comments as.
func (s *Server) me(w http.ResponseWriter, r *http.Request) {
	principal, err := commentHost{store: s.Store}.Principal(r.Context())
	respond(w, principal.Actor, err)
}

// fileComments pages every comment on a file, with the caller's read version of each.
func fileComments(ctx context.Context, service *comments.Service, target fileTarget) ([]comments.Comment, map[string]int64, error) {
	return service.TargetThreads(ctx, comments.ReadTarget{ScopeURI: target.Scope, URI: target.URI})
}

func (s *Server) listComments(w http.ResponseWriter, r *http.Request) {
	doc, err := s.Store.commentable(r.PathValue("id"))
	if err != nil {
		respond(w, nil, err)
		return
	}
	target, err := targetOf(r.Context(), doc.Path)
	if err != nil {
		respond(w, nil, err)
		return
	}
	service, err := s.OpenComments(r.Context())
	if err != nil {
		respond(w, nil, err)
		return
	}
	rows, _, err := fileComments(r.Context(), service, target)
	if err != nil {
		respond(w, nil, err)
		return
	}
	anchors, err := service.Anchors(r.Context(), rows)
	if err != nil {
		respond(w, nil, err)
		return
	}
	respond(w, struct {
		Scope    string                            `json:"scope"`
		Target   string                            `json:"target"`
		Revision string                            `json:"revision"`
		Comments []comments.Comment                `json:"comments"`
		Anchors  map[string]comments.PatchLocation `json:"anchors"`
	}{target.Scope, target.URI, doc.Revision, rows, anchors}, nil)
}
