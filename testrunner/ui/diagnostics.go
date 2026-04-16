package testui

import (
	"bytes"
	"fmt"
	"runtime/pprof"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/process"
)

const diagnosticsRefreshInterval = time.Second

type DiagnosticsSnapshot struct {
	Root        *ProcessNode `json:"root,omitempty"`
	GeneratedAt time.Time    `json:"generated_at,omitempty"`
}

type ProcessNode struct {
	PID          int            `json:"pid"`
	PPID         int            `json:"ppid,omitempty"`
	Name         string         `json:"name,omitempty"`
	Command      string         `json:"command,omitempty"`
	Status       string         `json:"status,omitempty"`
	CPUPercent   float64        `json:"cpu_percent,omitempty"`
	RSS          uint64         `json:"rss,omitempty"`
	VMS          uint64         `json:"vms,omitempty"`
	OpenFiles    *int32         `json:"open_files,omitempty"`
	IsRoot       bool           `json:"is_root,omitempty"`
	Children     []*ProcessNode `json:"children,omitempty"`
	StackCapture *StackCapture  `json:"stack_capture,omitempty"`
}

type ProcessDetails struct {
	PID          int           `json:"pid"`
	PPID         int           `json:"ppid,omitempty"`
	Name         string        `json:"name,omitempty"`
	Command      string        `json:"command,omitempty"`
	Status       string        `json:"status,omitempty"`
	CPUPercent   float64       `json:"cpu_percent,omitempty"`
	RSS          uint64        `json:"rss,omitempty"`
	VMS          uint64        `json:"vms,omitempty"`
	OpenFiles    *int32        `json:"open_files,omitempty"`
	IsRoot       bool          `json:"is_root,omitempty"`
	StackCapture *StackCapture `json:"stack_capture,omitempty"`
}

type StackCapture struct {
	Status      string    `json:"status"`
	Supported   bool      `json:"supported"`
	Text        string    `json:"text,omitempty"`
	Error       string    `json:"error,omitempty"`
	CollectedAt time.Time `json:"collected_at,omitempty"`
}

type StackCaptureRequest struct {
	PID int `json:"pid"`
}

type DiagnosticsCollector interface {
	Snapshot(rootPID int) (*DiagnosticsSnapshot, error)
	CaptureStack(rootPID, pid int) StackCapture
}

type DiagnosticsManager struct {
	mu         sync.Mutex
	rootPID    int
	collector  DiagnosticsCollector
	lastAt     time.Time
	cached     *DiagnosticsSnapshot
	stackByPID map[int]StackCapture
}

func NewDiagnosticsManager(rootPID int, collector DiagnosticsCollector) *DiagnosticsManager {
	if collector == nil {
		collector = systemDiagnosticsCollector{}
	}
	return &DiagnosticsManager{
		rootPID:    rootPID,
		collector:  collector,
		stackByPID: make(map[int]StackCapture),
	}
}

func (m *DiagnosticsManager) RootPID() int {
	if m == nil {
		return 0
	}
	return m.rootPID
}

