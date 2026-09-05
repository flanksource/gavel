package fixtures

import (
	"encoding/json"
	"time"

	"github.com/flanksource/captain/pkg/api"
	"github.com/goccy/go-yaml"
	g "github.com/onsi/ginkgo/v2"
	o "github.com/onsi/gomega"
)

var _ = g.Describe("AI fixture config codecs", func() {
	g.It("roundtrips the canonical spec and flat options through JSON", func() {
		config := FixtureAIConfig{
			Spec:  &api.Spec{Model: api.Model{Name: "sonnet", Mode: api.ModeCLI}, Sandbox: &api.SandboxRef{Mode: api.SandboxNative}},
			Model: "api:gpt-5", MaxTokens: 800, CacheTTL: time.Minute, CriteriaSection: "Review",
		}
		data, err := json.Marshal(config)
		o.Expect(err).NotTo(o.HaveOccurred())
		var decoded FixtureAIConfig
		o.Expect(json.Unmarshal(data, &decoded)).To(o.Succeed())
		o.Expect(decoded).To(o.Equal(config))
	})

	g.It("keeps fixture durations and native spec codecs in the same YAML document", func() {
		front, _, err := SplitFrontMatter("---\nai:\n  cacheTTL: 10m\n  spec:\n    model: sonnet\n    mode: cli\n    sandbox: native\n    fallbacks: [api:gpt-5]\n---\n")
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(front.AI).To(o.Equal(&FixtureAIConfig{CacheTTL: 10 * time.Minute, Spec: &api.Spec{
			Model:   api.Model{Name: "sonnet", Mode: api.ModeCLI, Fallbacks: []api.Model{{Name: "gpt-5", Mode: api.ModeAPI}}},
			Sandbox: &api.SandboxRef{Mode: api.SandboxNative},
		}}))
		encoded, err := yaml.Marshal(front)
		o.Expect(err).NotTo(o.HaveOccurred())
		decoded, _, err := SplitFrontMatter("---\n" + string(encoded) + "---\n")
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(decoded.AI).To(o.Equal(front.AI))
	})

	g.It("surfaces an invalid native spec instead of ignoring it", func() {
		_, _, err := SplitFrontMatter("---\nai:\n  spec:\n    sandbox: [native]\n---\n")
		o.Expect(err).To(o.HaveOccurred())
	})
})
