package runstatus

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/flanksource/clicky/sse"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/sdk/status"
)

type socketServer struct {
	path   string
	server *http.Server
	served chan error
}

// serve opens the status socket and records it in the status.
func (r *Reporter) serve() error {
	path := status.SocketPath(r.options.Cache, os.Getpid())
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("remove leftover socket %s: %w", path, err)
	}
	listener, err := net.Listen("unix", path)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", path, err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /status", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(r.Status()); err != nil {
			logger.Debugf("gavel status: %v", err)
		}
	})
	mux.HandleFunc("GET /status/stream", func(w http.ResponseWriter, req *http.Request) {
		wake, unsubscribe := r.changed.Subscribe()
		defer unsubscribe()
		err := sse.ServeSnapshot(w, req, sse.SnapshotOptions{
			Load:     func(context.Context) (any, error) { return r.Status(), nil },
			Interval: 5 * time.Second,
			Wake:     wake,
		})
		if err != nil {
			logger.Debugf("gavel status stream: %v", err)
		}
	})
	server := &socketServer{path: path, server: &http.Server{Handler: mux, ReadHeaderTimeout: 5 * time.Second}, served: make(chan error, 1)}
	go func() { server.served <- server.server.Serve(listener) }()
	r.mu.Lock()
	r.server = server
	r.current.Socket = path
	r.dirty = true
	r.mu.Unlock()
	return nil
}

// close stops the socket, ending its streams (which never go idle, so there is
// nothing to drain), and removes it.
func (s *socketServer) close() error {
	err := s.server.Close()
	if served := <-s.served; !errors.Is(served, http.ErrServerClosed) {
		err = errors.Join(err, served)
	}
	if removeErr := os.Remove(s.path); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
		err = errors.Join(err, removeErr)
	}
	return err
}
