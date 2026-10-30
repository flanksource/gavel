package changegraph

import (
	"path/filepath"
	"sort"
)

type FileSet map[string]struct{}

func NewFileSet() FileSet { return FileSet{} }

func (fs FileSet) Add(path string) {
	if path != "" {
		fs[filepath.ToSlash(path)] = struct{}{}
	}
}

func (fs FileSet) Union(other FileSet) {
	for path := range other {
		fs[path] = struct{}{}
	}
}

func (fs FileSet) Has(path string) bool {
	_, exists := fs[filepath.ToSlash(path)]
	return exists
}

func (fs FileSet) Len() int { return len(fs) }

func (fs FileSet) Sorted() []string {
	paths := make([]string, 0, len(fs))
	for path := range fs {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return paths
}
