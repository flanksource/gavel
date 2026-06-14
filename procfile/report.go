package procfile

import (
	"fmt"
	"time"

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
	return Report{Root: root, Procfile: pf, Running: running, SupervisorPID: st.SupervisorPID, Processes: procs}, nil
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
		if p.Restarts > 0 {
			t = t.Append("restarts ", "text-muted").Append(p.Restarts)
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
