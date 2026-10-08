package run

import (
	"github.com/flanksource/gavel/todos/types"
	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = ginkgo.Describe("run lifecycle options", func() {
	ginkgo.It("forwards ordered preset selectors and explicit-empty state to resolution and dispatch", func() {
		opts := runOptions(Request{Options: Options{Presets: []string{"organization", "review"}, PresetsSet: true}}, nil)
		Expect(opts.Presets).To(Equal([]string{"organization", "review"}))
		Expect(opts.PresetsSet).To(BeTrue())
	})

	ginkgo.It("forwards reuse of the previous run's branch to resolution and dispatch", func() {
		Expect(runOptions(Request{Options: Options{ReuseBranch: true}}, nil).ReuseBranch).To(BeTrue())
	})

	ginkgo.It("refuses to reuse a branch for a run that is one of several todos", func() {
		_, err := Resolve(ginkgo.GinkgoT().Context(), Request{
			Todo: &types.TODO{ID: "todo-1"}, Options: Options{ReuseBranch: true, Batch: []string{"todo-1", "todo-2"}},
		})

		Expect(err).To(MatchError(ContainSubstring("reuse-branch continues one todo's branch; this request runs 2 todos")))
	})
})
