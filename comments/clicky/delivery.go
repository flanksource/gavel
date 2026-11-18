package clicky

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/flanksource/clicky/route"
	"github.com/flanksource/clicky/rpc"
	"github.com/flanksource/gavel/comments"
)

// RegisterDeliveryRoutes mounts agent delivery beside the generated comments
// operations. A host without comments.Options.Delivery answers them 501.
// Automatic delivery is configured in the host's settings (comments.agents),
// read and saved per target through /api/v1/comments/agent; a host whose routes
// do not implement comments.CommentAgents answers it 501.
func RegisterDeliveryRoutes(router *route.Router, resolve Resolver) {
	router.Raw("GET /api/v1/comments/agent-sessions", serve(resolve, agentSessions), route.Meta{Entity: "comments", Verb: "agent-sessions", ReadOnly: true})
	router.Raw("GET /api/v1/comments/mentionables", serve(resolve, mentionables), route.Meta{Entity: "comments", Verb: "mentionables", ReadOnly: true})
	router.Raw("GET /api/v1/comments/agent", serve(resolve, commentAgent), route.Meta{Entity: "comments", Verb: "agent", ReadOnly: true})
	router.Raw("PUT /api/v1/comments/agent", serve(resolve, saveCommentAgent), route.Meta{Entity: "comments", Verb: "save-agent"})
	router.Raw("POST /api/v1/comments/deliver", serve(resolve, deliverBatch), route.Meta{Entity: "comments", Verb: "deliver-batch"})
	router.Raw("POST /api/v1/comments/delivery/preview", serve(resolve, previewCreateDelivery), route.Meta{Entity: "comments", Verb: "create-delivery-preview", ReadOnly: true})
	router.Raw("POST /api/v1/comments/{id}/deliver", serve(resolve, deliver), route.Meta{Entity: "comments", Verb: "deliver", IDParam: "id"})
	router.Raw("POST /api/v1/comments/{id}/deliver/preview", serve(resolve, previewDelivery), route.Meta{Entity: "comments", Verb: "deliver-preview", IDParam: "id", ReadOnly: true})
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
	return service.AgentSessions(ctx, comments.SessionQuery{ScopeURI: q.Get("scope"), GroupKey: q.Get("group"), All: q.Get("all") == "1"})
}

func mentionables(ctx context.Context, service *comments.Service, r *http.Request) (any, error) {
	q := r.URL.Query()
	return service.Mentionables(ctx, comments.Target{ScopeURI: q.Get("scope"), URI: q.Get("target")})
}

func commentAgent(ctx context.Context, service *comments.Service, r *http.Request) (any, error) {
	q := r.URL.Query()
	return service.CommentAgent(ctx, comments.Target{ScopeURI: q.Get("scope"), URI: q.Get("uri")})
}

func saveCommentAgent(ctx context.Context, service *comments.Service, _ *http.Request) (any, error) {
	var req comments.SaveCommentAgentRequest
	if err := decode(ctx, bodyLimit, &req); err != nil {
		return nil, err
	}
	return service.SaveCommentAgent(ctx, req)
}

func deliver(ctx context.Context, service *comments.Service, r *http.Request) (any, error) {
	var req comments.DeliverRequest
	if err := decode(ctx, bodyLimit, &req); err != nil {
		return nil, err
	}
	req.CommentID = r.PathValue("id")
	return service.Deliver(ctx, req)
}

func deliverBatch(ctx context.Context, service *comments.Service, _ *http.Request) (any, error) {
	var req comments.DeliverBatchRequest
	if err := decode(ctx, bodyLimit, &req); err != nil {
		return nil, err
	}
	return service.DeliverBatch(ctx, req)
}

func previewCreateDelivery(ctx context.Context, service *comments.Service, _ *http.Request) (any, error) {
	var draft comments.CreateDeliveryDraft
	if err := decode(ctx, bodyLimit, &draft); err != nil {
		return nil, err
	}
	return service.PreviewCreateDelivery(ctx, draft)
}

func previewDelivery(ctx context.Context, service *comments.Service, r *http.Request) (any, error) {
	var req struct {
		Target comments.DeliverTarget `json:"target"`
	}
	if err := decode(ctx, bodyLimit, &req); err != nil {
		return nil, err
	}
	return service.PreviewDelivery(ctx, r.PathValue("id"), req.Target)
}
