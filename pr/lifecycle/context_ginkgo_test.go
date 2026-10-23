package lifecycle_test

import (
	"testing"
	"time"

	git "github.com/flanksource/commons-db/git"
	"github.com/flanksource/gavel/lifecycle"
	prlifecycle "github.com/flanksource/gavel/pr/lifecycle"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestPRLifecycle(t *testing.T) { RegisterFailHandler(Fail); RunSpecs(t, "PR lifecycle") }

var _ = Describe("PR lifecycle applicability", func() {
	It("orders edited comments by their most recent activity", func() {
		definition, err := prlifecycle.Default()
		Expect(err).NotTo(HaveOccurred())
		engine, err := lifecycle.New(definition)
		Expect(err).NotTo(HaveOccurred())
		created := time.Date(2026, 10, 21, 10, 0, 0, 0, time.UTC)
		snapshot := prlifecycle.Snapshot{PR: git.PullRequest{Number: 7, URL: "https://git.example/acme/cli/pull/7", State: git.PullRequestOpen, HeadSHA: "current"}, Activity: []git.PullRequestActivity{
			{ID: "edited", Kind: git.PullRequestComment, CreatedAt: created, UpdatedAt: created.Add(2 * time.Hour)},
			{ID: "newer", Kind: git.PullRequestComment, CreatedAt: created.Add(time.Hour)},
		}}
		actions, err := prlifecycle.Actions(engine, snapshot, nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(actions[0].Event.ID).To(HavePrefix("comment:edited:"))
		Expect(actions[0].Event.At).To(Equal(created.Add(2 * time.Hour)))
	})
	It("offers only the first matching step for a mentioned comment", func() {
		def := lifecycle.Lifecycle{Name: "pr", Subject: map[string]string{"number": "int"}, Steps: []lifecycle.Step{
			{Name: "mention", Prompt: "pr.fix", Events: []string{"completed"}, When: `event.type == 'pr.comment' && 'reviewer' in event.mentions`, Outcomes: []lifecycle.Outcome{{Status: "succeeded", When: "true"}}},
			{Name: "comment", Prompt: "pr.fix", Events: []string{"completed"}, When: `event.type == 'pr.comment'`, Outcomes: []lifecycle.Outcome{{Status: "succeeded", When: "true"}}},
		}}
		engine, err := lifecycle.New(def)
		Expect(err).NotTo(HaveOccurred())
		snapshot := prlifecycle.Snapshot{PR: git.PullRequest{Number: 7, URL: "https://git.example/acme/cli/pull/7", State: git.PullRequestOpen, HeadSHA: "revision-a"}, Activity: []git.PullRequestActivity{{ID: "comment-a", Kind: git.PullRequestComment, Mentions: []string{"reviewer"}}}}
		actions, err := prlifecycle.Actions(engine, snapshot, nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(actions).To(HaveLen(1))
		Expect(actions[0].Step.Name).To(Equal("mention"))
	})
	It("ignores failures for an older PR head", func() {
		snapshot := prlifecycle.Snapshot{PR: git.PullRequest{Number: 7, URL: "https://git.example/acme/cli/pull/7", State: git.PullRequestOpen, HeadSHA: "current"}, Failures: []prlifecycle.Failure{{Kind: "test", RunID: "run-a", HeadSHA: "previous"}, {Kind: "lint", RunID: "run-b", HeadSHA: "current"}}}
		events, err := prlifecycle.Events(snapshot)
		Expect(err).NotTo(HaveOccurred())
		Expect(events).To(HaveLen(3))
		Expect(events[2].Type).To(Equal("lint.failure"))
	})
	It("offers no contexts for a closed PR", func() {
		events, err := prlifecycle.Events(prlifecycle.Snapshot{PR: git.PullRequest{Number: 7, URL: "https://git.example/acme/cli/pull/7", State: git.PullRequestClosed, HeadSHA: "current"}})
		Expect(err).NotTo(HaveOccurred())
		Expect(events).To(BeEmpty())
	})
})
