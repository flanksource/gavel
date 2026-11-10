package utils

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// outputDirNames are the directories gavel writes its own output to inside a repository.
var outputDirNames = map[string]bool{".gavel": true, ".tmp": true}

// MkdirOutput creates dir like os.MkdirAll, and makes every .gavel or .tmp directory on its path
// ignore itself with a .gitignore of "*", so gavel's output never shows as untracked to Git.
func MkdirOutput(dir string, perm fs.FileMode) error {
	if err := os.MkdirAll(dir, perm); err != nil {
		return err
	}
	for current := filepath.Clean(dir); ; {
		if outputDirNames[filepath.Base(current)] {
			if err := IgnoreOutputDir(current); err != nil {
				return err
			}
		}
		parent := filepath.Dir(current)
		if parent == current {
			return nil
		}
		current = parent
	}
}

// IgnoreOutputDir writes a .gitignore of "*" in the output directory dir when it has none. A
// missing dir is left alone.
func IgnoreOutputDir(dir string) error {
	info, err := os.Stat(dir)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	} else if err != nil || !info.IsDir() {
		return err
	}
	path := filepath.Join(dir, ".gitignore")
	if _, err := os.Stat(path); !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	if err := os.WriteFile(path, []byte("*\n"), 0o644); err != nil {
		return fmt.Errorf("make %s ignore itself: %w", dir, err)
	}
	return nil
}
