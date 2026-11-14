package editor

import (
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var ErrExists = errors.New("a file already exists at that path")

// editRoot is a directory files are placed in, as opened (display) and
// symlink-resolved (canonical).
type editRoot struct{ display, canonical string }

func (s *Store) addRoot(display string) error {
	canonical, err := filepath.EvalSymlinks(display)
	if err != nil {
		return err
	}
	for _, root := range s.roots {
		if root.display == display {
			return nil
		}
	}
	s.roots = append(s.roots, editRoot{display: display, canonical: canonical})
	return nil
}

// place is the deepest edited root holding the absolute path, with the path's
// name below it; a path outside every root is refused.
func (s *Store) place(path string) (editRoot, string, error) {
	if !filepath.IsAbs(path) {
		return editRoot{}, "", fmt.Errorf("path must be absolute: %q", path)
	}
	var best editRoot
	name := ""
	for _, root := range s.roots {
		rel, err := filepath.Rel(root.display, filepath.Clean(path))
		if err != nil || !filepath.IsLocal(rel) || len(root.display) < len(best.display) {
			continue
		}
		best, name = root, rel
	}
	if name == "" {
		displays := make([]string, 0, len(s.roots))
		for _, root := range s.roots {
			displays = append(displays, root.display)
		}
		return editRoot{}, "", fmt.Errorf("%s is outside the edited folders (%s)", path, strings.Join(displays, ", "))
	}
	return best, name, nil
}

// checkRevision refuses to move or delete a document changed on disk since
// the caller read it at revision.
func (s *Store) checkRevision(e entry, revision string) error {
	if e.Kind != KindDocument {
		return nil
	}
	current, err := s.read(e.ID)
	if err != nil {
		return err
	}
	if current.Revision != revision {
		return ErrConflict
	}
	return nil
}

// Create writes a new document at the absolute path, seeded from the template
// (empty for none), creating missing parent folders; an existing file is never replaced.
func (s *Store) Create(path, templateID string) (Document, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	root, name, err := s.place(path)
	if err != nil {
		return Document{}, err
	}
	if kind, _ := kindOf(name); kind != KindDocument {
		return Document{}, fmt.Errorf("cannot create %s: new files must end in one of %s", path, strings.Join(documentExtensions, ", "))
	}
	source, err := sourceFor(templateID, name)
	if err != nil {
		return Document{}, err
	}
	dir, err := os.OpenRoot(root.canonical)
	if err != nil {
		return Document{}, err
	}
	defer dir.Close()
	if err := dir.MkdirAll(filepath.Dir(name), 0o755); err != nil {
		return Document{}, err
	}
	file, err := dir.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if errors.Is(err, fs.ErrExist) {
		return Document{}, fmt.Errorf("%w: %s", ErrExists, path)
	}
	if err != nil {
		return Document{}, err
	}
	_, writeErr := file.WriteString(source)
	if err := errors.Join(writeErr, file.Sync(), file.Close()); err != nil {
		return Document{}, errors.Join(err, dir.Remove(name))
	}
	return s.read(s.add(entry{File: File{Path: filepath.Join(root.display, name), Kind: KindDocument}, root: root.canonical, name: name}))
}

// Rename moves the file id to the absolute path within its edited root, keeping
// its kind. The moved file has a new id, which is returned.
func (s *Store) Rename(id, path, revision string) (File, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, err := s.entry(id)
	if err != nil {
		return File{}, err
	}
	if err := s.checkRevision(e, revision); err != nil {
		return File{}, err
	}
	root, name, err := s.place(path)
	if err != nil {
		return File{}, err
	}
	if root.canonical != e.root {
		return File{}, fmt.Errorf("cannot move %s to %s: files stay within the folder they were opened from (%s)", e.Path, path, filepath.Dir(e.Path))
	}
	if kind, _ := kindOf(name); kind != e.Kind {
		return File{}, fmt.Errorf("cannot rename %s to %s: the new name must keep it a supported %s", e.Path, path, e.Kind)
	}
	dir, err := os.OpenRoot(e.root)
	if err != nil {
		return File{}, err
	}
	defer dir.Close()
	if _, err := dir.Lstat(name); err == nil {
		return File{}, fmt.Errorf("%w: %s", ErrExists, path)
	} else if !errors.Is(err, fs.ErrNotExist) {
		return File{}, err
	}
	if err := dir.MkdirAll(filepath.Dir(name), 0o755); err != nil {
		return File{}, err
	}
	if err := dir.Rename(e.name, name); err != nil {
		return File{}, err
	}
	delete(s.files, id)
	return s.files[s.add(entry{File: File{Path: filepath.Join(root.display, name), Kind: e.Kind}, root: e.root, name: name})].File, nil
}

// Delete removes the file id from disk and from the edited files.
func (s *Store) Delete(id, revision string) (File, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, err := s.entry(id)
	if err != nil {
		return File{}, err
	}
	if err := s.checkRevision(e, revision); err != nil {
		return File{}, err
	}
	dir, err := os.OpenRoot(e.root)
	if err != nil {
		return File{}, err
	}
	defer dir.Close()
	if err := dir.Remove(e.name); err != nil {
		return File{}, err
	}
	delete(s.files, id)
	return e.File, nil
}

type fileOpRequest struct {
	Path     string `json:"path"`
	Template string `json:"template"`
	Revision string `json:"revision"`
}

// fileOp answers a file operation once the problems are reloaded for the
// changed file list, so they never point at a removed id.
func (s *Server) fileOp(w http.ResponseWriter, status int, value any, err error) {
	if err == nil {
		if reloadErr := s.Reload(); reloadErr != nil {
			err = fmt.Errorf("the file was changed, but reloading its problems failed: %w", reloadErr)
		}
	}
	if err == nil && status != http.StatusOK {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
	}
	respond(w, value, err)
}

func (s *Server) createFile(w http.ResponseWriter, r *http.Request) {
	var input fileOpRequest
	if !decode(w, r, &input) {
		return
	}
	doc, err := s.Store.Create(input.Path, input.Template)
	s.fileOp(w, http.StatusCreated, doc, err)
}

func (s *Server) renameFile(w http.ResponseWriter, r *http.Request) {
	var input fileOpRequest
	if !decode(w, r, &input) {
		return
	}
	file, err := s.Store.Rename(r.PathValue("id"), input.Path, input.Revision)
	s.fileOp(w, http.StatusOK, file, err)
}

func (s *Server) deleteFile(w http.ResponseWriter, r *http.Request) {
	var input fileOpRequest
	if !decode(w, r, &input) {
		return
	}
	file, err := s.Store.Delete(r.PathValue("id"), input.Revision)
	s.fileOp(w, http.StatusOK, file, err)
}
