package postgres_test

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/comments"
	"github.com/flanksource/gavel/comments/postgres"
	"github.com/flanksource/gavel/internal/database"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// fakeProposals is a Proposals host keeping patches in memory by ref.
type fakeProposals struct {
	patches   map[string]string
	edits     map[string][]comments.FileEdit
	proposed  []string
	discarded []string
	applied   []bool
	applyErr  error
}

func (f *fakeProposals) Propose(_ context.Context, c comments.Comment, replyID string, input comments.ProposeInput) (comments.Proposal, error) {
	f.proposed = append(f.proposed, replyID)
	ref := "refs/gavel/comments/" + c.ID + "/" + replyID
	f.patches[ref], f.edits[ref] = string(input.Patch), input.Edits
	return comments.Proposal{Ref: ref, Base: "base", Commit: fmt.Sprintf("abc%04d0000", len(f.proposed)), Files: []comments.ProposalFile{{Path: "a.go", Additions: 1, Deletions: 1}}, State: comments.ProposalProposed}, nil
}
func (f *fakeProposals) Discard(_ context.Context, _ comments.Comment, p comments.Proposal) error {
	f.discarded = append(f.discarded, p.Ref)
	return nil
}
func (f *fakeProposals) Diff(_ context.Context, _ comments.Comment, p comments.Proposal) (comments.ProposalDiff, error) {
	return comments.ProposalDiff{Diff: f.patches[p.Ref], Commit: p.Commit}, nil
}
func (f *fakeProposals) Apply(_ context.Context, _ comments.Comment, _ comments.Proposal, stage bool) error {
	f.applied = append(f.applied, stage)
	return f.applyErr
}

// revisionDelivery records the deliver-agent operations a revision request queues.
type revisionDelivery struct{ targets []comments.DeliverTarget }

func (d *revisionDelivery) Execute(_ context.Context, op comments.Operation) (json.RawMessage, error) {
	var target comments.DeliverTarget
	if err := json.Unmarshal(op.Payload, &target); err != nil {
		return nil, err
	}
	d.targets = append(d.targets, target)
	return json.RawMessage(`{}`), nil
}
func (d *revisionDelivery) Check(comments.DeliverTarget) error { return nil }
func (d *revisionDelivery) Sessions(context.Context, comments.SessionQuery) ([]comments.AgentSession, error) {
	return nil, nil
}
func (d *revisionDelivery) Preview(context.Context, comments.Comment, comments.DeliverTarget) (json.RawMessage, error) {
	return nil, nil
}
func (d *revisionDelivery) RecordFailure(context.Context, string, comments.DeliverTarget, error) error {
	return nil
}

func proposalOf(m comments.Message) comments.Proposal {
	GinkgoHelper()
	var content struct {
		Proposal comments.Proposal `json:"proposal"`
	}
	Expect(json.Unmarshal(m.Content, &content)).To(Succeed(), string(m.Content))
	return content.Proposal
}

func statesOf(c comments.Comment) []comments.ProposalState {
	GinkgoHelper()
	states := []comments.ProposalState{}
	for _, m := range c.Messages {
		if m.ParentID == "" {
			states = append(states, proposalOf(m).State)
		}
	}
	return states
}

