package prompt

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/flanksource/gavel/utils"
)

// fileScope is what one review covers: Paths as given, relative to the work
// dir or absolute. Exact means every path is a file, so a violation must name
// one of them; with a directory in scope it must sit under one of the paths.
type fileScope struct {
	Paths []string
	Exact bool
}

// newFileScope classifies paths under dir. A path that does not exist, such as
// a deleted changed file, is still a file name and stays enumerable.
func newFileScope(dir string, paths []string) (fileScope, error) {
	scope := fileScope{Paths: paths, Exact: len(paths) > 0}
	if len(paths) == 0 {
		return scope, nil
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return fileScope{}, err
	}
	root, err := os.OpenRoot(abs)
	if err != nil {
		return fileScope{}, err
	}
	defer root.Close()
	for _, path := range paths {
		path, err := utils.ResolveWithin(abs, path)
		if err != nil {
			return fileScope{}, err
		}
		path, err = filepath.Rel(abs, path)
		if err != nil {
			return fileScope{}, err
		}
		info, err := root.Stat(path)
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			return fileScope{}, fmt.Errorf("prompt lint path: %w", err)
		}
		if err == nil && info.IsDir() {
			scope.Exact = false
		}
	}
	return scope, nil
}

// fileProperty is report_violation's file schema: the files under review as
// an enum when the scope is exact.
func (s fileScope) fileProperty() map[string]any {
	if !s.Exact {
		return map[string]any{"type": "string", "description": "Path relative to the work dir"}
	}
	return map[string]any{"type": "string", "enum": s.Paths, "description": "The file under review the finding is in"}
}

// check rejects a violation file, already resolved against dir, that is
// outside the scope. Without a path scope every file is in scope.
func (s fileScope) check(dir, file string) error {
	if len(s.Paths) == 0 {
		return nil
	}
	if file == "" {
		return fmt.Errorf("violation requires a file, one of %s", strings.Join(s.Paths, ", "))
	}
	for _, path := range s.Paths {
		path = scopePath(dir, path)
		if file == path || !s.Exact && strings.HasPrefix(file, path+string(filepath.Separator)) {
			return nil
		}
	}
	return fmt.Errorf("file %s is not under review; want one of %s", file, strings.Join(s.Paths, ", "))
}

func scopePath(dir, path string) string {
	if filepath.IsAbs(path) {
		return filepath.Clean(path)
	}
	return filepath.Join(dir, path)
}
