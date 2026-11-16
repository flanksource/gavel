package agents_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/comments"
	commentapi "github.com/flanksource/gavel/comments/clicky"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Delivering comment threads to the agents their body mentions", Ordered, func() {
	proposeRule := comments.Route{Source: "db:path:pages/proposals", Target: comments.DeliverTarget{Kind: "run", Prompt: "editor-comments", Edits: comments.EditsPropose}}
	assignRule := comments.Route{Source: "db:path:pages/assigned", Target: comments.DeliverTarget{Kind: "session", SessionID: "live-2"}, Assign: true}
	var f *fixture
	BeforeAll(func(ctx SpecContext) {
		f = newFixture(ctx, "comment_agent_mentions", pageRules{pageRoutes: pageRoutes{}, rules: map[string]comments.Route{"proposals": proposeRule, "assigned": assignRule}})
	})
	sessionOf := func(ctx context.Context, id string) func() string {
		return func() string { return f.runOf(ctx, id)().SessionID }
	}

	It("delivers a root mentioning @claude as a fresh run of claude's default agent model under the page rule's prompt and edits", func(ctx SpecContext) {
		_, before := f.bridge.Snapshot()
		c := f.create(ctx, "proposals", "@claude tighten the hero copy")
		Eventually(func() string { return f.runOf(ctx, c.ID)().State }).WithTimeout(10 * time.Second).Should(Equal("sent"))
		_, launched := f.bridge.Snapshot()
		Expect(launched).To(HaveLen(len(before) + 1))
		run := launched[len(launched)-1]
		var vars map[string]any
		Expect(json.Unmarshal([]byte(run.Options.Vars), &vars)).To(Succeed())
		Expect([]any{run.Reference, run.Options.Spec.Name, run.Options.Spec.Mode, vars["propose"], f.runOf(ctx, c.ID)().Edits}).
			To(Equal([]any{"editor-comments", agentDefaultModel("claude"), api.ModeAgent, true, comments.EditsPropose}))
	})

	It("delivers @agent to the page rule, @<session slug> to that session, and a reply's mention to its thread", func(ctx SpecContext) {
		_, before := f.bridge.Snapshot()
		ruled := f.create(ctx, "proposals", "@agent shorten the footer")
		Eventually(func() string { return f.runOf(ctx, ruled.ID)().State }).WithTimeout(10 * time.Second).Should(Equal("sent"))
		_, launched := f.bridge.Snapshot()

		toSlug := f.create(ctx, "mentions", "Please @rename-pass rename the button")
		Eventually(sessionOf(ctx, toSlug.ID)).WithTimeout(10 * time.Second).Should(Equal("live-1"))
		thread := f.create(ctx, "mentions", "The table overflows")
		_, err := f.svc.Act(ctx, comments.ActionRequest{ID: thread.ID, ExpectedVersion: thread.Version, Action: "reply", Body: "@docs-pass can you take this?"})
		Expect(err).NotTo(HaveOccurred())
		Eventually(sessionOf(ctx, thread.ID)).WithTimeout(10 * time.Second).Should(Equal("live-2"))
		Expect([]any{len(launched) - len(before), launched[len(launched)-1].Reference, launched[len(launched)-1].Options.Spec.Name}).To(Equal([]any{1, "editor-comments", ""}))
	})

	It("assigns a thread mentioning @agent to the rule's agent before delivering it, as the rule's own route does", func(ctx SpecContext) {
		stages, release := make(chan string), make(chan struct{})
		f.bridge.SetHold(func(stage string) {
			if stage == "live" {
				stages <- stage
				<-release
			}
		})
		DeferCleanup(func() { f.bridge.SetHold(nil) })
		c := f.create(ctx, "assigned", "@agent check the table")
		Eventually(stages).WithTimeout(10 * time.Second).Should(Receive(Equal("live")))
		Expect(*f.get(ctx, c.ID).Assignee).To(Equal(comments.RouteAgent(assignRule)))
		close(release)
		Eventually(sessionOf(ctx, c.ID)).WithTimeout(10 * time.Second).Should(Equal("live-2"))
	})

	It("delivers a batch reply's mention once the batch commits, and nothing for a batch that rolls back", func(ctx SpecContext) {
		committed, rolledBack := f.create(ctx, "batch-mentions", "The chart is cut off"), f.create(ctx, "batch-mentions", "The legend overlaps")
		_, err := f.svc.Batch(ctx, []comments.ActionRequest{{ID: committed.ID, ExpectedVersion: committed.Version, Action: "reply", Body: "@docs-pass please fix"}})
		Expect(err).NotTo(HaveOccurred())
		Eventually(sessionOf(ctx, committed.ID)).WithTimeout(10 * time.Second).Should(Equal("live-2"))

		_, err = f.svc.Batch(ctx, []comments.ActionRequest{
			{ID: rolledBack.ID, ExpectedVersion: rolledBack.Version, Action: "reply", Body: "@docs-pass please fix"},
			{ID: rolledBack.ID, ExpectedVersion: rolledBack.Version, Action: "reply", Body: "a stale second reply"},
		})
		Expect(err).To(MatchError(comments.ErrConflict))
		Consistently(func() []any { return []any{f.get(ctx, rolledBack.ID).Version, f.runOf(ctx, rolledBack.ID)().State} }).
			WithTimeout(time.Second).Should(Equal([]any{rolledBack.Version, ""}))
	})

	It("keeps unknown and agent-written mentions as prose, refuses two agents in one body, and lets an explicit target win", func(ctx SpecContext) {
		prose := f.create(ctx, "mentions", "ping @nobody about it")
		byAgent := f.create(commentapi.WithActor(ctx, comments.Actor{Kind: "agent", ID: "claude:live-1"}), "mentions", "Handing over to @docs-pass")
		two := pageTarget("mentions-two")
		_, err := f.svc.Create(ctx, comments.CreateRequest{Target: two, Body: "@claude or @rename-pass", WorkflowID: "human-review", RequestID: "create-two-mentions"})
		Expect(err).To(MatchError(comments.ErrInvalid))
		listed, err := f.svc.List(ctx, comments.Filter{ScopeURI: scope, TargetURI: two.URI, Limit: 10})
		Expect(err).NotTo(HaveOccurred())
		_, replyErr := f.svc.Act(ctx, comments.ActionRequest{ID: prose.ID, ExpectedVersion: prose.Version, Action: "reply", Body: "@claude and @docs-pass"})
		Expect([]any{len(listed), replyErr != nil && errors.Is(replyErr, comments.ErrInvalid), f.get(ctx, prose.ID).Version}).To(Equal([]any{0, true, prose.Version}))

		explicit := comments.DeliverTarget{Kind: "session", SessionID: "live-2"}
		c, err := f.svc.Create(ctx, comments.CreateRequest{Target: pageTarget("mentions-explicit"), Body: "@rename-pass should not get this", WorkflowID: "human-review", RequestID: "create-explicit-mention", Content: json.RawMessage(`{}`), Deliver: &explicit})
		Expect(err).NotTo(HaveOccurred())
		Eventually(sessionOf(ctx, c.ID)).WithTimeout(10 * time.Second).Should(Equal("live-2"))
		Consistently(func() []string { return []string{f.runOf(ctx, prose.ID)().State, f.runOf(ctx, byAgent.ID)().State} }).
			WithTimeout(time.Second).Should(Equal([]string{"", ""}))
	})

	It("matches @agent and runtimes without the scope's workdir, lists sessions only for an unmatched token, and has none in a scope without a workdir", func(ctx SpecContext) {
		createOn := func(svc *comments.Service, page, body string) error {
			_, err := svc.Create(ctx, comments.CreateRequest{Target: pageTarget(page), Body: body, WorkflowID: "human-review", RequestID: "create-workdir-" + page + "-" + body})
			return err
		}
		broken := f.service(comments.Options{Delivery: f.delivery(func(context.Context, comments.Target) (string, error) {
			return "", errors.New("project catalog offline")
		})})
		Expect(createOn(broken, "workdir", "@claude tighten it")).To(Succeed())
		Expect(createOn(broken, "workdir", "@nobody tighten it")).To(MatchError(ContainSubstring("project catalog offline")))

		remote := f.service(comments.Options{Delivery: f.delivery(func(context.Context, comments.Target) (string, error) {
			return "", fmt.Errorf("%w: a web project", comments.ErrNoWorkdir)
		})})
		Expect(createOn(remote, "no-workdir", "@rename-pass take it")).To(Succeed())
		listed, err := remote.List(ctx, comments.Filter{ScopeURI: scope, TargetURI: pageTarget("no-workdir").URI, Limit: 10})
		Expect(err).NotTo(HaveOccurred())
		Expect(listed).To(HaveLen(1))
		Consistently(func() string { return contentOf(f.get(ctx, listed[0].ID)).AgentRun.State }).WithTimeout(time.Second).Should(Equal(""), "a session mention in a scope without a workdir is prose")
	})

	It("stores a create's attachments before its mention delivers it, so the agent's message lists them", func(ctx SpecContext) {
		attaching := f.service(comments.Options{Delivery: f.newDelivery, Blobs: f.store, AttachmentLimit: 1 << 10})
		createWith := func(requestID string, attachment comments.CreateAttachment) (comments.Comment, error) {
			return attaching.Create(ctx, comments.CreateRequest{Target: pageTarget("screenshots"), Body: "@rename-pass the logo is clipped", WorkflowID: "human-review", RequestID: requestID, Attachments: []comments.CreateAttachment{attachment}})
		}
		before, _ := f.bridge.Snapshot()
		c, err := createWith("create-screenshot", comments.CreateAttachment{Filename: "clipped.png", ContentType: "image/png", Data: []byte("png bytes")})
		Expect(err).NotTo(HaveOccurred())
		Eventually(sessionOf(ctx, c.ID)).WithTimeout(10 * time.Second).Should(Equal("live-1"))
		sent, _ := f.bridge.Snapshot()
		Expect(sent).To(HaveLen(len(before) + 1))
		Expect([]any{len(c.Attachments), c.Attachments[0].Name, c.Version}).To(Equal([]any{1, "clipped.png", int64(2)}))
		Expect(sent[len(sent)-1].Text).To(ContainSubstring("clipped.png"))

		_, badType := createWith("create-bad-type", comments.CreateAttachment{Filename: "x.png", ContentType: "not a type", Data: []byte("x")})
		_, tooBig := createWith("create-too-big", comments.CreateAttachment{Filename: "x.png", ContentType: "image/png", Data: make([]byte, 2<<10)})
		listed, err := attaching.List(ctx, comments.Filter{ScopeURI: scope, TargetURI: pageTarget("screenshots").URI, Limit: 10})
		Expect(err).NotTo(HaveOccurred())
		Expect([]any{errors.Is(badType, comments.ErrInvalid), errors.Is(tooBig, comments.ErrInvalid), len(listed)}).To(Equal([]any{true, true, 1}))
	})
})
