package launch

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/flanksource/gavel/linters"
)

func retainPatch(root, id, patch string) error {
	if patch == "" {
		return nil
	}
	dir := filepath.Join(root, ".gavel", "artifacts", id)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "patch.diff"), []byte(patch), 0o600)
}

func retainFindingSources(root string, results []*linters.LinterResult) error {
	var failures []error
	for _, result := range results {
		for index := range result.Violations {
			finding := &result.Violations[index]
			if finding.File == "" || finding.Code != nil {
				continue
			}
			base := result.WorkDir
			if base == "" {
				base = root
			}
			path := finding.File
			if !filepath.IsAbs(path) {
				path = filepath.Join(base, path)
			}
			code, err := sourceExcerpt(root, path, finding.Line)
			if err != nil {
				failures = append(failures, fmt.Errorf("retain finding source %s: %w", finding.File, err))
				continue
			}
			finding.Code = &code
		}
	}
	return errors.Join(failures...)
}

func sourceExcerpt(root, path string, line int) (string, error) {
	if line < 0 {
		return "", errors.New("finding line must be nonnegative")
	}
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return "", err
	}
	canonical, err := filepath.EvalSymlinks(root)
	if err != nil {
		return "", err
	}
	relative, err := filepath.Rel(canonical, resolved)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", errors.New("finding source is outside the prepared workspace")
	}
	file, err := os.Open(resolved)
	if err != nil {
		return "", err
	}
	defer file.Close()
	const maxBytes = 1 << 20
	data, err := io.ReadAll(io.LimitReader(file, maxBytes+1))
	if err != nil {
		return "", err
	}
	if len(data) > maxBytes {
		return "", errors.New("finding source exceeds the 1 MiB capture limit")
	}
	lines := strings.Split(string(data), "\n")
	start, end := max(0, line-4), min(len(lines), line+3)
	if start >= len(lines) {
		return "", fmt.Errorf("finding line %d exceeds the source length", line)
	}
	return strings.Join(lines[start:end], "\n"), nil
}
