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
)

type CommentSelector struct {
	Kind     string                `json:"kind"`
	Version  int                   `json:"version"`
	Anchor   string                `json:"anchor,omitempty"`
	View     string                `json:"view"`
	Patch    *comments.PatchAnchor `json:"patch,omitempty"`
	Locators json.RawMessage       `json:"locators,omitempty"`
	HTML     string                `json:"html,omitempty"`
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
	doc, err := h.store.editedFile(t.ScopeURI, t.URI)
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
		return t, comments.ErrInvalid
	}
	if selector.Version != 1 || !slices.Contains([]string{"preview", "source", "rich", "form"}, selector.View) || len(selector.HTML) > 8192 {
		return t, comments.ErrInvalid
	}
	switch selector.Kind {
	case "document":
	case "dom":
		if selector.Anchor == "" {
			return t, comments.ErrInvalid
		}
	case "code-range":
		if selector.Patch == nil || (selector.View != "source" && selector.View != "preview") {
			return t, comments.ErrInvalid
		}
		capture, err := comments.CapturePatch(doc.Source, selector.Patch.StartLine, selector.Patch.EndLine)
		if err != nil {
			return t, err
		}
		expected, _ := json.Marshal(capture)
		actual, _ := json.Marshal(selector.Patch)
		if string(expected) != string(actual) {
			return t, comments.ErrConflict
		}
	default:
		return t, comments.ErrInvalid
	}
	if t.Kind != selector.Kind {
		return t, comments.ErrInvalid
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
	delivery, err := agents.Delivery(agents.Options{Bridge: s.Agents, Groups: store, Workdir: ui.CommentWorkdir})
	if err != nil {
		return nil, err
	}
	host := commentHost{store: s.Store}
	return comments.New(comments.Options{Store: store, Access: host, Targets: host, Workflows: []comments.Workflow{comments.HumanReview()}, Blobs: store, AttachmentLimit: 10 << 20, Routes: configRoutes{db: store, files: s.Store}, Delivery: delivery})
}

// me serves GET /api/me: the actor this editor authors comments as.
func (s *Server) me(w http.ResponseWriter, r *http.Request) {
	principal, err := commentHost{store: s.Store}.Principal(r.Context())
	respond(w, principal.Actor, err)
}

// fileComments pages every comment on a file, with the caller's read version of each.
func fileComments(ctx context.Context, service *comments.Service, target fileTarget) ([]comments.Comment, map[string]int64, error) {
	rows := []comments.Comment{}
	reads := map[string]int64{}
	for offset := 0; ; {
		page, err := service.Page(ctx, comments.Filter{ScopeURI: target.Scope, TargetURI: target.URI, Limit: 200, Offset: offset})
		if err != nil {
			return nil, nil, err
		}
		rows = append(rows, page.Items...)
		for _, read := range page.Read {
			reads[read.CommentID] = read.Version
		}
		offset += len(page.Items)
		if int64(offset) >= page.Total {
			return rows, reads, nil
		}
		if len(page.Items) == 0 {
			return nil, nil, fmt.Errorf("comments pagination stopped before total for %s", target.URI)
		}
	}
}

func (s *Server) listComments(w http.ResponseWriter, r *http.Request) {
	doc, err := s.Store.Read(r.PathValue("id"))
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
	anchors := map[string]comments.PatchLocation{}
	for _, c := range rows {
		var selector CommentSelector
		if err := json.Unmarshal(c.Target.Selector, &selector); err != nil {
			respond(w, nil, err)
			return
		}
		if selector.Patch != nil {
			anchors[c.ID] = comments.ResolvePatch(doc.Source, *selector.Patch)
		}
	}
	respond(w, struct {
		Scope    string                            `json:"scope"`
		Target   string                            `json:"target"`
		Revision string                            `json:"revision"`
		Comments []comments.Comment                `json:"comments"`
		Anchors  map[string]comments.PatchLocation `json:"anchors"`
	}{target.Scope, target.URI, doc.Revision, rows, anchors}, nil)
}
