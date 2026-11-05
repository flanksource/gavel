package launch

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/flanksource/gavel/comments"
	api "github.com/flanksource/gavel/comments/clicky"
	"github.com/flanksource/gavel/comments/postgres"
	"gorm.io/gorm"
)

type commentActorKey struct{}
type localCommentHost struct{ scope string }

func (h localCommentHost) Principal(ctx context.Context) (comments.Principal, error) {
	actor, ok := ctx.Value(commentActorKey{}).(comments.Actor)
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

// CommentsHandler embeds comments for a local workspace. It never provisions storage.
// Actor assertions are accepted only from loopback callers such as a local UI host.
func CommentsHandler(db *gorm.DB, root string) (http.Handler, error) {
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
	service, err := comments.New(comments.Options{Store: store, Access: host, Targets: host, Workflows: []comments.Workflow{comments.HumanReview()}, Blobs: store, AttachmentLimit: 32 << 20})
	if err != nil {
		return nil, err
	}
	handler := api.Handler(func(context.Context) (*comments.Service, error) { return service, nil })
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		address, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil || !net.ParseIP(address).IsLoopback() {
			http.Error(w, "workspace comments require a loopback caller", http.StatusForbidden)
			return
		}
		actor := comments.Actor{Kind: "user", ID: fmt.Sprintf("local:%d", os.Getuid()), Label: "You"}
		if value := r.Header.Get("X-Gavel-Comment-Actor"); value != "" {
			data, err := base64.RawURLEncoding.DecodeString(value)
			if err != nil || json.Unmarshal(data, &actor) != nil || actor.ID == "" || (actor.Kind != "user" && actor.Kind != "agent") {
				http.Error(w, "invalid local comment actor", 400)
				return
			}
		}
		handler.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), commentActorKey{}, actor)))
	}), nil
}
