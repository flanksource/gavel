package lifecycle

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/captain/pkg/runtimeprofiles"
	todoprompt "github.com/flanksource/gavel/todos/prompt"
	"github.com/flanksource/gavel/verify"
	g "github.com/onsi/ginkgo/v2"
	o "github.com/onsi/gomega"
)

var _ = g.Describe("lifecycle layer validation ownership", func() {
	g.DescribeTable("rejects malformed project layers before a request can override them",
		func(spec api.Spec) {
			in := LayerInput{
				Config:  verify.GavelConfig{AI: spec},
				Request: api.Spec{Budget: api.Budget{Cost: 5}, Permissions: api.Permissions{Mode: api.PermissionDefault}},
			}
			var configErr *ConfigurationError
			_, err := ResolveLayers(in)
			o.Expect(err).To(o.HaveOccurred())
			o.Expect(errors.As(err, &configErr)).To(o.BeTrue(), "%v", err)
			o.Expect(err.Error()).To(o.ContainSubstring(".gavel.yaml ai"))
		},
		g.Entry("negative budget", api.Spec{Budget: api.Budget{Cost: -1}}),
		g.Entry("unknown permission mode", api.Spec{Permissions: api.Permissions{Mode: "invalid-mode"}}),
	)

	g.It("validates project layers before opening a selected catalog", func() {
		host := Host{Catalog: func(context.Context) (*runtimeprofiles.Catalog, error) {
			g.Fail("invalid project configuration must fail before catalog access")
			return nil, nil
		}}
		_, err := host.resolveProfileLayers(context.Background(), LayerInput{
			Config:         verify.GavelConfig{AI: api.Spec{Budget: api.Budget{Cost: -1}}},
			RuntimeProfile: ProfileSelection{Requested: "requested"},
			Request:        api.Spec{Budget: api.Budget{Cost: 5}},
		})
		var configErr *ConfigurationError
		o.Expect(errors.As(err, &configErr)).To(o.BeTrue(), "%v", err)
	})

	g.It("keeps malformed requests separate from configuration errors", func() {
		_, err := ResolveLayers(LayerInput{Request: api.Spec{Budget: api.Budget{Cost: -1}}})
		var configErr *ConfigurationError
		o.Expect(err).To(o.HaveOccurred())
		o.Expect(errors.As(err, &configErr)).To(o.BeFalse())
	})

	g.It("attributes malformed project files to configuration", func() {
		dir := g.GinkgoT().TempDir()
		o.Expect(os.WriteFile(filepath.Join(dir, ".gavel.yaml"), []byte("ai: [\n"), 0600)).To(o.Succeed())
		_, err := NewHost(nil, dir, HostCLI)
		var configErr *ConfigurationError
		o.Expect(errors.As(err, &configErr)).To(o.BeTrue(), "%v", err)
	})

	g.It("attributes malformed prompt files to configuration", func() {
		dir := g.GinkgoT().TempDir()
		path := filepath.Join(dir, "run.prompt")
		o.Expect(os.WriteFile(path, []byte("---\npermissions: [\n---\nbody\n"), 0600)).To(o.Succeed())
		_, err := PromptLayers(dir, nil, todoprompt.Definition{Name: "run", Override: verify.PromptSpec{File: path}})
		var configErr *ConfigurationError
		o.Expect(errors.As(err, &configErr)).To(o.BeTrue(), "%v", err)
	})

	g.It("attributes catalog access failures to configuration", func() {
		host := Host{Catalog: func(context.Context) (*runtimeprofiles.Catalog, error) {
			return nil, errors.New("catalog database unavailable")
		}}
		_, err := host.resolveProfileLayers(context.Background(), LayerInput{RuntimeProfile: ProfileSelection{Requested: "requested"}})
		var configErr *ConfigurationError
		o.Expect(errors.As(err, &configErr)).To(o.BeTrue(), "%v", err)
		o.Expect(err.Error()).To(o.ContainSubstring("catalog database unavailable"))
	})

	g.DescribeTable("validates every project-owned layer before request precedence",
		func(in LayerInput, name string) {
			in.Request = api.Spec{Budget: api.Budget{MaxTurns: 5}}
			_, err := ResolveLayers(in)
			var configErr *ConfigurationError
			o.Expect(errors.As(err, &configErr)).To(o.BeTrue(), "%v", err)
			o.Expect(err.Error()).To(o.ContainSubstring(name))
		},
		g.Entry("prompt frontmatter", LayerInput{Frontmatter: []api.SpecLayer{
			api.PromptSpecLayer("custom.prompt", api.Spec{Budget: api.Budget{MaxTurns: -1}}),
		}}, "custom.prompt"),
		g.Entry("lifecycle step", LayerInput{Step: "run", StepSpec: api.Spec{Budget: api.Budget{MaxTurns: -1}}}, "lifecycle step run"),
		g.Entry("step config", LayerInput{Step: "run", Config: verify.GavelConfig{Todos: verify.TodosConfig{
			Run: verify.PromptSpec{Spec: api.Spec{Budget: api.Budget{MaxTurns: -1}}},
		}}}, ".gavel.yaml todos.run"),
	)
})
