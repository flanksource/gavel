package shims

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// LookPath is exec.LookPath that never resolves to a gavel shim, so gavel and
// the shims themselves always reach the real tool.
func LookPath(name string) (string, error) {
	if strings.ContainsRune(name, filepath.Separator) {
		path, err := exec.LookPath(name)
		if err != nil {
			return "", err
		}
		if IsShim(path) {
			return "", fmt.Errorf("%s is a gavel shim, not the real executable", path)
		}
		return path, nil
	}
	shadowed := false
	for _, dir := range filepath.SplitList(os.Getenv("PATH")) {
		if !filepath.IsAbs(dir) {
			continue
		}
		path, err := exec.LookPath(filepath.Join(dir, name))
		if err != nil {
			continue
		}
		if IsShim(path) {
			shadowed = true
			continue
		}
		return path, nil
	}
	if shadowed {
		return "", fmt.Errorf("real %s not found on PATH (only gavel shims): %w", name, exec.ErrNotFound)
	}
	return "", fmt.Errorf("%s not found on PATH: %w", name, exec.ErrNotFound)
}
