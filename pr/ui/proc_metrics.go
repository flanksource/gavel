package ui

import (
	"fmt"
	"time"

	"github.com/flanksource/clicky/metrics"
	"github.com/flanksource/gavel/procfile"
)

const (
	// procSampleInterval is how often the sampler records a CPU/memory point per
	// process. It matches clicky's own resource-monitor cadence so we don't
	// oversample the supervisor's cached readings.
	procSampleInterval = 2 * time.Second
	// procSampleIdle gates sampling to recent dashboard activity: if no
	// /api/proc/status request has arrived within this window nobody is watching
	// the gauges, so we stop dialing every project's supervisor.
	procSampleIdle = 20 * time.Second
)

// procRunKey rebuilds the metric key the frontend computes in ProcessTable.tsx
// (runKey): "<project>/<name>/<started|not-started>/<pid|0>". The two MUST stay
// in sync — the browser requests /api/proc/metrics/{this key}/{cpu|memory} and
// the sampler records under the same id. proc.started on the wire is Go's JSON
// marshaling of ProcState.Started (RFC3339Nano), reproduced here with Format.
func procRunKey(project string, ps procfile.ProcState) string {
	started := "not-started"
	if ps.Started != nil {
		started = ps.Started.Format(time.RFC3339Nano)
	}
	return fmt.Sprintf("%s/%s/%s/%d", project, ps.Name, started, ps.PID)
}

// sampleProcMetrics records one CPU and one memory point for every running
// process across all projects, plus a per-project "__total__" series that the
// workspace header gauges read. Recording is best-effort instrumentation.
func (s *Server) sampleProcMetrics() {
	now := time.Now()
	for _, p := range LoadProjects() {
		st := projectStatus(p)
		if !st.Running {
			continue
		}
		var totalCPU float64
		var totalMem uint64
		for _, proc := range st.Processes {
			if proc.PID == 0 {
				continue
			}
			key := procRunKey(p.Name, proc)
			s.recordProcPoint(key+"/cpu", now, proc.CPUPercent)
			s.recordProcPoint(key+"/memory", now, float64(proc.MemoryRSS))
			totalCPU += proc.CPUPercent
			totalMem += proc.MemoryRSS
		}
		total := p.Name + "/__total__"
		s.recordProcPoint(total+"/cpu", now, totalCPU)
		s.recordProcPoint(total+"/memory", now, float64(totalMem))
	}
}

func (s *Server) recordProcPoint(id string, at time.Time, value float64) {
	_ = s.procMetrics.Record(metrics.RecordRequest{ID: id, At: at, Value: value})
}

// procMetricsLoop samples on a fixed cadence while the dashboard is being
// watched. It starts from NewServer and runs for the life of the process.
func (s *Server) procMetricsLoop() {
	t := time.NewTicker(procSampleInterval)
	defer t.Stop()
	for range t.C {
		s.mu.RLock()
		last := s.lastProcPoll
		s.mu.RUnlock()
		if last.IsZero() || time.Since(last) > procSampleIdle {
			continue
		}
		s.sampleProcMetrics()
	}
}
