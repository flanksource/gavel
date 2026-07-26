package ui

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/flanksource/gavel/status"
)

type projectIgnoreRequest struct {
	Path      string `json:"path"`
	Directory bool   `json:"directory"`
}

type projectIgnoreResponse struct {
	Path      string `json:"path"`
	Directory bool   `json:"directory"`
	Rule      string `json:"rule"`
	Added     bool   `json:"added"`
}

func (s *Server) handleProjectIgnore(w http.ResponseWriter, r *http.Request) {
	project, err := GetProject(r.PathValue("name"))
	if err != nil {
		respondError(w, statusForProjectErr(err), err.Error())
		return
	}
	var request projectIgnoreRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}
	normalized, err := normalizeProjectIgnorePath(request.Path)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	result, err := gatherProjectStatus(project.ResolvedDir(), status.Options{NoRepomap: true})
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("gather project status: %v", err))
		return
	}
	if err := validateProjectIgnoreTarget(normalized, request.Directory, result.Files); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	rule := projectIgnoreRule(normalized, request.Directory)
	added, err := appendProjectIgnoreRule(project.ResolvedDir(), rule)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.notify()
	respondJSON(w, http.StatusOK, projectIgnoreResponse{
		Path: normalized, Directory: request.Directory, Rule: rule, Added: added,
	})
}

func normalizeProjectIgnorePath(requested string) (string, error) {
	cleaned := path.Clean(requested)
	if requested == "" || requested != strings.TrimSpace(requested) || strings.ContainsAny(requested, "\x00\r\n") ||
		path.IsAbs(requested) || cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, "../") || cleaned != requested {
		return "", fmt.Errorf("%q must be a normalized project-relative path", requested)
	}
	if cleaned == ".git" || strings.HasPrefix(cleaned, ".git/") || cleaned == ".gitignore" {
		return "", fmt.Errorf("%q cannot be added to .gitignore", requested)
	}
	return cleaned, nil
}

func validateProjectIgnoreTarget(requested string, directory bool, files []status.FileStatus) error {
	matched := false
	for _, file := range files {
		isMatch := file.Path == requested
		if directory {
			isMatch = strings.HasPrefix(file.Path, requested+"/")
		}
		if !isMatch {
			continue
		}
		matched = true
		if file.State != status.StateUntracked {
			return fmt.Errorf("%q is tracked and cannot be hidden by .gitignore", file.Path)
		}
	}
	if !matched {
		return fmt.Errorf("%q is not a current untracked project change", requested)
	}
	return nil
}

func projectIgnoreRule(requested string, directory bool) string {
	escaped := strings.NewReplacer(
		`\`, `\\`,
		` `, `\ `,
		`*`, `\*`,
		`?`, `\?`,
		`[`, `\[`,
	).Replace(requested)
	if directory {
		return "/" + escaped + "/"
	}
	return "/" + escaped
}

func appendProjectIgnoreRule(workDir, rule string) (bool, error) {
	ignorePath := filepath.Join(workDir, ".gitignore")
	contents, err := os.ReadFile(ignorePath)
	if err != nil && !os.IsNotExist(err) {
		return false, fmt.Errorf("read %s: %w", ignorePath, err)
	}
	for _, line := range strings.Split(string(contents), "\n") {
		if line == rule {
			return false, nil
		}
	}
	prefix := ""
	if len(contents) > 0 && contents[len(contents)-1] != '\n' {
		prefix = "\n"
	}
	file, err := os.OpenFile(ignorePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return false, fmt.Errorf("open %s: %w", ignorePath, err)
	}
	if _, err := file.WriteString(prefix + rule + "\n"); err != nil {
		_ = file.Close()
		return false, fmt.Errorf("append %s: %w", ignorePath, err)
	}
	if err := file.Close(); err != nil {
		return false, fmt.Errorf("close %s: %w", ignorePath, err)
	}
	return true, nil
}
