package postgres_test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/comments"
	"github.com/flanksource/gavel/comments/postgres"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"time"
)

type fixtureProvider struct {
	writes int
	remote comments.RemoteComment
}

func (*fixtureProvider) Name() string { return "fixture" }
func (p *fixtureProvider) Fetch(context.Context, comments.Principal, comments.Target) ([]comments.RemoteComment, error) {
	return []comments.RemoteComment{p.remote}, nil
}
func (p *fixtureProvider) Apply(_ context.Context, _ comments.Principal, action string, publication comments.Publication) (comments.RemoteComment, error) {
	p.writes++
	p.remote.Body = publication.Comment.Body
	p.remote.Binding.Revision = fmt.Sprint(p.writes)
	if action == "resolve-native" {
		p.remote.Resolved = true
	}
	if action == "reopen-native" {
		p.remote.Resolved = false
	}
	if action == "publish-reply" {
		p.remote.Messages = append(p.remote.Messages, comments.RemoteMessage{ID: "reply-1", Body: publication.Message.Body, Author: comments.Actor{Kind: "user", ID: "provider-user"}, CreatedAt: time.Now().UTC()})
		p.remote.Binding.MessageIDs = map[string]string{"reply-1": publication.Message.ID}
	}
	return p.remote, nil
}

var _ = Describe("explicit provider publication", func() {
	It("publishes only admitted snapshots and synchronizes native state without echoing writes", func(ctx SpecContext) {
		db := dbtest.ForGinkgo(dbtest.Options{Name: "comment_providers"})
		Expect(postgres.Provision(ctx, db.DSN())).To(Succeed())
		store, err := postgres.New(db.Gorm())
		Expect(err).NotTo(HaveOccurred())
		a := access{tenant: "publisher"}
		provider := &fixtureProvider{remote: comments.RemoteComment{Binding: comments.Binding{Provider: "fixture", RemoteKey: "repo/pr/1/thread/1", ThreadID: "thread-1", RootID: "root-1", Revision: "0", Resolvable: true}, Body: "Remote concern", Author: comments.Actor{Kind: "user", ID: "provider-user"}, CreatedAt: time.Now().UTC()}}
		svc, err := comments.New(comments.Options{Store: store, Access: a, Targets: a, Workflows: []comments.Workflow{comments.HumanReview(), comments.NativePR()}, Providers: []comments.Provider{provider}})
		Expect(err).NotTo(HaveOccurred())
		target := comments.Target{URI: "https://example.com/repo/pull/1", ScopeURI: "git://repo", Kind: "pull-request", Revision: "head-sha"}
		c, err := svc.Create(ctx, comments.CreateRequest{Target: target, Body: "Check this error", WorkflowID: comments.NativePR().ID, RequestID: "draft"})
		Expect(err).NotTo(HaveOccurred())
		Expect(provider.writes).To(BeZero())
		_, err = svc.Publish(ctx, comments.PublicationRequest{CommentID: c.ID, ExpectedVersion: 1, RequestID: "post", Provider: "fixture", Action: "publish"})
		Expect(err).NotTo(HaveOccurred())
		Expect(provider.writes).To(BeZero())
		_, err = svc.Act(ctx, comments.ActionRequest{ID: c.ID, ExpectedVersion: 1, Action: "update", Body: "changed after publish"})
		Expect(err).To(MatchError(comments.ErrConflict))
		op, err := svc.ProcessProvider(ctx, time.Minute)
		Expect(err).NotTo(HaveOccurred())
		Expect(op.State).To(Equal("done"))
		Expect(provider.writes).To(Equal(1))
		replayed, err := svc.Publish(ctx, comments.PublicationRequest{CommentID: c.ID, ExpectedVersion: 1, RequestID: "post", Provider: "fixture", Action: "publish"})
		Expect(err).NotTo(HaveOccurred())
		Expect(replayed.ID).To(Equal(op.ID))
		Expect(replayed.State).To(Equal("done"))
		Expect(provider.writes).To(Equal(1))
		synced, err := svc.Sync(ctx, "fixture", target)
		Expect(err).NotTo(HaveOccurred())
		Expect(synced).To(HaveLen(1))
		Expect(synced[0].ID).To(Equal(c.ID))
		_, err = svc.Act(ctx, comments.ActionRequest{ID: c.ID, ExpectedVersion: synced[0].Version, Action: "resolve"})
		Expect(err).To(MatchError(comments.ErrConflict))
		Expect(provider.writes).To(Equal(1))
		c, err = svc.Act(ctx, comments.ActionRequest{ID: c.ID, ExpectedVersion: synced[0].Version, Action: "reply", Body: "Updated", Content: json.RawMessage(`{"capture":{"revision":"head-sha"}}`)})
		Expect(err).NotTo(HaveOccurred())
		_, err = svc.Publish(ctx, comments.PublicationRequest{CommentID: c.ID, ExpectedVersion: c.Version, RequestID: "reply", Provider: "fixture", Action: "publish-reply", MessageID: c.Messages[0].ID})
		Expect(err).NotTo(HaveOccurred())
		op, err = svc.ProcessProvider(ctx, time.Minute)
		Expect(err).NotTo(HaveOccurred())
		Expect(op.State).To(Equal("done"))
		c, err = svc.Get(ctx, c.ID)
		Expect(err).NotTo(HaveOccurred())
		Expect(c.Messages).To(HaveLen(1))
		Expect(c.Messages[0].Body).To(Equal("Updated"))
		Expect(c.Messages[0].Content).To(MatchJSON(`{"capture":{"revision":"head-sha"}}`))
		_, err = svc.Publish(ctx, comments.PublicationRequest{CommentID: c.ID, ExpectedVersion: c.Version, RequestID: "resolve", Provider: "fixture", Action: "resolve-native"})
		Expect(err).NotTo(HaveOccurred())
		op, err = svc.ProcessProvider(ctx, time.Minute)
		Expect(err).NotTo(HaveOccurred())
		Expect(op.State).To(Equal("done"))
		c, err = svc.Get(ctx, c.ID)
		Expect(err).NotTo(HaveOccurred())
		Expect(c.Phase).To(Equal(comments.Final))
		Expect(c.Outcome.Actor).To(Equal(comments.Actor{Kind: "system", ID: "provider:fixture"}))
	})
})
