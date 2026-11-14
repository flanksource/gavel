package clicky

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/flanksource/gavel/comments"
	"net/http"
	"strconv"
	"time"
)

// Stream serves an authorized, resumable SSE feed. Hosts own routing and middleware.
func Stream(resolve Resolver) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		service, resolveErr := resolve(r.Context())
		if resolveErr != nil {
			writeError(w, resolveErr)
			return
		}
		cursor := r.Header.Get("Last-Event-ID")
		if cursor == "" {
			cursor = r.URL.Query().Get("after")
		}
		var after int64
		var err error
		if cursor != "" {
			after, err = strconv.ParseInt(cursor, 10, 64)
			if err != nil || after < 0 {
				http.Error(w, "invalid event cursor", http.StatusBadRequest)
				return
			}
		}
		scope := r.URL.Query().Get("scope")
		events, err := service.Events(r.Context(), scope, after, 200)
		if err != nil {
			status := http.StatusInternalServerError
			if errors.Is(err, comments.ErrForbidden) {
				status = http.StatusForbidden
			} else if errors.Is(err, comments.ErrInvalid) {
				status = http.StatusBadRequest
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(status)
			if encodeErr := json.NewEncoder(w).Encode(map[string]string{"error": err.Error()}); encodeErr != nil {
				return
			}
			return
		}
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming unsupported", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			for _, event := range events {
				data, err := json.Marshal(event)
				if err != nil {
					return
				}
				if _, err = fmt.Fprintf(w, "id: %d\nevent: comment\ndata: %s\n\n", event.Sequence, data); err != nil {
					return
				}
				after = event.Sequence
			}
			flusher.Flush()
			select {
			case <-r.Context().Done():
				return
			case <-ticker.C:
			}
			events, err = service.Events(r.Context(), scope, after, 200)
			if err != nil {
				return
			}
		}
	})
}
