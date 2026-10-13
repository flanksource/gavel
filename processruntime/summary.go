package processruntime

import (
	"fmt"
	"slices"
	"strings"
)

type Summary struct {
	Env     []string `json:"env,omitempty" yaml:"env,omitempty"`
	Nice    *int     `json:"nice,omitempty" yaml:"nice,omitempty"`
	CPU     *float64 `json:"cpu,omitempty" yaml:"cpu,omitempty"`
	Mem     *string  `json:"mem,omitempty" yaml:"mem,omitempty"`
	Timeout *string  `json:"timeout,omitempty" yaml:"timeout,omitempty"`
}

func (o Options) Summary() Summary {
	names := make([]string, 0, len(o.Env))
	for name := range o.Env {
		names = append(names, name)
	}
	slices.Sort(names)
	return Summary{Env: names, Nice: o.Nice, CPU: o.CPU, Mem: o.Mem, Timeout: o.Timeout}
}

func (s Summary) String() string {
	var fields []string
	if len(s.Env) > 0 {
		fields = append(fields, "env=["+strings.Join(s.Env, ", ")+"]")
	}
	if s.Nice != nil {
		fields = append(fields, fmt.Sprintf("nice=%d", *s.Nice))
	}
	if s.CPU != nil {
		fields = append(fields, fmt.Sprintf("cpu=%g%%", *s.CPU))
	}
	if s.Mem != nil {
		fields = append(fields, fmt.Sprintf("mem=%q", *s.Mem))
	}
	if s.Timeout != nil {
		fields = append(fields, fmt.Sprintf("timeout=%q", *s.Timeout))
	}
	return strings.Join(fields, " ")
}
