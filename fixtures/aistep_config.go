package fixtures

import (
	"github.com/flanksource/captain/pkg/api"
	ghodss "github.com/ghodss/yaml"
	"github.com/goccy/go-yaml"
)

type fixtureAIFields FixtureAIConfig

// Captain's spec uses JSON codecs for native scalar and connection types. Keep
// the fixture's duration codec while routing only the spec through that contract.
func (c *FixtureAIConfig) UnmarshalYAML(data []byte) error {
	var fields fixtureAIFields
	if err := yaml.UnmarshalWithOptions(data, &fields, yaml.CustomUnmarshaler[api.Spec](func(spec *api.Spec, data []byte) error {
		return ghodss.Unmarshal(data, spec)
	})); err != nil {
		return err
	}
	*c = FixtureAIConfig(fields)
	return nil
}

func (c FixtureAIConfig) MarshalYAML() ([]byte, error) {
	return yaml.MarshalWithOptions(fixtureAIFields(c), yaml.CustomMarshaler[api.Spec](func(spec api.Spec) ([]byte, error) {
		return ghodss.Marshal(spec)
	}))
}
