package lifecycle

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/flanksource/captain/pkg/aiflags"
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/captain/pkg/captainconfig"
	"github.com/flanksource/gavel/verify"
	g "github.com/onsi/ginkgo/v2"
	o "github.com/onsi/gomega"
)

var _ = g.Describe("lifecycle saved runtime defaults", func() {
	g.It("keeps the captured Captain catalog configuration when the file changes", func() {
		path := filepath.Join(g.GinkgoT().TempDir(), ".captain.yaml")
		captainconfig.SetPathForTesting(path)
		g.DeferCleanup(func() { captainconfig.SetPathForTesting("") })
		host := Host{WorkDir: g.GinkgoT().TempDir(), Saved: &captainconfig.Config{}}
		o.Expect(os.WriteFile(path, []byte("ai: [invalid"), 0600)).To(o.Succeed())
		_, err := host.RuntimeCatalog(context.Background())
		o.Expect(err).NotTo(o.HaveOccurred())
	})

	g.It("attributes invalid saved configuration to the host rather than the request", func() {
		_, err := ResolveLayers(LayerInput{
			Saved:   &captainconfig.AIDefaults{Temperature: 3},
			Request: api.Spec{Model: api.Model{Name: "api:claude-haiku-4-5"}},
		})
		var configuration *ConfigurationError
		o.Expect(errors.As(err, &configuration)).To(o.BeTrue())
		o.Expect(err).To(o.MatchError(o.ContainSubstring("temperature")))
	})

	g.It("fills the selected model while preserving the project budget and its source", func() {
		resolved, err := ResolveLayers(LayerInput{
			Config:       verify.GavelConfig{AI: api.Spec{Budget: api.Budget{Cost: 2}}},
			Saved:        &captainconfig.AIDefaults{DefaultModel: "api:claude-haiku-4-5", BudgetUSD: 1},
			RequireModel: true,
		})
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(resolved.Spec.Name).To(o.Equal("claude-haiku-4-5"))
		o.Expect(resolved.Spec.Mode).To(o.Equal(api.ModeAPI))
		o.Expect(resolved.Spec.Budget.Cost).To(o.Equal(float64(2)))
		o.Expect(resolved.Provenance["/budget/cost"].Source.Name).To(o.Equal(".gavel.yaml ai"))
	})

	g.It("keeps explicit zero, false and empty fallbacks from the request", func() {
		var request api.Spec
		o.Expect(json.Unmarshal([]byte(`{"temperature":0,"noCache":false,"fallbacks":[]}`), &request)).To(o.Succeed())
		resolved, err := ResolveLayers(LayerInput{
			Config:  verify.GavelConfig{AI: api.Spec{Model: api.Model{Fallbacks: api.ModelList{{Name: "api:sonnet"}}, NoCache: true}}},
			Saved:   &captainconfig.AIDefaults{DefaultModel: "api:claude-haiku-4-5", Temperature: 0.8, NoCache: true},
			Request: request, RequireModel: true,
		})
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(resolved.Spec.Temperature).To(o.Equal(new(float64)))
		o.Expect(resolved.Spec.NoCache).To(o.BeFalse())
		o.Expect(resolved.Spec.Fallbacks).To(o.BeEmpty())
		o.Expect(resolved.Provenance["/noCache"].Source.Name).To(o.Equal("request"))
	})

	g.It("requires a model only for a generating run", func() {
		input := LayerInput{Saved: &captainconfig.AIDefaults{}}
		resolved, err := ResolveLayers(input)
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(resolved.Spec.Name).To(o.BeEmpty())
		input.RequireModel = true
		_, err = ResolveLayers(input)
		o.Expect(err).To(o.MatchError(o.ContainSubstring("configure")))
		o.Expect(errors.Is(err, aiflags.ErrUnconfigured)).To(o.BeTrue())
	})
})
