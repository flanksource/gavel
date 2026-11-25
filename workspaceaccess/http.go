package workspaceaccess

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/flanksource/gavel/internal/database"
	"github.com/jackc/pgx/v5"
)

func Trusted(r *http.Request) bool {
	host, _, err := net.SplitHostPort(r.Host)
	if err != nil {
		host = r.Host
	}
	peer, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return false
	}
	ip := net.ParseIP(peer)
	localHost := host == "localhost" || host == "127.0.0.1" || host == "::1"
	return localHost && ip != nil && ip.IsLoopback() && (r.Header.Get("Origin") == "" || r.Header.Get("Origin") == "http://"+r.Host)
}
func Respond(w http.ResponseWriter, value any, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	if err != nil {
		var denied *Denied
		if errors.As(err, &denied) {
			w.WriteHeader(403)
			value = denied
		} else {
			w.WriteHeader(422)
			value = map[string]string{"error": err.Error()}
		}
	}
	_ = json.NewEncoder(w).Encode(value)
}
func (s *Service) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/workspace/access", func(w http.ResponseWriter, r *http.Request) {
		roots, err := s.Roots(r.Context())
		Respond(w, map[string]any{"roots": roots, "canManage": true}, err)
	})
	mux.HandleFunc("POST /api/workspace/access/preview", func(w http.ResponseWriter, r *http.Request) {
		var input PreviewRequest
		if !decode(w, r, &input) {
			return
		}
		path, err := Canonical(input.Path)
		Respond(w, map[string]string{"path": path}, err)
	})
	mux.HandleFunc("POST /api/workspace/access", func(w http.ResponseWriter, r *http.Request) {
		var input GrantRequest
		if !decode(w, r, &input) {
			return
		}
		root, err := s.Grant(r.Context(), input.Path, input.CanonicalPath)
		Respond(w, root, err)
	})
	mux.HandleFunc("DELETE /api/workspace/access/{id}", func(w http.ResponseWriter, r *http.Request) {
		err := s.Repository.Revoke(r.Context(), r.PathValue("id"))
		Respond(w, map[string]bool{"revoked": err == nil}, err)
	})
	mux.HandleFunc("GET /api/workspace/access/events", s.events)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !Trusted(r) {
			Respond(w, nil, &Denied{Code: "workspace_access_denied", ErrorMessage: "Workspace access management requires a local same-origin caller"})
			return
		}
		mux.ServeHTTP(w, r)
	})
}
func decode(w http.ResponseWriter, r *http.Request, value any) bool {
	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		http.Error(w, "application/json required", 415)
		return false
	}
	d := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096))
	d.DisallowUnknownFields()
	if err := d.Decode(value); err != nil {
		Respond(w, nil, err)
		return false
	}
	if err := d.Decode(new(any)); err != io.EOF {
		Respond(w, nil, errors.New("expected exactly one JSON object"))
		return false
	}
	return true
}
func (s *Service) events(w http.ResponseWriter, r *http.Request) {
	started := false
	f, ok := w.(http.Flusher)
	if !ok {
		Respond(w, nil, errors.New("streaming unavailable"))
		return
	}
	err := s.Follow(r.Context(), func() error {
		started = true
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-store")
		_, err := fmt.Fprint(w, "event: workspace-access\ndata: {}\n\n")
		f.Flush()
		return err
	})
	if err != nil && r.Context().Err() == nil {
		if !started {
			Respond(w, nil, err)
		} else {
			data, _ := json.Marshal(map[string]string{"error": err.Error()})
			_, _ = fmt.Fprintf(w, "event: workspace-access-error\ndata: %s\n\n", data)
			f.Flush()
		}
	}
}
func (s *Service) Follow(ctx context.Context, changed func() error) error {
	if _, err := database.Require(ctx, "workspace access notifications"); err != nil {
		return err
	}
	db, err := database.Shared(ctx)
	if err != nil {
		return err
	}
	conn, err := pgx.Connect(ctx, db.DSN())
	if err != nil {
		return err
	}
	defer conn.Close(context.Background())
	if _, err = conn.Exec(ctx, "LISTEN gavel_workspace_access_change"); err != nil {
		return err
	}
	for ctx.Err() == nil {
		if err := changed(); err != nil {
			return err
		}
		wait, cancel := context.WithTimeout(ctx, 15*time.Second)
		_, err = conn.WaitForNotification(wait)
		cancel()
		if err != nil && !errors.Is(err, context.DeadlineExceeded) {
			return err
		}
	}
	return ctx.Err()
}
