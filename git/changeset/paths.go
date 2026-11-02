package changeset

import (
	"context"
	"os"
	"path/filepath"
)

func selectionRoot(ctx context.Context, dir string, opts Options) (string, error) {
	if opts.Changed || opts.Since != "" || opts.Comparison != nil {
		return gitText(ctx, dir, "rev-parse", "--show-toplevel")
	}
	root, err := physicalPath(dir)
	if err != nil {
		return "", err
	}
	for candidate := root; ; candidate = filepath.Dir(candidate) {
		if _, err := os.Stat(filepath.Join(candidate, ".git")); err == nil {
			return candidate, nil
		} else if !os.IsNotExist(err) {
			return "", err
		}
		if filepath.Dir(candidate) == candidate {
			return root, nil
		}
	}
}

func physicalPath(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	resolved, err := filepath.EvalSymlinks(abs)
	if err == nil || !os.IsNotExist(err) {
		return resolved, err
	}
	parent, err := physicalPath(filepath.Dir(abs))
	if err != nil {
		return "", err
	}
	return filepath.Join(parent, filepath.Base(abs)), nil
}
