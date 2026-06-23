package ui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/github/cache"
	"github.com/flanksource/gavel/procfile"
)

// projectInfo is the wire shape for GET /api/projects. Dir is resolved (~
// expanded) and hasProcfile reports whether the directory currently contains a
// Procfile, so the frontend knows which repo headers can show process controls.
type projectInfo struct {
	Name        string   `json:"name"`
	Dir         string   `json:"dir"`
	Repos       []string `json:"repos"`
	HasProcfile bool     `json:"hasProcfile"`
	// TodoProvider echoes the configured provider ("grite"/"todos"/"" for auto)
	// so the dashboard can scope its per-workspace todo requests the same way.
	TodoProvider string `json:"todoProvider,omitempty"`
	// TodoBackend is the backend actually resolved for the directory ("grite" or
	// "todos"); TodoBackendAuto is true when it was auto-detected rather than
	// pinned by TodoProvider.
	TodoBackend     string     `json:"todoBackend"`
	TodoBackendAuto bool       `json:"todoBackendAuto"`
	TodoCounts      todoCounts `json:"todoCounts"`
}

// procStatus is the wire shape for /api/proc/status. hasProcfile=false is the
// normal "no Procfile in this directory" state, not an error; the supervisor
// fields are only meaningful when hasProcfile is true.
type procStatus struct {
	HasProcfile   bool                 `json:"hasProcfile"`
	Running       bool                 `json:"running"`
	SupervisorPID int                  `json:"supervisorPid,omitempty"`
	Processes     []procfile.ProcState `json:"processes,omitempty"`
	// Profiles are the profiles declared in the Procfile; Profile is the active
	// one (running supervisor's, else the .gavel.yaml default).
	Profiles []string `json:"profiles,omitempty"`
	Profile  string   `json:"profile,omitempty"`
	// GitChanges counts uncommitted changes (staged, unstaged, and untracked) in
	// the project's directory. Omitted when the directory is not a git work tree.
	GitChanges int    `json:"gitChanges,omitempty"`
	Error      string `json:"error,omitempty"`
}

// procControl is the request body for the start/stop/restart endpoints. Profile
// applies only when start spawns a new daemon (which set of processes auto-start).
type procControl struct {
	Project string   `json:"project"`
	Names   []string `json:"names,omitempty"`
	Profile string   `json:"profile,omitempty"`
}

