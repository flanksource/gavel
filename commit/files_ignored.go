package commit

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/flanksource/clicky/prompt"
)

var ErrIgnoredPathsCancelled = errors.New("commit cancelled: ignored file arguments were not approved")

func explicitIgnoredPaths(workDir string, files []string) (map[string]struct{}, error) {
	var paths []string
	for _, file := range files {
		// Git pathspecs without a literal filesystem entry are handled by git add.
		if _, err := os.Lstat(filepath.Join(workDir, file)); os.IsNotExist(err) {
			continue
		} else if err != nil {
			return nil, fmt.Errorf("inspect file argument %q: %w", file, err)
		}
		paths = append(paths, file)
	}
	return gitCheckIgnore(workDir, paths)
}

func approveIgnoredPaths(ctx context.Context, workDir string, ignored map[string]struct{}) error {
	paths := make([]string, 0, len(ignored))
	for file := range ignored {
		paths = append(paths, file)
	}
	tracked, err := trackedFiles(workDir, paths)
	if err != nil {
		return err
	}
	var pending []string
	for _, file := range paths {
		if _, ok := tracked[file]; !ok {
			pending = append(pending, file)
		}
	}
	if len(pending) == 0 {
		return nil
	}
	sort.Strings(pending)
	if !stdinIsTerminal() && !prompt.HasInteractiveSink() {
		return fmt.Errorf("ignored file arguments require approval: %q; use git add -f -- <paths>, then gavel commit --stage staged", pending)
	}
	choice, ok := promptSelectIndex(ctx, "These paths are ignored by Git:\n"+strings.Join(pending, "\n"), []string{
		"Force-add and continue", "Cancel commit",
	})
	if !ok || choice != 0 {
		return ErrIgnoredPathsCancelled
	}
	return nil
}