func (m *DiagnosticsManager) Snapshot() (*DiagnosticsSnapshot, error) {
	if m == nil {
		return nil, fmt.Errorf("diagnostics unavailable")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.cached == nil || time.Since(m.lastAt) >= diagnosticsRefreshInterval {
		if err := m.refreshLocked(); err != nil {
			return nil, err
		}
	}
	return cloneDiagnosticsSnapshot(m.cached), nil
}

func (m *DiagnosticsManager) CollectStack(pid int) (*ProcessDetails, error) {
	if m == nil {
		return nil, fmt.Errorf("diagnostics unavailable")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.cached == nil || time.Since(m.lastAt) >= diagnosticsRefreshInterval {
		if err := m.refreshLocked(); err != nil {
			return nil, err
		}
	}

	node := findProcessNode(m.cached.Root, pid)
	if node == nil {
		return nil, fmt.Errorf("process %d not found", pid)
	}

	capture := m.collector.CaptureStack(m.rootPID, pid)
	m.stackByPID[pid] = capture
	if err := m.refreshLocked(); err != nil {
		return nil, err
	}

	node = findProcessNode(m.cached.Root, pid)
	if node == nil {
		return nil, fmt.Errorf("process %d not found", pid)
	}
	return processDetailsFromNode(node), nil
}

func (m *DiagnosticsManager) refreshLocked() error {
	snapshot, err := m.collector.Snapshot(m.rootPID)
	if err != nil {
		return err
	}
	if snapshot == nil {
		return fmt.Errorf("diagnostics collector returned nil snapshot")
	}
	snapshot.GeneratedAt = time.Now().UTC()
	applyStackCaptures(snapshot.Root, m.stackByPID)
	m.cached = snapshot
	m.lastAt = time.Now()
	return nil
}

func applyStackCaptures(node *ProcessNode, stackByPID map[int]StackCapture) {
	if node == nil {
		return
	}
	if capture, ok := stackByPID[node.PID]; ok {
		copied := capture
		node.StackCapture = &copied
	}
	for _, child := range node.Children {
		applyStackCaptures(child, stackByPID)
	}
}

func cloneDiagnosticsSnapshot(snapshot *DiagnosticsSnapshot) *DiagnosticsSnapshot {
	if snapshot == nil {
		return nil
	}
	return &DiagnosticsSnapshot{
		GeneratedAt: snapshot.GeneratedAt,
		Root:        cloneProcessNode(snapshot.Root),
	}
}

func cloneProcessNode(node *ProcessNode) *ProcessNode {
	if node == nil {
		return nil
	}
	cloned := *node
	if node.StackCapture != nil {
		capture := *node.StackCapture
		cloned.StackCapture = &capture
	}
	if len(node.Children) > 0 {
		cloned.Children = make([]*ProcessNode, 0, len(node.Children))
		for _, child := range node.Children {
			cloned.Children = append(cloned.Children, cloneProcessNode(child))
		}
	}
	return &cloned
}

func findProcessNode(node *ProcessNode, pid int) *ProcessNode {
	if node == nil {
		return nil
	}
	if node.PID == pid {
		return node
	}
	for _, child := range node.Children {
		if found := findProcessNode(child, pid); found != nil {
			return found
		}
	}
	return nil
}

func processDetailsFromNode(node *ProcessNode) *ProcessDetails {
	if node == nil {
		return nil
	}
	details := &ProcessDetails{
		PID:        node.PID,
		PPID:       node.PPID,
		Name:       node.Name,
		Command:    node.Command,
		Status:     node.Status,
		CPUPercent: node.CPUPercent,
		RSS:        node.RSS,
		VMS:        node.VMS,
		OpenFiles:  node.OpenFiles,
		IsRoot:     node.IsRoot,
	}
	if node.StackCapture != nil {
		capture := *node.StackCapture
		details.StackCapture = &capture
	}
	return details
}

type systemDiagnosticsCollector struct{}

func (systemDiagnosticsCollector) Snapshot(rootPID int) (*DiagnosticsSnapshot, error) {
	proc, err := process.NewProcess(int32(rootPID))
	if err != nil {
		return nil, err
	}
	root, err := snapshotProcess(proc, true)
	if err != nil {
		return nil, err
	}
	return &DiagnosticsSnapshot{Root: root}, nil
}

func (systemDiagnosticsCollector) CaptureStack(rootPID, pid int) StackCapture {
	if pid != rootPID {
		return StackCapture{
			Status:    "unsupported",
			Supported: false,
			Error:     "stack capture is only supported for the live gavel process",
		}
	}

	var buf bytes.Buffer
	if err := pprof.Lookup("goroutine").WriteTo(&buf, 2); err != nil {
		return StackCapture{
			Status:    "error",
			Supported: true,
			Error:     err.Error(),
		}
	}

	return StackCapture{
		Status:      "ready",
		Supported:   true,
		Text:        buf.String(),
		CollectedAt: time.Now().UTC(),
	}
}

func snapshotProcess(proc *process.Process, isRoot bool) (*ProcessNode, error) {
	pid := int(proc.Pid)
	ppid, _ := proc.Ppid()
	name, _ := proc.Name()
	cmdline, _ := proc.Cmdline()
	if cmdline == "" {
		exe, _ := proc.Exe()
		cmdline = exe
	}
	statusList, _ := proc.Status()
	cpuPercent, _ := proc.CPUPercent()
	memInfo, _ := proc.MemoryInfo()
	openFiles, err := proc.NumFDs()
	if err != nil {
		openFiles = 0
	}

	node := &ProcessNode{
		PID:        pid,
		PPID:       int(ppid),
		Name:       firstNonEmpty(name, shortProcessName(cmdline)),
		Command:    cmdline,
		Status:     strings.Join(statusList, ", "),
		CPUPercent: cpuPercent,
		IsRoot:     isRoot,
	}
	if memInfo != nil {
		node.RSS = memInfo.RSS
		node.VMS = memInfo.VMS
	}
	if openFiles > 0 {
		value := openFiles
		node.OpenFiles = &value
	}

	children, _ := proc.Children()
	sort.Slice(children, func(i, j int) bool {
		return children[i].Pid < children[j].Pid
	})
	for _, child := range children {
		childNode, err := snapshotProcess(child, false)
		if err != nil {
			continue
		}
		node.Children = append(node.Children, childNode)
	}

	return node, nil
}

func shortProcessName(cmdline string) string {
	if cmdline == "" {
		return ""
	}
	fields := strings.Fields(cmdline)
	if len(fields) == 0 {
		return ""
	}
	return fields[0]
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
