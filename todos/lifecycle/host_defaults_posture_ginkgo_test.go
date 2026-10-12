package lifecycle_test

import (
	"context"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/todos/lifecycle"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// The regression: a home `.gavel.yaml` naming `ai.permissions.mode` used to
// become a ceiling on every layer above it, and the dashboard's run dialog then
// failed to load entirely — one step whose prompt declares an unrelated posture
// took down the whole run-context response, because a ceiling has to compare two
// postures and `auto` compares with nothing.
//
// Layers only default now, so any posture composes with any other, and the
// read-only steps are pinned after the fold instead.
var _ = Describe("Step defaults under a configured posture", func() {
	DescribeTable("resolves every lifecycle step whatever posture the project configured",
		func(configured api.PermissionMode) {
			host := newHost(&fakeProvider{})
			host.Config.AI = api.Spec{Permissions: api.Permissions{Mode: configured}}

			for _, step := range host.Def.Definition().Steps {
				defaults, err := host.StepDefaults(context.Background(), step)

				Expect(err).NotTo(HaveOccurred(), "step %s must resolve under %q", step.Name, configured)
				if lifecycle.Class(step) != "run" {
					Expect(defaults.Spec.Permissions.Mode).To(Equal(api.PermissionPlan),
						"read-only step %s must stay read-only under %q", step.Name, configured)
				}
			}
		},
		Entry("auto", api.PermissionAuto),
		Entry("dontAsk", api.PermissionDontAsk),
		Entry("bypassPermissions", api.PermissionBypass),
		Entry("acceptEdits", api.PermissionAcceptEdits),
		Entry("plan", api.PermissionPlan),
		Entry("unset", api.PermissionMode("")),
	)
})
