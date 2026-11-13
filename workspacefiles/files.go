package workspacefiles

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	git "github.com/flanksource/commons-db/git"
	"github.com/fsnotify/fsnotify"
	"github.com/google/uuid"
)

var ErrConflict = errors.New("file changed on disk; reload before saving")
var ErrExpired = errors.New("file watch lease expired")

type Options struct {
	Roots       []string
	Authorize   func(string) (string, error)
	IdleTimeout time.Duration
}
type Service struct {
	options     Options
	mu          sync.Mutex
	watcher     *fsnotify.Watcher
	leases      map[string]*lease
	directories map[string]int
	stop        chan struct{}
	done        chan struct{}
}
type FileRequest struct {
	Root string `json:"root"`
	Path string `json:"path"`
}
type SaveRequest struct {
	FileRequest
	Revision string `json:"revision"`
	Source   string `json:"source"`
}
type File struct {
	FileRequest
	Source   string `json:"source"`
	Revision string `json:"revision"`
}
type Repository struct {
	Root string `json:"root"`
	Name string `json:"name"`
}
type DirectoryEntry struct {
	Path      string `json:"path"`
	Name      string `json:"name"`
	Directory bool   `json:"directory"`
	Visible   bool   `json:"visible"`
	Rule      string `json:"rule,omitempty"`
}
type EntriesRequest struct {
	FileRequest
	Overrides         []string `json:"overrides"`
	OverrideDirectory string   `json:"overrideDirectory,omitempty"`
	IncludeIgnored    bool     `json:"includeIgnored"`
}

func New(options Options) (*Service, error) {
	if options.Authorize == nil {
		return nil, errors.New("filesystem authorization is required")
	}
	if options.IdleTimeout == 0 {
		options.IdleTimeout = 30 * time.Second
	}
	if options.IdleTimeout < 0 {
		return nil, errors.New("watch idle timeout must be positive")
	}
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	service := &Service{options: options, watcher: watcher, leases: map[string]*lease{}, directories: map[string]int{}, stop: make(chan struct{}), done: make(chan struct{})}
	go service.watch()
	return service, nil
}

func (s *Service) Repository(ctx context.Context, path string) (Repository, error) {
	if path == "" {
		return Repository{}, errors.New("repository root is required")
	}
	canonical, err := s.options.Authorize(path)
	if err != nil {
		return Repository{}, err
	}
	found, err := git.Discover(ctx, git.DiscoverOptions{Path: canonical})
	if err != nil {
		return Repository{}, err
	}
	if found.Bare || found.Root != canonical {
		return Repository{}, fmt.Errorf("select a Git working-tree root: %s", canonical)
	}
	return Repository{Root: canonical, Name: filepath.Base(canonical)}, nil
}

func validPath(path string) bool {
	return fs.ValidPath(path) && !strings.ContainsRune(path, '\\') && !strings.Contains("/"+path+"/", "/.git/")
}

func (s *Service) open(ctx context.Context, input FileRequest) (*os.Root, error) {
	if !validPath(input.Path) {
		return nil, fmt.Errorf("invalid repository path %q", input.Path)
	}
	repo, err := s.Repository(ctx, input.Root)
	if err != nil {
		return nil, err
	}
	return os.OpenRoot(repo.Root)
}

func revision(data []byte) string { hash := sha256.Sum256(data); return hex.EncodeToString(hash[:]) }

func read(root *os.Root, input FileRequest) (File, error) {
	info, err := root.Stat(input.Path)
	if err != nil {
		return File{}, err
	}
	if !info.Mode().IsRegular() || info.Size() > 8<<20 {
		return File{}, errors.New("document must be a regular file smaller than 8 MiB")
	}
	data, err := root.ReadFile(input.Path)
	if err != nil {
		return File{}, err
	}
	if !utf8.Valid(data) {
		return File{}, errors.New("document is not UTF-8 text")
	}
	return File{FileRequest: input, Source: string(data), Revision: revision(data)}, nil
}

func (s *Service) Read(ctx context.Context, input FileRequest) (File, error) {
	root, err := s.open(ctx, input)
	if err != nil {
		return File{}, err
	}
	defer root.Close()
	return read(root, input)
}

