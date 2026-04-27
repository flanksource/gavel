package testui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/flanksource/gavel/testrunner/parsers"
)

// gavelDir is the conventional snapshot directory written by snapshots.Save.
// Duplicated here (rather than imported from the snapshots package) because
// snapshots already imports this package.
const gavelDir = ".gavel"

// pointerJSON mirrors snapshots.Pointer's wire shape. Re-declared locally to
// avoid importing the snapshots package, which would create an import cycle.
type pointerJSON struct {
	Path        string `json:"path"`
	SHA         string `json:"sha"`
	Uncommitted string `json:"uncommitted,omitempty"`
}

type RunCounts struct {
	Total   int `json:"total"`
	Passed  int `json:"passed"`
	Failed  int `json:"failed"`
	Skipped int `json:"skipped"`
	Pending int `json:"pending"`
}

type RunIndexEntry struct {
	Name     string     `json:"name"`
	Path     string     `json:"path"`
	Pointer  string     `json:"pointer,omitempty"`
	Modified time.Time  `json:"modified"`
	SHA      string     `json:"sha,omitempty"`
	Started  *time.Time `json:"started,omitempty"`
	Ended    *time.Time `json:"ended,omitempty"`
	Counts   *RunCounts `json:"counts,omitempty"`
	Lint     int        `json:"lint,omitempty"`
	Error    string     `json:"error,omitempty"`
}

// SetGavelDir tells the server which directory holds previously-saved snapshots.
// When set, /api/runs scans this directory and /api/runs/{name} loads a single
// snapshot from it. Empty string disables both endpoints.
func (s *Server) SetGavelDir(root string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if root == "" {
		s.gavelDir = ""
		return
	}
	s.gavelDir = filepath.Join(root, gavelDir)
}

// GavelDir returns the directory backing /api/runs, or "" if no-arg index mode
// is disabled.
func (s *Server) GavelDir() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.gavelDir
}

func (s *Server) handleRunsIndex(w http.ResponseWriter, _ *http.Request) {
	dir := s.GavelDir()
	if dir == "" {
		http.NotFound(w, nil)
		return
	}
	entries, err := buildRunIndex(dir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entries) //nolint:errcheck
}

func (s *Server) handleRunSnapshot(w http.ResponseWriter, r *http.Request) {
	dir := s.GavelDir()
	if dir == "" {
		http.NotFound(w, r)
		return
	}
	name := strings.TrimPrefix(r.URL.Path, "/api/runs/")
	if name == "" || strings.ContainsRune(name, '/') || strings.Contains(name, "..") {
		http.Error(w, "invalid run name", http.StatusBadRequest)
		return
	}
	if !strings.HasSuffix(name, ".json") {
		name += ".json"
	}
	path := filepath.Join(dir, name)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// If this is a pointer, dereference once.
	if resolved, ok := tryResolvePointer(dir, data); ok {
		data = resolved
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data) //nolint:errcheck
}

// tryResolvePointer detects pointer files (the small {path,sha} shape written
// by snapshots.Save) and returns the bytes of the snapshot they reference.
// Returns (nil, false) when raw is not a pointer or the pointed file cannot be
// read — caller falls back to the original bytes.
func tryResolvePointer(dir string, raw []byte) ([]byte, bool) {
	var p pointerJSON
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, false
	}
	if p.Path == "" || p.SHA == "" {
		return nil, false
	}
	target := p.Path
	if !filepath.IsAbs(target) {
		target = filepath.Join(filepath.Dir(dir), target)
	}
	data, err := os.ReadFile(target)
	if err != nil {
		return nil, false
	}
	return data, true
}

// buildRunIndex scans dir, returns one entry per .json file with summary
// metadata. Pointers are resolved (stats come from the underlying snapshot)
// and rendered as labeled rows pointing to the resolved snapshot.
func buildRunIndex(dir string) ([]RunIndexEntry, error) {
	infos, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", dir, err)
	}

	pointers := make([]RunIndexEntry, 0)
	snapshots := make([]RunIndexEntry, 0)

	for _, fi := range infos {
		if fi.IsDir() {
			continue
		}
		name := fi.Name()
		if !strings.HasSuffix(name, ".json") {
			continue
		}
		full := filepath.Join(dir, name)
		stat, err := fi.Info()
		if err != nil {
			continue
		}
		entry := RunIndexEntry{
			Name:     strings.TrimSuffix(name, ".json"),
			Path:     filepath.Join(gavelDir, name),
			Modified: stat.ModTime(),
		}
		raw, err := os.ReadFile(full)
		if err != nil {
			entry.Error = err.Error()
			snapshots = append(snapshots, entry)
			continue
		}

		if resolved, ok := tryResolvePointer(dir, raw); ok {
			var ptr pointerJSON
			_ = json.Unmarshal(raw, &ptr)
			entry.Pointer = entry.Name
			entry.SHA = shortSHA(ptr.SHA)
			// Path points at the resolved snapshot so click-through deep links
			// to the same URL as a direct snapshot row.
			entry.Path = ptr.Path
			fillFromSnapshot(&entry, resolved)
			pointers = append(pointers, entry)
			continue
		}

		fillFromSnapshot(&entry, raw)
		snapshots = append(snapshots, entry)
	}

	sort.SliceStable(pointers, func(i, j int) bool {
		return pointerOrder(pointers[i].Pointer) < pointerOrder(pointers[j].Pointer) ||
			(pointerOrder(pointers[i].Pointer) == pointerOrder(pointers[j].Pointer) &&
				pointers[i].Pointer < pointers[j].Pointer)
	})
	sort.SliceStable(snapshots, func(i, j int) bool {
		return snapshots[i].Modified.After(snapshots[j].Modified)
	})

	out := make([]RunIndexEntry, 0, len(pointers)+len(snapshots))
	out = append(out, pointers...)
	out = append(out, snapshots...)
	return out, nil
}

func pointerOrder(name string) int {
	switch name {
	case "last":
		return 0
	case "main", "master":
		return 2
	default:
		return 1
	}
}

func fillFromSnapshot(entry *RunIndexEntry, raw []byte) {
	var snap Snapshot
	if err := json.Unmarshal(raw, &snap); err != nil {
		entry.Error = err.Error()
		return
	}
	if snap.Git != nil && entry.SHA == "" {
		entry.SHA = shortSHA(snap.Git.SHA)
	}
	if snap.Metadata != nil {
		if !snap.Metadata.Started.IsZero() {
			t := snap.Metadata.Started
			entry.Started = &t
		}
		if !snap.Metadata.Ended.IsZero() {
			t := snap.Metadata.Ended
			entry.Ended = &t
		}
	}
	sum := parsers.Tests(snap.Tests).Sum()
	if sum.Total > 0 || sum.Failed > 0 || sum.Skipped > 0 || sum.Pending > 0 {
		entry.Counts = &RunCounts{
			Total:   sum.Total,
			Passed:  sum.Passed,
			Failed:  sum.Failed,
			Skipped: sum.Skipped,
			Pending: sum.Pending,
		}
	}
	for _, lr := range snap.Lint {
		if lr == nil {
			continue
		}
		entry.Lint += len(lr.Violations)
	}
}

func shortSHA(sha string) string {
	if len(sha) > 7 {
		return sha[:7]
	}
	return sha
}
