package editor

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/flanksource/clicky/sse"
	"github.com/flanksource/gavel/lint"
	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/snapshots"
	"github.com/flanksource/gavel/testrunner/parsers"
	"github.com/flanksource/gavel/workspace"
)

// lintStreamInterval is how often the problems stream rechecks for edits that
// make the shown problems stale; reloads and runs wake it at once.
var lintStreamInterval = 2 * time.Second

// stateRoots groups the canonical paths of paths (file id → path) by the
// directory holding their .gavel state.
func stateRoots(paths map[string]string) (map[string][]string, error) {
	byDir := map[string]string{}
	roots := map[string][]string{}
	for _, path := range paths {
		dir := filepath.Dir(path)
		root, ok := byDir[dir]
		if !ok {
			dirs, err := workspace.Resolve(dir)
			if err != nil {
				return nil, err
			}
			root, byDir[dir] = dirs.StateRoot, dirs.StateRoot
		}
		roots[root] = append(roots[root], path)
	}
	for root := range roots {
		sort.Strings(roots[root])
	}
	return roots, nil
}

// Reload replaces the shown problems with those the workspaces' recorded runs
// hold for the edited files, applying the current .gavel.yaml ignore rules.
// A file changed on disk after the run that checked it ended is stale.
func (s *Server) Reload() error {
	paths := s.Store.Paths()
	roots, err := stateRoots(paths)
	if err != nil {
		return err
	}
	run := LintRun{Revisions: map[string]string{}}
	checked := map[string]time.Time{}
	for root, files := range roots {
		problems, err := snapshots.CurrentProblems(root, files)
		if err != nil {
			return fmt.Errorf("load recorded problems from %s: %w", filepath.Join(root, workspace.StateDirName), err)
		}
		if err := lint.FilterConfiguredViolations(problems.Lint, root); err != nil {
			return err
		}
		run.Results = append(run.Results, problems.Lint...)
		run.Tests = append(run.Tests, problems.Tests...)
		for file, ended := range problems.LintEnded {
			checked[file] = ended
		}
		// A file is as fresh as the older of the runs that checked it.
		for _, test := range problems.Tests {
			if ended, ok := checked[test.File]; !ok || problems.TestsEnded.Before(ended) {
				checked[test.File] = problems.TestsEnded
			}
		}
		if latest := latestOf(problems); !latest.IsZero() && (run.Recorded == nil || latest.After(*run.Recorded)) {
			run.Recorded = &latest
		}
	}
	staleAtLoad := map[string]bool{}
	for _, file := range s.Store.Files() {
		if file.Kind != KindDocument {
			continue
		}
		doc, err := s.Store.Read(file.ID)
		if err != nil {
			return err
		}
		run.Revisions[file.ID] = doc.Revision
		info, err := os.Stat(paths[file.ID])
		if err != nil {
			return err
		}
		if ended, ok := checked[paths[file.ID]]; ok && info.ModTime().After(ended) {
			staleAtLoad[file.ID] = true
		}
	}
	openedAs(s.Store, run.Results, run.Tests)
	s.mu.Lock()
	run.Running, run.Error, run.Active = s.run.Running, s.run.Error, s.run.Active
	s.run, s.staleAtLoad, s.stateRoots = run, staleAtLoad, slices.Sorted(maps.Keys(roots))
	s.mu.Unlock()
	s.updated.Notify()
	return nil
}

// openedAs rewrites the symlink-resolved paths of recorded problems to the
// paths the files were opened through (e.g. a symlinked checkout), so the
// UI places each finding on the file it shows rather than under a second root.
func openedAs(store *Store, results []*linters.LinterResult, tests []parsers.Test) {
	canonical := store.Paths()
	type alias struct{ real, opened string }
	var aliases []alias
	for _, file := range store.Files() {
		real, opened := canonical[file.ID], file.Path
		// The real and opened paths share their trailing components below the symlink.
		for filepath.Base(real) == filepath.Base(opened) && real != filepath.Dir(real) && opened != filepath.Dir(opened) {
			real, opened = filepath.Dir(real), filepath.Dir(opened)
		}
		if real != opened && !slices.Contains(aliases, alias{real, opened}) {
			aliases = append(aliases, alias{real, opened})
		}
	}
	rewrite := func(path string) string {
		for _, a := range aliases {
			if rest, ok := strings.CutPrefix(path, a.real); ok && (rest == "" || rest[0] == filepath.Separator) {
				return a.opened + rest
			}
		}
		return path
	}
	for _, result := range results {
		result.WorkDir = rewrite(result.WorkDir)
		for i := range result.Violations {
			if filepath.IsAbs(result.Violations[i].File) {
				result.Violations[i].File = rewrite(result.Violations[i].File)
			}
		}
	}
	for i := range tests {
		tests[i].WorkDir, tests[i].File = rewrite(tests[i].WorkDir), rewrite(tests[i].File)
	}
}

func latestOf(problems *snapshots.Problems) time.Time {
	latest := problems.TestsEnded
	for _, ended := range problems.LintEnded {
		if ended.After(latest) {
			latest = ended
		}
	}
	return latest
}

// lintFiles lints paths through the same seam as `gavel lint`, saving each
// state root's run to its .gavel, then reloads the shown problems.
func (s *Server) lintFiles(paths []string) error {
	roots, err := stateRoots(pathSet(paths))
	if err != nil {
		return err
	}
	var errs []error
	for root, files := range roots {
		fields := map[string]bool{"fix": true, "files": true, "changed": true, "since": true, "changed-files": true, "failed": true}
		if s.Schema != "" {
			fields["schema"] = true
		}
		started := time.Now().UTC()
		results, err := lint.Run(s.Context, lint.Options{WorkDir: root, Files: files, Schema: s.Schema, OptionFields: fields, Fix: false})
		if err != nil {
			errs = append(errs, err)
			continue
		}
		snapshot := lint.Snapshot(lint.SnapshotOptions{WorkDir: root, Version: s.Version, Started: started, Files: files, Results: results})
		if _, err := lint.SaveSnapshot(root, snapshot); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(append(errs, s.Reload())...)
}

func pathSet(paths []string) map[string]string {
	set := make(map[string]string, len(paths))
	for _, path := range paths {
		set[path] = path
	}
	return set
}

// lintState is the shown problems with each file's staleness: changed on disk
// after its run, or edited since the problems were loaded.
func (s *Server) lintState() LintRun {
	s.mu.Lock()
	defer s.mu.Unlock()
	run := s.run
	run.StaleFiles = nil
	for id, revision := range run.Revisions {
		doc, err := s.Store.Read(id)
		if s.staleAtLoad[id] || err != nil || doc.Revision != revision {
			run.StaleFiles = append(run.StaleFiles, id)
		}
	}
	sort.Strings(run.StaleFiles)
	run.Stale = len(run.StaleFiles) > 0
	run.Results = slices.Clone(run.Results)
	return run
}

func (s *Server) lintStream(w http.ResponseWriter, r *http.Request) {
	wake, unsubscribe := s.updated.Subscribe()
	defer unsubscribe()
	_ = sse.ServeSnapshot(w, r, sse.SnapshotOptions{
		Interval: lintStreamInterval,
		Wake:     wake,
		Load:     func(context.Context) (any, error) { return s.lintState(), nil },
	})
}
