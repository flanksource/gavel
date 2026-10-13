package processruntime

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	cexec "github.com/flanksource/clicky/exec"
)

type Options struct {
	Env     map[string]string `json:"env,omitempty" yaml:"env,omitempty"`
	Nice    *int              `json:"nice,omitempty" yaml:"nice,omitempty"`
	CPU     *float64          `json:"cpu,omitempty" yaml:"cpu,omitempty"`
	Mem     *string           `json:"mem,omitempty" yaml:"mem,omitempty"`
	Timeout *string           `json:"timeout,omitempty" yaml:"timeout,omitempty"`
}

func (o Options) Merge(override Options) Options {
	out := o
	out.Env = make(map[string]string, len(o.Env)+len(override.Env))
	for key, value := range o.Env {
		out.Env[key] = value
	}
	for key, value := range override.Env {
		out.Env[key] = value
	}
	if override.Nice != nil {
		out.Nice = override.Nice
	}
	if override.CPU != nil {
		out.CPU = override.CPU
	}
	if override.Mem != nil {
		out.Mem = override.Mem
	}
	if override.Timeout != nil {
		out.Timeout = override.Timeout
	}
	return out
}

func (o Options) Validate() error {
	for key, value := range o.Env {
		if key == "" || strings.ContainsAny(key, "=\x00") || strings.ContainsRune(value, '\x00') {
			return fmt.Errorf("invalid runtime env entry %q", key)
		}
	}
	if o.Nice != nil && (*o.Nice < -20 || *o.Nice > 19) {
		return fmt.Errorf("runtime nice must be between -20 and 19")
	}
	if o.CPU != nil && (*o.CPU < 0 || math.IsNaN(*o.CPU) || math.IsInf(*o.CPU, 0)) {
		return fmt.Errorf("runtime cpu must be finite and nonnegative")
	}
	if o.Mem != nil && *o.Mem != "" {
		mem, err := humanize.ParseBytes(*o.Mem)
		if err != nil {
			return fmt.Errorf("runtime mem: %w", err)
		}
		if mem == 0 {
			return fmt.Errorf("runtime mem must be positive or empty")
		}
	}
	if o.Timeout != nil && *o.Timeout != "" {
		d, err := time.ParseDuration(*o.Timeout)
		if err != nil || d <= 0 {
			return fmt.Errorf("runtime timeout must be a positive duration, got %q", *o.Timeout)
		}
	}
	return nil
}

func (o Options) Isolated() bool {
	return o.CPU != nil && *o.CPU > 0 || o.Mem != nil && *o.Mem != "" || o.Timeout != nil && *o.Timeout != ""
}

func (o Options) Apply(p *cexec.Process) error {
	if err := o.Validate(); err != nil {
		return err
	}
	p.WithEnv(o.Env)
	if o.Nice != nil {
		p.WithNiceness(*o.Nice)
	}
	limits := cexec.ResourceLimits{}
	if o.CPU != nil {
		limits.MaxCPUPercent = *o.CPU
	}
	if o.Mem != nil && *o.Mem != "" {
		mem, err := humanize.ParseBytes(*o.Mem)
		if err != nil {
			return fmt.Errorf("runtime mem: %w", err)
		}
		limits.MaxRSSBytes = mem
	}
	p.WithResourceLimits(limits)
	if o.Timeout != nil && *o.Timeout == "" {
		p.WithTimeout(0)
	}
	if o.Timeout != nil && *o.Timeout != "" {
		timeout, err := time.ParseDuration(*o.Timeout)
		if err != nil {
			return fmt.Errorf("runtime timeout: %w", err)
		}
		p.WithTimeout(timeout)
	}
	return nil
}
