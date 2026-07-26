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
	// A project whose counts failed to load renders its error rather than a
	// misleading 0/0.
	todos := p.Error
	if p.TodoCounts != nil {
		todos = fmt.Sprintf("%d/%d", p.TodoCounts.Open, p.TodoCounts.Total)
	}
	return map[string]any{
		"name":     p.Name,
		"dir":      p.Dir,
		"repos":    strings.Join(p.Repos, ", "),
		"procfile": yesNo(p.HasProcfile),
		"todos":    todos,
		"provider": p.TodoBackend,
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
// as a clicky entity and the status/action endpoints used by the Projects tab.
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

func projectActionRequestSchema() map[string]any {
	return map[string]any{
		"type":     "object",
		"required": []string{"action"},
		"properties": map[string]any{
			"action":  map[string]any{"type": "string", "enum": []string{"lint", "test"}},
			"files":   map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			"options": map[string]any{"type": "object", "additionalProperties": true},
		},
	}
}

func projectCommitQueueRequestSchema() map[string]any {
	return map[string]any{
		"type":     "object",
		"required": []string{"action"},
		"properties": map[string]any{
			"action":  map[string]any{"type": "string", "enum": []string{"commit", "open-pr"}},
			"files":   map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			"options": map[string]any{"type": "object", "additionalProperties": true, "description": "Advanced commit options; not supported for open-pr"},
		},
	}
}

func projectActionSchemaResponseSchema() map[string]any {
	return map[string]any{
		"type":     "object",
		"required": []string{"schemaVersion", "action", "schema", "defaults"},
		"properties": map[string]any{
			"schemaVersion": map[string]any{"type": "integer"},
			"action":        map[string]any{"type": "string", "enum": []string{"commit", "lint", "test"}},
			"schema":        map[string]any{"type": "object", "additionalProperties": true},
			"defaults":      map[string]any{"type": "object", "additionalProperties": true},
		},
	}
}

func projectDiffResponseSchema() map[string]any {
	return map[string]any{
		"type":     "object",
		"required": []string{"path", "diff", "truncated", "binary"},
		"properties": map[string]any{
			"path":      map[string]any{"type": "string"},
			"diff":      map[string]any{"type": "string"},
			"truncated": map[string]any{"type": "boolean"},
			"binary":    map[string]any{"type": "boolean"},
		},
	}
}

func projectActionStatusSchema() map[string]any {
	return map[string]any{
		"type":     "object",
		"required": []string{"running"},
		"properties": map[string]any{
			"action":    map[string]any{"type": "string", "enum": []string{"lint", "test"}},
			"running":   map[string]any{"type": "boolean"},
			"startedAt": map[string]any{"type": "string", "format": "date-time"},
			"endedAt":   map[string]any{"type": "string", "format": "date-time"},
			"exitCode":  map[string]any{"type": "integer"},
			"output":    map[string]any{"type": "string"},
			"error":     map[string]any{"type": "string"},
		},
	}
}

func commitQueueEntrySchema() map[string]any {
	return map[string]any{
		"type":     "object",
		"required": []string{"id", "action", "files", "status"},
		"properties": map[string]any{
			"id":        map[string]any{"type": "string", "description": "Task id, used to cancel this group"},
			"action":    map[string]any{"type": "string", "enum": []string{"commit", "open-pr"}},
			"files":     map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			"status":    map[string]any{"type": "string", "enum": []string{"pending", "running", "success", "warning", "failed", "canceled"}},
			"startedAt": map[string]any{"type": "string", "format": "date-time"},
			"endedAt":   map[string]any{"type": "string", "format": "date-time"},
			"exitCode":  map[string]any{"type": "integer"},
			"output":    map[string]any{"type": "string"},
			"error":     map[string]any{"type": "string"},
		},
	}
}

func commitQueueSchema() map[string]any {
	return map[string]any{
		"type":     "object",
		"required": []string{"running"},
		"properties": map[string]any{
			"runId":   map[string]any{"type": "string"},
			"href":    map[string]any{"type": "string"},
			"running": map[string]any{"type": "boolean"},
			"entries": map[string]any{
				"type":        "array",
				"items":       map[string]any{"$ref": "#/components/schemas/CommitQueueEntry"},
				"description": "Queued commit groups in execution order; one gavel commit each, run one at a time",
			},
		},
	}
}

func projectIgnoreRequestSchema() map[string]any {
	return map[string]any{
		"type":     "object",
		"required": []string{"path", "directory"},
		"properties": map[string]any{
			"path":      map[string]any{"type": "string"},
			"directory": map[string]any{"type": "boolean"},
		},
	}
}

func projectIgnoreResponseSchema() map[string]any {
	return map[string]any{
		"type":     "object",
		"required": []string{"path", "directory", "rule", "added"},
		"properties": map[string]any{
			"path":      map[string]any{"type": "string"},
			"directory": map[string]any{"type": "boolean"},
			"rule":      map[string]any{"type": "string"},
			"added":     map[string]any{"type": "boolean"},
		},
	}
}

func projectStatusSchema() map[string]any {
	return map[string]any{
		"type":     "object",
		"required": []string{"project", "workDir", "branch", "files", "resultsStale", "action", "commitQueue"},
		"properties": map[string]any{
			"project":      map[string]any{"$ref": "#/components/schemas/Project"},
			"workDir":      map[string]any{"type": "string"},
			"branch":       map[string]any{"type": "string"},
			"files":        map[string]any{"type": "array", "items": map[string]any{"$ref": "#/components/schemas/ProjectFileStatus"}},
			"resultsStale": map[string]any{"type": "boolean"},
			"action":       map[string]any{"$ref": "#/components/schemas/ProjectActionStatus"},
			"commitQueue":  map[string]any{"$ref": "#/components/schemas/CommitQueue"},
		},
	}
}

func projectFileStatusSchema() map[string]any {
	return map[string]any{
		"type":     "object",
		"required": []string{"path", "state", "adds", "dels", "testStatus", "lintStatus", "resultsStale"},
		"properties": map[string]any{
			"path":         map[string]any{"type": "string"},
			"previousPath": map[string]any{"type": "string"},
			"state":        map[string]any{"type": "string", "enum": []string{"staged", "unstaged", "both", "untracked", "conflict"}},
			"stagedKind":   map[string]any{"type": "string"},
			"workKind":     map[string]any{"type": "string"},
			"adds":         map[string]any{"type": "integer"},
			"dels":         map[string]any{"type": "integer"},
			"resultsStale": map[string]any{"type": "boolean"},
			"testStatus":   map[string]any{"type": "object"},
			"lintStatus":   map[string]any{"type": "object"},
			"problems":     map[string]any{"type": "array", "items": map[string]any{"type": "object"}},
		},
	}
}

func projectLifecycleOp(id, summary, statusCode, responseSchema string, extra ...map[string]any) map[string]any {
	op := map[string]any{
		"operationId": id,
		"summary":     summary,
		"tags":        []string{"project"},
		"parameters":  nameParam()["parameters"],
		"responses": map[string]any{statusCode: map[string]any{
			"description": "OK",
			"content": map[string]any{"application/json": map[string]any{
				"schema": map[string]any{"$ref": "#/components/schemas/" + responseSchema},
			}},
		}},
	}
	for _, fields := range extra {
		for key, value := range fields {
			op[key] = value
		}
	}
	return op
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
			"/api/projects/{name}/status": map[string]any{
				"get": projectLifecycleOp("projects_status", "Get project working-tree status", "200", "ProjectStatus"),
			},
			"/api/projects/{name}/commit-queue": map[string]any{
				"post": projectLifecycleOp("projects_commit_queue_add", "Queue a commit group; it runs as soon as the project's earlier groups finish", "202", "CommitQueue", map[string]any{
					"requestBody": map[string]any{
						"required": true,
						"content": map[string]any{"application/json": map[string]any{
							"schema": map[string]any{"$ref": "#/components/schemas/ProjectCommitQueueRequest"},
						}},
					},
				}),
			},
			"/api/projects/{name}/commit-queue/{id}": map[string]any{
				"delete": projectLifecycleOp("projects_commit_queue_cancel", "Cancel a queued or running commit group", "200", "CommitQueue", map[string]any{
					"parameters": append(nameParam()["parameters"].([]any), map[string]any{
						"name": "id", "in": "path", "required": true,
						"schema": map[string]any{"type": "string"},
					}),
				}),
			},
			"/api/projects/{name}/actions": map[string]any{
				"post": projectLifecycleOp("projects_action", "Run a project lint or test action (commits are queued instead)", "202", "ProjectActionStatus", map[string]any{
					"requestBody": map[string]any{
						"required": true,
						"content": map[string]any{"application/json": map[string]any{
							"schema": map[string]any{"$ref": "#/components/schemas/ProjectActionRequest"},
						}},
					},
				}),
			},
			"/api/projects/{name}/actions/schema": map[string]any{
				"get": projectLifecycleOp("projects_action_schema", "Get safe advanced commit, lint, or test options", "200", "ProjectActionSchema", map[string]any{
					"parameters": append(nameParam()["parameters"].([]any), map[string]any{
						"name": "action", "in": "query", "required": true,
						"schema": map[string]any{"type": "string", "enum": []string{"commit", "lint", "test"}},
					}),
				}),
			},
			"/api/project-runs/{runId}/api/tests/stream": map[string]any{
				"get": map[string]any{
					"operationId": "project_action_run_stream",
					"summary":     "Stream structured project test or lint run details",
					"tags":        []string{"project"},
					"parameters": []any{map[string]any{
						"name": "runId", "in": "path", "required": true, "schema": map[string]any{"type": "string"},
					}},
					"responses": map[string]any{"200": map[string]any{"description": "Server-sent test-runner snapshots"}},
				},
			},
			"/api/projects/{name}/diff": map[string]any{
				"get": projectLifecycleOp("projects_diff", "Get a working-tree diff for a changed file or folder", "200", "ProjectDiff", map[string]any{
					"parameters": append(nameParam()["parameters"].([]any), map[string]any{
						"name": "path", "in": "query", "required": true,
						"schema": map[string]any{"type": "string"},
					}),
				}),
			},
			"/api/projects/{name}/ignore": map[string]any{
				"post": projectLifecycleOp("projects_ignore", "Add an untracked file or directory to .gitignore", "200", "ProjectIgnoreResponse", map[string]any{
					"requestBody": map[string]any{
						"required": true,
						"content": map[string]any{"application/json": map[string]any{
							"schema": map[string]any{"$ref": "#/components/schemas/ProjectIgnoreRequest"},
						}},
					},
				}),
			},
		},
		"components": map[string]any{"schemas": map[string]any{
			"Project":                   projectSchema(),
			"ProjectFileStatus":         projectFileStatusSchema(),
			"ProjectStatus":             projectStatusSchema(),
			"ProjectActionRequest":      projectActionRequestSchema(),
			"ProjectCommitQueueRequest": projectCommitQueueRequestSchema(),
			"ProjectActionStatus":       projectActionStatusSchema(),
			"ProjectActionSchema":       projectActionSchemaResponseSchema(),
			"CommitQueue":               commitQueueSchema(),
			"CommitQueueEntry":          commitQueueEntrySchema(),
			"ProjectDiff":               projectDiffResponseSchema(),
			"ProjectIgnoreRequest":      projectIgnoreRequestSchema(),
			"ProjectIgnoreResponse":     projectIgnoreResponseSchema(),
		}},
	}
}