func (s *Server) handleProcStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	// Mark the dashboard as actively watched so procMetricsLoop keeps sampling.
	s.mu.Lock()
	s.lastProcPoll = time.Now()
	s.mu.Unlock()
	ps := LoadProjects()

	if name := r.URL.Query().Get("project"); name != "" {
		p, ok := projectByName(ps, name)
		if !ok {
			http.Error(w, `{"error":"unknown project"}`, http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(projectStatus(p)) //nolint:errcheck
		return
	}

	// No project param: return every project's status (see procStatusByKey).
	json.NewEncoder(w).Encode(procStatusByKey()) //nolint:errcheck
}

// procStatusByKey returns every project's status keyed by both project name (so
// repo-less projects in the pinned Projects bar resolve) and each repo slug (so
// the sidebar repo headers light up). Project names are bare and repos contain a
// slash, so the keyspaces don't collide. Shared by handleProcStatus so the
// single-shot poll carries the full wire shape (live cpu/mem + process tree).
func procStatusByKey() map[string]procStatus {
	byKey := make(map[string]procStatus)
	for _, p := range LoadProjects() {
		st := projectStatus(p)
		byKey[p.Name] = st
		for _, repo := range p.Repos {
			byKey[repo] = st
		}
	}
	return byKey
}

// streamProcStatusByKey is procStatusByKey projected for the SSE stream — see
// leanProcStatus for why the resource fields are dropped.
func streamProcStatusByKey() map[string]procStatus {
	return leanProcStatus(procStatusByKey())
}

// leanProcStatus clears the continuously-fluctuating resource fields from every
// process: CPUPercent, MemoryRSS and the per-process Tree. Those churn on every
// supervisor sample, so streaming them defeats handleProcStatusStream's
// change-detection — the dashboard would receive a re-render-firing data frame
// every cadence instead of a cheap keep-alive ping. The per-process gauges
// already own live cpu/mem via /api/proc/metrics and the expanded row fetches
// its tree on demand, so the stream only needs the stable supervision state.
// OpenFiles is kept: it changes slowly and feeds the always-visible Files
// column, so it rarely defeats the diff. Mutates and returns byKey.
func leanProcStatus(byKey map[string]procStatus) map[string]procStatus {
	for key, st := range byKey {
		if len(st.Processes) == 0 {
			continue
		}
		lean := make([]procfile.ProcState, len(st.Processes))
		for i, p := range st.Processes {
			p.CPUPercent = 0
			p.MemoryRSS = 0
			p.Tree = nil
			lean[i] = p
		}
		st.Processes = lean
		byKey[key] = st
	}
	return byKey
}

const (
	// procStreamFast is the push cadence while a process is starting/restarting,
	// so the dashboard tracks that progress promptly; procStreamSteady is the
	// idle cadence. They mirror the adaptive interval the client poll used to run.
	procStreamFast   = 1 * time.Second
	procStreamSteady = 3 * time.Second
)

// handleProcStatusStream pushes the proc-status map to the dashboard over SSE,
// replacing the client's /api/proc/status poll. A per-connection adaptive ticker
// is enough — projectStatus recomputes from the supervisor on each tick, so no
// broadcaster is needed — and an open stream is the "dashboard is being watched"
// signal that keeps procMetricsLoop sampling (the role the poll used to play).
func (s *Server) handleProcStatusStream(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	var last []byte
	for {
		s.mu.Lock()
		s.lastProcPoll = time.Now()
		s.mu.Unlock()

		byKey := streamProcStatusByKey()
		if b, err := json.Marshal(byKey); err == nil && !bytes.Equal(b, last) {
			fmt.Fprintf(w, "data: %s\n\n", b)
			last = b
		} else {
			// Comment frame: keeps the socket warm without firing a client re-render.
			fmt.Fprint(w, ": ping\n\n")
		}
		flusher.Flush()

		next := procStreamSteady
		if anyTransitioning(byKey) {
			next = procStreamFast
		}
		select {
		case <-r.Context().Done():
			return
		case <-time.After(next):
		}
	}
}

// anyTransitioning reports whether any process is mid start/restart, selecting
// the faster stream cadence so the UI tracks that progress.
func anyTransitioning(byKey map[string]procStatus) bool {
	for _, st := range byKey {
		for _, p := range st.Processes {
			if p.Status == procfile.StatusStarting || p.Status == procfile.StatusRestarting {
				return true
			}
		}
	}
	return false
}

// handleProcFavicon fetches a favicon from a localhost service that Gavel has
// already discovered as an open process port. The project+port guard keeps this
// from becoming an arbitrary URL proxy.
func (s *Server) handleProcFavicon(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	project := r.URL.Query().Get("project")
	portStr := r.URL.Query().Get("port")
	if project == "" || portStr == "" {
		http.Error(w, "project and port params are required", http.StatusBadRequest)
		return
	}
	port, err := strconv.Atoi(portStr)
	if err != nil || port <= 0 || port > 65535 {
		http.Error(w, "invalid port", http.StatusBadRequest)
		return
	}
	p, ok := projectByName(LoadProjects(), project)
	if !ok {
		http.Error(w, "unknown project", http.StatusNotFound)
		return
	}
	st := projectStatus(p)
	if !st.HasProcfile || !procStatusHasPort(st, port) {
		http.Error(w, "unknown process port", http.StatusNotFound)
		return
	}

	homepage := "http://localhost:" + strconv.Itoa(port)
	store := cache.Shared()
	data, mime, hit, err := store.GetFavicon(homepage)
	if err != nil {
		logger.Warnf("process favicon cache read %s: %v", homepage, err)
	}
	if !hit {
		data, mime, err = store.FetchFavicon(r.Context(), homepage)
		if err != nil {
			logger.Debugf("process favicon fetch %s: %v", homepage, err)
			http.Error(w, "favicon unavailable", http.StatusNotFound)
			return
		}
	}
	if len(data) == 0 {
		http.Error(w, "no favicon", http.StatusNotFound)
		return
	}
	if mime != "" {
		w.Header().Set("Content-Type", mime)
	}
	w.Header().Set("Cache-Control", "public, max-age=86400")
	_, _ = w.Write(data)
}

func procStatusHasPort(st procStatus, port int) bool {
	for _, p := range st.Processes {
		for _, candidate := range p.Ports {
			if candidate == port {
				return true
			}
		}
	}
	return false
}

// handleProcControl backs POST /api/proc/{start,stop,restart}. The action is
// taken from the final path segment so the three routes share one handler.
func (s *Server) handleProcControl(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	var body procControl
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}
	p, ok := projectByName(LoadProjects(), body.Project)
	if !ok {
		http.Error(w, `{"error":"unknown project"}`, http.StatusNotFound)
		return
	}
	dir := p.ResolvedDir()
	if dir == "" || procfile.Find(dir, "") == "" {
		http.Error(w, `{"error":"no Procfile for project"}`, http.StatusBadRequest)
		return
	}

	// body.Profile is honoured when start/restart spawns a fresh daemon (it
	// chooses which processes auto-start); it's ignored when acting on a process
	// by name on an already-running daemon.
	var actErr error
	switch path.Base(r.URL.Path) {
	case "start":
		_, actErr = procfile.Start(dir, "", body.Names, body.Profile)
	case "stop":
		_, actErr = procfile.Stop(dir, "", body.Names)
	case "restart":
		_, actErr = procfile.Restart(dir, "", body.Names, body.Profile)
	default:
		http.Error(w, `{"error":"unknown action"}`, http.StatusNotFound)
		return
	}

	if actErr != nil {
		writeJSONError(w, http.StatusInternalServerError, actErr)
		return
	}
	json.NewEncoder(w).Encode(projectStatus(p)) //nolint:errcheck
}

