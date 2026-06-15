package procfile

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/clicky/api/icons"
)

// Report is the merged view of a Procfile and the live supervisor state. Every
// Procfile entry appears exactly once; entries with no live state render as
// stopped. It backs both StatusReport and ListReport.
type Report struct {
	Root          string      `json:"root"`
	Procfile      string      `json:"procfile"`
	Running       bool        `json:"running"`
	SupervisorPID int         `json:"supervisorPid,omitempty"`
	Processes     []ProcState `json:"processes"`
	// Profiles is the set of profiles declared across the Procfile entries
	// (sorted, deduped). Empty when no entry declares a profile.
	Profiles []string `json:"profiles,omitempty"`
	// Profile is the active profile: the running supervisor's, or the .gavel.yaml
	// default when not running (i.e. what the next start would use).
	Profile string `json:"profile,omitempty"`
}

// StatusReport renders the full per-process status table.
type StatusReport struct{ Report }

// ListReport renders the configured processes and their commands.
type ListReport struct{ Report }

// gather merges the Procfile entries with the live (socket) or last-persisted
// (state.json) supervisor state for root.
func gather(root, pf string) (Report, error) {
	dir, err := StateDir(root)
	if err != nil {
		return Report{}, err
	}
	entries, err := Load(pf)
	if err != nil {
		return Report{}, err
	}
	st, err := ReadState(dir)
	if err != nil {
		return Report{}, err
	}
	running := st.Running()
	if running {
		if resp, err := sendControl(root, ctrlRequest{Action: actionStatus}); err == nil {
			st = resp.State
		}
	}

	// Active profile: the running supervisor's (from the socket), or the
	// .gavel.yaml default when not running (what the next start would use).
	profile := st.Profile
	if !running {
		if cfg, err := loadConfig(root); err == nil {
			profile = cfg.Profile
		}
	}

	byName := make(map[string]ProcState, len(st.Processes))
	for _, p := range st.Processes {
		byName[p.Name] = p
	}
	procs := make([]ProcState, 0, len(entries))
	for _, e := range entries {
		if p, ok := byName[e.Name]; ok {
			if p.Command == "" {
				p.Command = e.Command
			}
			procs = append(procs, p)
			continue
		}
		procs = append(procs, ProcState{Name: e.Name, Command: e.Command, Status: StatusStopped, LogFile: LogPath(dir, e.Name)})
	}
	return Report{
		Root:          root,
		Procfile:      pf,
		Running:       running,
		SupervisorPID: st.SupervisorPID,
		Processes:     procs,
		Profiles:      availableProfiles(entries),
		Profile:       profile,
	}, nil
}

// availableProfiles returns the sorted, deduped set of profiles declared across
// the Procfile entries. Empty when no entry declares a profile.
func availableProfiles(entries []Entry) []string {
	seen := map[string]struct{}{}
	for _, e := range entries {
		for _, p := range e.Profiles {
			seen[p] = struct{}{}
		}
	}
	if len(seen) == 0 {
		return nil
	}
	out := make([]string, 0, len(seen))
	for p := range seen {
		out = append(out, p)
	}
	sort.Strings(out)
	return out
}

func (r Report) header() api.Text {
	t := clicky.Text("Procfile: ", "text-muted").Append(r.Procfile)
	if r.Running {
		t = t.Space().Append("(supervisor pid ", "text-muted").Append(r.SupervisorPID).Append(")", "text-muted")
	} else {
		t = t.Space().Append("(not running)", "text-muted")
	}
	return t.NewLine()
}

// Pretty renders the status table: icon, name, status, pid, uptime, restarts.
func (r StatusReport) Pretty() api.Text {
	t := r.header()
	for _, p := range r.Processes {
		t = t.Add(statusIcon(p.Status)).Space().
			Append(fmt.Sprintf("%-12s", p.Name), "font-bold").
			Append(fmt.Sprintf("%-12s", p.Status))
		if p.PID > 0 {
			t = t.Append("pid ", "text-muted").Append(fmt.Sprintf("%-8d", p.PID))
		} else {
			t = t.Append(fmt.Sprintf("%-12s", ""))
		}
		if up := uptime(p.Started, p.Status); up != "" {
			t = t.Append("up ", "text-muted").Append(up).Space()
		}
		if len(p.Ports) > 0 {
			t = t.Append(PortsLabel(p.Ports), "text-blue-500").Space()
		}
		if p.Restarts > 0 {
			t = t.Append("restarts ", "text-muted").Append(p.Restarts).Space()
		}
		if p.Status == StatusRunning {
			if p.CPUPercent > 0 {
				t = t.Append("cpu ", "text-muted").Append(fmt.Sprintf("%.0f%% ", p.CPUPercent))
			}
			if p.MemoryRSS > 0 {
				t = t.Append("mem ", "text-muted").Append(humanize.Bytes(p.MemoryRSS)).Space()
			}
			if p.OpenFiles > 0 {
				t = t.Append("fds ", "text-muted").Append(p.OpenFiles)
			}
		}
		t = t.NewLine()
	}
	return t
}

// Pretty renders the configured processes and their commands.
func (r ListReport) Pretty() api.Text {
	t := r.header()
	for _, p := range r.Processes {
		t = t.Add(statusIcon(p.Status)).Space().
			Append(fmt.Sprintf("%-12s", p.Name), "font-bold").
			Append(p.Command, "text-muted").NewLine()
	}
	return t
}

// PortsLabel renders a process's listening ports as space-separated
// `http://localhost:PORT` URLs — terminals auto-linkify the scheme-qualified
// form, so the ports are click-to-open. Shared by the status table and the CLI
// start/restart progress view so both spell ports the same way.
func PortsLabel(ports []int) string {
	parts := make([]string, len(ports))
	for i, p := range ports {
		parts[i] = fmt.Sprintf("http://localhost:%d", p)
	}
	return strings.Join(parts, " ")
}

func statusIcon(status string) api.Text {
	switch status {
	case StatusRunning:
		return api.Text{}.Add(icons.Success)
	case StatusRestarting, StatusStarting:
		return api.Text{}.Add(icons.Warning)
	case StatusCrashed:
		return api.Text{}.Add(icons.Error)
	default:
		return api.Text{}.Add(icons.Pending)
	}
}

func uptime(started *time.Time, status string) string {
	if started == nil || status != StatusRunning {
		return ""
	}
	d := time.Since(*started).Truncate(time.Second)
	if d < 0 {
		return ""
	}
	return d.String()
}
