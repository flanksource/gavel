package azuredevops

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"time"
)

func policy(status string, blocking bool, kind string) policyEvaluation {
	p := policyEvaluation{Status: status}
	p.Configuration.Enabled, p.Configuration.Blocking = true, blocking
	p.Configuration.Type.ID, p.Configuration.Type.Name = kind, "Required policy"
	return p
}

var _ = Describe("Azure status normalization", func() {
	DescribeTable("distinguishes merge computation and policy readiness",
		func(status string, draft bool, policies []policyEvaluation, expected string) {
			result := readiness(pullRequest{Status: "active", MergeStatus: status, Draft: draft}, policies)
			Expect(result.State).To(Equal(expected))
			if expected != "ready" {
				Expect(result.Reasons).NotTo(BeEmpty())
			}
		},
		Entry("ready", "succeeded", false, nil, "ready"),
		Entry("conflict", "conflicts", false, nil, "conflicting"),
		Entry("computing", "queued", false, nil, "pending"),
		Entry("unknown", "notSet", false, nil, "unknown"),
		Entry("merge failure", "failure", false, nil, "blocked"),
		Entry("native policy rejection", "rejectedByPolicy", false, nil, "blocked"),
		Entry("draft", "succeeded", true, nil, "blocked"),
		Entry("required rejection", "succeeded", false, []policyEvaluation{policy("rejected", true, "review")}, "blocked"),
		Entry("waiting reviewer", "succeeded", false, []policyEvaluation{policy("queued", true, "review")}, "blocked"),
		Entry("waiting build", "succeeded", false, []policyEvaluation{policy("running", true, buildPolicyType)}, "pending"),
		Entry("broken policy", "succeeded", false, []policyEvaluation{policy("broken", true, "review")}, "blocked"),
		Entry("optional rejection", "succeeded", false, []policyEvaluation{policy("rejected", false, "review")}, "ready"),
		Entry("approved policy", "succeeded", false, []policyEvaluation{policy("approved", true, "review")}, "ready"),
		Entry("not applicable policy", "succeeded", false, []policyEvaluation{policy("notApplicable", true, "review")}, "ready"),
	)
	It("selects current-revision builds and the latest attempt of each pipeline", func() {
		pr := pullRequest{ID: 27, Source: "refs/heads/feature", SourceCommit: commitRef{ID: "current-head"}, MergeCommit: commitRef{ID: "current-merge"}}
		now := time.Now()
		var builds []build
		for i, version := range []string{"old-merge", "current-merge", "current-merge", "wrong-head", "current-head"} {
			b := build{ID: int64(i + 1), QueueTime: now.Add(time.Duration(i) * time.Minute), SourceVersion: version, SourceBranch: "refs/pull/27/merge"}
			b.Definition.ID = 1
			if i > 2 {
				b.SourceBranch = pr.Source
				b.Definition.ID = 2
			}
			builds = append(builds, b)
		}
		selected := currentBuilds(pr, builds)
		Expect(selected).To(HaveLen(2))
		Expect([]int64{selected[0].ID, selected[1].ID}).To(Equal([]int64{3, 5}))
	})
	It("does not claim policy completion for an unsupported policy state", func() {
		result := readiness(pullRequest{Status: "active", MergeStatus: "succeeded"}, []policyEvaluation{policy("unexpected", true, "review")})
		Expect(result.State).To(Equal("unknown"))
		Expect(result.Reasons).NotTo(BeEmpty())
	})
})
