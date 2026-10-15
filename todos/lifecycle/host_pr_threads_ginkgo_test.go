package lifecycle

import (
	"context"
	"errors"

	"github.com/flanksource/captain/pkg/ai/agent"
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/pr/model"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("PR turn thread resolution guards", func() {
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
		publish, err := completedPRTurn(hc)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(publish).To(gomega.Equal(allowed))
	},
		ginkgo.Entry("completed", &api.Response{Text: `{"summary":"Fixed parser","endStatus":"completed","addressedThreadIds":["native"]}`}, false, nil, true),
		ginkgo.Entry("ask", &api.Response{Text: `{"summary":"Need intent","endStatus":"ask","questions":[{"question":"Which behavior?"}]}`}, false, nil, false),
		ginkgo.Entry("failed", &api.Response{Text: `{"summary":"Unable to fix","endStatus":"failed"}`}, false, nil, false),
		ginkgo.Entry("failed generation", (*api.Response)(nil), true, nil, false),
		ginkgo.Entry("failed push", (*api.Response)(nil), false, errors.New("push failed"), false),
	)
})