var _ = Describe("proposed patches on replies", func() {
	const tenant, scope = "proposals", "git://repo"
	agent := comments.Actor{Kind: "agent", ID: "claude:live-1"}
	var (
		store     *postgres.Store
		proposals *fakeProposals
		delivery  *revisionDelivery
		human     *comments.Service
		bot       *comments.Service
	)
	newService := func(a comments.Access, opts comments.Options) *comments.Service {
		GinkgoHelper()
		opts.Store, opts.Access, opts.Targets, opts.Workflows = store, a, access{tenant: tenant}, []comments.Workflow{comments.HumanReview()}
		svc, err := comments.New(opts)
		Expect(err).NotTo(HaveOccurred())
		return svc
	}
	create := func(body string) comments.Comment {
		GinkgoHelper()
		c, err := human.Create(context.Background(), comments.CreateRequest{Target: comments.Target{URI: scope + "/a.go", ScopeURI: scope, Kind: "code"}, Body: body, WorkflowID: "human-review", RequestID: body})
		Expect(err).NotTo(HaveOccurred())
		return c
	}
	propose := func(svc *comments.Service, c comments.Comment, patch string) comments.Comment {
		GinkgoHelper()
		out, err := svc.Act(context.Background(), comments.ActionRequest{ID: c.ID, ExpectedVersion: c.Version, Action: "reply", Body: "Proposed fix", Patch: patch})
		Expect(err).NotTo(HaveOccurred())
		return out
	}
	act := func(svc *comments.Service, c comments.Comment, req comments.ActionRequest) (comments.Comment, error) {
		req.ID, req.ExpectedVersion = c.ID, c.Version
		return svc.Act(context.Background(), req)
	}
	last := func(c comments.Comment) comments.Message { return c.Messages[len(c.Messages)-1] }

	BeforeEach(func(ctx SpecContext) {
		db := dbtest.ForGinkgo(dbtest.Options{Name: "comment_proposals"})
		Expect(database.Migrate(ctx, db.DSN())).To(Succeed())
		var err error
		store, err = postgres.New(db.Gorm())
		Expect(err).NotTo(HaveOccurred())
		proposals, delivery = &fakeProposals{patches: map[string]string{}, edits: map[string][]comments.FileEdit{}}, &revisionDelivery{}
		opts := comments.Options{Proposals: proposals, Delivery: func(*comments.Service) comments.Delivery { return delivery }}
		human = newService(access{tenant: tenant}, opts)
		bot = newService(actorAccess{access: access{tenant: tenant}, actor: agent}, opts)
	})

	It("stores a reply's proposal, supersedes the earlier one, discards the ref of a reply never stored and serves its diff", func(ctx SpecContext) {
		c := create("Check the error")
		first := propose(bot, c, "patch one")
		Expect(proposals.proposed).To(Equal([]string{last(first).ID}))
		Expect(proposalOf(last(first))).To(Equal(comments.Proposal{Ref: "refs/gavel/comments/" + c.ID + "/" + last(first).ID, Base: "base", Commit: "abc00010000", Files: []comments.ProposalFile{{Path: "a.go", Additions: 1, Deletions: 1}}, State: comments.ProposalProposed}))

		second := propose(bot, first, "patch two")
		Expect(statesOf(second)).To(Equal([]comments.ProposalState{comments.ProposalSuperseded, comments.ProposalProposed}))
		stored, err := human.Get(ctx, c.ID)
		Expect(err).NotTo(HaveOccurred())
		Expect(statesOf(stored)).To(Equal(statesOf(second)))

		_, err = act(bot, first, comments.ActionRequest{Action: "reply", Body: "Stale", Patch: "patch three"})
		Expect(err).To(MatchError(comments.ErrConflict))
		Expect(proposals.discarded).To(Equal([]string{"refs/gavel/comments/" + c.ID + "/" + proposals.proposed[2]}))

		diff, err := human.ProposalDiff(ctx, c.ID, last(second).ID)
		Expect(err).NotTo(HaveOccurred())
		Expect(diff).To(Equal(comments.ProposalDiff{Diff: "patch two", Commit: "abc00020000"}))

		bare := newService(access{tenant: tenant}, comments.Options{})
		_, err = act(bare, second, comments.ActionRequest{Action: "reply", Body: "Fix", Patch: "patch"})
		Expect(err).To(MatchError(ContainSubstring("this host cannot hold proposed patches")))
	})

	It("hands a reply's edits to the host and refuses edits with a patch or outside a reply", func() {
		c := create("Rename the variable")
		edits := []comments.FileEdit{{Path: "a.go", OldString: "x := 1", NewString: "count := 1"}}
		out, err := act(bot, c, comments.ActionRequest{Action: "reply", Body: "Renamed", Edits: edits})
		Expect(err).NotTo(HaveOccurred())
		Expect(proposals.edits[proposalOf(last(out)).Ref]).To(Equal(edits))

		_, err = act(bot, out, comments.ActionRequest{Action: "reply", Body: "Both", Patch: "patch", Edits: edits})
		Expect(err).To(MatchError(ContainSubstring("a patch or edits, not both")))
		_, err = act(bot, out, comments.ActionRequest{Action: "resolve", Summary: "Done", Edits: edits})
		Expect(err).To(MatchError(ContainSubstring("only a reply carries a patch or edits")))
		_, err = act(bot, out, comments.ActionRequest{Action: "reply", Body: "Empty", Edits: []comments.FileEdit{}})
		Expect(err).To(MatchError(ContainSubstring("edits must list at least one edit")))
		Expect(proposals.proposed).To(HaveLen(1))
	})

	It("accepts by applying or staging and resolves the thread, leaving it untouched on a conflict, and refuses agents", func(ctx SpecContext) {
		c := propose(bot, create("Check the error"), "patch one")
		reply := last(c).ID
		capabilities := func(svc *comments.Service) []string {
			GinkgoHelper()
			caps, err := svc.Capabilities(ctx, c.ID)
			Expect(err).NotTo(HaveOccurred())
			return caps.Actions
		}
		Expect(capabilities(human)).To(ContainElements(comments.ActionAcceptProposal, comments.ActionRejectProposal))
		Expect(capabilities(human)).NotTo(ContainElement(comments.ActionRequestRevision), "no agent session to revise")
		Expect(capabilities(bot)).NotTo(ContainElement(comments.ActionAcceptProposal))

		_, err := act(bot, c, comments.ActionRequest{Action: comments.ActionAcceptProposal, MessageID: reply, Mode: comments.ProposalModeApply})
		Expect(err).To(MatchError(comments.ErrForbidden))
		_, err = act(human, c, comments.ActionRequest{Action: comments.ActionAcceptProposal, MessageID: reply, Mode: "merge"})
		Expect(err).To(MatchError(comments.ErrInvalid))
		proposals.applyErr = fmt.Errorf("%w: a.go does not apply", comments.ErrConflict)
		_, err = act(human, c, comments.ActionRequest{Action: comments.ActionAcceptProposal, MessageID: reply, Mode: comments.ProposalModeApply})
		Expect(err).To(MatchError(comments.ErrConflict))
		unchanged, err := human.Get(ctx, c.ID)
		Expect(err).NotTo(HaveOccurred())
		Expect([]any{unchanged.Version, unchanged.Status, statesOf(unchanged)}).To(Equal([]any{c.Version, "open", []comments.ProposalState{comments.ProposalProposed}}))

		proposals.applyErr = nil
		applied, err := act(human, c, comments.ActionRequest{Action: comments.ActionAcceptProposal, MessageID: reply, Mode: comments.ProposalModeApply})
		Expect(err).NotTo(HaveOccurred())
		decided := proposalOf(last(applied))
		Expect([]any{decided.State, *decided.DecidedBy, applied.Status, applied.Resolution.Summary}).To(Equal([]any{comments.ProposalApplied, comments.Actor{Kind: "user", ID: "reviewer"}, "resolved", "Applied proposal abc0001"}))

		staged, err := act(human, propose(bot, create("Name the constant"), "patch two"), comments.ActionRequest{Action: comments.ActionAcceptProposal, MessageID: proposals.proposed[1], Mode: comments.ProposalModeStage})
		Expect(err).NotTo(HaveOccurred())
		Expect([]any{proposalOf(last(staged)).State, staged.Status, staged.Resolution.Summary}).To(Equal([]any{comments.ProposalStaged, "resolved", "Staged proposal abc0002"}))
		Expect(proposals.applied).To(Equal([]bool{false, false, true}))
	})

	It("rejects, asks for revisions redelivered to the agent session, and holds an agent's resolve until a proposal is accepted", func(ctx SpecContext) {
		c, err := act(human, create("Check the error"), comments.ActionRequest{Action: "metadata", Content: json.RawMessage(`{"agentRun":{"sessionId":"live-1","edits":"propose"}}`)})
		Expect(err).NotTo(HaveOccurred())
		_, err = act(bot, c, comments.ActionRequest{Action: "resolve"})
		Expect(err).To(MatchError(ContainSubstring("this thread requires a proposal; reply with edits or a patch")))

		c = propose(bot, c, "patch one")
		rejected, err := act(human, c, comments.ActionRequest{Action: comments.ActionRejectProposal, MessageID: last(c).ID, Body: "Wrong file"})
		Expect(err).NotTo(HaveOccurred())
		Expect([]any{proposalOf(last(rejected)).State, proposalOf(last(rejected)).Feedback, rejected.Status}).To(Equal([]any{comments.ProposalRejected, "Wrong file", "open"}))
		_, err = act(human, rejected, comments.ActionRequest{Action: comments.ActionRequestRevision, MessageID: last(c).ID, Body: "Again"})
		Expect(err).To(MatchError(comments.ErrConflict))

		c = propose(bot, rejected, "patch two")
		reply := last(c).ID
		caps, err := human.Capabilities(ctx, c.ID)
		Expect(err).NotTo(HaveOccurred())
		Expect(caps.Actions).To(ContainElement(comments.ActionRequestRevision))
		_, err = act(human, c, comments.ActionRequest{Action: comments.ActionRequestRevision, MessageID: reply})
		Expect(err).To(MatchError(comments.ErrInvalid))
		revised, err := act(human, c, comments.ActionRequest{Action: comments.ActionRequestRevision, MessageID: reply, Body: "Rename it"})
		Expect(err).NotTo(HaveOccurred())
		feedback := last(revised)
		Expect([]any{proposalOf(revised.Messages[len(revised.Messages)-2]).State, feedback.ParentID, feedback.Body}).To(Equal([]any{comments.ProposalRevisionRequested, reply, "Rename it"}))
		Expect(delivery.targets).To(Equal([]comments.DeliverTarget{{Kind: "session", SessionID: "live-1", Edits: comments.EditsPropose}}))

		unrun := propose(bot, create("Name the constant"), "patch three")
		_, err = act(human, unrun, comments.ActionRequest{Action: comments.ActionRequestRevision, MessageID: last(unrun).ID, Body: "Rename it"})
		Expect(err).To(MatchError(ContainSubstring("no agent session")))
		stored, err := human.Get(ctx, unrun.ID)
		Expect(err).NotTo(HaveOccurred())
		Expect(stored.Version).To(Equal(unrun.Version))

		c = propose(bot, revised, "patch four")
		accepted, err := act(human, c, comments.ActionRequest{Action: comments.ActionAcceptProposal, MessageID: last(c).ID, Mode: comments.ProposalModeApply})
		Expect(err).NotTo(HaveOccurred())
		reopened, err := act(human, accepted, comments.ActionRequest{Action: "reopen"})
		Expect(err).NotTo(HaveOccurred())
		resolved, err := act(bot, reopened, comments.ActionRequest{Action: "resolve", Summary: "Done"})
		Expect(err).NotTo(HaveOccurred())
		Expect(resolved.Status).To(Equal("resolved"))
	})
})