func (s *Service) Save(ctx context.Context, input SaveRequest) (result File, resultErr error) {
	if input.Revision == "" || len(input.Source) > 8<<20 || !utf8.ValidString(input.Source) {
		return File{}, errors.New("save requires revision and UTF-8 source under 8 MiB")
	}
	root, err := s.open(ctx, input.FileRequest)
	if err != nil {
		return File{}, err
	}
	defer root.Close()
	s.mu.Lock()
	defer s.mu.Unlock()
	current, err := read(root, input.FileRequest)
	if err != nil {
		return File{}, err
	}
	if current.Revision != input.Revision {
		return File{}, ErrConflict
	}
	info, err := root.Lstat(input.Path)
	if err != nil {
		return File{}, err
	}
	if !info.Mode().IsRegular() {
		return File{}, errors.New("saving a symbolic link is not supported")
	}
	temp := filepath.Join(filepath.Dir(input.Path), ".gavel-save-"+uuid.NewString())
	file, err := root.OpenFile(temp, os.O_CREATE|os.O_EXCL|os.O_WRONLY, info.Mode().Perm())
	if err != nil {
		return File{}, err
	}
	defer func() {
		if err := root.Remove(temp); err != nil && !errors.Is(err, fs.ErrNotExist) {
			resultErr = errors.Join(resultErr, err)
		}
	}()
	_, writeErr := file.WriteString(input.Source)
	if err := errors.Join(writeErr, file.Sync(), file.Close()); err != nil {
		return File{}, err
	}
	latest, err := read(root, input.FileRequest)
	if err != nil {
		return File{}, err
	}
	if latest.Revision != input.Revision {
		return File{}, ErrConflict
	}
	if err := root.Rename(temp, input.Path); err != nil {
		return File{}, err
	}
	return File{FileRequest: input.FileRequest, Source: input.Source, Revision: revision([]byte(input.Source))}, nil
}

func (s *Service) Entries(ctx context.Context, input EntriesRequest) ([]DirectoryEntry, error) {
	if input.Path == "" {
		input.Path = "."
	}
	root, err := s.open(ctx, input.FileRequest)
	if err != nil {
		return nil, err
	}
	defer root.Close()
	dir, err := root.Open(input.Path)
	if err != nil {
		return nil, err
	}
	defer dir.Close()
	children, err := dir.ReadDir(-1)
	if err != nil {
		return nil, err
	}
	entries := make([]DirectoryEntry, 0, len(children))
	paths := make([]git.VisibilityPath, 0, len(children))
	for _, child := range children {
		path := filepath.ToSlash(filepath.Join(input.Path, child.Name()))
		if !validPath(path) || child.Type()&os.ModeSymlink != 0 {
			continue
		}
		if !child.IsDir() && !markdown(path) {
			continue
		}
		entries = append(entries, DirectoryEntry{Path: path, Name: child.Name(), Directory: child.IsDir()})
		paths = append(paths, git.VisibilityPath{Path: path, Directory: child.IsDir()})
	}
	visibility, err := git.CheckVisibility(ctx, git.VisibilityOptions{Directory: input.Root, Paths: paths, Overrides: input.Overrides, OverrideDirectory: input.OverrideDirectory})
	if err != nil {
		return nil, err
	}
	result := make([]DirectoryEntry, 0, len(entries))
	for index, entry := range entries {
		entry.Visible = visibility[index].Visible
		entry.Rule = visibility[index].Rule
		if entry.Visible || input.IncludeIgnored {
			result = append(result, entry)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Directory != result[j].Directory {
			return result[i].Directory
		}
		return result[i].Name < result[j].Name
	})
	return result, nil
}

func markdown(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".md" || ext == ".markdown" || ext == ".mdx"
}

func (s *Service) Directories(path string) ([]DirectoryEntry, error) {
	if path == "" {
		entries := make([]DirectoryEntry, 0, len(s.options.Roots))
		for _, root := range s.options.Roots {
			entries = append(entries, DirectoryEntry{Path: root, Name: root, Directory: true, Visible: true})
		}
		return entries, nil
	}
	canonical, err := s.options.Authorize(path)
	if err != nil {
		return nil, err
	}
	children, err := os.ReadDir(canonical)
	if err != nil {
		return nil, err
	}
	entries := []DirectoryEntry{}
	for _, child := range children {
		if child.IsDir() && child.Name() != ".git" {
			entries = append(entries, DirectoryEntry{Path: filepath.Join(canonical, child.Name()), Name: child.Name(), Directory: true, Visible: true})
		}
	}
	return entries, nil
}

func (s *Service) Close() error { close(s.stop); <-s.done; return s.watcher.Close() }
