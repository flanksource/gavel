package ui

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/flanksource/gavel/status"
)

// projectActionArgs builds the gavel command for request, run in dir.
func (s *Server) projectActionArgs(dir string, request projectActionRequest) ([]string, error) {
	if request.Options != nil {
		if len(request.Files) > 0 {
			return nil, errors.New("advanced project actions cannot include top-level files")
		}
		if request.Action != projectActionCommit && request.Action != projectActionLint && request.Action != projectActionTest {
			return nil, fmt.Errorf("advanced options are not supported for %s", request.Action)
		}
		if s.projectActionOptionsProvider == nil {
			return nil, errors.New("project action options provider is not configured")
		}
		positional := "paths"
		if request.Action != projectActionTest {
			positional = "files"
		}
		if raw, present := request.Options[positional]; present {
			paths, err := projectActionOptionPaths(raw)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", positional, err)
			}
			if request.Action == projectActionCommit {
				if len(paths) == 0 {
					return nil, errors.New("commit requires at least one selected file")
				}
				if err := validateProjectActionFiles(dir, paths); err != nil {
					return nil, err
				}
			} else if err := validateProjectOptionPaths(dir, paths); err != nil {
				return nil, err
			}
		} else if request.Action == projectActionCommit {
			return nil, errors.New("commit requires at least one selected file")
		}
		advanced, err := s.projectActionOptionsProvider.Args(string(request.Action), request.Options)
		if err != nil {
			return nil, err
		}
		base := []string{string(request.Action), "--work-dir", dir}
		if request.Action == projectActionCommit {
			base = append(base, "--precommit=fail")
		}
		return append(base, advanced...), nil
	}
	switch request.Action {
	case projectActionCommit:
		if len(request.Files) == 0 {
			return nil, errors.New("commit requires at least one selected file")
		}
		if err := validateProjectActionFiles(dir, request.Files); err != nil {
			return nil, err
		}
		return append([]string{"commit", "--work-dir", dir, "--precommit=fail"}, request.Files...), nil
	case projectActionLint:
		if len(request.Files) > 0 {
			if err := validateProjectActionFiles(dir, request.Files); err != nil {
				return nil, err
			}
		}
		return append([]string{"lint", "--work-dir", dir}, request.Files...), nil
	case projectActionTest:
		if len(request.Files) > 0 {
			return nil, errors.New("test action derives its scope from current changes; files must be empty")
		}
		return []string{"test", "--work-dir", dir, "--changed"}, nil
	default:
		return nil, fmt.Errorf("unknown project action %q", request.Action)
	}
}

func projectActionOptionPaths(value any) ([]string, error) {
	values, ok := value.([]any)
	if !ok {
		return nil, fmt.Errorf("expected an array, got %T", value)
	}
	paths := make([]string, 0, len(values))
	for _, value := range values {
		path, ok := value.(string)
		if !ok {
			return nil, fmt.Errorf("expected a string path, got %T", value)
		}
		paths = append(paths, path)
	}
	return paths, nil
}

func validateProjectOptionPaths(dir string, paths []string) error {
	for _, path := range paths {
		clean := filepath.Clean(path)
		if path == "" || filepath.IsAbs(path) || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
			return fmt.Errorf("project action path %q escapes the project", path)
		}
		absolute, err := filepath.Abs(filepath.Join(dir, clean))
		if err != nil {
			return fmt.Errorf("resolve project action path %q: %w", path, err)
		}
		relative, err := filepath.Rel(dir, absolute)
		if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			return fmt.Errorf("project action path %q escapes the project", path)
		}
	}
	return nil
}

func validateProjectActionFiles(dir string, requested []string) error {
	result, err := gatherProjectStatus(dir, status.Options{NoRepomap: true, NoResults: true})
	if err != nil {
		return fmt.Errorf("gather project status: %w", err)
	}
	available := make(map[string]status.FileState, len(result.Files))
	for _, file := range result.Files {
		available[file.Path] = file.State
	}
	seen := make(map[string]struct{}, len(requested))
	for _, path := range requested {
		state, ok := available[path]
		if !ok {
			return fmt.Errorf("%q is not a current project change", path)
		}
		if state == status.StateConflict {
			return fmt.Errorf("%q has unresolved conflicts", path)
		}
		if _, duplicate := seen[path]; duplicate {
			return fmt.Errorf("duplicate project file %q", path)
		}
		seen[path] = struct{}{}
	}
	return nil
}
