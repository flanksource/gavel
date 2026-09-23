package ui

import (
	"context"
	"fmt"
	"net/http"

	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/route"
	"github.com/flanksource/clicky/rpc"
	"github.com/flanksource/gavel/internal/database"
	"github.com/spf13/cobra"
)

// entityCommandRoot registers every dashboard entity and generates the command
// tree they share. The REST routes, the generated OpenAPI document and the
// assistant's tools are all derived from this one tree, so an operation reaches
// all three or none.
func entityCommandRoot() (*cobra.Command, error) {
	if err := registerTodoEntity(); err != nil {
		return nil, fmt.Errorf("register the todos entity: %w", err)
	}
	registerProjectEntity()
	root := &cobra.Command{Use: "gavel"}
	clicky.GenerateCLI(root)
	return root, nil
}

// registerEntityRoutes mounts the generated entity surface:
//
//	/api/v1/<entity>...     every entity operation and action
//	GET /api/entities       the catalog the selection toolbar renders from
//	GET /api/v1/openapi.json the generated document the chat's context picker
//	                        lists TODOs and projects from
//	/api/chat, /api/chat/   the assistant, whose tools are the same operations
//
// The generated document is served beside /api/openapi.json rather than merged
// into it: that hand-built document describes the /api/projects CRUD routes, and
// one document carrying two project surfaces would leave every client to guess
// which one it meant.
func (s *Server) registerEntityRoutes(mux *http.ServeMux) {
	root, err := entityCommandRoot()
	if err != nil {
		// A failed registration means the dashboard would silently serve a
		// toolbar and an assistant with no operations behind them.
		panic("ui: " + err.Error())
	}
	server := rpc.NewSwaggerServer(&rpc.ServeConfig{
		Title:      "gavel",
		SkipHealth: true,
		Executor:   &rpc.ExecutorConfig{Enabled: true, PathPrefix: "/api/v1"},
	}, root, nil)
	server.RegisterExecutionRoutes(route.NewRouter(mux))
	mux.HandleFunc("GET /api/entities", server.HandleEntities)
	mux.HandleFunc("GET /api/v1/openapi.json", server.HandleOpenAPIJSON)
	pool, err := database.Require(context.Background(), "Gavel chat")
	if err != nil {
		unavailable := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, fmt.Sprintf("Gavel chat requires the TODO database: %v", err), http.StatusServiceUnavailable)
		})
		mux.Handle("/api/chat", unavailable)
		mux.Handle("/api/chat/", unavailable)
		return
	}
	db, err := captaindb.Use(pool)
	if err != nil {
		panic("ui: opening the chat database: " + err.Error())
	}
	chat, err := newGavelChatServer(root, s.todoWorkDir(), db)
	if err != nil {
		panic("ui: registering the chat service: " + err.Error())
	}
	mux.Handle("/api/chat", chat.Handler())
	mux.Handle("/api/chat/", chat.Handler())
}
