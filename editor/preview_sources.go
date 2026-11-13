package editor

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type PreviewContext struct {
	Page          string            `json:"page"`
	Paths         []string          `json:"paths"`
	Dependencies  map[string]string `json:"dependencies"`
	RuntimeOrigin string            `json:"runtimeOrigin"`
}

func (s *Store) previewRoot(id string) (string, string, error) {
	s.mu.Lock()
	e, ok := s.files[id]
	s.mu.Unlock()
	if !ok {
		return "", "", os.ErrNotExist
	}
	file := filepath.Join(e.root, e.name)
	root := filepath.Dir(file)
	for dir := root; ; dir = filepath.Dir(dir) {
		_, err := os.Stat(filepath.Join(dir, "package.json"))
		if err == nil {
			root = dir
			break
		}
		if !os.IsNotExist(err) {
			return "", "", err
		}
		_, gitErr := os.Stat(filepath.Join(dir, ".git"))
		if gitErr != nil && !os.IsNotExist(gitErr) {
			return "", "", gitErr
		}
		if gitErr == nil || filepath.Dir(dir) == dir {
			break
		}
	}
	page, err := filepath.Rel(root, file)
	return root, filepath.ToSlash(page), err
}

func (s *Store) Preview(id string) (PreviewContext, error) {
	root, page, err := s.previewRoot(id)
	if err != nil {
		return PreviewContext{}, err
	}
	result := PreviewContext{Page: page, Paths: []string{}, Dependencies: map[string]string{}}
	err = filepath.WalkDir(root, func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if name != root && (strings.HasPrefix(entry.Name(), ".") || entry.Name() == "node_modules" || entry.Name() == "dist") {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 || strings.HasPrefix(entry.Name(), ".") {
			return nil
		}
		switch strings.ToLower(filepath.Ext(name)) {
		case ".ts", ".tsx", ".js", ".jsx", ".mjs", ".md", ".mdx", ".css", ".json", ".svg", ".png", ".jpg", ".jpeg", ".webp", ".gif", ".woff", ".woff2", ".ttf":
			rel, err := filepath.Rel(root, name)
			if err != nil {
				return err
			}
			result.Paths = append(result.Paths, filepath.ToSlash(rel))
		}
		return nil
	})
	if err != nil {
		return result, err
	}
	data, err := os.ReadFile(filepath.Join(root, "package.json"))
	if os.IsNotExist(err) {
		return result, nil
	}
	if err != nil {
		return result, err
	}
	var manifest struct {
		Dependencies map[string]string `json:"dependencies"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		return result, fmt.Errorf("preview package.json: %w", err)
	}
	for name, version := range manifest.Dependencies {
		if name == "react" || name == "react-dom" || name == "scheduler" || name == "jotai" || name == "@tanstack/react-query" {
			continue
		}
		result.Dependencies[name] = version
	}
	return result, nil
}

func (s *Server) previewContext(w http.ResponseWriter, r *http.Request) {
	result, err := s.Store.Preview(r.PathValue("id"))
	result.RuntimeOrigin = s.PreviewOrigin
	if err == nil {
		err = s.resolvePreview(r.Context(), result.Dependencies, r.URL.Query().Get("facet"))
	}
	respond(w, result, err)
}

func (s *Store) previewSource(w http.ResponseWriter, r *http.Request) {
	root, _, err := s.previewRoot(r.PathValue("id"))
	if err != nil {
		respond(w, nil, err)
		return
	}
	name := r.PathValue("path")
	if !fs.ValidPath(name) || strings.Contains(name, "\\") {
		http.Error(w, "invalid source path", 400)
		return
	}
	context, err := s.Preview(r.PathValue("id"))
	if err != nil {
		respond(w, nil, err)
		return
	}
	allowed := false
	for _, path := range context.Paths {
		if path == name {
			allowed = true
			break
		}
	}
	if !allowed {
		http.Error(w, "source outside approved index", 403)
		return
	}
	dir, err := os.OpenRoot(root)
	if err != nil {
		respond(w, nil, err)
		return
	}
	defer dir.Close()
	file, err := dir.Open(name)
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
	http.ServeContent(w, r, name, info.ModTime(), file)
}
