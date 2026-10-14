package linters

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
)

// ScratchDir returns (and creates) a gavel-owned directory for a linter's
// generated reports, merged configs and wrapper scripts. It lives under the
// user cache dir, keyed by the absolute workDir, so gavel's own tool artifacts
// never land in — and dirty — the tree under change.
func ScratchDir(workDir, linter string) (string, error) {
	if workDir == "" {
		return "", fmt.Errorf("%s scratch dir: WorkDir is empty", linter)
	}
	abs, err := filepath.Abs(workDir)
	if err != nil {
		return "", fmt.Errorf("%s scratch dir: resolve %q: %w", linter, workDir, err)
	}
	cache, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("%s scratch dir: %w", linter, err)
	}
	sum := sha256.Sum256([]byte(abs))
	dir := filepath.Join(cache, "gavel", "lint", linter, hex.EncodeToString(sum[:8]))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("%s scratch dir: create %s: %w", linter, dir, err)
	}
	return dir, nil
}
