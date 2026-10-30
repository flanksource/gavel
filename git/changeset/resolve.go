package changeset

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

func Resolve(ctx context.Context, opts Options) (*ChangeSet, error) {
	if !opts.Active() {
		return nil, nil
	}
	dir, err := filepath.Abs(opts.Dir)
	if err != nil {
		return nil, err
	}
	root, err := selectionRoot(ctx, dir, opts)
	if err != nil {
		return nil, err
	}
	env := opts.Environment
	if env.Lookup == nil {
		env.Lookup = os.Getenv
	}
	if env.ReadFile == nil {
		env.ReadFile = os.ReadFile
	}
	selected := &ChangeSet{Root: root, Mode: "files", Changes: []Change{}}
	since := opts.Since
	if since == "" && opts.Changed && opts.Comparison == nil {
		since = env.Lookup("GAVEL_CHANGED_BASE")
	}
	switch {
	case since != "":
		err = selected.fromSelector(ctx, since)
	case opts.Comparison != nil:
		err = selected.fromComparison(ctx, *opts.Comparison)
	case opts.Changed:
		var comparison *Comparison
		comparison, err = detectComparison(ctx, env, opts.Providers)
		if err == nil && comparison != nil {
			err = selected.fromComparison(ctx, *comparison)
		} else if err == nil {
			err = selected.fromSelector(ctx, "HEAD")
		}
	}
	if err != nil {
		return nil, fmt.Errorf("changeset in %s: %w", root, err)
	}
	for _, file := range opts.Files {
		if !filepath.IsAbs(file) {
			file = filepath.Join(dir, file)
		}
		file, err = physicalPath(file)
		if err != nil {
			return nil, err
		}
		rel, err := filepath.Rel(root, file)
		if err != nil {
			return nil, err
		}
		if rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			selected.Changes = append(selected.Changes, Change{Path: filepath.ToSlash(rel), Status: "explicit"})
		}
	}
	selected.normalize()
	return selected, nil
}

func (s *ChangeSet) fromSelector(ctx context.Context, since string) error {
	s.Mode = since
	if since == "HEAD" || since == "staged" {
		head, err := gitText(ctx, s.Root, "rev-parse", "--verify", "--quiet", "HEAD^{commit}")
		var exit *exec.ExitError
		if err != nil && (!errors.As(err, &exit) || exit.ExitCode() != 1) {
			return err
		}
		s.Base, s.Head = head, head
	}
	if since == "staged" {
		return s.collect(ctx, "diff", "--cached", "--name-status", "-z", "-M", "--")
	}
	if since != "HEAD" {
		base, err := commit(ctx, s.Root, since)
		if err != nil {
			return err
		}
		s.Head, err = commit(ctx, s.Root, "HEAD")
		if err != nil {
			return err
		}
		s.Base, err = gitText(ctx, s.Root, "merge-base", s.Head, base)
		if err != nil {
			return err
		}
		if err := s.collect(ctx, "diff", "--name-status", "-z", "-M", s.Base, s.Head, "--"); err != nil {
			return err
		}
	}
	if err := s.collect(ctx, "diff", "--cached", "--name-status", "-z", "-M", "--"); err != nil {
		return err
	}
	if err := s.collect(ctx, "diff", "--name-status", "-z", "-M", "--"); err != nil {
		return err
	}
	files, err := gitOutput(ctx, s.Root, "ls-files", "--others", "--exclude-standard", "-z")
	if err != nil {
		return err
	}
	for _, path := range strings.Split(string(files), "\x00") {
		if path != "" {
			s.Changes = append(s.Changes, Change{Path: path, Status: "untracked"})
		}
	}
	return nil
}

func (s *ChangeSet) normalize() {
	seen := map[Change]bool{}
	kept := s.Changes[:0]
	for _, change := range s.Changes {
		if !seen[change] {
			kept = append(kept, change)
			seen[change] = true
		}
	}
	s.Changes = kept
	sort.Slice(s.Changes, func(i, j int) bool {
		if s.Changes[i].Path != s.Changes[j].Path {
			return s.Changes[i].Path < s.Changes[j].Path
		}
		return s.Changes[i].Status < s.Changes[j].Status
	})
}
