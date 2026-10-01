package ui

import (
	"context"
	"fmt"
	"net/http"
	"sync"

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
func (s *Server) registerEntityRoutes(router *route.Router) {
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
	server.RegisterExecutionRoutes(router)
	router.MountGenerated("GET /api/entities", http.HandlerFunc(server.HandleEntities))
	router.MountGenerated("GET /api/v1/openapi.json", http.HandlerFunc(server.HandleOpenAPIJSON))
	chat := &lazyChatHandler{root: root, workDir: s.todoWorkDir()}
	router.MountGenerated("/api/chat", chat)
	router.MountGenerated("/api/chat/", chat)
}

// lazyChatHandler opens the chat service on its first request rather than when
// the mux is built: `gavel serve` opens the process database (with migrations)
// after the handler exists, and opening it here first would pin it unmigrated.
// A failed open is not cached, so chat recovers once the database is reachable.
type lazyChatHandler struct {
	root    *cobra.Command
	workDir string
	mu      sync.Mutex
	handler http.Handler
}

func (h *lazyChatHandler) ServeHTTP(w http.ResponseWriter, request *http.Request) {
	handler, err := h.resolve(request.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf("Gavel chat requires the TODO database: %v", err), http.StatusServiceUnavailable)
		return
	}
	handler.ServeHTTP(w, request)
}

func (h *lazyChatHandler) resolve(ctx context.Context) (http.Handler, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.handler != nil {
		return h.handler, nil
	}
	pool, err := database.Require(ctx, "Gavel chat")
	if err != nil {
		return nil, err
	}
	db, err := captaindb.Use(pool)
	if err != nil {
		return nil, fmt.Errorf("open the chat database: %w", err)
	}
	chat, err := newGavelChatServer(h.root, h.workDir, db)
	if err != nil {
		return nil, err
	}
	h.handler = chat.Handler()
	return h.handler, nil
}
