package changeset

import (
	"context"
	"os"
	"path/filepath"
	"sort"
)

type Options struct {
	Dir         string
	Changed     bool
	Since       string
	Files       []string
	Comparison  *Comparison
	Providers   []Provider
	Environment Environment
}

func (o Options) Active() bool {
	return o.Changed || o.Since != "" || len(o.Files) > 0 || o.Comparison != nil
}

type Comparison struct {
	Provider string `json:"provider,omitempty"`
	Kind     string `json:"kind"`
	Base     string `json:"base,omitempty"`
	Head     string `json:"head,omitempty"`
}

type Change struct {
	Path         string `json:"path"`
	PreviousPath string `json:"previous_path,omitempty"`
	Status       string `json:"status"`
}

type ChangeSet struct {
	Root     string   `json:"root"`
	Mode     string   `json:"mode"`
	Provider string   `json:"provider,omitempty"`
	Base     string   `json:"base,omitempty"`
	Head     string   `json:"head,omitempty"`
	Changes  []Change `json:"files"`
}

func (s *ChangeSet) Paths(dir string) []string {
	dir, err := physicalPath(dir)
	if err != nil {
		panic(err)
	}
	paths := map[string]bool{}
	for _, change := range s.Changes {
		for _, path := range []string{change.Path, change.PreviousPath} {
			if path == "" {
				continue
			}
			rel, err := filepath.Rel(dir, filepath.Join(s.Root, filepath.FromSlash(path)))
			if err != nil {
				panic(err)
			}
			paths[filepath.ToSlash(rel)] = true
		}
	}
	result := make([]string, 0, len(paths))
	for path := range paths {
		result = append(result, path)
	}
	sort.Strings(result)
	return result
}

func (s *ChangeSet) CurrentPaths() ([]string, error) {
	var paths []string
	for _, path := range s.Paths(s.Root) {
		_, err := os.Lstat(filepath.Join(s.Root, filepath.FromSlash(path)))
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, err
		}
		paths = append(paths, path)
	}
	return paths, nil
}

type Environment struct {
	Lookup   func(string) string
	ReadFile func(string) ([]byte, error)
}

type Provider interface {
	Name() string
	Detect(Environment) bool
	Comparison(context.Context, Environment) (Comparison, error)
}
