package clicky

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/flanksource/gavel/comments"
)

// ActorHeader carries a base64url JSON comments.Actor that a local caller acts as,
// e.g. an agent session replying to a thread it was handed.
const ActorHeader = "X-Gavel-Comment-Actor"

type actorKey struct{}

// WithActor sets the actor a local comment host acts as for ctx.
func WithActor(ctx context.Context, actor comments.Actor) context.Context {
	return context.WithValue(ctx, actorKey{}, actor)
}

// ActorFrom returns the actor LocalActor or WithActor set on ctx.
func ActorFrom(ctx context.Context) (comments.Actor, bool) {
	actor, ok := ctx.Value(actorKey{}).(comments.Actor)
	return actor, ok
}

// LocalUser is the person running this machine's local comment hosts: the OS
// account, labelled with their git identity (`git config user.name`, read once).
var LocalUser = sync.OnceValues(func() (comments.Actor, error) {
	return localUser(os.Getuid(), gitConfig)
})

// localUser labels uid's actor with git's user.name, or "You" while git has none.
// The id stays the OS account so existing authorship and read markers keep matching.
func localUser(uid int, git func(key string) (string, bool, error)) (comments.Actor, error) {
	actor := comments.Actor{Kind: "user", ID: fmt.Sprintf("local:%d", uid), Label: "You"}
	name, ok, err := git("user.name")
	if err != nil {
		return comments.Actor{}, fmt.Errorf("read the local comment author's git identity: %w", err)
	}
	if ok {
		actor.Label = name
	}
	return actor, nil
}

// gitConfig reads a git config value, false when it is unset (git exits 1).
func gitConfig(key string) (string, bool, error) {
	out, err := exec.Command("git", "config", "--get", key).Output()
	var exit *exec.ExitError
	if errors.As(err, &exit) && exit.ExitCode() == 1 {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("git config --get %s: %w", key, err)
	}
	value := strings.TrimSpace(string(out))
	return value, value != "", nil
}

// EncodeActor renders actor as an ActorHeader value.
func EncodeActor(actor comments.Actor) string {
	data, err := json.Marshal(actor)
	if err != nil {
		panic(err) // an Actor is three strings; marshalling cannot fail
	}
	return base64.RawURLEncoding.EncodeToString(data)
}

// LocalActor serves handler to loopback callers only, acting as fallback unless
// the request asserts an ActorHeader user or agent. A local host trusts its
// own machine's callers to say who they are; nothing else may reach it.
func LocalActor(handler http.Handler, fallback comments.Actor) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		address, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil || !net.ParseIP(address).IsLoopback() {
			http.Error(w, "local comments require a loopback caller", http.StatusForbidden)
			return
		}
		actor := fallback
		if value := r.Header.Get(ActorHeader); value != "" {
			if actor, err = DecodeActor(value); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		}
		handler.ServeHTTP(w, r.WithContext(WithActor(r.Context(), actor)))
	})
}

// ActorEnv carries an EncodeActor value that `gavel comments` acts as, so an
// agent session replies to a thread it was handed as itself.
const ActorEnv = "GAVEL_COMMENT_ACTOR"

// DecodeActor reads an EncodeActor value: base64url JSON of a user or agent with an id.
func DecodeActor(value string) (comments.Actor, error) {
	var actor comments.Actor
	data, err := base64.RawURLEncoding.DecodeString(value)
	if err == nil {
		err = json.Unmarshal(data, &actor)
	}
	if err == nil && (actor.ID == "" || (actor.Kind != "user" && actor.Kind != "agent")) {
		err = fmt.Errorf("want kind user or agent and an id, got %+v", actor)
	}
	if err != nil {
		return comments.Actor{}, fmt.Errorf("invalid local comment actor: %w", err)
	}
	return actor, nil
}
