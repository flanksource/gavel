// Command testhost serves the comments API over an isolated test database for
// browser and backend tests. Each request is scoped by X-Gavel-Comment-Scope
// (tenant and scope) and may assert X-Gavel-Comment-Actor. Agent delivery runs
// against an in-memory Captain bridge with one live session, "fake-session".
package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	captaincli "github.com/flanksource/captain/pkg/cli"
	"github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/comments"
	"github.com/flanksource/gavel/comments/agents"
	"github.com/flanksource/gavel/comments/agents/agentstest"
	api "github.com/flanksource/gavel/comments/clicky"
	"github.com/flanksource/gavel/comments/postgres"
	"github.com/flanksource/gavel/internal/database"
	"gorm.io/gorm"
)

type requestKey struct{}
type host struct {
	scope string
	actor comments.Actor
}

// Principal acts as the actor a delivery asserts (an agent acknowledging), else the request's.
func (h host) Principal(ctx context.Context) (comments.Principal, error) {
	if actor, ok := api.ActorFrom(ctx); ok {
		return comments.Principal{TenantID: h.scope, Actor: actor}, nil
	}
	return comments.Principal{TenantID: h.scope, Actor: h.actor}, nil
}
func (h host) Authorize(_ context.Context, _ comments.Principal, _ string, t comments.Target) error {
	if h.scope == "" || t.ScopeURI != h.scope {
		return comments.ErrForbidden
	}
	return nil
}
func (h host) Resolve(ctx context.Context, p comments.Principal, t comments.Target) (comments.Target, error) {
	return t, h.Authorize(ctx, p, "target", t)
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run() (result error) {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	db, closeDB, err := dbtest.Open(dbtest.Options{Name: "playground_comments"})
	if err != nil {
		return err
	}
	defer func() { result = errors.Join(result, closeDB()) }()
	if err = database.Migrate(ctx, db.DSN()); err != nil {
		return err
	}
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	bridge := &agentstest.Bridge{Dir: dir, Sessions: []captaincli.SessionRecord{{ID: "fake-session", Source: "claude", Model: "claude-test", Slug: "fake-session", Live: &captaincli.SessionLiveWire{CWD: dir}}}}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	baseURL := "http://" + listener.Addr().String()
	mux, err := newHandler(db.Gorm(), bridge, dir, baseURL)
	if err != nil {
		return errors.Join(err, listener.Close())
	}
	return serve(ctx, listener, baseURL, mux)
}

// newHandler serves the comments API plus POST /fixtures over db, delivering
// threads through bridge with managed sessions running in dir and attachment
// links made absolute against baseURL.
func newHandler(db *gorm.DB, bridge agents.Bridge, dir, baseURL string) (http.Handler, error) {
	store, err := postgres.New(db)
	if err != nil {
		return nil, err
	}
	delivery, err := agents.Delivery(agents.Options{Bridge: bridge, Groups: store, BaseURL: baseURL, Workdir: func(context.Context, comments.Target) (string, error) { return dir, nil }})
	if err != nil {
		return nil, err
	}
	resolve := func(ctx context.Context) (*comments.Service, error) {
		r := ctx.Value(requestKey{}).(*http.Request)
		a := comments.Actor{Kind: "user", ID: "user:You", Label: "You"}
		if value := r.Header.Get("X-Gavel-Comment-Actor"); value != "" {
			data, err := base64.RawURLEncoding.DecodeString(value)
			if err != nil {
				return nil, err
			}
			if err = json.Unmarshal(data, &a); err != nil {
				return nil, err
			}
		}
		h := host{scope: r.Header.Get("X-Gavel-Comment-Scope"), actor: a}
		return comments.New(comments.Options{Store: store, Access: h, Targets: h, Workflows: []comments.Workflow{comments.HumanReview()}, Blobs: store, AttachmentLimit: 32 << 20, Routes: store, Delivery: delivery})
	}
	if err = api.Register(resolve); err != nil {
		return nil, err
	}
	handler := api.Handler(resolve)
	mux := http.NewServeMux()
	mux.Handle("POST /fixtures", fixtures(db))
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), requestKey{}, r)))
	}))
	return mux, nil
}

// fixtures replaces the scope's comments with the posted ones.
func fixtures(db *gorm.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var rows []comments.Comment
		if err := json.NewDecoder(r.Body).Decode(&rows); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		tenant := r.Header.Get("X-Gavel-Comment-Scope")
		if err := db.Transaction(func(tx *gorm.DB) error { return replaceFixtures(r.Context(), tx, tenant, rows) }); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, "{}")
	})
}

func replaceFixtures(ctx context.Context, tx *gorm.DB, tenant string, rows []comments.Comment) error {
	for _, table := range []string{"comment_operations", "comment_bindings", "comment_reads", "comment_attachments", "comment_events", "comment_messages", "comments", "comment_sequences", "comment_blobs", "comment_lint_runs", "comment_routes", "comment_agent_sessions"} {
		if err := tx.Exec("DELETE FROM "+table+" WHERE tenant_id=?", tenant).Error; err != nil {
			return err
		}
	}
	store, err := postgres.New(tx)
	if err != nil {
		return err
	}
	for _, c := range rows {
		c.TenantID = tenant
		messages := c.Messages
		c.Messages = nil
		if _, err = store.Create(ctx, c, c.ID, c.ID, comments.Event{CommentID: c.ID, Kind: "fixture", Actor: c.Author, At: c.CreatedAt}); err != nil {
			return err
		}
		for i, m := range messages {
			data, err := json.Marshal(m)
			if err != nil {
				return err
			}
			if err = tx.Exec("INSERT INTO comment_messages (tenant_id,comment_id,id,sequence,payload) VALUES (?,?,?,?,?::jsonb)", tenant, c.ID, m.ID, i+1, string(data)).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func serve(ctx context.Context, listener net.Listener, baseURL string, handler http.Handler) error {
	server := &http.Server{Handler: handler}
	failure := make(chan error, 1)
	go func() { failure <- server.Serve(listener) }()
	if err := json.NewEncoder(os.Stdout).Encode(map[string]string{"url": baseURL}); err != nil {
		return err
	}
	select {
	case err := <-failure:
		return err
	case <-ctx.Done():
		return server.Shutdown(context.Background())
	}
}
