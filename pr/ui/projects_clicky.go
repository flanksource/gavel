package ui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/flanksource/clicky/api"
	"github.com/flanksource/clicky/formatters"
)

// clickyMediaType is the content type that selects the clicky table/detail
// representation of the projects entity (the shape @flanksource/clicky-ui's
// <Clicky/> renders), as opposed to the plain JSON the dashboard consumes.
const clickyMediaType = "application/json+clicky"

// wantsClicky reports whether the caller asked for the clicky representation,
// either via the Accept header or an explicit ?format=clicky override.
func wantsClicky(r *http.Request) bool {
	return strings.Contains(r.Header.Get("Accept"), clickyMediaType) ||
		r.URL.Query().Get("format") == "clicky"
}

// Columns and Row make projectInfo an api.TableProvider so a slice of them
// renders as a clicky table node via api.NewTableFrom.
func (p projectInfo) Columns() []api.ColumnDef {
	return []api.ColumnDef{
		{Name: "name", Label: "Name"},
		{Name: "dir", Label: "Directory"},
		{Name: "repos", Label: "Repos"},
		{Name: "procfile", Label: "Procfile"},
		{Name: "todos", Label: "Todos"},
		{Name: "provider", Label: "Todo Backend"},
	}
}

func (p projectInfo) Row() map[string]any {
	backend := p.TodoBackend
	if p.TodoBackendAuto {
		backend += " (auto)"
	}
	return map[string]any{
		"name":     p.Name,
		"dir":      p.Dir,
		"repos":    strings.Join(p.Repos, ", "),
		"procfile": yesNo(p.HasProcfile),
		"todos":    fmt.Sprintf("%d/%d", p.TodoCounts.Open, p.TodoCounts.Total),
		"provider": backend,
	}
}

func yesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

// writeProjectsClicky renders the projects as the clicky document consumed by
// clicky-ui (a "table" node), reusing the same node converter the clicky-json
// formatter uses so the shape stays in lock-step with the library.
func writeProjectsClicky(w http.ResponseWriter, infos []projectInfo) {
	w.Header().Set("Content-Type", clickyMediaType)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(formatters.NewClickyDocument(api.NewTableFrom(infos))) //nolint:errcheck
}

// handleOpenAPI serves the OpenAPI document that describes the projects surface
// as a clicky entity (x-clicky surfaces + list/get/create/update/delete
// operations), so a clicky-ui EntityExplorer or the clicky API explorer can
// drive the same endpoints the dashboard does.
func (s *Server) handleOpenAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	respondJSON(w, http.StatusOK, projectsOpenAPI())
}

// projectSchema is the JSON Schema for a project create/update body, also the
// shape clicky-ui renders its create/edit form from.
func projectSchema() map[string]any {
	return map[string]any{
		"type":     "object",
		"required": []string{"name", "dir"},
		"properties": map[string]any{
			"name": map[string]any{"type": "string", "description": "Unique workspace name"},
			"dir":  map[string]any{"type": "string", "description": "Absolute path to the local checkout (a leading ~ is expanded)"},
			"repos": map[string]any{
				"type":        "array",
				"items":       map[string]any{"type": "string"},
				"description": "GitHub repos this directory backs (owner/repo)",
			},
			"todoProvider": map[string]any{
				"type":        "string",
				"enum":        []string{"", "grite", "todos"},
				"description": "Pinned todo backend; empty means auto-detect",
			},
		},
	}
}

// clickyOp builds one OpenAPI operation carrying its x-clicky entity metadata.
// idParam is empty for collection-scoped operations. extra is merged in for
// per-operation additions (parameters, requestBody).
func clickyOp(id, summary, verb, scope, idParam string, extra ...map[string]any) map[string]any {
	xclicky := map[string]any{"surface": "projects", "verb": verb, "scope": scope}
	if idParam != "" {
		xclicky["idParam"] = idParam
	}
	op := map[string]any{
		"operationId": id,
		"summary":     summary,
		"tags":        []string{"project"},
		"x-clicky":    xclicky,
		"responses":   map[string]any{"200": map[string]any{"description": "OK"}},
	}
	for _, e := range extra {
		for k, v := range e {
			op[k] = v
		}
	}
	return op
}

func nameParam() map[string]any {
	return map[string]any{"parameters": []any{map[string]any{
		"name":     "name",
		"in":       "path",
		"required": true,
		"schema":   map[string]any{"type": "string"},
	}}}
}

func jsonBody() map[string]any {
	return map[string]any{"requestBody": map[string]any{
		"required": true,
		"content":  map[string]any{"application/json": map[string]any{"schema": projectSchema()}},
	}}
}

func projectsOpenAPI() map[string]any {
	return map[string]any{
		"openapi": "3.0.0",
		"info":    map[string]any{"title": "Gavel PR Dashboard", "version": "1.0.0"},
		"x-clicky": map[string]any{"surfaces": []any{map[string]any{
			"key":         "projects",
			"entity":      "project",
			"title":       "Projects",
			"description": "Local workspace directories tracked by the dashboard.",
		}}},
		"paths": map[string]any{
			"/api/projects": map[string]any{
				"get":  clickyOp("projects_list", "List projects", "list", "collection", ""),
				"post": clickyOp("projects_create", "Create project", "create", "collection", "", jsonBody()),
			},
			"/api/projects/{name}": map[string]any{
				"get":    clickyOp("projects_get", "Get project", "get", "entity", "name", nameParam()),
				"put":    clickyOp("projects_update", "Update project", "update", "entity", "name", nameParam(), jsonBody()),
				"delete": clickyOp("projects_delete", "Delete project", "delete", "entity", "name", nameParam()),
			},
		},
		"components": map[string]any{"schemas": map[string]any{"Project": projectSchema()}},
	}
}
