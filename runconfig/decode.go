package runconfig

import (
	"bytes"
	"encoding/json"

	"github.com/ghodss/yaml"
)

func (o *Test) UnmarshalJSON(data []byte) error {
	type plain Test
	return strictDecode(data, (*plain)(o))
}

func (o *Lint) UnmarshalJSON(data []byte) error {
	type plain Lint
	return strictDecode(data, (*plain)(o))
}

func strictDecode(data []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}

func YAMLFields(data []byte) (map[string]bool, error) {
	encoded, err := yaml.YAMLToJSON(data)
	if err != nil {
		return nil, err
	}
	return JSONFields(encoded)
}
