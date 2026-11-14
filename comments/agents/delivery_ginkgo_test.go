package agents_test

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

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

func contentOf(c comments.Comment) agentContent {
	var content agentContent
	Expect(json.Unmarshal(c.Content, &content)).To(Succeed())
	return content
}

var _ = Describe("Delivering comment threads to agents", Ordered, func() {
	var (
		svc         *comments.Service
		store       *postgres.Store
		newDelivery func(*comments.Service) comments.Delivery
		bridge      *agentstest.Bridge
		dir         string
	)
	live := func(id, slug string) captaincli.SessionRecord {
		return captaincli.SessionRecord{ID: id, Source: "claude", Model: "claude-opus", Slug: slug, Live: &captaincli.SessionLiveWire{CWD: dir}}
	}
	create := func(ctx context.Context, page, body string) comments.Comment {
		c, err := svc.Create(ctx, comments.CreateRequest{
			Target: comments.Target{URI: scope + "/pages/" + page, ScopeURI: scope, Kind: "web-page", Label: page},
			Body:   body, WorkflowID: "human-review", RequestID: "create-" + page + "-" + body,
			Content: json.RawMessage(`{"page":{"url":"https://example.com/` + page + `"}}`),
		})
		Expect(err).NotTo(HaveOccurred())
		return c
	}
	get := func(ctx context.Context, id string) comments.Comment {
		c, err := svc.Get(ctx, id)
		Expect(err).NotTo(HaveOccurred())
		return c
	}

	BeforeAll(func(ctx SpecContext) {
		db := dbtest.ForGinkgo(dbtest.Options{Name: "comment_agents"})
		Expect(database.Migrate(ctx, db.DSN())).To(Succeed())
		var err error
		store, err = postgres.New(db.Gorm())
		Expect(err).NotTo(HaveOccurred())
		dir = GinkgoT().TempDir()
		bridge = &agentstest.Bridge{Dir: dir}
		bridge.Sessions = []captaincli.SessionRecord{live("live-1", "rename-pass"), live("live-2", "docs-pass")}
		newDelivery, err = agents.Delivery(agents.Options{Bridge: bridge, Groups: store, Workdir: func(context.Context, comments.Target) (string, error) { return dir, nil }})
		Expect(err).NotTo(HaveOccurred())
		svc, err = comments.New(comments.Options{Store: store, Access: host{}, Targets: host{}, Workflows: []comments.Workflow{comments.HumanReview()}, Routes: store, Delivery: newDelivery})
		Expect(err).NotTo(HaveOccurred())
	})

	It("hands a thread to a live session as its agent, delivers it with the page it is on, and records the run", func(ctx SpecContext) {
		c := create(ctx, "home", "Rename the button")
		delivered, err := svc.Deliver(ctx, comments.DeliverRequest{CommentID: c.ID, ExpectedVersion: c.Version, RequestID: "send-home", Target: comments.DeliverTarget{Kind: "session", SessionID: "live-1"}})
		Expect(err).NotTo(HaveOccurred())
		agent := comments.Actor{Kind: "agent", ID: "claude:live-1", Label: "claude-opus"}
		run := contentOf(delivered).AgentRun
		Expect([]any{delivered.Status, *delivered.Assignee, contentOf(delivered).Page["url"], run.SessionID, run.State, run.Target}).
			To(Equal([]any{"in_progress", agent, "https://example.com/home", "live-1", "sent", agents.AgentRunTarget{Label: "rename-pass", Channel: "claude-socket"}}))
		sent, _ := bridge.Snapshot()
		Expect(sent).To(HaveLen(1))
		Expect(sent[0].SessionID).To(Equal("live-1"))
		Expect(sent[0].Text).To(ContainSubstring("thread " + c.ID + " on https://example.com/home"))
		Expect(sent[0].Text).To(ContainSubstring("export GAVEL_COMMENT_ACTOR=" + commentapi.EncodeActor(agent)))
		Expect(sent[0].Text).To(ContainSubstring("Rename the button"))
	})

	It("records a failed delivery on the thread as an errored agentRun and returns ErrDelivery", func(ctx SpecContext) {
		bridge.FailDeliver = "socket closed"
		DeferCleanup(func() { bridge.FailDeliver = "" })
		c := create(ctx, "about", "Fix the heading")
		failed, err := svc.Deliver(ctx, comments.DeliverRequest{CommentID: c.ID, ExpectedVersion: c.Version, RequestID: "send-about", Target: comments.DeliverTarget{Kind: "session", SessionID: "live-1"}})
		Expect(err).To(MatchError(comments.ErrDelivery))
		Expect(err.Error()).To(ContainSubstring("socket closed"))
		run := contentOf(failed).AgentRun
		Expect([]any{run.State, run.SessionID}).To(Equal([]any{"error", "live-1"}))
		Expect(run.Error).To(ContainSubstring("socket closed"))
		_, err = svc.Deliver(ctx, comments.DeliverRequest{CommentID: c.ID, ExpectedVersion: 1, RequestID: "send-about-stale", Target: comments.DeliverTarget{Kind: "session", SessionID: "live-1"}})
		Expect(err).To(MatchError(comments.ErrConflict))
	})

	It("starts one managed session for a group and sends the group's next thread into it", func(ctx SpecContext) {
		target := comments.DeliverTarget{Kind: "run", Prompt: "editor-comments", GroupKey: "page:pricing", RunOptions: json.RawMessage(`{"instructions":"Keep the copy short"}`)}
		first := create(ctx, "pricing", "Shorten the intro")
		started, err := svc.Deliver(ctx, comments.DeliverRequest{CommentID: first.ID, ExpectedVersion: first.Version, RequestID: "run-pricing-1", Target: target})
		Expect(err).NotTo(HaveOccurred())
		_, launched := bridge.Snapshot()
		Expect(launched).To(HaveLen(1))
		var vars map[string]string
		Expect(json.Unmarshal([]byte(launched[0].Options.Vars), &vars)).To(Succeed())
		Expect([]any{launched[0].Reference, vars["dir"], vars["scope"], vars["target"]}).To(Equal([]any{"editor-comments", dir, scope, first.Target.URI}))
		Expect(vars["message"]).To(ContainSubstring("## Additional instructions\n\nKeep the copy short"))
		run := contentOf(get(ctx, started.ID)).AgentRun
		Expect([]any{run.State, run.SessionID}).To(Equal([]any{"sent", "managed-1"}))

		second := create(ctx, "pricing", "Add a FAQ")
		_, err = svc.Deliver(ctx, comments.DeliverRequest{CommentID: second.ID, ExpectedVersion: second.Version, RequestID: "run-pricing-2", Target: target})
		Expect(err).NotTo(HaveOccurred())
		sent, launched := bridge.Snapshot()
		Expect(launched).To(HaveLen(1))
		Expect(sent[len(sent)-1].SessionID).To(Equal("managed-1"))
		Expect(contentOf(get(ctx, second.ID)).AgentRun.SessionID).To(Equal("managed-1"))
	})

	It("starts a fresh session for a group whose session is still live, and the group's next thread reuses the fresh one", func(ctx SpecContext) {
		target := comments.DeliverTarget{Kind: "run", Prompt: "editor-comments", GroupKey: "page:faq"}
		deliver := func(page, requestID string, fresh bool) agents.AgentRun {
			c := create(ctx, page, requestID)
			target.Fresh = fresh
			_, err := svc.Deliver(ctx, comments.DeliverRequest{CommentID: c.ID, ExpectedVersion: c.Version, RequestID: requestID, Target: target})
			Expect(err).NotTo(HaveOccurred())
			return contentOf(get(ctx, c.ID)).AgentRun
		}
		_, before := bridge.Snapshot()
		first := deliver("faq", "run-faq-1", false)
		retried := deliver("faq", "run-faq-fresh", true)
		next := deliver("faq", "run-faq-next", false)
		_, launched := bridge.Snapshot()
		recorded, err := store.GroupSession(ctx, "tenant-agents", scope, "page:faq")
		Expect(err).NotTo(HaveOccurred())
		freshID := fmt.Sprintf("managed-%d", len(before)+2)
		Expect([]any{len(launched) - len(before), first.SessionID, retried.SessionID, next.SessionID, recorded}).
			To(Equal([]any{2, fmt.Sprintf("managed-%d", len(before)+1), freshID, freshID, freshID}))
	})

	It("redelivers a thread whose hand-off errored into a fresh session and replaces its agentRun", func(ctx SpecContext) {
		bridge.FailDeliver = "socket closed"
		c := create(ctx, "careers", "Fix the form")
		_, err := svc.Deliver(ctx, comments.DeliverRequest{CommentID: c.ID, ExpectedVersion: c.Version, RequestID: "send-careers", Target: comments.DeliverTarget{Kind: "session", SessionID: "live-1"}})
		bridge.FailDeliver = ""
		Expect(err).To(MatchError(comments.ErrDelivery))
		failed := get(ctx, c.ID)
		Expect(contentOf(failed).AgentRun.State).To(Equal("error"))

		_, err = svc.Deliver(ctx, comments.DeliverRequest{CommentID: c.ID, ExpectedVersion: failed.Version, RequestID: "retry-careers", Target: comments.DeliverTarget{Kind: "run", Prompt: "editor-comments", Fresh: true}})
		Expect(err).NotTo(HaveOccurred())
		_, launched := bridge.Snapshot()
		run := contentOf(get(ctx, c.ID)).AgentRun
		Expect([]any{run.State, run.Error, run.SessionID}).To(Equal([]any{"sent", "", fmt.Sprintf("managed-%d", len(launched))}))
	})

	It("rejects fresh on a session target", func() {
		Expect(comments.DeliverTarget{Kind: "session", SessionID: "live-1", Fresh: true}.Validate()).To(MatchError(comments.ErrInvalid))
	})

	It("refuses run options it cannot run before queueing anything", func(ctx SpecContext) {
		c := create(ctx, "terms", "Fix it")
		_, err := svc.Deliver(ctx, comments.DeliverRequest{CommentID: c.ID, ExpectedVersion: c.Version, RequestID: "run-invalid", Target: comments.DeliverTarget{Kind: "run", Prompt: "editor-comments", RunOptions: json.RawMessage(`{"spec":"opus"}`)}})
		Expect(err).To(MatchError(comments.ErrInvalid))
		Expect(get(ctx, c.ID).Version).To(Equal(c.Version))
	})

	It("routes a new human root to the longest matching route, skips disabled subtrees and agent-authored roots", func(ctx SpecContext) {
		save := func(prefix, session string, enabled bool) {
			_, err := svc.SaveRoute(ctx, comments.Route{ScopeURI: scope, TargetPrefix: scope + prefix, WorkflowID: "human-review", Target: comments.DeliverTarget{Kind: "session", SessionID: session}, Enabled: enabled})
			Expect(err).NotTo(HaveOccurred())
		}
		save("/pages/", "live-1", true)
		save("/pages/docs", "live-2", true)
		save("/pages/private", "live-1", false)
		sessionOf := func(id string) func() string {
			return func() string { return contentOf(get(ctx, id)).AgentRun.SessionID }
		}
		general, docs := create(ctx, "blog", "Tighten the post"), create(ctx, "docs/intro", "Add an example")
		Eventually(sessionOf(general.ID)).WithTimeout(10 * time.Second).Should(Equal("live-1"))
		Eventually(sessionOf(docs.ID)).WithTimeout(10 * time.Second).Should(Equal("live-2"))
		private := create(ctx, "private/notes", "Do not send")
		byAgent := create(commentapi.WithActor(ctx, comments.Actor{Kind: "agent", ID: "claude:live-1"}), "blog-agent", "Agent note")
		Consistently(func() []string { return []string{sessionOf(private.ID)(), sessionOf(byAgent.ID)()} }).
			WithTimeout(time.Second).Should(Equal([]string{"", ""}))
		routes, err := svc.Routes(ctx, scope)
		Expect(err).NotTo(HaveOccurred())
		Expect(routes).To(HaveLen(3))
	})

	It("lets the creator attach screenshots with its create version while a routed delivery claims and sends the thread", func(ctx SpecContext) {
		routes := gatedRoutes{RouteStore: store, reached: make(chan struct{}), release: make(chan struct{})}
		routed, err := comments.New(comments.Options{Store: store, Access: host{}, Targets: host{}, Workflows: []comments.Workflow{comments.HumanReview()}, Blobs: store, AttachmentLimit: 1 << 20, Routes: routes, Delivery: newDelivery})
		Expect(err).NotTo(HaveOccurred())
		_, err = routed.SaveRoute(ctx, comments.Route{ScopeURI: scope, TargetPrefix: scope + "/pages/race", WorkflowID: "human-review", Target: comments.DeliverTarget{Kind: "session", SessionID: "live-1"}, Enabled: true})
		Expect(err).NotTo(HaveOccurred())
		stages, release := make(chan string), map[string]chan struct{}{"live": make(chan struct{}), "deliver": make(chan struct{})}
		bridge.SetHold(func(stage string) { stages <- stage; <-release[stage] })
		DeferCleanup(func() { bridge.SetHold(nil) })

		c, err := routed.Create(ctx, comments.CreateRequest{Target: comments.Target{URI: scope + "/pages/race", ScopeURI: scope, Kind: "web-page"}, Body: "The logo is clipped", WorkflowID: "human-review", RequestID: "create-race"})
		Expect(err).NotTo(HaveOccurred())
		attach := func(name string) {
			_, err := routed.Upload(ctx, comments.Upload{CommentID: c.ID, ExpectedVersion: c.Version, Name: name, MediaType: "image/png", Body: strings.NewReader("png " + name)})
			Expect(err).NotTo(HaveOccurred(), "uploading %s with the create version %d", name, c.Version)
		}
		Eventually(routes.reached).Should(Receive())
		attach("before-queue.png")
		close(routes.release)
		Eventually(stages).WithTimeout(10 * time.Second).Should(Receive(Equal("live")))
		attach("before-claim.png")
		close(release["live"])
		Eventually(stages).WithTimeout(10 * time.Second).Should(Receive(Equal("deliver")))
		attach("before-send.png")
		close(release["deliver"])

		Eventually(func() string { return contentOf(get(ctx, c.ID)).AgentRun.State }).WithTimeout(10 * time.Second).Should(Equal("sent"))
		final := get(ctx, c.ID)
		names := []string{}
		for _, a := range final.Attachments {
			names = append(names, a.Name)
		}
		Expect(names).To(ConsistOf("before-queue.png", "before-claim.png", "before-send.png"))
		Expect([]any{*final.Assignee, final.Status}).To(Equal([]any{comments.Actor{Kind: "agent", ID: "claude:live-1", Label: "claude-opus"}, "in_progress"}))
	})

	It("assigns a thread to an assign route's agent before its delivery reaches a session, then to the session's agent", func(ctx SpecContext) {
		route, err := svc.SaveRoute(ctx, comments.Route{ScopeURI: scope, TargetPrefix: scope + "/pages/assigned", WorkflowID: "human-review", Target: comments.DeliverTarget{Kind: "session", SessionID: "live-2"}, Enabled: true, Assign: true})
		Expect(err).NotTo(HaveOccurred())
		stages, release := make(chan string), make(chan struct{})
		bridge.SetHold(func(stage string) {
			if stage == "live" {
				stages <- stage
				<-release
			}
		})
		DeferCleanup(func() { bridge.SetHold(nil) })
		assigneeOf := func(id string) func() comments.Actor {
			return func() comments.Actor {
				if c := get(ctx, id); c.Assignee != nil {
					return *c.Assignee
				}
				return comments.Actor{}
			}
		}

		c := create(ctx, "assigned", "Check the table")
		Eventually(stages).WithTimeout(10 * time.Second).Should(Receive(Equal("live")))
		Expect(assigneeOf(c.ID)()).To(Equal(comments.Actor{Kind: "agent", ID: "route:" + route.ID, Label: "live-2"}))
		close(release)
		Eventually(assigneeOf(c.ID)).WithTimeout(10 * time.Second).Should(Equal(comments.Actor{Kind: "agent", ID: "claude:live-2", Label: "claude-opus"}))
	})
})

// gatedRoutes holds an automatic delivery's route lookup until the test releases it.
type gatedRoutes struct {
	comments.RouteStore
	reached, release chan struct{}
}

func (g gatedRoutes) Routes(ctx context.Context, tenant, scope string) ([]comments.Route, error) {
	g.reached <- struct{}{}
	<-g.release
	return g.RouteStore.Routes(ctx, tenant, scope)
}
