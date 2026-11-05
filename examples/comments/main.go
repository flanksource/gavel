package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	framework "github.com/flanksource/clicky"
	"github.com/flanksource/clicky/route"
	"github.com/flanksource/clicky/rpc"
	commonsdb "github.com/flanksource/commons-db/db"
	"github.com/flanksource/gavel/comments"
	api "github.com/flanksource/gavel/comments/clicky"
	"github.com/flanksource/gavel/comments/postgres"
	"github.com/spf13/cobra"
)

type localHost struct{ scope string }

func (h localHost) Principal(context.Context) (comments.Principal, error) {
	return comments.Principal{TenantID: "example", Actor: comments.Actor{Kind: "user", ID: "local-example"}}, nil
}
func (h localHost) Authorize(_ context.Context, _ comments.Principal, _ string, t comments.Target) error {
	if t.ScopeURI != h.scope || t.URI != "" && !strings.HasPrefix(t.URI, h.scope+"/") {
		return comments.ErrForbidden
	}
	return nil
}
func (h localHost) Resolve(ctx context.Context, p comments.Principal, t comments.Target) (comments.Target, error) {
	if err := h.Authorize(ctx, p, "target", t); err != nil {
		return t, err
	}
	if t.Kind == "code" {
		selector, err := comments.ParseCodeSelector(t.Selector)
		if err != nil {
			return t, err
		}
		if selector.Kind == "code-range" && selector.Revision != t.Revision {
			return t, comments.ErrConflict
		}
	}
	return t, nil
}
func main() {
	if err := run(); err != nil {
		log.Print(err)
		os.Exit(1)
	}
}
func run() error {
	dsn := flag.String("db", os.Getenv("COMMENTS_EXAMPLE_DSN"), "PostgreSQL connection string for an empty example database")
	port := flag.Uint("port", 8099, "Loopback HTTP port")
	scope := flag.String("scope", "example://workspace", "Host-owned target namespace")
	provision := flag.Bool("provision", false, "Explicitly provision empty comments tables before serving")
	flag.Parse()
	if *dsn == "" {
		return fmt.Errorf("set COMMENTS_EXAMPLE_DSN or --db")
	}
	if *port == 0 || *port > 65535 {
		return fmt.Errorf("port must be 1..65535")
	}
	ctx := context.Background()
	if *provision {
		if err := postgres.Provision(ctx, *dsn); err != nil {
			return err
		}
	}
	db, err := commonsdb.NewGorm(*dsn, commonsdb.DefaultGormConfig())
	if err != nil {
		return err
	}
	pool, err := db.DB()
	if err != nil {
		return err
	}
	defer pool.Close()
	store, err := postgres.New(db)
	if err != nil {
		return err
	}
	host := localHost{scope: *scope}
	service, err := comments.New(comments.Options{Store: store, Access: host, Targets: host, Workflows: []comments.Workflow{comments.HumanReview()}, Blobs: store, AttachmentLimit: 10 << 20})
	if err != nil {
		return err
	}
	resolve := func(context.Context) (*comments.Service, error) { return service, nil }
	if err = api.Register(resolve); err != nil {
		return err
	}
	root := &cobra.Command{Use: "comments-example"}
	framework.GenerateCLI(root)
	router := route.NewRouter(nil)
	server := rpc.NewSwaggerServer(&rpc.ServeConfig{Executor: &rpc.ExecutorConfig{Enabled: true, PathPrefix: "/api/v1"}}, root, nil)
	server.RegisterExecutionRoutes(router)
	router.Raw("POST /api/v1/comments/batch", api.Batch(resolve), route.Meta{Entity: "comments", Verb: "batch"})
	router.Raw("GET /api/v1/comments/events", api.Stream(resolve), route.Meta{Entity: "comments", Verb: "events", ReadOnly: true})
	router.Raw("POST /api/v1/comments/{id}/attachments", api.Attachments(resolve), route.Meta{Entity: "comments", Verb: "attach", IDParam: "id"})
	router.Raw("GET /api/v1/comments/{id}/attachments/{attachment}", api.Attachments(resolve), route.Meta{Entity: "comments", Verb: "download", IDParam: "id", ReadOnly: true})
	router.MountGenerated("GET /api/v1/openapi.json", api.OpenAPI(http.HandlerFunc(server.HandleOpenAPIJSON), "/api/v1"))
	log.Printf("comments example: http://127.0.0.1:%d/api/v1/comments", *port)
	return http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", *port), router)
}
