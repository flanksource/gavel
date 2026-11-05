package clicky

import (
	"context"
	"encoding/json"
	"net/http"

	framework "github.com/flanksource/clicky"
	"github.com/flanksource/clicky/route"
	"github.com/flanksource/clicky/rpc"
	"github.com/flanksource/gavel/comments"
	"github.com/spf13/cobra"
)

type resolverKey struct{}

// Handler exposes registered comment operations with an isolated host resolver.
// Register must have been called before generating the command catalog.
func Handler(resolve Resolver) http.Handler {
	if resolve == nil {
		panic("comments HTTP handler requires a resolver")
	}
	root := &cobra.Command{Use: "comments-host"}
	framework.GenerateCLI(root)
	if command, _, err := root.Find([]string{"comments"}); err != nil || command == root {
		panic("register comments before constructing its HTTP handler")
	}
	for _, command := range root.Commands() {
		if command.Name() != "comments" {
			root.RemoveCommand(command)
		}
	}
	router := route.NewRouter(nil)
	server := rpc.NewSwaggerServer(&rpc.ServeConfig{Executor: &rpc.ExecutorConfig{Enabled: true, PathPrefix: "/api/v1"}}, root, nil)
	server.RegisterExecutionRoutes(router)
	router.Raw("GET /api/v1/comments/events", Stream(resolve), route.Meta{Entity: "comments", Verb: "events", ReadOnly: true})
	router.Raw("POST /api/v1/comments/{id}/attachments", Attachments(resolve), route.Meta{Entity: "comments", Verb: "attach"})
	router.Raw("GET /api/v1/comments/{id}/attachments/{attachment}", Attachments(resolve), route.Meta{Entity: "comments", Verb: "download", ReadOnly: true})
	router.Raw("POST /api/v1/comments/batch", Batch(resolve), route.Meta{Entity: "comments", Verb: "batch"})
	router.MountGenerated("GET /api/v1/comments/openapi.json", OpenAPI(http.HandlerFunc(server.HandleOpenAPIJSON), "/api/v1"))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		router.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), resolverKey{}, resolve)))
	})
}

func Batch(resolve Resolver) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var requests []comments.ActionRequest
		ctx := rpc.ContextWithRequest(r.Context(), r)
		if err := decode(ctx, nil, &requests); err != nil {
			writeError(w, err)
			return
		}
		service, err := resolve(ctx)
		if err != nil {
			writeError(w, err)
			return
		}
		out, err := service.Batch(ctx, requests)
		if err != nil {
			writeError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(out); err != nil {
			return
		}
	})
}
