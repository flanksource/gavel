package changeset

import (
	"context"
	"fmt"
	"path/filepath"
	"slices"
	"sort"
	"sync"
)

type Selection struct {
	options Options
	mu      sync.Mutex
	sets    map[string]*ChangeSet
}

func NewSelection(opts Options) *Selection {
	opts.Files = slices.Clone(opts.Files)
	return &Selection{options: opts, sets: map[string]*ChangeSet{}}
}

func (s *Selection) Fresh() *Selection {
	if s == nil {
		return nil
	}
	return NewSelection(s.options)
}

func (s *Selection) Active() bool { return s != nil && s.options.Active() }

func (s *Selection) Resolve(ctx context.Context, dir string) (*ChangeSet, error) {
	if !s.Active() {
		return nil, nil
	}
	root, err := selectionRoot(ctx, dir, s.options)
	if err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if selected, exists := s.sets[root]; exists {
		return selected, nil
	}
	opts := s.options
	opts.Dir = root
	opts.Files = make([]string, 0, len(s.options.Files))
	for _, file := range s.options.Files {
		if !filepath.IsAbs(file) {
			file, err = filepath.Abs(filepath.Join(s.options.Dir, file))
			if err != nil {
				return nil, err
			}
		}
		opts.Files = append(opts.Files, file)
	}
	selected, err := Resolve(ctx, opts)
	if err != nil {
		return nil, err
	}
	s.sets[root] = selected
	return selected, nil
}

func (s *Selection) Snapshot() []*ChangeSet {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	sets := make([]*ChangeSet, 0, len(s.sets))
	for _, selected := range s.sets {
		sets = append(sets, selected)
	}
	sort.Slice(sets, func(i, j int) bool { return sets[i].Root < sets[j].Root })
	return sets
}

func (s *ChangeSet) Summary() string {
	provider := s.Provider
	if provider == "" {
		provider = "local"
	}
	return fmt.Sprintf("Changes: %s %s base=%s head=%s files=%d root=%s", provider, s.Mode, s.Base, s.Head, len(s.Paths(s.Root)), s.Root)
}
