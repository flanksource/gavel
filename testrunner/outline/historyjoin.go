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
	file      string
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
		key := historyKey{
			framework: entry.Framework,
			pkg:       cleanRelPath(entry.PackagePath),
			file:      cleanRelPath(entry.File),
			suite:     strings.Join(entry.Suite, "\x00"),
			name:      entry.Name,
		}
		if key.file != "" {
			exactKey := key
			exactKey.pkg = ""
			byKey[exactKey] = entry
		}
		key.file = ""
		byKey[key] = entry
	}

	for _, leaf := range report.Leaves() {
		if leaf.Dynamic {
			continue // dynamic names never match recorded runs
		}
		key := historyKey{
			framework: leaf.Framework,
			file:      cleanRelPath(leaf.File),
			suite:     strings.Join(leaf.Suite, "\x00"),
			name:      leaf.Name,
		}
		if key.file != "" {
			if entry := byKey[key]; entry != nil {
				leaf.History = entry
				continue
			}
		}
		key.file = ""
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
