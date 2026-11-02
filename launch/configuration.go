package launch

import "encoding/json"

func (c Configuration) MarshalJSON() ([]byte, error) {
	sections := map[string]any{}
	if c.Test != nil {
		sections["test"] = c.Test
	}
	if c.Lint != nil {
		sections["lint"] = c.Lint
	}
	return json.Marshal(sections)
}
