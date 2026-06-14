package procfile

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"
)

// Control actions exchanged over the supervisor's unix control socket.
const (
	actionStatus  = "status"
	actionStart   = "start"
	actionStop    = "stop"
	actionRestart = "restart"
)

type ctrlRequest struct {
	Action string   `json:"action"`
	Names  []string `json:"names,omitempty"`
}

type ctrlResponse struct {
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
	State State  `json:"state"`
}

// ControlSocketPath returns the unix socket path for the supervisor rooted at
// root. It lives under the temp dir — not .gavel/proc — keyed by a hash of root
// so it stays within the ~104-byte sun_path limit no matter how deep root is.
func ControlSocketPath(root string) string {
	sum := sha1.Sum([]byte(root))
	return filepath.Join(os.TempDir(), "gavel-proc-"+hex.EncodeToString(sum[:6])+".sock")
}

// State builds a live snapshot, the same shape persisted to state.json.
func (s *Supervisor) State() State {
	st := State{
		Root:          s.root,
		Procfile:      s.procfile,
		SupervisorPID: os.Getpid(),
		Socket:        s.socket,
		Started:       s.started,
	}
	for _, m := range s.procs {
		st.Processes = append(st.Processes, m.snapshot())
	}
	return st
}

func (s *Supervisor) serveControl() error {
	sock := ControlSocketPath(s.root)
	_ = os.Remove(sock)
	l, err := net.Listen("unix", sock)
	if err != nil {
		return fmt.Errorf("listen on control socket %s: %w", sock, err)
	}
	s.socket = sock
	s.mu.Lock()
	s.listener = l
	s.mu.Unlock()
	go s.acceptLoop(l)
	return nil
}

func (s *Supervisor) acceptLoop(l net.Listener) {
	for {
		conn, err := l.Accept()
		if err != nil {
			return // listener closed during teardown
		}
		go s.handleConn(conn)
	}
}

func (s *Supervisor) handleConn(conn net.Conn) {
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(10 * time.Second))
	var req ctrlRequest
	if err := json.NewDecoder(conn).Decode(&req); err != nil {
		_ = json.NewEncoder(conn).Encode(ctrlResponse{Error: fmt.Sprintf("decode request: %v", err)})
		return
	}
	_ = json.NewEncoder(conn).Encode(s.handle(req))
}

func (s *Supervisor) handle(req ctrlRequest) ctrlResponse {
	switch req.Action {
	case actionStatus:
		return ctrlResponse{OK: true, State: s.State()}
	case actionStart, actionStop, actionRestart:
		targets, err := s.targets(req.Names)
		if err != nil {
			return ctrlResponse{Error: err.Error()}
		}
		for _, m := range targets {
			switch req.Action {
			case actionStart:
				s.startProc(m)
			case actionStop:
				s.stopProc(m)
			case actionRestart:
				s.restartProc(m)
			}
		}
		s.persist()
		return ctrlResponse{OK: true, State: s.State()}
	default:
		return ctrlResponse{Error: fmt.Sprintf("unknown action %q", req.Action)}
	}
}

func (s *Supervisor) targets(names []string) ([]*managed, error) {
	if len(names) == 0 {
		return s.procs, nil
	}
	out := make([]*managed, 0, len(names))
	for _, n := range names {
		m, ok := s.byName[n]
		if !ok {
			return nil, fmt.Errorf("unknown process %q", n)
		}
		out = append(out, m)
	}
	return out, nil
}

// sendControl dials the supervisor's control socket for root and performs req.
// A failure to connect means no supervisor is running for that root.
func sendControl(root string, req ctrlRequest) (ctrlResponse, error) {
	conn, err := net.DialTimeout("unix", ControlSocketPath(root), 2*time.Second)
	if err != nil {
		return ctrlResponse{}, fmt.Errorf("connect to supervisor: %w", err)
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(10 * time.Second))
	if err := json.NewEncoder(conn).Encode(req); err != nil {
		return ctrlResponse{}, fmt.Errorf("send request: %w", err)
	}
	var resp ctrlResponse
	if err := json.NewDecoder(conn).Decode(&resp); err != nil {
		return ctrlResponse{}, fmt.Errorf("read response: %w", err)
	}
	if !resp.OK && resp.Error != "" {
		return resp, fmt.Errorf("%s", resp.Error)
	}
	return resp, nil
}
