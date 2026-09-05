package lifecycle

import (
	"github.com/flanksource/captain/pkg/api"
	g "github.com/onsi/ginkgo/v2"
	o "github.com/onsi/gomega"
)

var _ = g.Describe("Request permission guard", func() {
	g.DescribeTable("rejects requests that widen configured permissions",
		func(configured, requested api.Permissions, field string) {
			layers := []api.SpecLayer{
				api.PromptSpecLayer(".gavel.yaml todos.run", api.Spec{Permissions: configured}),
				api.RequestSpecLayer("request", api.Spec{Permissions: requested}),
			}
			err := ValidateRequestPermissions(layers)
			o.Expect(err).To(o.HaveOccurred())
			o.Expect(err.Error()).To(o.And(o.ContainSubstring("request"), o.ContainSubstring(".gavel.yaml todos.run"), o.ContainSubstring(field)))
		},
		g.Entry("denied tool becomes allowed",
			api.Permissions{Tools: api.Tools{"Bash": api.ToolPolicyDeny}}, api.Permissions{Tools: api.Tools{"Bash": api.ToolPolicyAllow}}, "permissions.tools.Bash"),
		g.Entry("denied tool becomes automatic",
			api.Permissions{Tools: api.Tools{"Bash": api.ToolPolicyDeny}}, api.Permissions{Tools: api.Tools{"Bash": api.ToolPolicyAuto}}, "permissions.tools.Bash"),
		g.Entry("denied tool becomes an approval request",
			api.Permissions{Tools: api.Tools{"Bash": api.ToolPolicyDeny}}, api.Permissions{Tools: api.Tools{"Bash": api.ToolPolicyAsk}}, "permissions.tools.Bash"),
		g.Entry("explicit empty tool policy removes denial",
			api.Permissions{Tools: api.Tools{"Bash": api.ToolPolicyDeny}}, api.Permissions{Tools: api.Tools{"Bash": ""}}, "permissions.tools.Bash"),
		g.Entry("plan becomes default",
			api.Permissions{Mode: api.PermissionPlan}, api.Permissions{Mode: api.PermissionDefault}, "permissions.mode"),
		g.Entry("plan becomes acceptEdits",
			api.Permissions{Mode: api.PermissionPlan}, api.Permissions{Mode: api.PermissionAcceptEdits}, "permissions.mode"),
		g.Entry("plan becomes bypassPermissions",
			api.Permissions{Mode: api.PermissionPlan}, api.Permissions{Mode: api.PermissionBypass}, "permissions.mode"),
		g.Entry("default becomes acceptEdits",
			api.Permissions{Mode: api.PermissionDefault}, api.Permissions{Mode: api.PermissionAcceptEdits}, "permissions.mode"),
		g.Entry("default becomes bypassPermissions",
			api.Permissions{Mode: api.PermissionDefault}, api.Permissions{Mode: api.PermissionBypass}, "permissions.mode"),
		g.Entry("acceptEdits becomes bypassPermissions",
			api.Permissions{Mode: api.PermissionAcceptEdits}, api.Permissions{Mode: api.PermissionBypass}, "permissions.mode"),
		g.Entry("dontAsk becomes bypassPermissions",
			api.Permissions{Mode: api.PermissionDontAsk}, api.Permissions{Mode: api.PermissionBypass}, "permissions.mode"),
		g.Entry("dontAsk becomes default",
			api.Permissions{Mode: api.PermissionDontAsk}, api.Permissions{Mode: api.PermissionDefault}, "permissions.mode"),
		g.Entry("dontAsk becomes auto",
			api.Permissions{Mode: api.PermissionDontAsk}, api.Permissions{Mode: api.PermissionAuto}, "permissions.mode"),
		g.Entry("dontAsk becomes acceptEdits",
			api.Permissions{Mode: api.PermissionDontAsk}, api.Permissions{Mode: api.PermissionAcceptEdits}, "permissions.mode"),
		g.Entry("plan becomes incomparable auto",
			api.Permissions{Mode: api.PermissionPlan}, api.Permissions{Mode: api.PermissionAuto}, "permissions.mode"),
	)

	g.DescribeTable("allows unchanged or narrower requests",
		func(configured, requested api.Permissions) {
			o.Expect(ValidateRequestPermissions([]api.SpecLayer{
				api.PromptSpecLayer(".gavel.yaml todos.run", api.Spec{Permissions: configured}),
				api.RequestSpecLayer("request", api.Spec{Permissions: requested}),
			})).To(o.Succeed())
		},
		g.Entry("allowed tool becomes denied",
			api.Permissions{Tools: api.Tools{"Bash": api.ToolPolicyAllow}}, api.Permissions{Tools: api.Tools{"Bash": api.ToolPolicyDeny}}),
		g.Entry("acceptEdits becomes plan", api.Permissions{Mode: api.PermissionAcceptEdits}, api.Permissions{Mode: api.PermissionPlan}),
		g.Entry("bypass becomes acceptEdits", api.Permissions{Mode: api.PermissionBypass}, api.Permissions{Mode: api.PermissionAcceptEdits}),
		g.Entry("bypass becomes plan", api.Permissions{Mode: api.PermissionBypass}, api.Permissions{Mode: api.PermissionPlan}),
		g.Entry("plan remains plan", api.Permissions{Mode: api.PermissionPlan}, api.Permissions{Mode: api.PermissionPlan}),
		g.Entry("dontAsk remains dontAsk", api.Permissions{Mode: api.PermissionDontAsk}, api.Permissions{Mode: api.PermissionDontAsk}),
		g.Entry("dontAsk remains when unset", api.Permissions{Mode: api.PermissionDontAsk}, api.Permissions{}),
		g.Entry("request leaves permissions unset", api.Permissions{Tools: api.Tools{"Bash": api.ToolPolicyDeny}}, api.Permissions{}),
		g.Entry("no configured policy", api.Permissions{}, api.Permissions{Mode: api.PermissionAcceptEdits}),
		g.Entry("skills disabled withdraws an enabled skill",
			api.Permissions{Skills: api.ResourcePolicies{"./skills": api.ResourceEnabled}}, api.Permissions{Skills: api.ResourcePolicies{"./skills": api.ResourceDisabled}}),
	)

	g.It("uses the effective baseline and identifies the last layer setting the denied tool", func() {
		layers := []api.SpecLayer{
			{Name: ".gavel.yaml ai", Scope: api.SpecLayerGlobal, Source: api.SpecLayerSourcePreset,
				Spec: api.Spec{Permissions: api.Permissions{Tools: api.Tools{"Bash": api.ToolPolicyAllow}}}},
			api.RequestSpecLayer("prior run spec", api.Spec{Permissions: api.Permissions{Tools: api.Tools{"Bash": api.ToolPolicyDeny}}}),
			api.RequestSpecLayer("host dashboard", api.Spec{Permissions: api.Permissions{Mode: api.PermissionDefault}}),
			api.RequestSpecLayer("request", api.Spec{Permissions: api.Permissions{Tools: api.Tools{"Bash": api.ToolPolicyAllow}}}),
		}
		err := ValidateRequestPermissions(layers)
		o.Expect(err).To(o.MatchError(o.And(o.ContainSubstring(`layer "prior run spec"`), o.ContainSubstring("permissions.tools.Bash"))))
		o.Expect(layers[0].Spec.Permissions.Tools).To(o.Equal(api.Tools{"Bash": api.ToolPolicyAllow}))
	})

	g.It("lets later authored configuration override earlier defaults", func() {
		o.Expect(ValidateRequestPermissions([]api.SpecLayer{
			{Name: ".gavel.yaml ai", Scope: api.SpecLayerGlobal, Source: api.SpecLayerSourcePreset,
				Spec: api.Spec{Permissions: api.Permissions{Tools: api.Tools{"Bash": api.ToolPolicyDeny}}}},
			api.PromptSpecLayer(".gavel.yaml todos.run", api.Spec{Permissions: api.Permissions{Tools: api.Tools{"Bash": api.ToolPolicyAllow}}}),
			api.RequestSpecLayer("request", api.Spec{Permissions: api.Permissions{Tools: api.Tools{"Bash": api.ToolPolicyAllow}}}),
		})).To(o.Succeed())
	})
})

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

	g.It("protects a profile's plan mode and rejects an explicit request to widen it", func() {
		layers := []api.SpecLayer{
			{Name: "Review profile", Source: api.SpecLayerSourceProfile, Scope: api.SpecLayerSurface,
				Spec: api.Spec{Permissions: api.Permissions{Mode: api.PermissionPlan}}},
			api.RequestSpecLayer("host dashboard", api.Spec{Permissions: api.Permissions{Mode: api.PermissionDefault}}),
			api.RequestSpecLayer("request", api.Spec{Permissions: api.Permissions{Mode: api.PermissionBypass}}),
		}
		err := ValidateRequestPermissions(RestrictHostPermissions(layers))
		o.Expect(err).To(o.MatchError(o.ContainSubstring("permissions.mode")))
	})
})
