package outline

import (
	"errors"
	"path"
	"strings"

	"github.com/flanksource/gavel/testrunner/history"
	"github.com/flanksource/gavel/testrunner/parsers"
)

type historyKey struct {
	framework parsers.Framework
	pkg       string
	suite     string
	name      string
}

// joinHistory attaches aggregated run history to matching leaf entries.
// History entries are keyed by package path (run snapshots rarely record the
// source file), so each leaf walks its file's ancestor directories until one
// matches: the file's own dir for go packages, the npm package root for
// vitest. Returns the number of recorded runs; a missing .gavel history is
// not an error because the outline is primarily static.
func joinHistory(report *Report, opts Options) (int, error) {
	hist, err := history.Load(history.Options{WorkDir: opts.WorkDir, Paths: opts.Paths})
	if err != nil {
		if errors.Is(err, history.ErrNoHistory) {
			return 0, nil
		}
		return 0, err
	}

	byKey := map[historyKey]*history.Entry{}
	for i := range hist.Tests {
		entry := &hist.Tests[i]
		byKey[historyKey{
			framework: entry.Framework,
			pkg:       cleanRelPath(entry.PackagePath),
			suite:     strings.Join(entry.Suite, "\x00"),
			name:      entry.Name,
		}] = entry
	}

	for _, leaf := range report.Leaves() {
		if leaf.Dynamic {
			continue // dynamic names never match recorded runs
		}
		key := historyKey{
			framework: leaf.Framework,
			suite:     strings.Join(leaf.Suite, "\x00"),
			name:      leaf.Name,
		}
		for dir := cleanRelPath(path.Dir(leaf.File)); ; dir = parentDir(dir) {
			key.pkg = dir
			if entry := byKey[key]; entry != nil {
				leaf.History = entry
				break
			}
			if dir == "" {
				break
			}
		}
	}
	return hist.RunCount, nil
}

func cleanRelPath(p string) string {
	p = strings.Trim(strings.TrimPrefix(strings.TrimSpace(p), "./"), "/")
	if p == "." {
		return ""
	}
	return p
}

func parentDir(dir string) string {
	if dir == "" {
		return ""
	}
	return cleanRelPath(path.Dir(dir))
}
