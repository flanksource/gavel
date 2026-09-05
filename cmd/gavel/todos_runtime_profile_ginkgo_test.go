package main

import (
	"context"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/captain/pkg/runtimeprofiles"
	"github.com/flanksource/gavel/todos/lifecycle"
	"github.com/flanksource/gavel/todos/run"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TODO runtime profile CLI", func() {
	It("passes an explicit profile as a selector without altering the request spec", func() {
		resetTodosRunFlags(GinkgoTB())
		flag := todosRunCmd.Flags().Lookup("runtime-profile")
		Expect(flag).NotTo(BeNil())
		previous, changed := flag.Value.String(), flag.Changed
		DeferCleanup(func() {
			Expect(flag.Value.Set(previous)).To(Succeed())
			flag.Changed = changed
		})
		Expect(flag.DefValue).To(BeEmpty())
		Expect(todosRunCmd.Flags().Set("runtime-profile", "Review profile")).To(Succeed())
		Expect(todosRunOptions().RuntimeProfile).To(Equal("Review profile"))
		Expect(api.IsEmpty(todosRequestSpec())).To(BeTrue())
	})

	It("prints the resolved profile identity in dry-run without dispatching", func() {
		resetTodosRunFlags(GinkgoTB())
		resolution := &lifecycle.Resolution{
			Prompt:   "Preview the selected profile",
			Warnings: []string{"permissions.plugins is unsupported by this runtime"},
			RuntimeProfile: &runtimeprofiles.Resolution{Profile: runtimeprofiles.Profile{
				ID: "profile-review-id", Name: "Review profile",
			}},
		}
		started := stubTodoRunSeams(GinkgoTB(), resolution, "run", "requested")
		dryRun = true
		var err error
		out := captureStdout(GinkgoTB(), func() {
			err = runTodoStep(context.Background(), GinkgoT().TempDir(), nil,
				runTodo("aaaaaaaa-0000-4000-8000-000000000005", "Preview profile"),
				run.Options{RuntimeProfile: "review"})
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(ContainSubstring("Runtime profile: Review profile (profile-review-id)"))
		Expect(out).To(ContainSubstring("Warning: permissions.plugins is unsupported by this runtime"))
		Expect(*started).To(BeEmpty())
	})
})
