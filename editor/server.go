package editor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/flanksource/clicky/route"
	"github.com/flanksource/clicky/sse"
	clickytask "github.com/flanksource/clicky/task"
	"github.com/flanksource/gavel/comments"
	"github.com/flanksource/gavel/comments/agents"
	"github.com/flanksource/gavel/internal/fileread"
	"github.com/flanksource/gavel/linters"
	jsonlint "github.com/flanksource/gavel/linters/json"
	"github.com/flanksource/gavel/testrunner/parsers"
	testui "github.com/flanksource/gavel/testrunner/ui"
	"github.com/flanksource/gavel/verify"
)

// LintRun is the problems shown for the edited files: the lint violations and
// failed tests the workspaces' recorded runs hold for them (see Reload), the
// editor's own lint run in progress, and gavel runs in progress elsewhere.
type LintRun struct {
	Running bool                    `json:"running"`
	Results []*linters.LinterResult `json:"results"`
	Tests   []parsers.Test          `json:"tests"`
	Error   string                  `json:"error,omitempty"`
	// Revisions are the files' revisions when the problems were loaded.
	Revisions map[string]string `json:"revisions"`
	// StaleFiles are the ids of files changed since the run that checked them.
	StaleFiles []string `json:"staleFiles"`
	Stale      bool     `json:"stale"`
	// Recorded is when the newest run the problems come from ended.
	Recorded *time.Time `json:"recorded,omitempty"`
	// Active are gavel test/lint processes running against the files' workspaces.
	Active []ActiveRun `json:"active"`
}
type Server struct {
	Comments       http.Handler
	CommentService func(context.Context) (*comments.Service, error)
	Store          *Store
	Schema         string
	Assets         fs.FS
	// DevProxy, when set, serves the page and its modules from the editor's
	// Vite dev server (`gavel edit --dev`) instead of Assets; /api/* stays here.
	DevProxy http.Handler
	Context  context.Context
	// Agents reaches the live agent sessions threads are sent to.
	Agents agents.Bridge
	// Version is the gavel version recorded on the lint runs the editor saves.
	Version string
	mu      sync.Mutex
	run     LintRun
	// staleAtLoad marks files changed on disk after the run that checked them.
	staleAtLoad map[string]bool
	stateRoots  []string
	updated     sse.Notifier
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/files", func(w http.ResponseWriter, r *http.Request) { respond(w, s.Store.Files(), nil) })
	mux.HandleFunc("GET /api/me", s.me)
	mux.HandleFunc("GET /api/file/{id}/comments", s.listComments)
	mux.HandleFunc("POST /api/file/{id}/comments/read", s.markFileCommentsRead)
	mux.HandleFunc("GET /api/comments", s.commentSummary)
	mux.HandleFunc("POST /api/projects", s.createProject)
	mux.HandleFunc("GET /api/file/{id}/agent-run/context", s.agentRunContextOf)
	mux.HandleFunc("GET /api/file/{id}/agent", s.pageAgentOf)
	mux.HandleFunc("GET /api/file/{id}/agent-run/permissions", s.permissionCatalog)
	mux.Handle("/api/captain/prompt/runs/", captainPromptRuns())
	catalog := runtimeCatalog()
	mux.Handle("GET /api/chat/models", catalog)
	mux.Handle("GET /api/chat/runtimes", catalog)
	// Captain in-process: a sent thread's session is read and followed over SSE
	// (GET /{id}?follow=1) here, with no captain serve.
	mux.Handle("/api/captain/sessions/", captainSessions())
	if s.Comments != nil {
		mux.Handle("/api/v1/comments", s.Comments)
		mux.Handle("/api/v1/comments/", s.Comments)
	}
	mux.HandleFunc("GET /api/file/{id}", func(w http.ResponseWriter, r *http.Request) {
		doc, err := s.Store.Read(r.PathValue("id"))
		respond(w, doc, err)
	})
	mux.HandleFunc("GET /api/file/{id}/image", s.image)
	mux.HandleFunc("PUT /api/file/{id}", s.write)
	mux.HandleFunc("POST /api/schema/{id}", s.schema)
	mux.HandleFunc("POST /api/lint", s.startLint)
	mux.HandleFunc("GET /api/lint", s.lintStatus)
	mux.HandleFunc("GET /api/lint/stream", s.lintStream)
	mux.HandleFunc("POST /api/lint/ignore", s.ignoreLint)
	// The task-manager API behind the navbar indicator: lint runs as a clicky
	// task group, so its progress shows up here alongside any other task.
	clickytask.RegisterHandlers(route.NewRouter(mux), "/api/v1")
	if s.DevProxy != nil {
		// Vite would answer an unrouted /api/ path 200 with index.html, which the
		// UI can only report as invalid JSON, so it is a JSON 404 here instead.
		mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "no API route for " + r.Method + " " + r.URL.Path})
		})
		mux.Handle("/", s.DevProxy)
	} else {
		mux.Handle("/", http.FileServerFS(s.Assets))
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if r.Method != "GET" && r.Method != "HEAD" {
			if origin := r.Header.Get("Origin"); origin != "" && origin != "http://"+r.Host {
				http.Error(w, "cross-origin request rejected", http.StatusForbidden)
				return
			}
			attachment := strings.HasPrefix(r.URL.Path, "/api/v1/comments/") && strings.HasSuffix(r.URL.Path, "/attachments") && r.Method == "POST"
			if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") && !attachment {
				http.Error(w, "application/json required", http.StatusUnsupportedMediaType)
				return
			}
		}
		mux.ServeHTTP(w, r)
	})
}
func decode(w http.ResponseWriter, r *http.Request, target any) bool {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return false
	}
	return true
}
func respond(w http.ResponseWriter, value any, err error) {
	var noProject NoProjectError
	if errors.As(err, &noProject) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": noProject.Error(), "code": "no_project", "suggestion": noProject.Suggestion})
		return
	}
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, os.ErrNotExist) {
			status = http.StatusNotFound
		}
		if errors.Is(err, ErrConflict) {
			status = http.StatusConflict
		}
		http.Error(w, err.Error(), status)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(value); err != nil {
		return
	}
}
func (s *Server) write(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Source   string `json:"source"`
		Revision string `json:"revision"`
	}
	if !decode(w, r, &input) {
		return
	}
	doc, err := s.Store.Write(Document{File: File{ID: r.PathValue("id")}, Source: input.Source, Revision: input.Revision})
	respond(w, doc, err)
}

