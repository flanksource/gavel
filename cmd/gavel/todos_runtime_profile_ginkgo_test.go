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

var _ = Describe("TODO runtime preset CLI", func() {
	It("passes ordered presets as selectors without altering the request spec", func() {
		resetTodosRunFlags(GinkgoTB())
		flag := todosRunCmd.Flags().Lookup("preset")
		Expect(flag).NotTo(BeNil())
		Expect(todosRunCmd.Flags().Set("preset", "organization")).To(Succeed())
		Expect(todosRunCmd.Flags().Set("preset", "review")).To(Succeed())
		Expect(todosRunOptions().Presets).To(Equal([]string{"organization", "review"}))
		Expect(todosRunOptions().PresetsSet).To(BeTrue())
		Expect(api.IsEmpty(todosRequestSpec())).To(BeTrue())
	})

	It("prints resolved preset identities in dry-run without dispatching", func() {
		resetTodosRunFlags(GinkgoTB())
		resolution := &lifecycle.Resolution{
			Prompt:   "Preview the selected presets",
			Warnings: []string{"permissions.plugins is unsupported by this runtime"},
			RuntimePresets: &runtimeprofiles.PresetResolution{Presets: []runtimeprofiles.Preset{{
				ID: "preset-review-id", Name: "Review preset",
			}}},
		}
		started := stubTodoRunSeams(GinkgoTB(), resolution, "run", "requested")
		dryRun = true
		var err error
		out := captureStdout(GinkgoTB(), func() {
			err = runTodoStep(context.Background(), GinkgoT().TempDir(), nil,
				runTodo("aaaaaaaa-0000-4000-8000-000000000005", "Preview profile"),
				run.Options{Presets: []string{"review"}, PresetsSet: true})
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(ContainSubstring("Runtime presets:"))
		Expect(out).To(ContainSubstring("Review preset (preset-review-id)"))
		Expect(out).To(ContainSubstring("Warning: permissions.plugins is unsupported by this runtime"))
		Expect(*started).To(BeEmpty())
	})
})
