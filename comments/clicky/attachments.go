package clicky

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"strconv"

	"github.com/flanksource/gavel/comments"
)

// Attachments uses the route values id and attachment. POST sends the raw bytes,
// with name, expectedVersion and optional messageId query parameters.
func Attachments(resolve Resolver) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s, err := resolve(r.Context())
		if err != nil {
			writeError(w, err)
			return
		}
		switch r.Method {
		case http.MethodPost:
			version, err := strconv.ParseInt(r.URL.Query().Get("expectedVersion"), 10, 64)
			if err != nil {
				writeError(w, comments.ErrInvalid)
				return
			}
			c, err := s.Upload(r.Context(), comments.Upload{CommentID: r.PathValue("id"), MessageID: r.URL.Query().Get("messageId"), ExpectedVersion: version, Name: r.URL.Query().Get("name"), MediaType: r.Header.Get("Content-Type"), Body: r.Body})
			if err != nil {
				writeError(w, err)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			if err = json.NewEncoder(w).Encode(c); err != nil {
				return
			}
		case http.MethodGet:
			meta, reader, err := s.Download(r.Context(), r.PathValue("id"), r.PathValue("attachment"))
			if err != nil {
				writeError(w, err)
				return
			}
			defer reader.Close()
			w.Header().Set("Content-Type", meta.MediaType)
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": meta.Name}))
			w.Header().Set("Content-Length", strconv.FormatInt(meta.Size, 10))
			w.Header().Set("ETag", `"`+meta.SHA256+`"`)
			if _, err = io.Copy(w, reader); err != nil {
				return
			}
		default:
			w.Header().Set("Allow", "GET, POST")
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
}
func writeError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, comments.ErrInvalid):
		status = 400
	case errors.Is(err, comments.ErrForbidden):
		status = 403
	case errors.Is(err, comments.ErrNotFound):
		status = 404
	case errors.Is(err, comments.ErrConflict):
		status = 409
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()}); err != nil {
		return
	}
}
