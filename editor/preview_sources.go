package editor

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"slices"
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

// PreviewExtensionRefs rewrites every absolute file `ref` in a schema's
// x-clicky-extensions (as jsonlint.EditorSchema resolves them) into a path
// inside the document's preview root, the path the runtime's source provider
// serves. A ref outside that root, or to a file the root does not index,
// fails rather than leaving the form without its extension.
func (s *Store) PreviewExtensionRefs(id string, schema any) error {
	var preview *PreviewContext
	var root string
	var walk func(any) error
	walk = func(value any) error {
		switch node := value.(type) {
		case map[string]any:
			for key, child := range node {
				if key != "x-clicky-extensions" {
					if err := walk(child); err != nil {
						return err
					}
					continue
				}
				sources, _ := child.([]any)
				for _, item := range sources {
					source, _ := item.(map[string]any)
					ref, ok := source["ref"].(string)
					if !ok {
						continue
					}
					if preview == nil {
						var err error
						if root, _, err = s.previewRoot(id); err != nil {
							return err
						}
						context, err := s.Preview(id)
						if err != nil {
							return err
						}
						preview = &context
					}
					path, err := previewExtensionPath(ref, root, preview.Paths)
					if err != nil {
						return err
					}
					source["ref"] = path
				}
			}
		case []any:
			for _, child := range node {
				if err := walk(child); err != nil {
					return err
				}
			}
		}
		return nil
	}
	return walk(schema)
}

func previewExtensionPath(ref, root string, paths []string) (string, error) {
	parsed, err := url.Parse(ref)
	if err != nil {
		return "", fmt.Errorf("form extension %s: %w", ref, err)
	}
	if parsed.Scheme != "file" {
		return "", fmt.Errorf("form extension %s must be a file in the project", ref)
	}
	rel, err := filepath.Rel(root, filepath.FromSlash(parsed.Path))
	if err != nil {
		return "", fmt.Errorf("form extension %s: %w", ref, err)
	}
	rel = filepath.ToSlash(rel)
	if rel == ".." || strings.HasPrefix(rel, "../") {
		return "", fmt.Errorf("form extension %s is outside the preview root %s", ref, root)
	}
	if !slices.Contains(paths, rel) {
		return "", fmt.Errorf("form extension %s is not a previewable project source under %s", ref, root)
	}
	return rel, nil
}
