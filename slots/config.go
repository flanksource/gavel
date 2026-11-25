package slots

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type Value string

func (v *Value) UnmarshalJSON(data []byte) error {
	if len(data) > 0 && data[0] == '"' {
		return json.Unmarshal(data, (*string)(v))
	}
	*v = Value(data)
	return nil
}

type Config struct {
	Global         *Value `json:"global,omitempty" yaml:"global,omitempty"`
	ProjectDefault *Value `json:"projectDefault,omitempty" yaml:"projectDefault,omitempty"`
	Project        *Value `json:"project,omitempty" yaml:"project,omitempty"`
}

type Limits struct {
	Global  int `json:"global,omitempty"`
	Project int `json:"project,omitempty"`
}

func (l Limits) Enabled() bool { return l.Global > 0 || l.Project > 0 }

func Parse(key string, value Value, cores int) (int, error) {
	raw := string(value)
	percentage := strings.HasSuffix(raw, "%")
	if percentage {
		raw = strings.TrimSuffix(raw, "%")
	}
	for _, digit := range raw {
		if digit < '0' || digit > '9' {
			return 0, fmt.Errorf("%s: invalid slot value %q", key, value)
		}
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 1 || percentage && n > 100 {
		return 0, fmt.Errorf("%s: expected a positive integer or a percentage from 1%% to 100%%, got %q", key, value)
	}
	if percentage {
		if cores < 1 {
			return 0, fmt.Errorf("%s: core count must be positive, got %d", key, cores)
		}
		n = max(1, cores/100*n+cores%100*n/100)
	}
	return n, nil
}

func (c Config) Validate(cores int) error {
	for _, field := range []struct {
		key   string
		value *Value
	}{{"slots.global", c.Global}, {"slots.projectDefault", c.ProjectDefault}, {"slots.project", c.Project}} {
		if field.value != nil {
			if _, err := Parse(field.key, *field.value, cores); err != nil {
				return err
			}
		}
	}
	return nil
}

func Resolve(global, project Config, cores int) (Limits, error) {
	if err := global.Validate(cores); err != nil {
		return Limits{}, err
	}
	if err := project.Validate(cores); err != nil {
		return Limits{}, err
	}
	var limits Limits
	var err error
	if global.Global != nil {
		limits.Global, err = Parse("slots.global", *global.Global, cores)
		if err != nil {
			return Limits{}, err
		}
	}
	projectValue, key := global.ProjectDefault, "slots.projectDefault"
	if project.Project != nil {
		projectValue, key = project.Project, "slots.project"
	}
	if projectValue != nil {
		limits.Project, err = Parse(key, *projectValue, cores)
		if err != nil {
			return Limits{}, err
		}
	}
	return limits, nil
}
