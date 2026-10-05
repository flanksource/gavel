package prcreate

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/flanksource/captain/pkg/aiflags"
	"github.com/flanksource/captain/pkg/captainconfig"
	commitpkg "github.com/flanksource/gavel/commit"
	"github.com/flanksource/gavel/verify"
)

// contentInput describes the last `picked` commits on the worktree's HEAD,
// oldest first, for AI PR content generation.
func contentInput(wtPath string, picked int, in Input) (commitpkg.PRContentInput, error) {
	revs, err := captureGit(wtPath, "rev-list", "--reverse", "--max-count="+strconv.Itoa(picked), "HEAD")
	if err != nil {
		return commitpkg.PRContentInput{}, fmt.Errorf("list cherry-picked commits: %w", err)
	}
	shas := strings.Fields(revs)
	if len(shas) != picked {
		return commitpkg.PRContentInput{}, fmt.Errorf("expected %d cherry-picked commits on HEAD, found %d", picked, len(shas))
	}
	return ContentInput(wtPath, shas, in.Flags, in.Saved)
}

// ContentInput describes shas (in the given order) of the repository at dir,
// with dir's .gavel.yaml prompt configuration, for AI PR content generation.
func ContentInput(dir string, shas []string, flags aiflags.ModelFlags, saved *captainconfig.Config) (commitpkg.PRContentInput, error) {
	commits := make([]commitpkg.PRCommitInput, 0, len(shas))
	for _, rev := range shas {
		commit, err := commitInput(dir, rev)
		if err != nil {
			return commitpkg.PRContentInput{}, err
		}
		commits = append(commits, commit)
	}
	options, err := loadContentOptions(dir, Input{Flags: flags, Saved: saved})
	if err != nil {
		return commitpkg.PRContentInput{}, err
	}
	return commitpkg.PRContentInput{Commits: commits, Options: options}, nil
}

func commitInput(wtPath, rev string) (commitpkg.PRCommitInput, error) {
	msg, err := captureGit(wtPath, "show", "-s", "--format=%B", rev)
	if err != nil {
		return commitpkg.PRCommitInput{}, fmt.Errorf("read PR commit message: %w", err)
	}
	rawFiles, err := captureGit(wtPath, "show", "--name-only", "--format=", rev)
	if err != nil {
		return commitpkg.PRCommitInput{}, fmt.Errorf("read PR commit files: %w", err)
	}
	var files []string
	for _, line := range strings.Split(rawFiles, "\n") {
		if s := strings.TrimSpace(line); s != "" {
			files = append(files, s)
		}
	}
	return commitpkg.PRCommitInput{Message: msg, Files: files}, nil
}

func loadContentOptions(dir string, in Input) (commitpkg.Options, error) {
	cfg, err := verify.LoadGavelConfig(dir)
	if err != nil {
		return commitpkg.Options{}, fmt.Errorf("load .gavel.yaml for PR content: %w", err)
	}
	return commitpkg.Options{WorkDir: dir, AI: cfg.AI, PR: cfg.PR, Flags: in.Flags, Saved: in.Saved}, nil
}
