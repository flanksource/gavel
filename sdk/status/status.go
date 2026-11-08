// Package status is the contract a running gavel process publishes about
// itself: the test, lint and fixture progress its live UI shows, and the
// machine-wide slots it holds or waits for. Each process writes
// <cache>/status/<pid>.json on change and, once it has run for a while, also
// serves the same Status over <cache>/status/<pid>.sock.
package status

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"
)

// Status is one gavel process at one moment.
type Status struct {
	PID     int       `json:"pid"`
	Command string    `json:"command"`
	WorkDir string    `json:"workDir"`
	Started time.Time `json:"started"`
	Updated time.Time `json:"updated"`
	// Socket is the unix socket serving /status and /status/stream, once open.
	Socket string `json:"socket,omitempty"`
	// Tests counts the tests the run has reported so far.
	Tests *Summary `json:"tests,omitempty"`
	// Groups are the live task groups: test packages, linters, prompt lint, fixtures.
	Groups []Group `json:"groups"`
	// Running are the tasks running now, oldest first.
	Running []Item  `json:"running"`
	Slots   []Lease `json:"slots"`
}

// Summary counts outcomes; TimedOut is a subset of Failed.
type Summary struct {
	Total    int `json:"total"`
	Passed   int `json:"passed"`
	Failed   int `json:"failed"`
	Warned   int `json:"warned,omitempty"`
	Aborted  int `json:"aborted,omitempty"`
	Skipped  int `json:"skipped,omitempty"`
	Pending  int `json:"pending,omitempty"`
	Running  int `json:"running,omitempty"`
	TimedOut int `json:"timedOut,omitempty"`
}

// Done is how many have finished, whatever their outcome.
func (s Summary) Done() int { return s.Total - s.Pending - s.Running }

// Group kinds name what a task group runs, for a short label.
const (
	GroupPackages = "packages"
	GroupLint     = "lint"
	GroupPrompt   = "prompt"
	GroupFixtures = "fixtures"
)

type Group struct {
	Name string `json:"name"`
	// Kind is one of the Group kinds; empty for a group nested under one.
	Kind     string    `json:"kind,omitempty"`
	Summary  Summary   `json:"summary"`
	Started  time.Time `json:"started,omitzero"`
	Finished time.Time `json:"finished,omitzero"`
}

type Item struct {
	Name        string    `json:"name"`
	Group       string    `json:"group"`
	Description string    `json:"description,omitempty"`
	Started     time.Time `json:"started,omitzero"`
	// Done of Total is the task's own bounded progress; Total 0 means unbounded.
	Done  int `json:"done,omitempty"`
	Total int `json:"total,omitempty"`
	// PID is the subprocess the task runs, when it runs one.
	PID int `json:"pid,omitempty"`
}

const (
	LeaseHeld    = "held"
	LeaseWaiting = "waiting"
)

// Lease is a slot the process holds, or a pool it is queued on.
type Lease struct {
	// Pool is "global" or "project/<id>".
	Pool string `json:"pool"`
	// Project is the canonical git common directory of a project pool.
	Project string `json:"project,omitempty"`
	Limit   int    `json:"limit"`
	// Index is the slot held, -1 while waiting.
	Index   int       `json:"index"`
	State   string    `json:"state"`
	Since   time.Time `json:"since"`
	Command string    `json:"command"`
}

// CacheDir is gavel's user cache directory: $XDG_CACHE_HOME/gavel when set, on
// every platform, else the OS user cache directory's gavel.
func CacheDir() (string, error) {
	if xdg := os.Getenv("XDG_CACHE_HOME"); xdg != "" {
		if !filepath.IsAbs(xdg) {
			return "", fmt.Errorf("XDG_CACHE_HOME must be an absolute path, got %q", xdg)
		}
		return filepath.Join(xdg, "gavel"), nil
	}
	cache, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("resolve user cache directory: %w", err)
	}
	return filepath.Join(cache, "gavel"), nil
}

func Dir(cache string) string { return filepath.Join(cache, "status") }

func Path(cache string, pid int) string {
	return filepath.Join(Dir(cache), strconv.Itoa(pid)+".json")
}

func SocketPath(cache string, pid int) string {
	return filepath.Join(Dir(cache), strconv.Itoa(pid)+".sock")
}

func Read(path string) (Status, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Status{}, err
	}
	return decode(path, data)
}

// List reads the status of every live gavel process, by pid. A file whose
// process has exited, or that vanished while being read, is skipped.
func List(cache string) ([]Status, error) {
	entries, err := os.ReadDir(Dir(cache))
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var statuses []Status
	for _, entry := range entries {
		pid, ok := pidOf(entry.Name(), ".json")
		if !ok || !Alive(pid) {
			continue
		}
		status, err := Read(filepath.Join(Dir(cache), entry.Name()))
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, err
		}
		statuses = append(statuses, status)
	}
	slices.SortFunc(statuses, func(a, b Status) int { return a.PID - b.PID })
	return statuses, nil
}

func pidOf(name, suffix string) (int, bool) {
	stem, ok := strings.CutSuffix(name, suffix)
	if !ok {
		return 0, false
	}
	pid, err := strconv.Atoi(stem)
	return pid, err == nil && pid > 0
}

// Stale lists the status files and sockets in cache whose process has exited.
func Stale(cache string) ([]string, error) {
	entries, err := os.ReadDir(Dir(cache))
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var stale []string
	for _, entry := range entries {
		// Temporary files (<pid>.json.tmp-*) of a process that crashed mid-write count too.
		stem, _, _ := strings.Cut(entry.Name(), ".")
		if pid, err := strconv.Atoi(stem); err == nil && pid > 0 && !Alive(pid) {
			stale = append(stale, filepath.Join(Dir(cache), entry.Name()))
		}
	}
	return stale, nil
}
