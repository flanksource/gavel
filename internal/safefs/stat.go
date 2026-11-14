package safefs

import (
	"fmt"
	"os"
	"path/filepath"
)

// LocalPath accepts relative paths and absolute paths contained by directory.
func LocalPath(directory, path string) (string, error) {
	if directory == "" || path == "" {
		return "", fmt.Errorf("path containment requires a workspace and path")
	}
	if filepath.IsAbs(path) {
		root, err := filepath.Abs(directory)
		if err != nil {
			return "", err
		}
		path, err = filepath.Rel(root, path)
		if err != nil {
			return "", err
		}
	}
	if !filepath.IsLocal(path) {
		return "", fmt.Errorf("path %q escapes workspace %q", path, directory)
	}
	return filepath.Clean(path), nil
}

// Stat confines traversal, including symlinks, to directory.
func Stat(directory, path string) (os.FileInfo, error) {
	local, err := LocalPath(directory, path)
	if err != nil {
		return nil, err
	}
	root, err := os.OpenRoot(directory)
	if err != nil {
		return nil, err
	}
	defer root.Close()
	info, err := root.Stat(local)
	if err != nil {
		return nil, fmt.Errorf("inspect %q in workspace %q: %w", path, directory, err)
	}
	return info, nil
}
