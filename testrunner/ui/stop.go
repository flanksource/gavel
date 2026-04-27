package testui

import (
	"encoding/json"
	"net/http"

	"github.com/flanksource/clicky"
)

// StopRequest is the payload accepted by POST /api/stop.
type StopRequest struct {
	Scope  string `json:"scope"`
	TaskID string `json:"task_id,omitempty"`
}

// SetStopFunc installs the callback used by the UI to stop the active run.
func (s *Server) SetStopFunc(fn func()) {
	s.mu.Lock()
	s.stopFn = fn
	s.mu.Unlock()
	s.notify()
}

func (s *Server) requestStop(message string) {
	s.mu.Lock()
	s.stopRequested = true
	s.stopMessage = message
	s.mu.Unlock()
	s.notify()
}

func (s *Server) stopFunc() func() {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.stopFn
}

func (s *Server) handleStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req StopRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json: "+err.Error(), http.StatusBadRequest)
		return
	}
	if req.Scope == "" {
		req.Scope = "global"
	}

	if s.stopFunc() == nil {
		http.Error(w, "stop not supported", http.StatusNotImplemented)
		return
	}

	s.mu.RLock()
	running := s.snapshot().Status.Running
	s.mu.RUnlock()
	if !running {
		http.Error(w, "run is not active", http.StatusConflict)
		return
	}

	switch req.Scope {
	case "global":
		s.requestStop("Stopped by user")
		s.stopFunc()()
		clicky.CancelAllGlobalTasks()
		w.WriteHeader(http.StatusAccepted)
	case "task":
		if req.TaskID == "" {
			http.Error(w, "task_id is required", http.StatusBadRequest)
			return
		}
		if !clicky.StopTask(req.TaskID) {
			http.Error(w, "task not found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	default:
		http.Error(w, "unsupported stop scope", http.StatusBadRequest)
	}
}
