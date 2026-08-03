package procfile

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/flanksource/clicky/metrics"
	clickytask "github.com/flanksource/clicky/task"
)

// Control actions exchanged over the supervisor's unix control socket.
const (
	actionStatus       = "status"
	actionStart        = "start"
	actionStop         = "stop"
	actionRestart      = "restart"
	actionTaskRuns     = "task-runs"
	actionTaskGet      = "task-get"
	actionTaskCtrl     = "task-control"
	actionTaskItemCtrl = "task-item-control"
	actionTaskMetric   = "task-metric"
)

type ctrlRequest struct {
	Action      string                   `json:"action"`
	Names       []string                 `json:"names,omitempty"`
	TaskID      string                   `json:"taskId,omitempty"`
	ChildTaskID string                   `json:"childTaskId,omitempty"`
	TaskAction  clickytask.ControlAction `json:"taskAction,omitempty"`
	TaskFilter  clickytask.RunFilter     `json:"taskFilter,omitempty"`
	Metric      metrics.QueryRequest     `json:"metric,omitempty"`
}

type ctrlResponse struct {
	OK        bool                      `json:"ok"`
	Error     string                    `json:"error,omitempty"`
	State     State                     `json:"state"`
	TaskRuns  []clickytask.RunMeta      `json:"taskRuns,omitempty"`
	Snapshots []clickytask.TaskSnapshot `json:"snapshots,omitempty"`
	Points    []metrics.Point           `json:"points,omitempty"`
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
		InstanceID:    s.instanceID,
		Socket:        s.socket,
		Profile:       s.profile,
		Started:       s.started,
	}
	for _, m := range s.procs {
		st.Processes = append(st.Processes, m.snapshot())
	}
	return st
}

func (s *Supervisor) serveControl() error {
	sock := ControlSocketPath(s.root)
	if _, err := os.Stat(sock); err == nil {
		if conn, dialErr := net.DialTimeout("unix", sock, 500*time.Millisecond); dialErr == nil {
			_ = conn.Close()
			return fmt.Errorf("gavel proc is already running for %s", s.root)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("inspect control socket %s: %w", sock, err)
	}
	if err := os.Remove(sock); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove stale control socket %s: %w", sock, err)
	}
	l, err := net.Listen("unix", sock)
	if err != nil {
		return fmt.Errorf("listen on control socket %s: %w", sock, err)
	}
	unixListener, ok := l.(*net.UnixListener)
	if !ok {
		_ = l.Close()
		return fmt.Errorf("control socket %s is not a unix listener", sock)
	}
	unixListener.SetUnlinkOnClose(false)
	s.socket = sock
	s.mu.Lock()
	s.listener = l
	s.mu.Unlock()
	go s.acceptLoop(l)
	return nil
}

func controlSocketOwnedBy(path, instanceID string) bool {
	if path == "" || instanceID == "" {
		return false
	}
	resp, err := sendControlAt(path, ctrlRequest{Action: actionStatus}, 500*time.Millisecond)
	return err == nil && resp.State.InstanceID == instanceID
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
	case actionTaskRuns:
		return ctrlResponse{OK: true, TaskRuns: clickytask.Runs(req.TaskFilter)}
	case actionTaskGet:
		return ctrlResponse{OK: true, Snapshots: clickytask.SnapshotByID(req.TaskID)}
	case actionTaskCtrl:
		if err := clickytask.ControlRun(context.Background(), req.TaskID, req.TaskAction); err != nil {
			return ctrlResponse{Error: err.Error()}
		}
		return ctrlResponse{OK: true}
	case actionTaskItemCtrl:
		if err := clickytask.ControlTask(context.Background(), req.TaskID, req.ChildTaskID, req.TaskAction); err != nil {
			return ctrlResponse{Error: err.Error()}
		}
		return ctrlResponse{OK: true}
	case actionTaskMetric:
		points, err := clickytask.Metrics().Query(req.Metric)
		if err != nil {
			return ctrlResponse{Error: err.Error()}
		}
		return ctrlResponse{OK: true, Points: points}
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
	return sendControlAt(ControlSocketPath(root), req, 10*time.Second)
}

func sendControlAt(path string, req ctrlRequest, timeout time.Duration) (ctrlResponse, error) {
	conn, err := net.DialTimeout("unix", path, timeout)
	if err != nil {
		return ctrlResponse{}, fmt.Errorf("connect to supervisor: %w", err)
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(timeout))
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

// TaskRuns returns the task runs owned by the detached supervisor for root.
func TaskRuns(root string, filter clickytask.RunFilter) ([]clickytask.RunMeta, error) {
	response, err := sendControl(root, ctrlRequest{Action: actionTaskRuns, TaskFilter: filter})
	return response.TaskRuns, err
}

// TaskSnapshot returns one detached supervisor task generation.
func TaskSnapshot(root, id string) ([]clickytask.TaskSnapshot, error) {
	response, err := sendControl(root, ctrlRequest{Action: actionTaskGet, TaskID: id})
	return response.Snapshots, err
}

// TaskControl performs a lifecycle action on a detached supervisor task.
func TaskControl(root, id string, action clickytask.ControlAction) error {
	_, err := sendControl(root, ctrlRequest{Action: actionTaskCtrl, TaskID: id, TaskAction: action})
	return err
}

// TaskControlTask performs a lifecycle action on one child task owned by a detached supervisor.
func TaskControlTask(root, runID, taskID string, action clickytask.ControlAction) error {
	_, err := sendControl(root, ctrlRequest{
		Action: actionTaskItemCtrl, TaskID: runID, ChildTaskID: taskID, TaskAction: action,
	})
	return err
}

// TaskMetrics queries one live metric series from the detached supervisor.
func TaskMetrics(root string, request metrics.QueryRequest) ([]metrics.Point, error) {
	response, err := sendControl(root, ctrlRequest{Action: actionTaskMetric, Metric: request})
	return response.Points, err
}
