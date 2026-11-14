package clicky

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/flanksource/clicky/route"
	"github.com/flanksource/clicky/rpc"
	"github.com/flanksource/gavel/comments"
)

// RegisterDeliveryRoutes mounts agent delivery and comment routes beside the
// generated comments operations. A host without comments.Options.Delivery
// answers them 501.
func RegisterDeliveryRoutes(router *route.Router, resolve Resolver) {
	router.Raw("GET /api/v1/comments/agent-sessions", serve(resolve, agentSessions), route.Meta{Entity: "comments", Verb: "agent-sessions", ReadOnly: true})
	router.Raw("POST /api/v1/comments/{id}/deliver", serve(resolve, deliver), route.Meta{Entity: "comments", Verb: "deliver", IDParam: "id"})
	router.Raw("POST /api/v1/comments/{id}/deliver/preview", serve(resolve, previewDelivery), route.Meta{Entity: "comments", Verb: "deliver-preview", IDParam: "id", ReadOnly: true})
	router.Raw("GET /api/v1/comments/routes", serve(resolve, listRoutes), route.Meta{Entity: "comments", Verb: "routes", ReadOnly: true})
	router.Raw("PUT /api/v1/comments/routes", serve(resolve, saveRoute), route.Meta{Entity: "comments", Verb: "set-route"})
	router.Raw("DELETE /api/v1/comments/routes/{routeId}", serve(resolve, deleteRoute), route.Meta{Entity: "comments", Verb: "delete-route", IDParam: "routeId"})
}

// handlerFunc answers one request with a JSON body, or nil for 204.
type handlerFunc func(ctx context.Context, service *comments.Service, r *http.Request) (any, error)

func serve(resolve Resolver, handle handlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := rpc.ContextWithRequest(r.Context(), r)
		service, err := resolve(ctx)
		if err != nil {
			writeError(w, err)
			return
		}
		out, err := handle(ctx, service, r)
		if err != nil {
			writeError(w, err)
			return
		}
		if out == nil {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(out); err != nil {
			return
		}
	})
}

func agentSessions(ctx context.Context, service *comments.Service, r *http.Request) (any, error) {
	q := r.URL.Query()
	return service.AgentSessions(ctx, comments.SessionQuery{ScopeURI: q.Get("scope"), All: q.Get("all") == "1"})
}

func deliver(ctx context.Context, service *comments.Service, r *http.Request) (any, error) {
	var req comments.DeliverRequest
	if err := decode(ctx, nil, &req); err != nil {
		return nil, err
	}
	req.CommentID = r.PathValue("id")
	return service.Deliver(ctx, req)
}

func previewDelivery(ctx context.Context, service *comments.Service, r *http.Request) (any, error) {
	var req struct {
		Target comments.DeliverTarget `json:"target"`
	}
	if err := decode(ctx, nil, &req); err != nil {
		return nil, err
	}
	return service.PreviewDelivery(ctx, r.PathValue("id"), req.Target)
}

func listRoutes(ctx context.Context, service *comments.Service, r *http.Request) (any, error) {
	return service.Routes(ctx, r.URL.Query().Get("scope"))
}

func saveRoute(ctx context.Context, service *comments.Service, _ *http.Request) (any, error) {
	var route comments.Route
	if err := decode(ctx, nil, &route); err != nil {
		return nil, err
	}
	return service.SaveRoute(ctx, route)
}

func deleteRoute(ctx context.Context, service *comments.Service, r *http.Request) (any, error) {
	return nil, service.DeleteRoute(ctx, r.PathValue("routeId"))
}
