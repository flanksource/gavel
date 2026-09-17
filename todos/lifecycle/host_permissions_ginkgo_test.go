package lifecycle

import (
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/todos/types"
	g "github.com/onsi/ginkgo/v2"
	o "github.com/onsi/gomega"
)

// A read-only class is an invariant of the step, not a ceiling attached to one
// of its layers: whatever the stack resolved to, the posture a plan, triage or
// verify envelope runs under is pinned after every layer has spoken.
var _ = g.Describe("Read-only class posture", func() {
	g.DescribeTable("pins the resolved posture last", func(class types.RunMode, resolved, expected api.PermissionMode) {
		spec := api.Spec{Permissions: api.Permissions{Mode: resolved}}
		ApplyClassInvariants(&spec, class)
		o.Expect(spec.Permissions.Mode).To(o.Equal(expected))
	},
		g.Entry("plan class over a request that asked to edit", types.ModePlan, api.PermissionAcceptEdits, api.PermissionPlan),
		g.Entry("plan class over an unordered posture", types.ModePlan, api.PermissionAuto, api.PermissionPlan),
		g.Entry("plan class over a host default", types.ModePlan, api.PermissionDefault, api.PermissionPlan),
		g.Entry("verify class over a bypass", types.ModeVerify, api.PermissionBypass, api.PermissionPlan),
		g.Entry("run class keeps what the layers resolved", types.ModeRun, api.PermissionAcceptEdits, api.PermissionAcceptEdits),
	)

	g.It("records the pin so a reader sees why the posture is not the request's", func() {
		spec := api.Spec{Permissions: api.Permissions{Mode: api.PermissionAcceptEdits}}
		spec.Explicit = api.FieldPresence{"/permissions/mode": true}

		ApplyClassInvariants(&spec, types.ModePlan)

		o.Expect(spec.Permissions.Mode).To(o.Equal(api.PermissionPlan))
		o.Expect(spec.Explicit).To(o.HaveKey("/permissions/mode"))
	})
})
