package lifecycle

import (
	"github.com/flanksource/captain/pkg/api"
	g "github.com/onsi/ginkgo/v2"
	o "github.com/onsi/gomega"
)

var _ = g.Describe("Host permission restriction", func() {
	g.DescribeTable("narrows host approval defaults without widening an authored mode",
		func(authored, expected api.PermissionMode) {
			layers := []api.SpecLayer{
				api.PromptSpecLayer(".gavel.yaml todos.run", api.Spec{Permissions: api.Permissions{Mode: authored}}),
				api.RequestSpecLayer("host dashboard", api.Spec{Permissions: api.Permissions{Mode: api.PermissionDefault}}),
			}
			ordered := RestrictHostPermissions(layers)
			o.Expect(ordered[1].Spec.Permissions.Mode).To(o.Equal(expected))
			o.Expect(layers[1].Spec.Permissions.Mode).To(o.Equal(api.PermissionDefault))
		},
		g.Entry("plan remains read-only", api.PermissionPlan, api.PermissionPlan),
		g.Entry("dontAsk remains constrained", api.PermissionDontAsk, api.PermissionDontAsk),
		g.Entry("acceptEdits requires approval", api.PermissionAcceptEdits, api.PermissionDefault),
		g.Entry("bypass requires approval", api.PermissionBypass, api.PermissionDefault),
		g.Entry("unset gets host default", api.PermissionMode(""), api.PermissionDefault),
	)
})
