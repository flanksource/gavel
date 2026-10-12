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
		Expect(todosRunCmd.Flags().Lookup("preset")).NotTo(BeNil())
		options := TodosRunOptions{Presets: []string{"organization", "review"}}
		Expect(todosRunOptions(options).Presets).To(Equal([]string{"organization", "review"}))
		Expect(todosRunOptions(options).PresetsSet).To(BeTrue())
		Expect(api.IsEmpty(todosRequestSpec(options))).To(BeTrue())
	})

	It("prints resolved preset identities in dry-run without dispatching", func() {
		resolution := &lifecycle.Resolution{
			Prompt:   "Preview the selected presets",
			Warnings: []string{"permissions.plugins is unsupported by this runtime"},
			RuntimePresets: &runtimeprofiles.PresetResolution{Presets: []runtimeprofiles.Preset{{
				ID: "preset-review-id", Name: "Review preset",
			}}},
		}
		started := stubTodoRunSeams(GinkgoTB(), resolution, "run", "requested")
		var err error
		out := captureStdout(GinkgoTB(), func() {
			err = runTodoStep(context.Background(), GinkgoT().TempDir(), nil,
				runTodo("aaaaaaaa-0000-4000-8000-000000000005", "Preview profile"),
				run.Options{Presets: []string{"review"}, PresetsSet: true}, runStepPolicy{DryRun: true, AllowCommit: true})
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(ContainSubstring("Runtime presets:"))
		Expect(out).To(ContainSubstring("Review preset (preset-review-id)"))
		Expect(out).To(ContainSubstring("Warning: permissions.plugins is unsupported by this runtime"))
		Expect(*started).To(BeEmpty())
	})
})
