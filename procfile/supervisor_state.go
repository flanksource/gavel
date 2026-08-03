package procfile

import (
	"fmt"
	"os"

	"github.com/dustin/go-humanize"
	cexec "github.com/flanksource/clicky/exec"
	"github.com/flanksource/gavel/verify"
)

func (m *managed) snapshot() ProcState {
	m.mu.Lock()
	p := m.proc
	m.mu.Unlock()

	st := ProcState{
		Name:    m.entry.Name,
		Command: m.entry.Command,
		LogFile: m.logPath,
		Status:  StatusStopped,
	}
	if p == nil {
		return st
	}
	res := p.Resources()
	st.PID = p.Pid()
	st.Status = string(p.Status())
	st.Started = p.Started()
	st.Restarts = p.Restarts()
	st.ExitCode = p.ExitCode()
	st.TaskRunID = p.TaskRunID()
	st.Ports = p.Ports()
	st.CPUPercent = res.CPUPercent
	st.MemoryRSS = res.RSSBytes
	st.MemoryVMS = res.VMSBytes
	st.OpenFiles = res.OpenFiles
	peak := p.Peak()
	st.PeakCPU = peak.CPUPercent
	st.PeakRSS = peak.RSSBytes
	st.PeakVMS = peak.VMSBytes
	st.PeakFiles = peak.OpenFiles
	st.Tree = procTree(p.Tree())
	return st
}

// procTree maps clicky's per-process samples to the wire ProcNode shape.
func procTree(samples []cexec.ProcessSample) []ProcNode {
	if len(samples) == 0 {
		return nil
	}
	nodes := make([]ProcNode, len(samples))
	for i, s := range samples {
		nodes[i] = ProcNode{
			PID:        int(s.PID),
			PPID:       int(s.PPID),
			Command:    s.Command,
			Status:     s.Status,
			Root:       s.IsRoot,
			CPUPercent: s.CPUPercent,
			MemoryRSS:  s.RSSBytes,
			MemoryVMS:  s.VMSBytes,
			OpenFiles:  s.OpenFiles,
		}
	}
	return nodes
}

// resolveProfile picks the active profile: the --profile flag wins, else the
// .gavel.yaml default.
func resolveProfile(opts Options) string {
	if opts.Profile != "" {
		return opts.Profile
	}
	return opts.Config.Profile
}

// resolvePolicy resolves the restart policy + cap for an entry, falling back to
// the global config. Maps the config's verify.RestartPolicy to clicky's enum
// (identical values).
func resolvePolicy(cfg verify.ProcfileConfig, e Entry) (cexec.RestartPolicy, int) {
	policy := cfg.AutoRestart
	if e.AutoRestart != "" {
		policy = e.AutoRestart
	}
	maxR := cfg.MaxRestarts
	if e.MaxRestarts != nil {
		maxR = *e.MaxRestarts
	}
	return cexec.RestartPolicy(policy), maxR
}

// resolveLimits resolves the resource limits for an entry, falling back to the
// global config. An unparseable mem is a loud error so a typo fails the start.
func resolveLimits(cfg verify.ProcfileConfig, e Entry) (cexec.ResourceLimits, error) {
	mem := cfg.Mem
	if e.Mem != "" {
		mem = e.Mem
	}
	cpu := cfg.CPU
	if e.CPU != 0 {
		cpu = e.CPU
	}
	var rss uint64
	if mem != "" {
		parsed, err := humanize.ParseBytes(mem)
		if err != nil {
			return cexec.ResourceLimits{}, fmt.Errorf("invalid mem %q for process %q: %w", mem, e.Name, err)
		}
		rss = parsed
	}
	return cexec.ResourceLimits{MaxRSSBytes: rss, MaxCPUPercent: cpu}, nil
}

func contains(xs []string, x string) bool {
	for _, v := range xs {
		if v == x {
			return true
		}
	}
	return false
}

func truncateFile(path string) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("truncate %s: %w", path, err)
	}
	return f.Close()
}
