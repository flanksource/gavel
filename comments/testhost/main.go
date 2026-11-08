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

	"github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/comments"
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

func (h host) Principal(context.Context) (comments.Principal, error) {
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
		store, err := postgres.New(db.Gorm())
		if err != nil {
			return nil, err
		}
		return comments.New(comments.Options{Store: store, Access: h, Targets: h, Workflows: []comments.Workflow{comments.HumanReview()}, Blobs: store, AttachmentLimit: 32 << 20})
	}
	if err = api.Register(resolve); err != nil {
		return err
	}
	handler := api.Handler(resolve)
	mux := http.NewServeMux()
	mux.HandleFunc("POST /fixtures", func(w http.ResponseWriter, r *http.Request) {
		var rows []comments.Comment
		if err := json.NewDecoder(r.Body).Decode(&rows); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		tenant := r.Header.Get("X-Gavel-Comment-Scope")
		err := db.Gorm().Transaction(func(tx *gorm.DB) error {
			for _, table := range []string{"comment_operations", "comment_bindings", "comment_reads", "comment_attachments", "comment_events", "comment_messages", "comments", "comment_sequences", "comment_blobs", "comment_lint_runs"} {
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
				if _, err = store.Create(r.Context(), c, c.ID, c.ID, comments.Event{CommentID: c.ID, Kind: "fixture", Actor: c.Author, At: c.CreatedAt}); err != nil {
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
		})
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, "{}")
	})
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), requestKey{}, r)))
	}))
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	server := &http.Server{Handler: mux}
	failure := make(chan error, 1)
	go func() { failure <- server.Serve(listener) }()
	if err = json.NewEncoder(os.Stdout).Encode(map[string]string{"url": "http://" + listener.Addr().String()}); err != nil {
		return err
	}
	select {
	case err = <-failure:
		return err
	case <-ctx.Done():
		return server.Shutdown(context.Background())
	}
}
