package commit

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/utils"
)

// findGoModRoots is swappable so tests can skip the .gitignore-respecting
// filesystem walk and feed a fixed list of module directories.
var findGoModRoots = func(root string) []string {
	return utils.FindAllProjectRoots(root, []string{"go.mod"})
}

// applyGoModTidy runs `go mod tidy` in every Go module in the repo and stages
// any go.mod / go.sum updates into the in-flight commit. On by default; opt
// out via opts.TidyFlag ("true"/"false") or opts.Config.Tidy.Enabled. Hard
// failure aborts the commit so a broken go.sum can't slip through. ctx is
// kept for signature parity with the other apply* helpers.
func applyGoModTidy(ctx context.Context, opts Options, source stagedSource) (stagedSource, error) {
	_ = ctx
	if !tidyEnabled(opts) {
		return source, nil
	}

	roots := findGoModRoots(opts.WorkDir)
	if len(roots) == 0 {
		return source, nil
	}

	var toStage []string
	for _, modDir := range roots {
		displayDir, err := filepath.Rel(opts.WorkDir, modDir)
		if err != nil || displayDir == "" {
			displayDir = modDir
		}

		modPath := filepath.Join(modDir, "go.mod")
		sumPath := filepath.Join(modDir, "go.sum")

		beforeMod, beforeModExists, err := hashFile(modPath)
		if err != nil {
			return source, fmt.Errorf("hash go.mod in %s: %w", displayDir, err)
		}
		beforeSum, beforeSumExists, err := hashFile(sumPath)
		if err != nil {
			return source, fmt.Errorf("hash go.sum in %s: %w", displayDir, err)
		}

		if err := runGoModTidy(modDir); err != nil {
			return source, fmt.Errorf("go mod tidy in %s: %w", displayDir, err)
		}

		afterMod, afterModExists, err := hashFile(modPath)
		if err != nil {
			return source, fmt.Errorf("hash go.mod in %s: %w", displayDir, err)
		}
		afterSum, afterSumExists, err := hashFile(sumPath)
		if err != nil {
			return source, fmt.Errorf("hash go.sum in %s: %w", displayDir, err)
		}

		var changed []string
		if fileChanged(beforeMod, beforeModExists, afterMod, afterModExists) {
			changed = append(changed, "go.mod")
			toStage = append(toStage, modPath)
		}
		if fileChanged(beforeSum, beforeSumExists, afterSum, afterSumExists) {
			changed = append(changed, "go.sum")
			toStage = append(toStage, sumPath)
		}
		if len(changed) > 0 {
			logger.Infof("go mod tidy: updated %s/%s", displayDir, strings.Join(changed, ", "))
		}
	}

	if len(toStage) == 0 {
		return source, nil
	}

	if err := addFiles(opts.WorkDir, toStage); err != nil {
		return source, fmt.Errorf("stage tidy updates: %w", err)
	}
	refreshed, err := readStagedSource(opts.WorkDir)
	if err != nil {
		return source, fmt.Errorf("re-read staged source after go mod tidy: %w", err)
	}
	refreshed.PendingRestores = append(refreshed.PendingRestores, source.PendingRestores...)
	return refreshed, nil
}

// tidyEnabled resolves the on/off state. CLI flag wins over config; config
// nil = on (default).
func tidyEnabled(opts Options) bool {
	switch strings.ToLower(strings.TrimSpace(opts.TidyFlag)) {
	case "true":
		return true
	case "false":
		return false
	}
	if opts.Config.Tidy.Enabled != nil {
		return *opts.Config.Tidy.Enabled
	}
	return true
}

// hashFile returns the sha256 of the file at path. The second return is false
// when the file does not exist (a non-error case for go.sum, which may be
// absent before or after tidy).
func hashFile(path string) ([]byte, bool, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return nil, false, err
	}
	return h.Sum(nil), true, nil
}

// fileChanged reports whether a file's content or existence flipped between
// before and after snapshots.
func fileChanged(before []byte, beforeExists bool, after []byte, afterExists bool) bool {
	if beforeExists != afterExists {
		return true
	}
	if !beforeExists {
		return false
	}
	if len(before) != len(after) {
		return true
	}
	for i := range before {
		if before[i] != after[i] {
			return true
		}
	}
	return false
}
