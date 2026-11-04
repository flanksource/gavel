// Tests how a step's sandbox resolves against the saved Captain config: a named backend gets its
// kind and options, and a step that names no sandbox gets none, whatever the saved default.
package lifecycle

import (
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/captain/pkg/captainconfig"
	g "github.com/onsi/ginkgo/v2"
	o "github.com/onsi/gomega"
)

var _ = g.Describe("a step's sandbox", func() {
	var host *Host
	g.BeforeEach(func() {
		host = &Host{Saved: &captainconfig.Config{Sandbox: captainconfig.SandboxDefaults{
			Default: "beta-substrate",
			Backends: map[string]captainconfig.SandboxBackend{
				"beta-substrate": {Kind: "substrate", Options: map[string]any{"envAPI": "env.example:443"}},
			},
		}}}
	})

	g.It("is the named backend, with its kind and options", func() {
		normalized, err := host.normalizeSandbox(api.Spec{Sandbox: &api.SandboxRef{Backend: "beta-substrate"}})
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(normalized.Spec.Sandbox).To(o.Equal(&api.SandboxRef{Mode: api.SandboxSubstrate, Backend: "beta-substrate"}))

		selection, err := host.sandboxSelection(normalized.Spec)
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(selection).To(o.Equal(&api.SandboxConfig{Kind: api.SandboxSubstrate, Name: "beta-substrate", Options: map[string]any{"envAPI": "env.example:443"}}))
	})

	g.It("is none for a step that names none, whatever the saved default", func() {
		normalized, err := host.normalizeSandbox(api.Spec{})
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(normalized.Spec.Sandbox).To(o.BeNil())
		selection, err := host.sandboxSelection(normalized.Spec)
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(selection).To(o.BeNil())
	})

	g.It("names a backend the saved config doesn't have", func() {
		_, err := host.normalizeSandbox(api.Spec{Sandbox: &api.SandboxRef{Backend: "prod-pool"}})
		o.Expect(err).To(o.MatchError(o.ContainSubstring(`unknown sandbox "prod-pool"`)))
	})
})
