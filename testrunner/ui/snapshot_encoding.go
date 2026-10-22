package testui

import (
	"encoding/json"
	"sync"
)

type snapshotEncodingCache struct {
	mu       sync.Mutex
	revision uint64
	tasks    string
	payload  string
	running  bool
}

func (s *Server) encodedSnapshot() (string, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	tasks, tasksRunning := s.liveTaskTestsLocked()
	key, err := json.Marshal(tasks)
	if err != nil {
		return "", false, err
	}
	s.encoding.mu.Lock()
	defer s.encoding.mu.Unlock()
	revision := s.revision.Load()
	if s.encoding.payload != "" && s.encoding.revision == revision && s.encoding.tasks == string(key) {
		return s.encoding.payload, s.encoding.running, nil
	}
	snapshot := s.snapshotWithTasks(tasks, tasksRunning)
	payload, err := json.Marshal(snapshot)
	if err != nil {
		return "", false, err
	}
	s.encoding.revision = revision
	s.encoding.tasks = string(key)
	s.encoding.payload = string(payload)
	s.encoding.running = snapshot.Status.Running
	return s.encoding.payload, s.encoding.running, nil
}