// handleProcLogs tails the last N lines of a project's process logs as plain
// text. ?name selects a single process; omitting it returns every process.
func (s *Server) handleProcLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	p, ok := projectByName(LoadProjects(), r.URL.Query().Get("project"))
	if !ok {
		http.Error(w, "unknown project", http.StatusNotFound)
		return
	}
	dir := p.ResolvedDir()
	if dir == "" || procfile.Find(dir, "") == "" {
		http.Error(w, "no Procfile for project", http.StatusNotFound)
		return
	}

	lines := 200
	if v := r.URL.Query().Get("lines"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			lines = n
		}
	}
	var names []string
	if proc := r.URL.Query().Get("name"); proc != "" {
		names = []string{proc}
	}

	// Buffer so an error (e.g. unknown process name) yields a clean status
	// instead of a half-written 200.
	var buf bytes.Buffer
	if err := procfile.Logs(dir, "", names, lines, false, &buf); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write(buf.Bytes()) //nolint:errcheck
}

// projectStatus resolves a project's directory and returns its Procfile status.
// A directory without a Procfile is reported as hasProcfile=false (not an error)
// so projects that aren't running anything render cleanly.
func projectStatus(p Project) procStatus {
	dir := p.ResolvedDir()

	// Uncommitted-change count is a property of the workspace directory, not of
	// Procfile supervision, so it is surfaced for every workspace in the sidebar
	// regardless of whether the directory has a Procfile.
	var st procStatus
	if dir != "" {
		if n, err := gitChangeCount(dir); err != nil {
			logger.Debugf("git status %s: %v", dir, err)
		} else {
			st.GitChanges = n
		}
	}

	if dir == "" || procfile.Find(dir, "") == "" {
		return st
	}
	st.HasProcfile = true
	rep, err := procfile.Status(dir, "")
	if err != nil {
		st.Error = err.Error()
		return st
	}
	st.Running = rep.Running
	st.SupervisorPID = rep.SupervisorPID
	st.Processes = rep.Processes
	st.Profiles = rep.Profiles
	st.Profile = rep.Profile
	return st
}

// gitChangeCount returns the number of uncommitted changes (staged, unstaged,
// and untracked) in dir. A non-nil error means dir is not a git work tree (or
// git is unavailable); callers treat that as "no git info" rather than zero
// changes.
func gitChangeCount(dir string) (int, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	count := 0
	for _, line := range strings.Split(string(out), "\n") {
		if strings.TrimSpace(line) != "" {
			count++
		}
	}
	return count, nil
}

func writeJSONError(w http.ResponseWriter, status int, err error) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": err.Error()}) //nolint:errcheck
}
