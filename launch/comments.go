package launch

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/flanksource/gavel/comments"
	"github.com/flanksource/gavel/comments/agents"
	api "github.com/flanksource/gavel/comments/clicky"
	"github.com/flanksource/gavel/comments/postgres"
	"gorm.io/gorm"
)

type localCommentHost struct{ scope string }

func (h localCommentHost) Principal(ctx context.Context) (comments.Principal, error) {
	actor, ok := api.ActorFrom(ctx)
	if !ok {
		return comments.Principal{}, comments.ErrForbidden
	}
	return comments.Principal{TenantID: fmt.Sprintf("local:%d", os.Getuid()), Actor: actor}, nil
}
func (h localCommentHost) Authorize(_ context.Context, _ comments.Principal, _ string, t comments.Target) error {
	if t.ScopeURI != h.scope || t.URI != "" && !strings.HasPrefix(t.URI, h.scope+"/") {
		return comments.ErrForbidden
	}
	return nil
}
func (h localCommentHost) Resolve(ctx context.Context, p comments.Principal, t comments.Target) (comments.Target, error) {
	if t.URI == "" || t.Kind == "" {
		return t, comments.ErrInvalid
	}
	return t, h.Authorize(ctx, p, "target", t)
}

// CommentsHandler embeds comments for a local workspace, delivering threads to
// agents through bridge. It never provisions storage.
// Actor assertions are accepted only from loopback callers such as a local UI host.
func CommentsHandler(db *gorm.DB, root string, bridge agents.Bridge) (http.Handler, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	digest := sha256.Sum256([]byte(root))
	host := localCommentHost{scope: "gavel://workspace/" + hex.EncodeToString(digest[:])}
	store, err := postgres.New(db)
	if err != nil {
		return nil, err
	}
	delivery, err := agents.Delivery(agents.Options{Bridge: bridge, Groups: store, Workdir: func(context.Context, comments.Target) (string, error) { return root, nil }})
	if err != nil {
		return nil, err
	}
	service, err := comments.New(comments.Options{Store: store, Access: host, Targets: host, Workflows: []comments.Workflow{comments.HumanReview()}, Blobs: store, AttachmentLimit: 32 << 20, Routes: store, Delivery: delivery})
	if err != nil {
		return nil, err
	}
	handler := api.Handler(func(context.Context) (*comments.Service, error) { return service, nil })
	you, err := api.LocalUser()
	if err != nil {
		return nil, err
	}
	return api.LocalActor(handler, you), nil
}
