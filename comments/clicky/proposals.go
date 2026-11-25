package clicky

import (
	"context"
	"net/http"

	"github.com/flanksource/clicky/route"
	"github.com/flanksource/gavel/comments"
)

// RegisterProposalRoutes mounts the patch a reply proposes beside the generated
// comments operations. Proposals live in the host's git repository, so a host
// without one answers 400.
func RegisterProposalRoutes(router *route.Router, resolve Resolver) {
	router.Raw("GET /api/v1/comments/{id}/proposals/{replyId}/diff", serve(resolve, proposalDiff), route.Meta{Entity: "comments", Verb: "proposal-diff", IDParam: "id", ReadOnly: true})
}

func proposalDiff(ctx context.Context, service *comments.Service, r *http.Request) (any, error) {
	return service.ProposalDiff(ctx, r.PathValue("id"), r.PathValue("replyId"))
}
