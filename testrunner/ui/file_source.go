package testui

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/flanksource/gavel/internal/fileread"
)

// handleFileSource serves GET /api/file?work_dir=&path= — the raw text of a
// file under one of the run's known work dirs, so the UI can overlay findings.
func (s *Server) handleFileSource(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	workDir, path := r.URL.Query().Get("work_dir"), r.URL.Query().Get("path")
	if workDir == "" || path == "" {
		http.Error(w, "work_dir and path are required", http.StatusBadRequest)
		return
	}
	if !filepath.IsLocal(path) {
		http.Error(w, fmt.Sprintf("path %q must be relative to work_dir and stay inside it", path), http.StatusBadRequest)
		return
	}
	s.mu.Lock()
	root := s.resolveWorkDirLocked(workDir)
	mapped, ok := s.mappedSourceLocked(root, path)
	s.mu.Unlock()
	if root == "" {
		http.Error(w, fmt.Sprintf("work_dir %q is not a work dir of this run", workDir), http.StatusBadRequest)
		return
	}
	if ok {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte(mapped))
		return
	}
	data, err := fileread.Read(root, path)
	if err != nil {
		writeFileSourceError(w, path, err)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write(data)
}

// mappedSourceLocked is the content a lint result run from root maps path to
// (linters.LinterResult.Sources), which wins over a file of that name.
func (s *Server) mappedSourceLocked(root, path string) (string, bool) {
	if root == "" {
		return "", false
	}
	for _, result := range s.lint {
		if result == nil || result.WorkDir == "" {
			continue
		}
		if abs, err := filepath.Abs(result.WorkDir); err != nil || filepath.Clean(abs) != root {
			continue
		}
		if source, ok := result.Sources[filepath.ToSlash(path)]; ok && source.Content != "" {
			return source.Content, true
		}
	}
	return "", false
}

// resolveWorkDirLocked returns the absolute form of workDir when the run
// vouches for it (a test or lint work dir, or the git root), and "" otherwise.
func (s *Server) resolveWorkDirLocked(workDir string) string {
	requested, err := filepath.Abs(workDir)
	if err != nil {
		return ""
	}
	return matchWorkDir(s.knownWorkDirsLocked(), filepath.Clean(requested))
}

func writeFileSourceError(w http.ResponseWriter, path string, err error) {
	status := http.StatusBadRequest
	switch {
	case errors.Is(err, os.ErrNotExist):
		status = http.StatusNotFound
	case errors.Is(err, fileread.ErrTooLarge):
		status = http.StatusRequestEntityTooLarge
	case errors.Is(err, fileread.ErrBinary):
		status = http.StatusUnsupportedMediaType
	case errors.Is(err, os.ErrPermission):
		status = http.StatusForbidden
	}
	http.Error(w, fmt.Sprintf("read %s: %v", path, err), status)
}
