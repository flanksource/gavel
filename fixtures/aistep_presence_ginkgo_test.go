package fixtures

import (
	"encoding/json"

	"github.com/goccy/go-yaml"
	g "github.com/onsi/ginkgo/v2"
	o "github.com/onsi/gomega"
)

var _ = g.Describe("AI fixture explicit zero values", func() {
	g.It("preserves authored zero temperature and false cache policy through JSON", func() {
		var config FixtureAIConfig
		o.Expect(json.Unmarshal([]byte(`{"temperature":0,"noCache":false}`), &config)).To(o.Succeed())
		encoded, err := json.Marshal(config)
		o.Expect(err).NotTo(o.HaveOccurred())
		var fields map[string]any
		o.Expect(json.Unmarshal(encoded, &fields)).To(o.Succeed())
		o.Expect(fields).To(o.Equal(map[string]any{"temperature": float64(0), "noCache": false}))
	})

	g.It("preserves authored zero temperature and false cache policy through YAML", func() {
		var config FixtureAIConfig
		o.Expect(yaml.Unmarshal([]byte("temperature: 0\nnoCache: false\n"), &config)).To(o.Succeed())
		encoded, err := yaml.Marshal(config)
		o.Expect(err).NotTo(o.HaveOccurred())
		var fields map[string]any
		o.Expect(yaml.Unmarshal(encoded, &fields)).To(o.Succeed())
		o.Expect(fields).To(o.HaveKeyWithValue("temperature", o.BeNumerically("==", 0)))
		o.Expect(fields).To(o.HaveKeyWithValue("noCache", false))
	})

	g.It("lets explicit flat zero and false override nonzero snapshot defaults", func() {
		var config FixtureAIConfig
		o.Expect(json.Unmarshal([]byte(`{"spec":{"temperature":0.7,"noCache":true},"temperature":0,"noCache":false}`), &config)).To(o.Succeed())
		spec := config.Spec.Merge(config.SpecOverride())
		agent := config.ToAgentConfig(spec)
		zero := 0.0
		o.Expect(agent.Model.Temperature).To(o.Equal(&zero))
		o.Expect(agent.Model.NoCache).To(o.BeFalse())
		o.Expect(agent.NoCache).To(o.BeFalse())
	})

	g.It("keeps omitted flat fields absent and retains snapshot values", func() {
		var config FixtureAIConfig
		o.Expect(json.Unmarshal([]byte(`{"spec":{"temperature":0.7,"noCache":true}}`), &config)).To(o.Succeed())
		spec := config.Spec.Merge(config.SpecOverride())
		agent := config.ToAgentConfig(spec)
		temperature := 0.7
		o.Expect(agent.Model.Temperature).To(o.Equal(&temperature))
		o.Expect(agent.Model.NoCache).To(o.BeTrue())
		o.Expect(agent.NoCache).To(o.BeTrue())
	})
})