// image serves GET /api/file/{id}/image?path=<rel>: the image entry itself, or
// an image relative to a document, e.g. one its markdown embeds.
func (s *Server) image(w http.ResponseWriter, r *http.Request) {
	name, data, err := s.Store.Image(r.PathValue("id"), r.URL.Query().Get("path"))
	if err != nil {
		respond(w, nil, err)
		return
	}
	_ = fileread.WriteImage(w, name, data)
}
func (s *Server) schema(w http.ResponseWriter, r *http.Request) {
	doc, err := s.Store.Read(r.PathValue("id"))
	if err != nil {
		respond(w, nil, err)
		return
	}
	var input struct {
		Source string `json:"source"`
	}
	if !decode(w, r, &input) {
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	override := s.Schema
	if override == "" {
		cfg, err := verify.LoadGavelConfig(filepath.Dir(doc.Path))
		if err != nil {
			respond(w, nil, err)
			return
		}
		if configured := cfg.LintOptionsFor("json").Schema; configured != nil {
			override = *configured
		}
	}
	schema, err := jsonlint.EditorSchema(ctx, []byte(input.Source), doc.Path, override)
	respond(w, schema, err)
}
func (s *Server) startLint(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Files []string `json:"files"`
	}
	if !decode(w, r, &input) {
		return
	}
	if len(input.Files) == 0 {
		http.Error(w, "select files to lint", http.StatusBadRequest)
		return
	}
	canonical := s.Store.Paths()
	paths := []string{}
	for _, id := range input.Files {
		if _, err := s.Store.Read(id); err != nil {
			respond(w, nil, err)
			return
		}
		paths = append(paths, canonical[id])
	}
	s.mu.Lock()
	if s.run.Running {
		s.mu.Unlock()
		http.Error(w, "lint is already running", http.StatusConflict)
		return
	}
	s.run.Running, s.run.Error = true, ""
	s.mu.Unlock()
	s.updated.Notify()
	go func() {
		err := s.lintFiles(paths)
		s.mu.Lock()
		s.run.Running = false
		if err != nil {
			s.run.Error = err.Error()
		}
		s.mu.Unlock()
		s.updated.Notify()
	}()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	respond(w, s.lintState(), nil)
}

// ignoreLint saves an ignore rule to the .gavel.yaml of a work dir from the
// current run, then drops the violations it matches from that run. Only a run's
// own work dirs are accepted, so a request can never write config elsewhere.
func (s *Server) ignoreLint(w http.ResponseWriter, r *http.Request) {
	var input testui.IgnoreRequest
	if !decode(w, r, &input) {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if !slices.ContainsFunc(s.run.Results, func(result *linters.LinterResult) bool {
		return result != nil && result.WorkDir != "" && result.WorkDir == input.WorkDir
	}) {
		http.Error(w, fmt.Sprintf("work_dir %q is not part of the current lint run", input.WorkDir), http.StatusBadRequest)
		return
	}
	rule := verify.LintIgnoreRule{Source: input.Source, Rule: input.Rule, File: input.File, Line: input.Line, Fingerprint: input.Fingerprint}
	count, err := verify.SaveLintIgnore(input.WorkDir, rule)
	if err != nil {
		http.Error(w, fmt.Sprintf("save %s/.gavel.yaml: %v", input.WorkDir, err), http.StatusBadRequest)
		return
	}
	filtered := linters.FilterIgnoredViolations(s.run.Results, []verify.LintIgnoreRule{rule})
	s.updated.Notify()
	respond(w, testui.IgnoreResponse{RuleCount: count, Filtered: filtered}, nil)
}

func (s *Server) lintStatus(w http.ResponseWriter, r *http.Request) {
	respond(w, s.lintState(), nil)
}
