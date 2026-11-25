package agents_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/flanksource/captain/pkg/api"
	captaincli "github.com/flanksource/captain/pkg/cli"
	"github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/comments"
	"github.com/flanksource/gavel/comments/agents"
	"github.com/flanksource/gavel/comments/agents/agentstest"
	commentapi "github.com/flanksource/gavel/comments/clicky"
	"github.com/flanksource/gavel/comments/postgres"
	"github.com/flanksource/gavel/internal/database"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestAgents(t *testing.T) { RegisterFailHandler(Fail); RunSpecs(t, "Comment agents") }

const scope = "gavel://project/acme"

var reviewer = comments.Actor{Kind: "user", ID: "user:reviewer", Label: "Reviewer"}

// host is a single-tenant host acting as the asserted actor, else the reviewer.
type host struct{}

func (host) Principal(ctx context.Context) (comments.Principal, error) {
	actor, ok := commentapi.ActorFrom(ctx)
	if !ok {
		actor = reviewer
	}
	return comments.Principal{TenantID: "tenant-agents", Actor: actor}, nil
}
func (host) Authorize(_ context.Context, _ comments.Principal, _ string, t comments.Target) error {
	if t.ScopeURI != scope {
		return comments.ErrForbidden
	}
	return nil
}
func (h host) Resolve(_ context.Context, _ comments.Principal, t comments.Target) (comments.Target, error) {
	return t, nil
}

type agentContent struct {
	Page     map[string]string `json:"page"`
	AgentRun agents.AgentRun   `json:"agentRun"`
}

// contentOf decodes c's content; a thread created without any has none yet.
func contentOf(c comments.Comment) agentContent {
	var content agentContent
	if len(c.Content) > 0 {
		Expect(json.Unmarshal(c.Content, &content)).To(Succeed())
	}
	return content
}

// pageRoutes resolves a root's route from its page (the target URI below
// <scope>/pages/), as a host's settings would; a page without one is not routed.
type pageRoutes map[string]comments.Route

func (p pageRoutes) RouteFor(_ context.Context, c comments.Comment) (comments.Route, bool, error) {
	r, ok := p[strings.TrimPrefix(c.Target.URI, scope+"/pages/")]
	return r, ok, nil
}

// gatedRoutes holds an automatic delivery's route lookup until the test releases it.
type gatedRoutes struct {
	comments.RouteResolver
	reached, release chan struct{}
}

func (g gatedRoutes) RouteFor(ctx context.Context, c comments.Comment) (comments.Route, bool, error) {
	g.reached <- struct{}{}
	<-g.release
	return g.RouteResolver.RouteFor(ctx, c)
}

// pageRules adds each page's comment agent rule, which applies whether or not
// the page routes new threads automatically; a broken page's rule lookup fails.
type pageRules struct {
	pageRoutes
	rules  map[string]comments.Route
	broken map[string]error
}

func (p pageRules) AgentRule(_ context.Context, t comments.Target) (comments.Route, bool, error) {
	page := strings.TrimPrefix(t.URI, scope+"/pages/")
	if err := p.broken[page]; err != nil {
		return comments.Route{}, false, err
	}
	r, ok := p.rules[page]
	return r, ok, nil
}

// agentDefaultModel is the default model of family's agent mode in captain's runtime catalog.
func agentDefaultModel(family string) string {
	GinkgoHelper()
	for _, f := range api.RuntimeCatalog() {
		for _, mode := range f.Modes {
			if f.Family == family && mode.Mode == string(api.ModeAgent) && !mode.Disabled {
				return mode.DefaultModel
			}
		}
	}
	Fail("captain's runtime catalog has no enabled agent mode for " + family)
	return ""
}

func sessionRoute(source, session string) comments.Route {
	return comments.Route{Source: source, Target: comments.DeliverTarget{Kind: "session", SessionID: session}}
}

// fixture is a migrated database, a bridge with two live sessions working in
// dir ("rename-pass" live-1 and "docs-pass" live-2), and a service on them.
type fixture struct {
	svc         *comments.Service
	store       *postgres.Store
	newDelivery func(*comments.Service) comments.Delivery
	bridge      *agentstest.Bridge
	dir         string
}

func newFixture(ctx context.Context, name string, routes comments.RouteResolver) *fixture {
	GinkgoHelper()
	db := dbtest.ForGinkgo(dbtest.Options{Name: name})
	Expect(database.Migrate(ctx, db.DSN())).To(Succeed())
	f := &fixture{dir: GinkgoT().TempDir()}
	var err error
	f.store, err = postgres.New(db.Gorm())
	Expect(err).NotTo(HaveOccurred())
	f.bridge = &agentstest.Bridge{Dir: f.dir, Sessions: []captaincli.SessionRecord{f.live("live-1", "rename-pass"), f.live("live-2", "docs-pass")}}
	f.newDelivery = f.delivery(func(context.Context, comments.Target) (string, error) { return f.dir, nil })
	f.svc = f.service(comments.Options{Routes: routes, Delivery: f.newDelivery})
	return f
}

func (f *fixture) live(id, slug string) captaincli.SessionRecord {
	return captaincli.SessionRecord{ID: id, Source: "claude", Model: "claude-opus", Slug: slug, Live: &captaincli.SessionLiveWire{CWD: f.dir}}
}

// delivery is agent delivery over the fixture's bridge and store, its sessions working in workdir.
func (f *fixture) delivery(workdir func(context.Context, comments.Target) (string, error)) func(*comments.Service) comments.Delivery {
	GinkgoHelper()
	newDelivery, err := agents.Delivery(agents.Options{Bridge: f.bridge, Groups: f.store, Workdir: workdir})
	Expect(err).NotTo(HaveOccurred())
	return newDelivery
}

// service is a human-review service on the fixture's store and host, with options' routes, delivery and blobs.
func (f *fixture) service(options comments.Options) *comments.Service {
	GinkgoHelper()
	options.Store, options.Access, options.Targets, options.Workflows = f.store, host{}, host{}, []comments.Workflow{comments.HumanReview()}
	svc, err := comments.New(options)
	Expect(err).NotTo(HaveOccurred())
	return svc
}

func pageTarget(page string) comments.Target {
	return comments.Target{URI: scope + "/pages/" + page, ScopeURI: scope, Kind: "web-page", Label: page}
}

func (f *fixture) create(ctx context.Context, page, body string) comments.Comment {
	GinkgoHelper()
	c, err := f.svc.Create(ctx, comments.CreateRequest{
		Target: pageTarget(page), Body: body, WorkflowID: "human-review", RequestID: "create-" + page + "-" + body,
		Content: json.RawMessage(`{"page":{"url":"https://example.com/` + page + `"}}`),
	})
	Expect(err).NotTo(HaveOccurred())
	return c
}

func (f *fixture) get(ctx context.Context, id string) comments.Comment {
	GinkgoHelper()
	c, err := f.svc.Get(ctx, id)
	Expect(err).NotTo(HaveOccurred())
	return c
}

// runOf polls thread id's agentRun.
func (f *fixture) runOf(ctx context.Context, id string) func() agents.AgentRun {
	return func() agents.AgentRun { return contentOf(f.get(ctx, id)).AgentRun }
}
