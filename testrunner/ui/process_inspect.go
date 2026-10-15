package testui

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/flanksource/clicky/process"
)

const processInspectTimeout = 10 * time.Second

// handleProcessInspect serves /api/process/{env,files,sockets,ports}?pid=N,
// read live when the Processes tab opens the facet. Only processes in this
// server's own tree are inspected, and their environment, which carries the
// run's credentials, only to a client on this machine.
func (s *Server) handleProcessInspect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "GET required", http.StatusMethodNotAllowed)
		return
	}
	facet := strings.TrimPrefix(r.URL.Path, "/api/process/")
	if !slices.Contains([]string{"env", "files", "sockets", "ports"}, facet) {
		http.Error(w, fmt.Sprintf("unknown process facet %q: want env, files, sockets or ports", facet), http.StatusNotFound)
		return
	}
	pid, err := strconv.Atoi(r.URL.Query().Get("pid"))
	if err != nil || pid <= 0 {
		http.Error(w, "pid query param required and must be a positive integer", http.StatusBadRequest)
		return
	}
	if facet == "env" && !loopbackClient(r.RemoteAddr) {
		http.Error(w, "a process environment is only served to a client on this machine", http.StatusForbidden)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), processInspectTimeout)
	defer cancel()
	if err := ownProcess(ctx, pid); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if facet == "env" {
		env, err := process.Environment(ctx, pid)
		if err != nil {
			http.Error(w, fmt.Sprintf("read environment of pid %d: %v", pid, err), http.StatusInternalServerError)
			return
		}
		writeJSON(w, env)
		return
	}
	descriptors, err := process.Descriptors(ctx, pid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	keep := map[string]func(process.Descriptor) bool{
		"files":   func(d process.Descriptor) bool { return !d.IsSocket() },
		"sockets": process.Descriptor.IsSocket,
		"ports":   process.Descriptor.IsListening,
	}[facet]
	selected := []process.Descriptor{}
	for _, d := range descriptors {
		if keep(d) {
			selected = append(selected, d)
		}
	}
	writeJSON(w, selected)
}

// ownProcess fails unless pid is this server or one of its descendants: the
// processes a run started.
func ownProcess(ctx context.Context, pid int) error {
	snapshot, err := process.Discover(ctx, process.SnapshotOptions{})
	if err != nil {
		return fmt.Errorf("list processes: %w", err)
	}
	for _, p := range snapshot.Subtree(os.Getpid()) {
		if p.PID == pid {
			return nil
		}
	}
	return fmt.Errorf("pid %d is not part of this run (exited, or started outside gavel)", pid)
}

func loopbackClient(remoteAddr string) bool {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return false
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}
