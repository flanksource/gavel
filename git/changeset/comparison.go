package changeset

import (
	"context"
	"fmt"
	"strings"
)

func (s *ChangeSet) fromComparison(ctx context.Context, comparison Comparison) error {
	s.Mode, s.Provider = comparison.Kind, comparison.Provider
	head, err := commit(ctx, s.Root, comparison.Head)
	if err != nil {
		return err
	}
	checkout, err := commit(ctx, s.Root, "HEAD")
	if err != nil {
		return err
	}
	if comparison.Kind == "pull-request" {
		return s.fromPullRequest(ctx, comparison, head, checkout)
	}
	if head != checkout {
		return fmt.Errorf("%s expects checkout %s, found %s", comparison.Provider, head, checkout)
	}
	s.Head = head
	if comparison.Kind == "commit" {
		parents, err := gitText(ctx, s.Root, "rev-list", "--parents", "-n", "1", head)
		if err != nil {
			return err
		}
		fields := strings.Fields(parents)
		if len(fields) == 1 {
			raw, err := gitText(ctx, s.Root, "cat-file", "-p", head)
			if err != nil {
				return err
			}
			if strings.HasPrefix(raw, "parent ") || strings.Contains(strings.SplitN(raw, "\n\n", 2)[0], "\nparent ") {
				return fmt.Errorf("commit %s is a shallow boundary; check out its parent history", head)
			}
			s.Base, err = gitText(ctx, s.Root, "hash-object", "-t", "tree", "--stdin")
			if err != nil {
				return err
			}
			return s.collect(ctx, "diff-tree", "--root", "--no-commit-id", "--name-status", "-z", "-r", "-M", head, "--")
		}
		comparison.Base = fields[1]
	} else if comparison.Kind != "range" {
		return fmt.Errorf("unsupported comparison kind %q", comparison.Kind)
	}
	s.Base, err = commit(ctx, s.Root, comparison.Base)
	if err != nil {
		return err
	}
	return s.collect(ctx, "diff", "--name-status", "-z", "-M", s.Base, s.Head, "--")
}

func (s *ChangeSet) fromPullRequest(ctx context.Context, comparison Comparison, source, checkout string) error {
	base, err := commit(ctx, s.Root, comparison.Base)
	if err != nil {
		return err
	}
	s.Head = checkout
	if checkout == source {
		s.Base, err = gitText(ctx, s.Root, "merge-base", source, base)
	} else {
		parents, parentErr := gitText(ctx, s.Root, "rev-list", "--parents", "-n", "1", checkout)
		if parentErr != nil {
			return parentErr
		}
		fields := strings.Fields(parents)
		if len(fields) != 3 || fields[2] != source {
			return fmt.Errorf("%s checkout %s is neither source %s nor its CI merge; check out the source or merge commit", comparison.Provider, checkout, source)
		}
		if _, err := gitOutput(ctx, s.Root, "merge-base", "--is-ancestor", base, fields[1]); err != nil {
			return fmt.Errorf("%s CI merge does not contain the reported target: %w", comparison.Provider, err)
		}
		s.Base = base
	}
	if err != nil {
		return err
	}
	return s.collect(ctx, "diff", "--name-status", "-z", "-M", s.Base, s.Head, "--")
}
