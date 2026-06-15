package procfile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/flanksource/gavel/utils"
)

// Process status values recorded in state.json.
const (
	StatusStarting   = "starting"
	StatusRunning    = "running"
	StatusStopped    = "stopped"
	StatusCrashed    = "crashed"
	StatusExited     = "exited"
	StatusRestarting = "restarting"
)

const (
	stateFileName      = "state.json"
	supervisorPidName  = "supervisor.pid"
	stateDirComponents = ".gavel" + string(filepath.Separator) + "proc"
)

// ProcState is the live state of a single supervised process.
type ProcState struct {
	Name     string     `json:"name"`
	Command  string     `json:"command"`
	PID      int        `json:"pid,omitempty"`
	Status   string     `json:"status"`
	Started  *time.Time `json:"started,omitempty"`
	Restarts int        `json:"restarts"`
	ExitCode *int       `json:"exitCode,omitempty"`
	LogFile  string     `json:"logFile"`
	// Ports are the TCP ports the process (and its process group) is listening
	// on, detected after start. Empty for processes that bind no port.
	Ports []int `json:"ports,omitempty"`
	// CPUPercent, MemoryRSS and OpenFiles are the latest resource sample of the
	// process group, taken live by the supervisor. OpenFiles is -1 on platforms
	// that cannot report it. All zero for a stopped process.
	CPUPercent float64 `json:"cpuPercent,omitempty"`
	MemoryRSS  uint64  `json:"memoryRss,omitempty"`
	OpenFiles  int     `json:"openFiles,omitempty"`
	// Tree is the per-process breakdown of the process group behind the
	// aggregate fields above (the leader and its descendants). Empty for a
	// stopped process or a supervisor that predates resource sampling.
	Tree []ProcNode `json:"tree,omitempty"`
}

// ProcNode is one process within a supervised process group's tree, with its
// own resource sample. OpenFiles is -1 where the platform cannot report it.
type ProcNode struct {
	PID        int     `json:"pid"`
	PPID       int     `json:"ppid"`
	Command    string  `json:"command"`
	CPUPercent float64 `json:"cpuPercent,omitempty"`
	MemoryRSS  uint64  `json:"memoryRss,omitempty"`
	OpenFiles  int     `json:"openFiles,omitempty"`
}

// State is the supervisor-owned snapshot persisted to .gavel/proc/state.json.
// The supervisor is the sole writer; CLI commands only read it.
type State struct {
	Root          string      `json:"root"`
	Procfile      string      `json:"procfile"`
	SupervisorPID int         `json:"supervisorPid"`
	Socket        string      `json:"socket,omitempty"`
	Profile       string      `json:"profile,omitempty"`
	Started       *time.Time  `json:"started,omitempty"`
	Processes     []ProcState `json:"processes"`
}

// Running reports whether the supervisor process is currently alive.
func (s State) Running() bool {
	return utils.ProcessAlive(s.SupervisorPID)
}

// Process returns the recorded state for name, if present.
func (s State) Process(name string) (ProcState, bool) {
	for _, p := range s.Processes {
		if p.Name == name {
			return p, true
		}
	}
	return ProcState{}, false
}

// StateDir returns <root>/.gavel/proc, creating it if necessary.
func StateDir(root string) (string, error) {
	dir := filepath.Join(root, stateDirComponents)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create state dir %s: %w", dir, err)
	}
	return dir, nil
}

// StatePath is the path of the JSON state file inside dir.
func StatePath(dir string) string { return filepath.Join(dir, stateFileName) }

// SupervisorPidPath is the path of the supervisor pidfile inside dir.
func SupervisorPidPath(dir string) string { return filepath.Join(dir, supervisorPidName) }

// PidPath is the per-process pidfile path inside dir.
func PidPath(dir, name string) string { return filepath.Join(dir, name+".pid") }

// LogPath is the per-process log file path inside dir.
func LogPath(dir, name string) string { return filepath.Join(dir, name+".log") }

// SupervisorLogPath is where the detached supervisor's own stdout/stderr lands.
func SupervisorLogPath(dir string) string { return filepath.Join(dir, "supervisor.log") }

// WriteState atomically persists s to dir/state.json.
func WriteState(dir string, s State) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal state: %w", err)
	}
	return writeFileAtomic(StatePath(dir), append(data, '\n'))
}

// ReadState reads dir/state.json. A missing file is reported as a zero State
// (SupervisorPID 0, Running() false) with no error so status/list work before
// the first start. A present-but-unreadable file is a loud error.
func ReadState(dir string) (State, error) {
	data, err := os.ReadFile(StatePath(dir))
	if err != nil {
		if os.IsNotExist(err) {
			return State{}, nil
		}
		return State{}, fmt.Errorf("read state %s: %w", StatePath(dir), err)
	}
	var s State
	if err := json.Unmarshal(data, &s); err != nil {
		return State{}, fmt.Errorf("parse state %s: %w", StatePath(dir), err)
	}
	return s, nil
}

// WritePid writes pid to path (with a trailing newline).
func WritePid(path string, pid int) error {
	return os.WriteFile(path, []byte(strconv.Itoa(pid)+"\n"), 0o644)
}

// ReadPid reads a pid from path. A missing file returns os.ErrNotExist.
func ReadPid(path string) (int, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	s := strings.TrimSpace(string(b))
	pid, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("pidfile %s has invalid pid %q: %w", path, s, err)
	}
	return pid, nil
}

// Clean removes the state file, the supervisor pidfile, and every per-process
// pidfile in dir. Log files are intentionally kept so `proc logs` works after a
// stop. Missing files are ignored.
func Clean(dir string) error {
	_ = os.Remove(StatePath(dir))
	_ = os.Remove(SupervisorPidPath(dir))
	matches, err := filepath.Glob(filepath.Join(dir, "*.pid"))
	if err != nil {
		return fmt.Errorf("glob pidfiles in %s: %w", dir, err)
	}
	for _, m := range matches {
		_ = os.Remove(m)
	}
	return nil
}

func writeFileAtomic(path string, data []byte) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", tmp, err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("rename %s -> %s: %w", tmp, path, err)
	}
	return nil
}
