package lifecycle

import (
	"context"
	"os"
	"path/filepath"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/captain/pkg/runtimeprofiles"
	"github.com/flanksource/gavel/verify"
	g "github.com/onsi/ginkgo/v2"
	o "github.com/onsi/gomega"
)

var _ = g.Describe("lifecycle runtime profile layers", func() {
	var host Host
	g.BeforeEach(func() {
		dir := g.GinkgoT().TempDir()
		for name, value := range map[string]string{
			"requested": "name: requested\nspec:\n  budget: {maxTurns: 13}\n  permissions: {mode: plan}\n",
			"pinned":    "name: pinned\nspec:\n  budget: {maxTurns: 9}\n",
			"default":   "name: default\nspec:\n  budget: {maxTurns: 5}\n",
		} {
			o.Expect(os.WriteFile(filepath.Join(dir, name+".yaml"), []byte(value), 0600)).To(o.Succeed())
		}
		source, err := runtimeprofiles.NewFileSource(runtimeprofiles.FileSourceOptions{Kind: runtimeprofiles.KindProfile, Dir: dir})
		o.Expect(err).NotTo(o.HaveOccurred())
		catalog, err := runtimeprofiles.NewCatalog(source)
		o.Expect(err).NotTo(o.HaveOccurred())
		host = Host{Catalog: func(context.Context) (*runtimeprofiles.Catalog, error) { return catalog, nil }}
	})

	g.DescribeTable("selects request before pin before default",
		func(selection ProfileSelection, turns int) {
			result, err := host.resolveProfileLayers(context.Background(), LayerInput{RuntimeProfile: selection})
			o.Expect(err).NotTo(o.HaveOccurred())
			o.Expect(result.Resolved.Spec.Budget.MaxTurns).To(o.Equal(turns))
			o.Expect(result.Profile).NotTo(o.BeNil())
		},
		g.Entry("request", ProfileSelection{Requested: "requested", Pinned: "pinned", Default: "default"}, 13),
		g.Entry("pin", ProfileSelection{Pinned: "pinned", Default: "default"}, 9),
		g.Entry("default", ProfileSelection{Default: "default"}, 5),
	)

	g.It("places the profile above project defaults and below prompt and request layers", func() {
		result, err := host.resolveProfileLayers(context.Background(), LayerInput{
			RuntimeProfile: ProfileSelection{Requested: "requested"},
			Config:         verify.GavelConfig{AI: api.Spec{Budget: api.Budget{MaxTurns: 3, Cost: 2}}},
			Frontmatter:    []api.SpecLayer{api.PromptSpecLayer("prompt", api.Spec{Budget: api.Budget{MaxTurns: 17}})},
			Request:        api.Spec{Budget: api.Budget{MaxTurns: 19}},
		})
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(result.Resolved.Spec.Budget).To(o.Equal(api.Budget{MaxTurns: 19, Cost: 2}))
		o.Expect(result.Resolved.Trace).To(o.HaveLen(4))
		o.Expect(result.Resolved.Trace[1].Source).To(o.Equal(api.SpecLayerSourceProfile))
	})

	g.It("keeps selected profile restrictions as request ceilings", func() {
		_, err := host.resolveProfileLayers(context.Background(), LayerInput{
			RuntimeProfile: ProfileSelection{Requested: "requested"},
			Request:        api.Spec{Permissions: api.Permissions{Mode: api.PermissionAcceptEdits}},
		})

		o.Expect(err).To(o.MatchError(o.And(o.ContainSubstring("permissions.mode"), o.ContainSubstring("requested run spec"), o.ContainSubstring("request"))))
	})

	g.It("never opens a catalog when no profile is selected", func() {
		host.Catalog = func(context.Context) (*runtimeprofiles.Catalog, error) {
			g.Fail("unexpected catalog lookup")
			return nil, nil
		}
		in := LayerInput{Request: api.Spec{Budget: api.Budget{MaxTurns: 7}}}
		want, err := ResolveLayers(in)
		o.Expect(err).NotTo(o.HaveOccurred())
		actual, err := host.resolveProfileLayers(context.Background(), in)
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(actual.Resolved).To(o.Equal(want))
	})

	g.It("reports the selection origin when a profile does not exist", func() {
		_, err := host.resolveProfileLayers(context.Background(), LayerInput{RuntimeProfile: ProfileSelection{Requested: "missing"}})
		o.Expect(err).To(o.MatchError(o.ContainSubstring(`selected by request`)))
	})
})
