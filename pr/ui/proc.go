package ui

import (
	"bytes"
	"encoding/json"
	"net/http"
	"path"
	"strconv"

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
}

// procStatus is the wire shape for /api/proc/status. hasProcfile=false is the
// normal "no Procfile in this directory" state, not an error; the supervisor
// fields are only meaningful when hasProcfile is true.
type procStatus struct {
	HasProcfile   bool                 `json:"hasProcfile"`
	Running       bool                 `json:"running"`
	SupervisorPID int                  `json:"supervisorPid,omitempty"`
	Processes     []procfile.ProcState `json:"processes,omitempty"`
	Error         string               `json:"error,omitempty"`
}

// procControl is the request body for the start/stop/restart endpoints.
type procControl struct {
	Project string   `json:"project"`
	Names   []string `json:"names,omitempty"`
}

func (s *Server) handleProjects(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case http.MethodGet:
		ps := LoadProjects()
		out := make([]projectInfo, 0, len(ps))
		for _, p := range ps {
			dir := p.ResolvedDir()
			out = append(out, projectInfo{
				Name:        p.Name,
				Dir:         dir,
				Repos:       p.Repos,
				HasProcfile: dir != "" && procfile.Find(dir, "") != "",
			})
		}
		json.NewEncoder(w).Encode(out) //nolint:errcheck
	case http.MethodPost:
		var p Project
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
			return
		}
		if p.Name == "" || p.Dir == "" {
			http.Error(w, `{"error":"name and dir are required"}`, http.StatusBadRequest)
			return
		}
		SaveProjects(upsertProject(LoadProjects(), p))
		json.NewEncoder(w).Encode(p) //nolint:errcheck
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleProcStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
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

	// No project param: return every project's status keyed by both repo (so the
	// sidebar repo headers light up) and project name (so repo-less projects in
	// the pinned Projects bar resolve too). Project names are bare and repos
	// contain a slash, so the keyspaces don't collide.
	byKey := make(map[string]procStatus)
	for _, p := range ps {
		st := projectStatus(p)
		byKey[p.Name] = st
		for _, repo := range p.Repos {
			byKey[repo] = st
		}
	}
	json.NewEncoder(w).Encode(byKey) //nolint:errcheck
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

	var action func(string, string, []string) (*procfile.StatusReport, error)
	switch path.Base(r.URL.Path) {
	case "start":
		action = procfile.Start
	case "stop":
		action = procfile.Stop
	case "restart":
		action = procfile.Restart
	default:
		http.Error(w, `{"error":"unknown action"}`, http.StatusNotFound)
		return
	}

	if _, err := action(dir, "", body.Names); err != nil {
		writeJSONError(w, http.StatusInternalServerError, err)
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
	if dir == "" || procfile.Find(dir, "") == "" {
		return procStatus{HasProcfile: false}
	}
	rep, err := procfile.Status(dir, "")
	if err != nil {
		return procStatus{HasProcfile: true, Error: err.Error()}
	}
	return procStatus{
		HasProcfile:   true,
		Running:       rep.Running,
		SupervisorPID: rep.SupervisorPID,
		Processes:     rep.Processes,
	}
}

func writeJSONError(w http.ResponseWriter, status int, err error) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": err.Error()}) //nolint:errcheck
}
