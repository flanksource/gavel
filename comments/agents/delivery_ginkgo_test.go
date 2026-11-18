package agents_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/comments"
	"github.com/flanksource/gavel/comments/agents"
	"github.com/flanksource/gavel/comments/agents/agentstest"
	commentapi "github.com/flanksource/gavel/comments/clicky"
	"github.com/flanksource/gavel/comments/postgres"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Delivering comment threads to agents", Ordered, func() {
	routes := pageRoutes{
		"blog":       sessionRoute("/repo/.gavel.yaml", "live-1"),
		"docs/intro": sessionRoute("db:path:pages/docs", "live-2"),
		"blog-agent": sessionRoute("/repo/.gavel.yaml", "live-1"),
		"race":       sessionRoute("/repo/.gavel.yaml", "live-1"),
		"assigned":   {Source: "db:path:pages", Target: comments.DeliverTarget{Kind: "session", SessionID: "live-2"}, Assign: true},
	}
	// proposeRule is the "proposals" page's rule: it does not route threads
	// automatically, but every delivery of them inherits its propose policy.
	proposeRule := comments.Route{Source: "db:path:pages/proposals", Target: comments.DeliverTarget{Kind: "run", Prompt: "editor-comments", Edits: comments.EditsPropose}}
	var (
		f           *fixture
		svc         *comments.Service
		store       *postgres.Store
		newDelivery func(*comments.Service) comments.Delivery
		bridge      *agentstest.Bridge
		dir         string
	)
	create := func(ctx context.Context, page, body string) comments.Comment { return f.create(ctx, page, body) }
	get := func(ctx context.Context, id string) comments.Comment { return f.get(ctx, id) }

	BeforeAll(func(ctx SpecContext) {
		ruled := pageRules{pageRoutes: routes, rules: map[string]comments.Route{"proposals": proposeRule}, broken: map[string]error{"broken-rule": errors.New("rule store offline")}}
		f = newFixture(ctx, "comment_agents", ruled)
		svc, store, newDelivery, bridge, dir = f.svc, f.store, f.newDelivery, f.bridge, f.dir
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

	It("records a propose delivery's edits policy in the agentRun and sends only the propose commands", func(ctx SpecContext) {
		c := create(ctx, "contact", "Reword the form label")
		delivered, err := svc.Deliver(ctx, comments.DeliverRequest{CommentID: c.ID, ExpectedVersion: c.Version, RequestID: "send-contact", Target: comments.DeliverTarget{Kind: "session", SessionID: "live-1", Edits: comments.EditsPropose}})
		Expect(err).NotTo(HaveOccurred())
		Expect(contentOf(delivered).AgentRun.Edits).To(Equal(comments.EditsPropose))
		sent, _ := bridge.Snapshot()
		message := sent[len(sent)-1].Text
		Expect([]any{strings.Contains(message, "Read the files under "+dir), strings.Contains(message, "--edits '"), strings.Contains(message, dir+"/.tmp"), strings.Contains(message, "gavel comments resolve")}).
			To(Equal([]any{true, true, false, false}))
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
		var vars map[string]any
		Expect(json.Unmarshal([]byte(launched[0].Options.Vars), &vars)).To(Succeed())
		Expect([]any{launched[0].Reference, vars["dir"], vars["scope"], vars["target"], vars["propose"]}).To(Equal([]any{"editor-comments", dir, scope, first.Target.URI, false}))
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

	It("lists a managed session after it ends as resumable, and resumes it with its recorded prompt when a thread is delivered to it", func(ctx SpecContext) {
		target := comments.DeliverTarget{Kind: "run", Prompt: "editor-comments", Fresh: true, RunOptions: json.RawMessage(`{"instructions":"Keep it terse"}`)}
		first := create(ctx, "changelog", "Write the changelog\nfor the release")
		_, err := svc.Deliver(ctx, comments.DeliverRequest{CommentID: first.ID, ExpectedVersion: first.Version, RequestID: "run-changelog", Target: target})
		Expect(err).NotTo(HaveOccurred())
		session := contentOf(get(ctx, first.ID)).AgentRun.SessionID
		managedOf := func() comments.AgentSession {
			GinkgoHelper()
			sessions, err := svc.AgentSessions(ctx, comments.SessionQuery{ScopeURI: scope})
			Expect(err).NotTo(HaveOccurred())
			found := []comments.AgentSession{}
			for _, s := range sessions {
				if s.ID == session {
					found = append(found, s)
				}
			}
			Expect(found).To(HaveLen(1), "session %s is listed once", session)
			return found[0]
		}
		Expect([]any{managedOf().Managed, managedOf().Resumable}).To(Equal([]any{true, false}))
		Expect(bridge.End(session)).To(Succeed())
		Expect(managedOf()).To(Equal(comments.AgentSession{ID: session, Title: "Write the changelog", CWD: dir, Status: "ended", Managed: true, Resumable: true}))

		second := create(ctx, "changelog", "Mention the fix")
		_, err = svc.Deliver(ctx, comments.DeliverRequest{CommentID: second.ID, ExpectedVersion: second.Version, RequestID: "resume-changelog", Target: comments.DeliverTarget{Kind: "session", SessionID: session}})
		Expect(err).NotTo(HaveOccurred())
		_, launched := bridge.Snapshot()
		resumed := launched[len(launched)-1]
		var vars map[string]any
		Expect(json.Unmarshal([]byte(resumed.Options.Vars), &vars)).To(Succeed())
		run := contentOf(get(ctx, second.ID)).AgentRun
		Expect([]any{resumed.Options.Resume, resumed.Reference, run.SessionID, run.State, run.Target.Label}).To(Equal([]any{session, "editor-comments", session, "sent", "Write the changelog"}))
		Expect(vars["message"]).To(And(ContainSubstring("Mention the fix"), ContainSubstring("## Additional instructions\n\nKeep it terse")))
		Expect([]any{managedOf().Status, managedOf().Resumable}).To(Equal([]any{"", false}))
	})

	It("refuses delivery to an ended session gavel did not launch", func(ctx SpecContext) {
		c := create(ctx, "legal", "Fix the footer")
		_, err := svc.Deliver(ctx, comments.DeliverRequest{CommentID: c.ID, ExpectedVersion: c.Version, RequestID: "send-legal", Target: comments.DeliverTarget{Kind: "session", SessionID: "ended-elsewhere"}})
		Expect(err).To(MatchError(comments.ErrDelivery))
		Expect(err.Error()).To(ContainSubstring("agent session ended-elsewhere is no longer live"))
	})

	It("resumes a group's ended managed session instead of launching a fresh one", func(ctx SpecContext) {
		target := comments.DeliverTarget{Kind: "run", Prompt: "editor-comments", GroupKey: "page:roadmap"}
		first := create(ctx, "roadmap", "Reorder the milestones")
		_, err := svc.Deliver(ctx, comments.DeliverRequest{CommentID: first.ID, ExpectedVersion: first.Version, RequestID: "run-roadmap-1", Target: target})
		Expect(err).NotTo(HaveOccurred())
		session := contentOf(get(ctx, first.ID)).AgentRun.SessionID
		Expect(bridge.End(session)).To(Succeed())

		second := create(ctx, "roadmap", "Date the next milestone")
		_, err = svc.Deliver(ctx, comments.DeliverRequest{CommentID: second.ID, ExpectedVersion: second.Version, RequestID: "run-roadmap-2", Target: target})
		Expect(err).NotTo(HaveOccurred())
		_, launched := bridge.Snapshot()
		recorded, err := store.GroupSession(ctx, "tenant-agents", scope, "page:roadmap")
		Expect(err).NotTo(HaveOccurred())
		Expect([]string{launched[len(launched)-1].Options.Resume, contentOf(get(ctx, second.ID)).AgentRun.SessionID, recorded}).To(Equal([]string{session, session, session}))
	})

	It("delivers a batch in one launch that every thread records, and a stale thread fails the batch before anything is sent", func(ctx SpecContext) {
		target := comments.DeliverTarget{Kind: "run", Prompt: "editor-comments", Fresh: true}
		a, b := create(ctx, "batch", "Fix the logo"), create(ctx, "batch", "Fix the footer")
		_, before := bridge.Snapshot()
		delivered, err := svc.DeliverBatch(ctx, comments.DeliverBatchRequest{RequestID: "batch-1", Target: target, Threads: []comments.DeliverThread{{ID: a.ID, ExpectedVersion: a.Version}, {ID: b.ID, ExpectedVersion: b.Version}}})
		Expect(err).NotTo(HaveOccurred())
		_, launched := bridge.Snapshot()
		Expect(launched).To(HaveLen(len(before) + 1))
		var vars map[string]any
		Expect(json.Unmarshal([]byte(launched[len(launched)-1].Options.Vars), &vars)).To(Succeed())
		Expect(vars["message"]).To(And(ContainSubstring("# Thread "+a.ID), ContainSubstring("Fix the logo"), ContainSubstring("# Thread "+b.ID), ContainSubstring("Fix the footer")))
		session := fmt.Sprintf("managed-%d", len(launched))
		runs := []any{}
		for _, c := range delivered {
			run := contentOf(get(ctx, c.ID)).AgentRun
			runs = append(runs, []string{c.ID, run.SessionID, run.State})
		}
		Expect(runs).To(Equal([]any{[]string{a.ID, session, "sent"}, []string{b.ID, session, "sent"}}))

		replayed, err := svc.DeliverBatch(ctx, comments.DeliverBatchRequest{RequestID: "batch-1", Target: target, Threads: []comments.DeliverThread{{ID: a.ID, ExpectedVersion: a.Version}, {ID: b.ID, ExpectedVersion: b.Version}}})
		Expect(err).NotTo(HaveOccurred())
		Expect(replayed).To(HaveLen(2))
		c, stale := create(ctx, "batch", "Fix the header"), get(ctx, a.ID)
		_, err = svc.DeliverBatch(ctx, comments.DeliverBatchRequest{RequestID: "batch-stale", Target: target, Threads: []comments.DeliverThread{{ID: c.ID, ExpectedVersion: c.Version}, {ID: a.ID, ExpectedVersion: stale.Version - 1}}})
		Expect(err).To(MatchError(comments.ErrConflict))
		_, after := bridge.Snapshot()
		Expect([]any{len(after), contentOf(get(ctx, c.ID)).AgentRun.State}).To(Equal([]any{len(launched), ""}))
		_, err = svc.DeliverBatch(ctx, comments.DeliverBatchRequest{RequestID: "batch-empty", Target: target})
		Expect(err).To(MatchError(comments.ErrInvalid))
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

	It("routes a new human root through the route its target resolves to, and skips unrouted and agent-authored roots", func(ctx SpecContext) {
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
	})

	It("delivers a root created with an explicit target there, once, instead of through its page's route", func(ctx SpecContext) {
		explicit := comments.DeliverTarget{Kind: "session", SessionID: "live-2"}
		before, _ := bridge.Snapshot()
		createTo := func(page, body string) comments.Comment {
			c, err := svc.Create(ctx, comments.CreateRequest{
				Target: comments.Target{URI: scope + "/pages/" + page, ScopeURI: scope, Kind: "web-page", Label: page},
				Body:   body, WorkflowID: "human-review", RequestID: "create-explicit-" + page,
				Content: json.RawMessage(`{"page":{"url":"https://example.com/` + page + `"}}`),
				Deliver: &explicit,
			})
			Expect(err).NotTo(HaveOccurred())
			return c
		}
		routed, unrouted := createTo("blog", "Send this to the docs pass"), createTo("private/explicit", "Deliver despite no route")
		sessionOf := func(id string) string { return contentOf(get(ctx, id)).AgentRun.SessionID }
		Eventually(func() []string { return []string{sessionOf(routed.ID), sessionOf(unrouted.ID)} }).
			WithTimeout(10 * time.Second).Should(Equal([]string{"live-2", "live-2"}))
		Consistently(func() []string {
			sent, _ := bridge.Snapshot()
			sessions := []string{}
			for _, d := range sent[len(before):] {
				sessions = append(sessions, d.SessionID)
			}
			return sessions
		}).WithTimeout(time.Second).Should(Equal([]string{"live-2", "live-2"}))
	})

	It("rejects a create whose explicit delivery target is malformed without creating the root", func(ctx SpecContext) {
		target := comments.Target{URI: scope + "/pages/bad-target", ScopeURI: scope, Kind: "web-page"}
		_, err := svc.Create(ctx, comments.CreateRequest{Target: target, Body: "Nowhere", WorkflowID: "human-review", RequestID: "create-bad-target", Deliver: &comments.DeliverTarget{Kind: "session"}})
		Expect(err).To(MatchError(comments.ErrInvalid))
		listed, err := svc.List(ctx, comments.Filter{ScopeURI: scope, TargetURI: target.URI, Limit: 10})
		Expect(err).NotTo(HaveOccurred())
		Expect(listed).To(BeEmpty())
	})

	It("lets the creator attach screenshots with its create version while a routed delivery claims and sends the thread", func(ctx SpecContext) {
		gated := gatedRoutes{RouteResolver: routes, reached: make(chan struct{}), release: make(chan struct{})}
		routed, err := comments.New(comments.Options{Store: store, Access: host{}, Targets: host{}, Workflows: []comments.Workflow{comments.HumanReview()}, Blobs: store, AttachmentLimit: 1 << 20, Routes: gated, Delivery: newDelivery})
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
		Eventually(gated.reached).Should(Receive())
		attach("before-queue.png")
		close(gated.release)
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
		Expect(assigneeOf(c.ID)()).To(Equal(comments.Actor{Kind: "agent", ID: "route:db:path:pages", Label: "live-2"}))
		close(release)
		Eventually(assigneeOf(c.ID)).WithTimeout(10 * time.Second).Should(Equal(comments.Actor{Kind: "agent", ID: "claude:live-2", Label: "claude-opus"}))
	})

	runOf := func(ctx context.Context, id string) func() agents.AgentRun {
		return func() agents.AgentRun { return contentOf(get(ctx, id)).AgentRun }
	}
	toolNames := func(tools []api.ToolDefinition) []string {
		names := []string{}
		for _, tool := range tools {
			names = append(names, tool.Name)
		}
		return names
	}
	proposeTools := []string{"list_threads", "reply_to_thread", "propose_change"}

	It("inherits the page rule's propose policy for a session and a batch delivery that leave edits empty, unless the target sets edits", func(ctx SpecContext) {
		toSession := func(c comments.Comment, requestID, edits string) agents.AgentRun {
			_, err := svc.Deliver(ctx, comments.DeliverRequest{CommentID: c.ID, ExpectedVersion: c.Version, RequestID: requestID, Target: comments.DeliverTarget{Kind: "session", SessionID: "live-1", Edits: edits}})
			Expect(err).NotTo(HaveOccurred())
			return runOf(ctx, c.ID)()
		}
		inherited := toSession(create(ctx, "proposals", "Reword the hero"), "inherit-session", "")
		sent, _ := bridge.Snapshot()
		proposeCommands := strings.Contains(sent[len(sent)-1].Text, "--edits '")
		overridden := toSession(create(ctx, "proposals", "Fix the hero typo"), "override-session", comments.EditsDirect)

		a, b := create(ctx, "proposals", "Shorten the intro"), create(ctx, "proposals", "Drop the banner")
		_, err := svc.DeliverBatch(ctx, comments.DeliverBatchRequest{RequestID: "inherit-batch", Target: comments.DeliverTarget{Kind: "run", Prompt: "editor-comments", Fresh: true}, Threads: []comments.DeliverThread{{ID: a.ID, ExpectedVersion: a.Version}, {ID: b.ID, ExpectedVersion: b.Version}}})
		Expect(err).NotTo(HaveOccurred())
		_, launched := bridge.Snapshot()
		batch := launched[len(launched)-1]
		Expect([]any{inherited.Edits, proposeCommands, overridden.Edits, runOf(ctx, a.ID)().Edits, runOf(ctx, b.ID)().Edits, batch.Options.Spec.Permissions.Tools["Edit"], toolNames(batch.Tools)}).
			To(Equal([]any{comments.EditsPropose, true, comments.EditsDirect, comments.EditsPropose, comments.EditsPropose, api.ToolPolicyDeny, proposeTools}))
	})

	It("previews a run target on a propose page with the propose tools", func(ctx SpecContext) {
		c := create(ctx, "proposals", "Preview the hero")
		raw, err := svc.PreviewDelivery(ctx, c.ID, comments.DeliverTarget{Kind: "run", Prompt: "editor-comments"})
		Expect(err).NotTo(HaveOccurred())
		var preview agents.RunPreview
		Expect(json.Unmarshal(raw, &preview)).To(Succeed())
		thread := []string{}
		for _, tool := range preview.Tools {
			if tool.Source == "thread" {
				thread = append(thread, tool.Name)
			}
		}
		Expect(thread).To(Equal(proposeTools))
	})

	It("refuses a batch whose threads inherit different edits policies, unless its target sets one", func(ctx SpecContext) {
		proposed, direct := create(ctx, "proposals", "Merge the cards"), create(ctx, "plain", "Merge the rows")
		threads := []comments.DeliverThread{{ID: proposed.ID, ExpectedVersion: proposed.Version}, {ID: direct.ID, ExpectedVersion: direct.Version}}
		_, mixed := svc.DeliverBatch(ctx, comments.DeliverBatchRequest{RequestID: "mixed-edits", Target: comments.DeliverTarget{Kind: "session", SessionID: "live-1"}, Threads: threads})
		Expect(mixed).To(MatchError(comments.ErrInvalid))
		_, err := svc.DeliverBatch(ctx, comments.DeliverBatchRequest{RequestID: "mixed-edits-chosen", Target: comments.DeliverTarget{Kind: "session", SessionID: "live-1", Edits: comments.EditsDirect}, Threads: threads})
		Expect(err).NotTo(HaveOccurred())
		Expect([]string{runOf(ctx, proposed.ID)().Edits, runOf(ctx, direct.ID)().Edits}).To(Equal([]string{comments.EditsDirect, comments.EditsDirect}))
	})

	It("records a background delivery that failed resolving the page's rule on the thread, with the rule's error", func(ctx SpecContext) {
		explicit := comments.DeliverTarget{Kind: "session", SessionID: "live-1"}
		c, err := svc.Create(ctx, comments.CreateRequest{Target: pageTarget("broken-rule"), Body: "Fix the nav", WorkflowID: "human-review", RequestID: "create-broken-rule", Deliver: &explicit})
		Expect(err).NotTo(HaveOccurred())
		Eventually(runOf(ctx, c.ID)).WithTimeout(10 * time.Second).Should(And(HaveField("State", "error"), HaveField("Error", ContainSubstring("rule store offline"))))
	})

	It("records a queued delivery that reached the host without an edits policy as an errored agentRun on its thread", func(ctx SpecContext) {
		c := create(ctx, "home", "Queued before edits were resolved")
		op, err := svc.Queue(ctx, comments.QueueRequest{CommentID: c.ID, ExpectedVersion: c.Version, Kind: comments.DeliverAgent, RequestID: "queued-no-edits", Payload: json.RawMessage(`{"kind":"session","sessionId":"live-1"}`)})
		Expect(err).NotTo(HaveOccurred())
		_, err = svc.Process(ctx, comments.ProcessRequest{Executor: newDelivery(svc), Lease: time.Minute, OperationID: op.ID})
		Expect(err).NotTo(HaveOccurred())
		Expect(runOf(ctx, c.ID)()).To(And(HaveField("State", "error"), HaveField("Error", ContainSubstring("edits policy"))))
	})
})
