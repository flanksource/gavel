package workspacefiles

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

func (s *Service) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/workspace/directories", func(w http.ResponseWriter, r *http.Request) {
		entries, err := s.Directories(r.URL.Query().Get("path"))
		respond(w, entries, err)
	})
	mux.HandleFunc("GET /api/workspace/repository", func(w http.ResponseWriter, r *http.Request) {
		repo, err := s.Repository(r.Context(), r.URL.Query().Get("path"))
		respond(w, repo, err)
	})
	mux.HandleFunc("POST /api/workspace/entries", func(w http.ResponseWriter, r *http.Request) {
		var input EntriesRequest
		if !decode(w, r, &input) {
			return
		}
		entries, err := s.Entries(r.Context(), input)
		respond(w, entries, err)
	})
	mux.HandleFunc("GET /api/workspace/file", func(w http.ResponseWriter, r *http.Request) {
		file, err := s.Read(r.Context(), fileQuery(r))
		respond(w, file, err)
	})
	mux.HandleFunc("PUT /api/workspace/file", func(w http.ResponseWriter, r *http.Request) {
		var input SaveRequest
		if !decode(w, r, &input) {
			return
		}
		file, err := s.Save(r.Context(), input)
		respond(w, file, err)
	})
	mux.HandleFunc("GET /api/workspace/asset", s.asset)
	mux.HandleFunc("GET /api/workspace/blame", s.blame)
	mux.HandleFunc("POST /api/workspace/watch", func(w http.ResponseWriter, r *http.Request) {
		var input FileRequest
		if !decode(w, r, &input) {
			return
		}
		lease, err := s.Acquire(r.Context(), input, r.Header.Get("Authorization"))
		respond(w, lease, err)
	})
	mux.HandleFunc("POST /api/workspace/watch/{id}", func(w http.ResponseWriter, r *http.Request) {
		lease, err := s.Renew(r.PathValue("id"), r.Header.Get("Authorization"))
		respond(w, lease, err)
	})
	mux.HandleFunc("DELETE /api/workspace/watch/{id}", func(w http.ResponseWriter, r *http.Request) {
		respond(w, map[string]bool{"released": true}, s.Release(r.PathValue("id"), r.Header.Get("Authorization")))
	})
	mux.HandleFunc("GET /api/workspace/watch/{id}/stream", s.stream)
	return mux
}

func fileQuery(r *http.Request) FileRequest {
	return FileRequest{Root: r.URL.Query().Get("root"), Path: r.URL.Query().Get("path")}
}

func decode(w http.ResponseWriter, r *http.Request, input any) bool {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 10<<20))
	decoder.DisallowUnknownFields()
	err := decoder.Decode(input)
	if err == nil {
		if decoder.Decode(new(any)) != io.EOF {
			err = errors.New("expected exactly one JSON object")
		}
	}
	if err != nil {
		w.WriteHeader(422)
		if err := json.NewEncoder(w).Encode(map[string]string{"error": err.Error()}); err != nil {
			return false
		}
		return false
	}
	return true
}

func respond(w http.ResponseWriter, value any, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	if err != nil {
		status := 422
		if errors.Is(err, fs.ErrNotExist) {
			status = 404
		}
		if errors.Is(err, ErrConflict) {
			status = 409
		}
		if errors.Is(err, ErrExpired) {
			status = 410
		}
		w.WriteHeader(status)
		value = map[string]string{"error": err.Error()}
	}
	if err := json.NewEncoder(w).Encode(value); err != nil {
		return
	}
}

func (s *Service) asset(w http.ResponseWriter, r *http.Request) {
	input := fileQuery(r)
	ext := strings.ToLower(filepath.Ext(input.Path))
	allowed := map[string]string{".png": "image/png", ".jpg": "image/jpeg", ".jpeg": "image/jpeg", ".gif": "image/gif", ".webp": "image/webp", ".avif": "image/avif", ".svg": "image/svg+xml"}
	contentType, ok := allowed[ext]
	if !ok {
		respond(w, nil, errors.New("only local image assets are supported"))
		return
	}
	root, err := s.open(r.Context(), input)
	if err != nil {
		respond(w, nil, err)
		return
	}
	defer root.Close()
	file, err := root.Open(input.Path)
	if err != nil {
		respond(w, nil, err)
		return
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		respond(w, nil, err)
		return
	}
	if !info.Mode().IsRegular() || info.Size() > 32<<20 {
		respond(w, nil, errors.New("asset must be a regular image under 32 MiB"))
		return
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Security-Policy", "sandbox; default-src 'none'")
	http.ServeContent(w, r, info.Name(), info.ModTime(), file)
}

func (s *Service) stream(w http.ResponseWriter, r *http.Request) {
	id, owner := r.PathValue("id"), r.Header.Get("Authorization")
	events, err := s.Subscribe(id, owner)
	if err != nil {
		respond(w, nil, err)
		return
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		respond(w, nil, errors.New("streaming unsupported"))
		return
	}
	s.mu.Lock()
	entry := s.leases[id]
	if entry == nil {
		s.mu.Unlock()
		respond(w, nil, ErrExpired)
		return
	}
	input := entry.File.FileRequest
	s.mu.Unlock()
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-store")
	send := func(event string, value any) bool {
		data, err := json.Marshal(value)
		if err != nil {
			return false
		}
		_, err = fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, data)
		flusher.Flush()
		return err == nil
	}
	snapshot := func() bool {
		file, err := s.Read(r.Context(), input)
		if err != nil {
			return send("file-error", map[string]string{"error": err.Error()})
		}
		return send("file", file)
	}
	if !snapshot() {
		return
	}
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-r.Context().Done():
			return
		case <-s.stop:
			return
		case event := <-events:
			if event.Error != "" {
				send("file-error", event)
				return
			}
			if event.Expired {
				send("expired", event)
				return
			}
			if !snapshot() {
				return
			}
		case <-ticker.C:
			if _, err := fmt.Fprint(w, ": keepalive\n\n"); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}
