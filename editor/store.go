package editor

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"sync"

	"github.com/flanksource/gavel/internal/fileread"
	"github.com/flanksource/gavel/utils"
)

var ErrConflict = errors.New("file changed on disk; compare or reload before saving")

const (
	KindDocument = "document"
	KindImage    = "image"
)

var documentExtensions = []string{".md", ".mdx", ".json", ".yaml", ".yml"}

type File struct {
	ID   string `json:"id"`
	Path string `json:"path"`
	// Kind is KindDocument for editable source or KindImage for a preview-only image.
	Kind string `json:"kind"`
}
type Document struct {
	File
	Source   string `json:"source"`
	Revision string `json:"revision"`
}
type entry struct {
	File
	root string
	name string
}
type Store struct {
	mu    sync.Mutex
	files map[string]entry
}

func NewStore(paths []string) (*Store, error) {
	if len(paths) == 0 {
		return nil, fmt.Errorf("provide at least one file or folder")
	}
	s := &Store{files: map[string]entry{}}
	for _, path := range paths {
		display, err := filepath.Abs(path)
		if err != nil {
			return nil, err
		}
		canonical, err := filepath.EvalSymlinks(path)
		if err != nil {
			return nil, err
		}
		canonical, err = filepath.Abs(canonical)
		if err != nil {
			return nil, err
		}
		info, err := os.Stat(canonical)
		if err != nil {
			return nil, err
		}
		if !info.IsDir() {
			kind, ok := kindOf(canonical)
			if !ok {
				return nil, fmt.Errorf("unsupported file: %s", path)
			}
			s.add(entry{File: File{Path: display, Kind: kind}, root: filepath.Dir(canonical), name: filepath.Base(canonical)})
			continue
		}
		matches, err := utils.GlobFilesBounded(canonical, supportedGlobs())
		if err != nil {
			return nil, err
		}
		for _, match := range matches {
			real, err := filepath.EvalSymlinks(match)
			if err != nil {
				return nil, err
			}
			rel, err := filepath.Rel(canonical, real)
			if err != nil {
				return nil, err
			}
			if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
				continue
			}
			kind, ok := kindOf(rel)
			if !ok {
				return nil, fmt.Errorf("glob matched an unsupported file: %s", match)
			}
			s.add(entry{File: File{Path: filepath.Join(display, rel), Kind: kind}, root: canonical, name: rel})
		}
	}
	if len(s.files) == 0 {
		return nil, fmt.Errorf("no supported files found")
	}
	return s, nil
}

func (s *Store) add(file entry) {
	id := fmt.Sprintf("%x", sha256.Sum256([]byte(filepath.Join(file.root, file.name))))
	if _, ok := s.files[id]; !ok {
		file.ID = id
		s.files[id] = file
	}
}

// kindOf classifies path by extension, and is false for an unsupported file.
func kindOf(path string) (string, bool) {
	if _, ok := fileread.ContentType(path); ok {
		return KindImage, true
	}
	if slices.Contains(documentExtensions, strings.ToLower(filepath.Ext(path))) {
		return KindDocument, true
	}
	return "", false
}

func supportedGlobs() []string {
	globs := []string{}
	for _, ext := range append(slices.Clone(documentExtensions), fileread.ImageExtensions()...) {
		globs = append(globs, "**/*"+ext)
	}
	return globs
}
func (s *Store) Files() []File {
	files := make([]File, 0, len(s.files))
	for _, e := range s.files {
		files = append(files, e.File)
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	return files
}

// Paths maps each file id to its canonical (symlink-resolved) absolute path.
func (s *Store) Paths() map[string]string {
	s.mu.Lock()
	defer s.mu.Unlock()
	paths := make(map[string]string, len(s.files))
	for id, e := range s.files {
		paths[id] = filepath.Join(e.root, e.name)
	}
	return paths
}
func revision(data []byte) string { return fmt.Sprintf("%x", sha256.Sum256(data)) }

// entry looks up id and confirms its selected path still resolves to the file
// discovered at startup, so a swapped symlink cannot redirect reads.
func (s *Store) entry(id string) (entry, error) {
	e, ok := s.files[id]
	if !ok {
		return entry{}, os.ErrNotExist
	}
	resolved, err := filepath.EvalSymlinks(e.Path)
	if err != nil {
		return entry{}, err
	}
	if resolved != filepath.Join(e.root, e.name) {
		return entry{}, fmt.Errorf("selected path changed its symlink target: %s", e.Path)
	}
	return e, nil
}

// Image returns the root-relative name and bytes of an image for id: the entry
// itself when rel is empty, or the image at rel relative to the entry's
// directory, confined to the entry's root.
func (s *Store) Image(id, rel string) (string, []byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, err := s.entry(id)
	if err != nil {
		return "", nil, err
	}
	name := filepath.ToSlash(e.name)
	if rel != "" {
		name = path.Join(path.Dir(name), rel)
	} else if e.Kind != KindImage {
		return "", nil, fmt.Errorf("%s is not an image; pass the path of an image relative to it", e.Path)
	}
	data, err := fileread.Read(e.root, filepath.FromSlash(name), fileread.Options{Image: true})
	return name, data, err
}
func (s *Store) Read(id string) (Document, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.read(id)
}
func (s *Store) read(id string) (Document, error) {
	e, err := s.entry(id)
	if err != nil {
		return Document{}, err
	}
	if e.Kind == KindImage {
		return Document{}, fmt.Errorf("%s is an image; it has no editable source", e.Path)
	}
	root, err := os.OpenRoot(e.root)
	if err != nil {
		return Document{}, err
	}
	defer root.Close()
	info, err := root.Stat(e.name)
	if err != nil {
		return Document{}, err
	}
	if !info.Mode().IsRegular() {
		return Document{}, fmt.Errorf("not a regular file: %s", e.Path)
	}
	data, err := root.ReadFile(e.name)
	if err != nil {
		return Document{}, err
	}
	return Document{File: e.File, Source: string(data), Revision: revision(data)}, nil
}
func (s *Store) Write(doc Document) (result Document, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	current, err := s.read(doc.ID)
	if err != nil {
		return Document{}, err
	}
	if current.Revision != doc.Revision {
		return Document{}, ErrConflict
	}
	e := s.files[doc.ID]
	root, err := os.OpenRoot(e.root)
	if err != nil {
		return Document{}, err
	}
	defer root.Close()
	info, err := root.Stat(e.name)
	if err != nil {
		return Document{}, err
	}
	temp := filepath.Join(filepath.Dir(e.name), ".gavel-edit-"+revision([]byte(doc.Source)))
	file, err := root.OpenFile(temp, os.O_WRONLY|os.O_CREATE|os.O_EXCL, info.Mode().Perm())
	if err != nil {
		return Document{}, err
	}
	defer func() {
		cleanupErr := root.Remove(temp)
		if !errors.Is(cleanupErr, os.ErrNotExist) {
			err = errors.Join(err, cleanupErr)
		}
	}()
	if err := file.Chmod(info.Mode().Perm()); err != nil {
		return Document{}, errors.Join(err, file.Close())
	}
	_, writeErr := file.WriteString(doc.Source)
	if err := errors.Join(writeErr, file.Sync(), file.Close()); err != nil {
		return Document{}, err
	}
	// Check again after preparing the replacement so external edits are not knowingly overwritten.
	current, err = s.read(doc.ID)
	if err != nil {
		return Document{}, err
	}
	if current.Revision != doc.Revision {
		return Document{}, ErrConflict
	}
	if err := root.Rename(temp, e.name); err != nil {
		return Document{}, err
	}
	return s.read(doc.ID)
}
