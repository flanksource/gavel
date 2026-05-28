package commit

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/flanksource/commons/logger"
	"golang.org/x/mod/modfile"
)

// goModUpgrade describes a single replace -> require upgrade applied to a
// go.mod file. It is reused for the pre-commit rewrite (drop replace, pin
// require) and the post-commit restore (re-add replace, leave require).
type goModUpgrade struct {
	OldPath    string // module path, e.g. "github.com/flanksource/foo"
	OldVersion string // Replace.Old.Version; "" for unpinned replaces (common case)
	OldTarget  string // local filesystem target, e.g. "../foo"
	NewVersion string // tagged version to pin, e.g. "v1.4.2"
}

// pendingRestore is work performed AFTER `git commit` succeeds: the committed
// snapshot of GoModFile has the replace directives dropped and require
// pinned; restoreLocalReplaces puts the local replaces back as an unstaged
// edit so the developer keeps building against their local checkout.
type pendingRestore struct {
	GoModFile string // path relative to git root
	Replaces  []goModUpgrade
}

// applyGoModReplaceUpgrade rewrites the working-tree go.mod at gitRoot/file:
// drops each replace listed in ups and pins/adds the corresponding require to
// the new version. Then runs `go mod tidy` in the module directory and
// re-stages go.mod (plus go.sum if it exists).
func applyGoModReplaceUpgrade(workDir, gitRoot, file string, ups []goModUpgrade) error {
	if len(ups) == 0 {
		return nil
	}
	absGoMod := filepath.Join(gitRoot, file)
	data, err := os.ReadFile(absGoMod)
	if err != nil {
		return fmt.Errorf("read %s: %w", absGoMod, err)
	}
	info, err := os.Stat(absGoMod)
	if err != nil {
		return fmt.Errorf("stat %s: %w", absGoMod, err)
	}
	f, err := modfile.Parse(file, data, nil)
	if err != nil {
		return fmt.Errorf("parse %s: %w", absGoMod, err)
	}
	for _, u := range ups {
		if err := f.DropReplace(u.OldPath, u.OldVersion); err != nil {
			return fmt.Errorf("drop replace %s %s: %w", u.OldPath, u.OldVersion, err)
		}
		if err := f.AddRequire(u.OldPath, u.NewVersion); err != nil {
			return fmt.Errorf("add require %s %s: %w", u.OldPath, u.NewVersion, err)
		}
	}
	f.Cleanup()
	out, err := f.Format()
	if err != nil {
		return fmt.Errorf("format %s: %w", absGoMod, err)
	}
	if err := os.WriteFile(absGoMod, out, info.Mode().Perm()); err != nil {
		return fmt.Errorf("write %s: %w", absGoMod, err)
	}

	modDir := filepath.Dir(absGoMod)
	if err := runGoModTidy(modDir); err != nil {
		return fmt.Errorf("tidy %s: %w", modDir, err)
	}

	toStage := []string{file}
	goSumRel := filepath.Join(filepath.Dir(file), "go.sum")
	if _, err := os.Stat(filepath.Join(gitRoot, goSumRel)); err == nil {
		toStage = append(toStage, goSumRel)
	}
	if err := addFiles(workDir, toStage); err != nil {
		return fmt.Errorf("stage %v: %w", toStage, err)
	}
	return nil
}

// restoreLocalReplaces re-adds the local replace directives listed in
// restores to the working-tree go.mod files. It runs AFTER the commit
// subprocess succeeds — the committed snapshot is already in git history;
// this only touches the working tree, intentionally leaving the resulting
// edits unstaged so the developer's next `git status` shows the dev-mode
// local replaces.
//
// Failures are logged as warnings and do not propagate to the caller: the
// commit already happened, so surfacing an error wouldn't undo it. The
// warning names the exact replace directive the user must re-add manually.
func restoreLocalReplaces(gitRoot string, restores []pendingRestore) {
	if len(restores) == 0 {
		return
	}
	for _, r := range restores {
		if err := restoreOneGoMod(gitRoot, r); err != nil {
			for _, u := range r.Replaces {
				logger.Warnf("could not restore local replace for %s (%s): %v — re-add manually: replace %s => %s",
					u.OldPath, r.GoModFile, err, u.OldPath, u.OldTarget)
			}
		}
	}
}

func restoreOneGoMod(gitRoot string, r pendingRestore) error {
	absGoMod := filepath.Join(gitRoot, r.GoModFile)
	data, err := os.ReadFile(absGoMod)
	if err != nil {
		return fmt.Errorf("read %s: %w", absGoMod, err)
	}
	info, err := os.Stat(absGoMod)
	if err != nil {
		return fmt.Errorf("stat %s: %w", absGoMod, err)
	}
	f, err := modfile.Parse(r.GoModFile, data, nil)
	if err != nil {
		return fmt.Errorf("parse %s: %w", absGoMod, err)
	}
	for _, u := range r.Replaces {
		if err := f.AddReplace(u.OldPath, u.OldVersion, u.OldTarget, ""); err != nil {
			return fmt.Errorf("add replace %s => %s: %w", u.OldPath, u.OldTarget, err)
		}
	}
	f.Cleanup()
	out, err := f.Format()
	if err != nil {
		return fmt.Errorf("format %s: %w", absGoMod, err)
	}
	if err := os.WriteFile(absGoMod, out, info.Mode().Perm()); err != nil {
		return fmt.Errorf("write %s: %w", absGoMod, err)
	}

	modDir := filepath.Dir(absGoMod)
	if err := runGoModTidy(modDir); err != nil {
		// Non-fatal here: go.mod has the replace back, go.sum may be slightly
		// stale until the user's next build/tidy. Logged for visibility.
		logger.Warnf("go mod tidy in %s after restore: %v — working-tree go.sum may be stale", modDir, err)
	}
	return nil
}
