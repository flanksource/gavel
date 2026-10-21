package lifecycle

import (
	"context"
	"errors"

	"github.com/flanksource/captain/pkg/ai/agent"
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/pr/model"
	"github.com/flanksource/gavel/pr/provider"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("PR turn thread resolution guards", func() {
	ginkgo.It("hands the PR hook a sandbox job rebuilds the threads its own pr_thread_addressed records", func(ctx context.Context) {
		ginkgo.GinkgoT().Setenv("GAVEL_SERVER", "https://gavel.example")
		ginkgo.GinkgoT().Setenv("GAVEL_RUN_TOKEN", "run-token")
		ginkgo.GinkgoT().Setenv("GAVEL_RUN_ID", "run-1")
		hooks := RunHooks{PR: &PRHooks{
			URL: "https://github.com/acme/app/pull/7", Number: 7, Write: true, WorkDir: "/workspace/src",
			Target: provider.RunTarget{RepositoryURL: "https://github.com/acme/app", Branch: "topic", HeadSHA: "abc"},
			Events: []string{"completed", "blocked"}, OpenThreads: []string{"t1", "t2"},
		}}
		built, local, err := hooks.Build()
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		var pr *prExecution
		for _, hook := range built {
			if found, ok := hook.(*prExecution); ok {
				pr = found
			}
		}
		gomega.Expect(pr).NotTo(gomega.BeNil())
		gomega.Expect(pr.events).To(gomega.Equal([]string{"completed", "blocked"}))
		gomega.Expect(pr.openThreads).To(gomega.Equal(map[string]bool{"t1": true, "t2": true}))
		gomega.Expect(local).To(gomega.HaveLen(1))

		_, err = local[0].Handler(ctx, map[string]any{"threadId": "t1"})

		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(pr.tools.TakeThreads()).To(gomega.Equal([]string{"t1"}))
	})

	ginkgo.It("gives a sandbox job no thread tool for a PR it only reads", func() {
		ginkgo.GinkgoT().Setenv("GAVEL_SERVER", "https://gavel.example")
		ginkgo.GinkgoT().Setenv("GAVEL_RUN_TOKEN", "run-token")
		ginkgo.GinkgoT().Setenv("GAVEL_RUN_ID", "run-1")
		_, local, err := RunHooks{PR: &PRHooks{URL: "https://github.com/acme/app/pull/7", Number: 7, WorkDir: "/workspace/src"}}.Build()
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(local).To(gomega.BeEmpty())
	})

	ginkgo.It("rejects unknown, unresolvable, or not-open-at-prepare threads before resolving any", func() {
		prepared := &model.StatusSnapshot{PR: &model.PRInfo{ReviewThreads: []model.PRComment{
			{ThreadID: "addressed", CanResolve: true}, {ThreadID: "other"}, {ThreadID: "done", CanResolve: true, IsResolved: true},
		}}}
		open := openThreads(prepared)
		gomega.Expect(open).To(gomega.Equal(map[string]bool{"addressed": true}))
		snapshot := &model.StatusSnapshot{PR: &model.PRInfo{State: "OPEN", HeadRefOID: "pushed", ReviewThreads: []model.PRComment{
			{ThreadID: "addressed", CanResolve: true}, {ThreadID: "other"}, {ThreadID: "done", CanResolve: true}, {ThreadID: "injected", CanResolve: true},
		}}}
		gomega.Expect(validateAddressedThreads(snapshot, "pushed", []string{"addressed"}, open)).To(gomega.Succeed())
		gomega.Expect(validateAddressedThreads(snapshot, "pushed", []string{"addressed", "foreign"}, open)).To(gomega.MatchError(gomega.ContainSubstring("does not belong")))
		gomega.Expect(validateAddressedThreads(snapshot, "pushed", []string{"other"}, open)).To(gomega.MatchError(gomega.ContainSubstring("cannot be resolved")))
		gomega.Expect(validateAddressedThreads(snapshot, "pushed", []string{"injected"}, open)).To(gomega.MatchError(gomega.ContainSubstring("was not open when this run was prepared")))
		gomega.Expect(validateAddressedThreads(snapshot, "pushed", []string{"done"}, open)).To(gomega.MatchError(gomega.ContainSubstring("was not open when this run was prepared")))
		gomega.Expect(validateAddressedThreads(snapshot, "old", []string{"addressed"}, open)).To(gomega.HaveOccurred())
	})
	ginkgo.DescribeTable("does not publish incomplete or failed turns", func(response *api.Response, failed bool, postErr error, allowed bool) {
		hc := &agent.HookContext{Context: context.Background(), Phase: agent.PhaseTurn, Response: response, Failed: failed, PostErr: postErr}
		publish, err := (&prExecution{events: []string{"completed", "blocked"}}).completedTurn(hc)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(publish).To(gomega.Equal(allowed))
	},
		ginkgo.Entry("succeeded", &api.Response{Text: `{"success":true,"event":"completed","comment":"Fixed parser"}`}, false, nil, true),
		ginkgo.Entry("claims no success", &api.Response{Text: `{"success":false,"event":"blocked","comment":"Unable to fix"}`}, false, nil, false),
		ginkgo.Entry("failed generation", (*api.Response)(nil), true, nil, false),
		ginkgo.Entry("failed push", (*api.Response)(nil), false, errors.New("push failed"), false),
	)
})
