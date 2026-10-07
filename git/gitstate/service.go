package gitstate

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/flanksource/commons/logger"
	"golang.org/x/sync/singleflight"
)

const (
	DefaultFresh    = 10 * time.Second
	DefaultMaxStale = 2 * time.Minute
)

// Options configures a Service. Zero fields take the defaults: Compute,
// DefaultFresh, DefaultMaxStale, time.Now and context.Background.
type Options struct {
	Compute func(ctx context.Context, dir string) (State, error)
	// Fresh is how long a state is served without recomputing it.
	Fresh time.Duration
	// MaxStale is how long a state is served while a background refresh runs;
	// past it, Get blocks on a new computation.
	MaxStale time.Duration
	Now      func() time.Time
	// Context bounds every computation. It is the process's root context, never
	// a request's: a computation is shared by every caller waiting on it, so one
	// canceled request must not fail it for the rest.
	Context context.Context
}

// Service memoizes State per project directory. Concurrent callers share one
// computation, a state younger than Fresh is served as is, and one younger than
// MaxStale is served while a single background refresh replaces it. Errors are
// never cached. Server-side git mutations call Invalidate so the next read
// recomputes instead of waiting out the freshness window.
type Service struct {
	compute         func(ctx context.Context, dir string) (State, error)
	fresh, maxStale time.Duration
	now             func() time.Time
	ctx             context.Context

	group singleflight.Group

	mu      sync.Mutex
	entries map[string]State
	// refreshing holds the keys with a background refresh in flight, so a burst
	// of stale reads starts one refresh rather than one each.
	refreshing map[string]bool
	// generation advances on every Invalidate; a computation started under an
	// older generation still answers its callers but is not stored.
	generation uint64
}

func New(opts Options) *Service {
	s := &Service{
		compute: opts.Compute, fresh: opts.Fresh, maxStale: opts.MaxStale, now: opts.Now, ctx: opts.Context,
		entries: map[string]State{}, refreshing: map[string]bool{},
	}
	if s.compute == nil {
		s.compute = Compute
	}
	if s.fresh == 0 {
		s.fresh = DefaultFresh
	}
	if s.maxStale == 0 {
		s.maxStale = DefaultMaxStale
	}
	if s.now == nil {
		s.now = time.Now
	}
	if s.ctx == nil {
		s.ctx = context.Background()
	}
	return s
}

// Get returns the git state of the repository at dir, memoized.
func (s *Service) Get(ctx context.Context, dir string) (State, error) {
	key := filepath.Clean(dir)
	s.mu.Lock()
	state, ok := s.entries[key]
	generation := s.generation
	age := s.now().Sub(state.ComputedAt)
	if ok && age >= s.fresh && age < s.maxStale && !s.refreshing[key] {
		s.refreshing[key] = true
		go s.refresh(key, generation)
	}
	s.mu.Unlock()
	if ok && age < s.maxStale {
		return state, nil
	}
	return s.load(ctx, key, generation)
}

// load waits for the shared computation of key under generation, returning
// early when ctx ends; the computation itself runs on, for its other waiters.
func (s *Service) load(ctx context.Context, key string, generation uint64) (State, error) {
	result := s.group.DoChan(flightKey(key, generation), func() (any, error) {
		return s.computeAndStore(key, generation)
	})
	select {
	case <-ctx.Done():
		return State{}, ctx.Err()
	case res := <-result:
		if res.Err != nil {
			return State{}, res.Err
		}
		return res.Val.(State), nil
	}
}

func (s *Service) refresh(key string, generation uint64) {
	defer func() {
		s.mu.Lock()
		delete(s.refreshing, key)
		s.mu.Unlock()
	}()
	if _, err := s.load(s.ctx, key, generation); err != nil && !errors.Is(err, context.Canceled) {
		logger.Warnf("refresh git state of %s: %v", key, err)
	}
}

func (s *Service) computeAndStore(key string, generation uint64) (State, error) {
	// A caller that missed the cache may reach the flight only after an earlier
	// one stored its result; it reuses that state instead of computing again.
	s.mu.Lock()
	state, ok := s.entries[key]
	s.mu.Unlock()
	if ok && s.now().Sub(state.ComputedAt) < s.fresh {
		return state, nil
	}
	started := s.now()
	state, err := s.compute(s.ctx, key)
	if err != nil {
		return State{}, err
	}
	state.ComputedAt = started.UTC()
	s.mu.Lock()
	if s.generation == generation {
		s.entries[key] = state
	}
	s.mu.Unlock()
	return state, nil
}

// flightKey scopes a computation to the generation it started under, so a Get
// after an Invalidate never joins a computation that predates it.
func flightKey(key string, generation uint64) string {
	return key + "\x00" + strconv.FormatUint(generation, 10)
}

// Invalidate drops the state of the project at path, or of the project one of
// whose worktrees is at path, and discards every computation in flight.
func (s *Service) Invalidate(path string) {
	paths := []string{filepath.Clean(path)}
	if resolved, err := filepath.EvalSymlinks(path); err == nil && resolved != paths[0] {
		paths = append(paths, resolved)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.generation++
	for key, state := range s.entries {
		if state.owns(key, paths) {
			delete(s.entries, key)
		}
	}
}

func (state State) owns(key string, paths []string) bool {
	for _, path := range paths {
		if key == path {
			return true
		}
		for _, wt := range state.Worktrees {
			if filepath.Clean(wt.Path) == path {
				return true
			}
		}
	}
	return false
}

// WorktreeOf finds the live worktree of the repository at dir whose resolved
// path is path. A path missing from the memoized state may be a worktree
// created since it was computed, so the state is re-read once before path is
// reported as no worktree.
func (s *Service) WorktreeOf(ctx context.Context, dir, path string) (Worktree, bool, error) {
	for attempt := 0; ; attempt++ {
		state, err := s.Get(ctx, dir)
		if err != nil {
			return Worktree{}, false, err
		}
		wt, found, err := state.worktreeAt(path)
		if err != nil || found || attempt > 0 {
			return wt, found, err
		}
		s.Invalidate(dir)
	}
}

func (state State) worktreeAt(path string) (Worktree, bool, error) {
	for _, wt := range state.Worktrees {
		if wt.Prunable {
			continue
		}
		resolved, err := filepath.EvalSymlinks(wt.Path)
		if errors.Is(err, os.ErrNotExist) {
			// Removed since the state was computed, so it cannot be path, which exists.
			continue
		}
		if err != nil {
			return Worktree{}, false, fmt.Errorf("resolve worktree %s: %w", wt.Path, err)
		}
		if resolved == path {
			return wt, true, nil
		}
	}
	return Worktree{}, false, nil
}
