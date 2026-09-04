package ui

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"sync"

	"github.com/flanksource/gavel/todos"
)

const todoBatchConcurrency = 4

type todoBatchRequest struct {
	Dirs *[]string `json:"dirs"`
}

type todoBatchResponse struct {
	Results []todoBatchResult `json:"results"`
}

type todoBatchResult struct {
	Dir    string          `json:"dir"`
	Counts *todoCounts     `json:"counts,omitempty"`
	Items  []todoSummary   `json:"items"`
	Error  *todoBatchError `json:"error,omitempty"`
}

type todoBatchError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (s *Server) handleTodoBatch(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var payload todoBatchRequest
	if err := decodeTodoRequest(r, &payload); err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	if payload.Dirs == nil {
		writeTodoError(w, http.StatusBadRequest, fmt.Errorf("dirs is required and must be an array"))
		return
	}
	if err := validateTodoBatchDirs(*payload.Dirs); err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	json.NewEncoder(w).Encode(todoBatchResponse{Results: s.loadTodoBatch(r.Context(), *payload.Dirs)}) //nolint:errcheck
}

func validateTodoBatchDirs(dirs []string) error {
	seen := make(map[string]int, len(dirs))
	for i, dir := range dirs {
		if dir == "" || strings.TrimSpace(dir) != dir {
			return fmt.Errorf("dirs[%d] must be a non-blank normalized workspace path", i)
		}
		if !filepath.IsAbs(dir) {
			return fmt.Errorf("dirs[%d] must be absolute: %q", i, dir)
		}
		if filepath.Clean(dir) != dir {
			return fmt.Errorf("dirs[%d] must be normalized: %q", i, dir)
		}
		if first, ok := seen[dir]; ok {
			return fmt.Errorf("dirs[%d] duplicates dirs[%d]: %q", i, first, dir)
		}
		seen[dir] = i
	}
	return nil
}

func (s *Server) loadTodoBatch(ctx context.Context, dirs []string) []todoBatchResult {
	results := make([]todoBatchResult, len(dirs))
	for i, dir := range dirs {
		results[i].Dir = dir
	}
	if len(dirs) == 0 {
		return results
	}

	workers := min(len(dirs), todoBatchConcurrency)
	jobs := make(chan int)
	var wg sync.WaitGroup
	wg.Add(workers)
	for range workers {
		go func() {
			defer wg.Done()
			for i := range jobs {
				results[i] = s.loadTodoBatchWorkspace(ctx, dirs[i])
			}
		}()
	}

	next := 0
	for next < len(dirs) {
		select {
		case jobs <- next:
			next++
		case <-ctx.Done():
			for i := next; i < len(dirs); i++ {
				results[i].Error = todoBatchErrorFor(ctx.Err())
			}
			next = len(dirs)
		}
	}
	close(jobs)
	wg.Wait()
	return results
}

func (s *Server) loadTodoBatchWorkspace(ctx context.Context, dir string) todoBatchResult {
	result := todoBatchResult{Dir: dir}
	if err := ctx.Err(); err != nil {
		result.Error = todoBatchErrorFor(err)
		return result
	}
	provider, source, err := s.todoProviderContext(ctx, todoSource{Dir: dir})
	if err != nil {
		result.Error = todoBatchErrorFor(err)
		return result
	}
	items, err := provider.List(ctx, todos.DiscoveryFilters{})
	if err != nil {
		result.Error = todoBatchErrorFor(err)
		return result
	}
	if err := ctx.Err(); err != nil {
		result.Error = todoBatchErrorFor(err)
		return result
	}

	counts := summarizeTodos(items)
	result.Counts = &counts
	result.Items = make([]todoSummary, 0, len(items))
	stats := commitDiffStats(ctx, source.Dir)
	if err := ctx.Err(); err != nil {
		return todoBatchResult{Dir: dir, Error: todoBatchErrorFor(err)}
	}
	for _, item := range items {
		summary := summarizeTodo(item, false)
		summary.Diff = diffStatFor(stats, item.ID)
		result.Items = append(result.Items, summary)
	}
	return result
}

func todoBatchErrorFor(err error) *todoBatchError {
	code := "load_failed"
	switch {
	case errors.Is(err, context.Canceled):
		code = "canceled"
	case errors.Is(err, context.DeadlineExceeded):
		code = "deadline_exceeded"
	}
	return &todoBatchError{Code: code, Message: err.Error()}
}
