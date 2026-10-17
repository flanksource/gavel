package shims

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/flanksource/gavel/processruntime"
)

// InstallOptions configures Install.
type InstallOptions struct {
	Dir       string
	GavelPath string
	Tools     []string
	Settings  Settings
	DryRun    bool
	Force     bool
}

// Result describes one installed (or planned) shim.
type Result struct {
	Tool   string `json:"tool"`
	Shim   string `json:"shim"`
	Action string `json:"action"`
	Real   string `json:"real,omitempty"`
	Active bool   `json:"active"`
	Note   string `json:"note,omitempty"`
}

// InstallReport is the outcome of Install.
type InstallReport struct {
	Dir      string   `json:"dir"`
	Shims    []Result `json:"shims"`
	PathHint string   `json:"pathHint,omitempty"`
}

// Install writes one shim per tool into opts.Dir. It refuses to overwrite a
// file that is not a gavel shim unless opts.Force is set.
func Install(opts InstallOptions) (InstallReport, error) {
	if opts.Dir == "" || opts.GavelPath == "" {
		return InstallReport{}, errors.New("shim dir and gavel path are required")
	}
	if _, _, err := opts.Settings.Runtime(); err != nil {
		return InstallReport{}, err
	}
	for _, tool := range opts.Tools {
		if _, ok := Profiles[tool]; !ok {
			return InstallReport{}, fmt.Errorf("unknown tool %q (supported: %v)", tool, ToolNames())
		}
	}
	if !opts.DryRun {
		if err := os.MkdirAll(opts.Dir, 0o755); err != nil {
			return InstallReport{}, fmt.Errorf("create shim dir: %w", err)
		}
	}
	report := InstallReport{Dir: opts.Dir}
	for _, tool := range opts.Tools {
		result, err := installShim(opts, tool)
		if err != nil {
			return report, err
		}
		report.Shims = append(report.Shims, result)
	}
	if slices.ContainsFunc(report.Shims, func(r Result) bool { return !r.Active }) {
		report.PathHint = fmt.Sprintf("export PATH=%q:$PATH", opts.Dir)
	}
	return report, nil
}

func installShim(opts InstallOptions, tool string) (Result, error) {
	path := filepath.Join(opts.Dir, tool)
	result := Result{Tool: tool, Shim: path, Action: "created"}
	if _, err := os.Lstat(path); err == nil {
		if !IsShim(path) && !opts.Force {
			return result, fmt.Errorf("%s exists and is not a gavel shim (use --force to overwrite)", path)
		}
		result.Action = "updated"
	}
	if opts.DryRun {
		result.Action = "would be " + result.Action
	} else if err := os.WriteFile(path, []byte(Render(tool, opts.GavelPath, opts.Settings)), 0o755); err != nil {
		return result, fmt.Errorf("write shim %s: %w", path, err)
	} else if err := os.Chmod(path, 0o755); err != nil {
		return result, fmt.Errorf("chmod shim %s: %w", path, err)
	}
	if real, err := LookPath(tool); err == nil {
		result.Real = real
	} else {
		result.Note = fmt.Sprintf("%s is not installed; the shim fails until it is", tool)
	}
	if found, err := exec.LookPath(tool); err == nil {
		result.Active = found == path
	}
	return result, nil
}

// Runtime converts the shim settings into the process runtime limits, the
// throttles derived from them, and the hard timeout (0 = none).
func (s Settings) Runtime() (processruntime.Options, Limits, error) {
	if s.CPU < 0 {
		return processruntime.Options{}, Limits{}, fmt.Errorf("cpu limit must be nonnegative, got %v", s.CPU)
	}
	opts := processruntime.Options{Nice: &s.Nice}
	limits := Limits{Slots: Slots(s.CPU)}
	if s.CPU > 0 {
		opts.CPU = &s.CPU
	}
	if s.Mem != "" {
		mem, err := humanize.ParseBytes(s.Mem)
		if err != nil {
			return opts, limits, fmt.Errorf("mem limit %q: %w", s.Mem, err)
		}
		opts.Mem, limits.MemBytes = &s.Mem, mem
	}
	if d, err := time.ParseDuration(s.Timeout); err != nil || d < 0 {
		return opts, limits, fmt.Errorf("timeout must be a nonnegative duration (0 disables it), got %q", s.Timeout)
	}
	return opts, limits, opts.Validate()
}

// TimeoutDuration returns the hard timeout, or 0 when it is disabled.
func (s Settings) TimeoutDuration() time.Duration {
	d, _ := time.ParseDuration(s.Timeout)
	return d
}

// SettingsFromEnv reads the settings a shim script passes to shim-exec and
// fails when any of them is missing, so a hand-edited shim cannot silently
// run unlimited.
func SettingsFromEnv(getenv func(string) string) (Settings, error) {
	var missing []string
	for _, key := range []string{EnvTimeout, EnvCPU, EnvNice} {
		if getenv(key) == "" {
			missing = append(missing, key)
		}
	}
	if len(missing) > 0 {
		return Settings{}, fmt.Errorf("shim-exec is run by gavel shims; missing %v (re-run gavel lint install-shims)", missing)
	}
	cpu, err := strconv.ParseFloat(getenv(EnvCPU), 64)
	if err != nil {
		return Settings{}, fmt.Errorf("%s: %w", EnvCPU, err)
	}
	nice, err := strconv.Atoi(getenv(EnvNice))
	if err != nil {
		return Settings{}, fmt.Errorf("%s: %w", EnvNice, err)
	}
	settings := Settings{Timeout: getenv(EnvTimeout), CPU: cpu, Mem: getenv(EnvMem), Nice: nice}
	if _, _, err := settings.Runtime(); err != nil {
		return Settings{}, err
	}
	return settings, nil
}
