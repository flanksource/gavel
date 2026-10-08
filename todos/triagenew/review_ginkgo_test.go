package triagenew

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"testing"
)

type reviewProvider struct {
	todos.Provider
	items   map[string]*types.TODO
	applied int
}

func (p *reviewProvider) Get(_ context.Context, ref string) (*types.TODO, error) {
	if todo := p.items[ref]; todo != nil {
		copy := *todo
		return &copy, nil
	}
	return nil, fmt.Errorf("unknown TODO %s", ref)
}
func (p *reviewProvider) ValidateTriageNew(_ context.Context, _ *types.TODO, env *types.TriageNewEnvelope) error {
	return env.Validate()
}
func (p *reviewProvider) ApplyTriageNew(_ context.Context, _ *types.TODO, _ *types.TriageNewEnvelope, _ todos.TriageNewApplyOptions) error {
	p.applied++
	return nil
}

func TestReview(t *testing.T) { RegisterFailHandler(Fail); RunSpecs(t, "Triage new review") }

var _ = Describe("Triage new review", func() {
	var provider *reviewProvider
	var review *Review
	var env types.TriageNewEnvelope
	BeforeEach(func() {
		provider = &reviewProvider{items: map[string]*types.TODO{"source": {ID: "source", Version: 1}, "target": {ID: "target", Version: 2}}}
		review = &Review{Provider: provider, SourceID: "source", Approved: func(context.Context, map[string]any) error { return nil }}
		env = types.TriageNewEnvelope{ResultEnvelope: types.ResultEnvelope{Summary: "Same scope", EndStatus: types.EndCompleted}, Title: "Repair parser", Labels: []string{"bug"}, Action: "child-of", Target: "target", Rationale: "Part of parser repair"}
	})
	prepare := func() Proposal {
		result, err := review.prepare(context.Background(), toMap(env))
		Expect(err).NotTo(HaveOccurred())
		return result.(Proposal)
	}
	It("applies only the exact reviewed proposal", func() {
		proposal := prepare()
		_, err := review.review(context.Background(), toMap(proposal))
		Expect(err).NotTo(HaveOccurred())
		Expect(review.Apply(context.Background(), provider.items["source"], &proposal.Result)).To(Succeed())
		Expect(provider.applied).To(Equal(1))
	})
	It("holds an unreviewed proposal without writing", func() {
		proposal := prepare()
		Expect(review.Apply(context.Background(), provider.items["source"], &proposal.Result)).To(MatchError(ContainSubstring("approved")))
		Expect(provider.applied).To(BeZero())
	})
	It("rejects bypass of the durable approval gate", func() {
		proposal := prepare()
		review.Approved = func(context.Context, map[string]any) error { return fmt.Errorf("approval denied") }
		_, err := review.review(context.Background(), toMap(proposal))
		Expect(err).To(HaveOccurred())
		Expect(review.Apply(context.Background(), provider.items["source"], &proposal.Result)).To(HaveOccurred())
		Expect(provider.applied).To(BeZero())
	})
	It("rejects changed proposal input", func() {
		proposal := prepare()
		proposal.Result.Target = "elsewhere"
		_, err := review.review(context.Background(), toMap(proposal))
		Expect(err).To(HaveOccurred())
		Expect(provider.applied).To(BeZero())
	})
	It("rejects a changed target before applying", func() {
		proposal := prepare()
		_, err := review.review(context.Background(), toMap(proposal))
		Expect(err).NotTo(HaveOccurred())
		provider.items["target"].Version++
		Expect(review.Apply(context.Background(), provider.items["source"], &proposal.Result)).To(MatchError(ContainSubstring("changed since review")))
		Expect(provider.applied).To(BeZero())
	})
	It("rejects final output changed after review", func() {
		proposal := prepare()
		_, err := review.review(context.Background(), toMap(proposal))
		Expect(err).NotTo(HaveOccurred())
		proposal.Result.Title = "Changed scope"
		Expect(review.Apply(context.Background(), provider.items["source"], &proposal.Result)).To(HaveOccurred())
		Expect(provider.applied).To(BeZero())
	})
	It("makes no write after cancellation", func() {
		proposal := prepare()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		Expect(review.Apply(ctx, provider.items["source"], &proposal.Result)).To(HaveOccurred())
		Expect(provider.applied).To(BeZero())
	})
})

func toMap(value any) map[string]any {
	var input map[string]any
	raw, err := json.Marshal(value)
	Expect(err).NotTo(HaveOccurred())
	Expect(json.Unmarshal(raw, &input)).To(Succeed())
	return input
}
