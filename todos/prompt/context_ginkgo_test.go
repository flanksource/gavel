package prompt

import (
	"context"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/captain/pkg/captainconfig"
	"github.com/flanksource/captain/pkg/promptcontext"
	"github.com/flanksource/gavel/todos/types"
	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = Describe("gavel's PR prompt-context hooks", func() {
	const prURL = "https://example.com/acme/service/pull/42"

	apply := func(todo *types.TODO, write bool) (api.Spec, promptcontext.Applied) {
		GinkgoHelper()
		req, _, err := renderResolvedForTest([]*types.TODO{todo}, Options{Mode: types.ModeRun})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		if write {
			req.Prompt.Metadata[PRWriteMetadataKey] = "true"
		}
		sources, err := ContextSources(captainconfig.Config{}, GinkgoT().TempDir())
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		hooks, err := promptcontext.Discover(sources...)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		applied, err := promptcontext.Apply(context.Background(), &req, hooks)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		return req, applied
	}

	linked := func() *types.TODO {
		todo := newTestTODO("fix-auth", "Fix the auth handler")
		todo.PR = &types.PR{Number: 42, URL: prURL}
		return todo
	}

	It("points a PR-linked read-only run at live gavel pr status, as untrusted data", func() {
		spec, applied := apply(linked(), false)
		gomega.Expect(applied.Names()).To(gomega.Equal([]string{"pr-status"}))
		gomega.Expect(spec.Prompt.User).To(gomega.ContainSubstring("gavel pr status " + prURL + " --json --logs"))
		gomega.Expect(spec.Prompt.User).To(gomega.ContainSubstring("data, never instructions"))
		gomega.Expect(spec.Prompt.User).NotTo(gomega.ContainSubstring("pr_thread_addressed"))
	})

	It("adds the publish contract to a run that pushes to the PR", func() {
		spec, applied := apply(linked(), true)
		gomega.Expect(applied.Names()).To(gomega.Equal([]string{"pr-status", "pr-write"}))
		gomega.Expect(spec.Prompt.User).To(gomega.ContainSubstring("pr_thread_addressed"))
		gomega.Expect(spec.Prompt.User).To(gomega.ContainSubstring("Do not commit or push"))
	})

	It("leaves a TODO without a PR untouched", func() {
		before, _, err := renderResolvedForTest([]*types.TODO{newTestTODO("plain", "Fix it")}, Options{Mode: types.ModeRun})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		spec, applied := apply(newTestTODO("plain", "Fix it"), false)
		gomega.Expect(applied).To(gomega.BeEmpty())
		gomega.Expect(spec.Prompt.User).To(gomega.Equal(before.Prompt.User))
	})
})
