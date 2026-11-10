// The run API on a socket in a substrate environment: the environment's program serves it to what
// runs under it, forwarding each request to the run's caller over the run's host call.
package runclient

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/flanksource/gavel/internal/safefs"
)

// SocketEnv names the socket the environment's program serves the run API on.
const SocketEnv = "GAVEL_RUN_API_SOCKET"

// ServeRun serves the run API through call on a socket in a directory of its own under the
// temporary directory, whose path fits in the 108 bytes a socket's may take wherever the run's
// directory is, and names it in SocketEnv for what the run starts, which inherits this process's
// environment. It returns a func that stops serving, removes the socket and unnames it.
func ServeRun(call func(name string, data json.RawMessage) (json.RawMessage, error)) (func() error, error) {
	dir, err := os.MkdirTemp("", "gavel-run-api-")
	if err != nil {
		return nil, fmt.Errorf("make the run API's socket directory: %w", err)
	}
	socket := filepath.Join(dir, "run-api.sock")
	stop, err := ServeOverHostCall(socket, call)
	if err != nil {
		return nil, errors.Join(err, removeSocketDir(dir))
	}
	stopAll := func() error { return errors.Join(stop(), removeSocketDir(dir), os.Unsetenv(SocketEnv)) }
	if err := os.Setenv(SocketEnv, socket); err != nil {
		return nil, errors.Join(fmt.Errorf("name the run API's socket for the run: %w", err), stopAll())
	}
	return stopAll, nil
}

// removeSocketDir removes the directory ServeRun made for the socket.
func removeSocketDir(dir string) error {
	return safefs.SafeRemove(dir, safefs.RemoveContext{AllowedRoot: os.TempDir()})
}

// ServeOverHostCall serves the run API on a unix socket at path, answering each request through
// call, the run's host call to its caller, which answers it as the run. It returns a func that
// stops serving.
func ServeOverHostCall(path string, call func(name string, data json.RawMessage) (json.RawMessage, error)) (func() error, error) {
	listener, err := net.Listen("unix", path)
	if err != nil {
		return nil, fmt.Errorf("serve the run API on %s: %w", path, err)
	}
	client := OverHostCall(call)
	server := &http.Server{ReadHeaderTimeout: 10 * time.Second, Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path, ok := runPath(r.URL.Path)
		if !ok {
			writeError(w, http.StatusNotFound, fmt.Sprintf("%s is not a run's endpoint", r.URL.Path))
			return
		}
		if r.URL.RawQuery != "" {
			path += "?" + r.URL.RawQuery
		}
		response, err := client.exchange(r.Method, path, r.Body)
		if err != nil {
			writeError(w, http.StatusBadGateway, err.Error())
			return
		}
		w.WriteHeader(response.Status)
		_, _ = w.Write(response.Body)
	})}
	go func() { _ = server.Serve(listener) }()
	return server.Close, nil
}

// runPath is what of the run API urlPath asks for, without /api/runs/<run>/: the caller answers as
// its own run, whichever the request names.
func runPath(urlPath string) (string, bool) {
	rest, ok := strings.CutPrefix(urlPath, "/api/runs/")
	if !ok {
		return "", false
	}
	_, path, ok := strings.Cut(rest, "/")
	return path, ok && path != ""
}

// writeError answers with status and reason, as the run API refuses a request.
func writeError(w http.ResponseWriter, status int, reason string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": reason})
}
