package prcreate

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

func defaultBase(dir string) (string, error) {
	repo, err := gogit.PlainOpenWithOptions(dir, &gogit.PlainOpenOptions{DetectDotGit: true, EnableDotGitCommonDir: true})
	if err != nil {
		return "", fmt.Errorf("resolve default branch in %s: %w", dir, err)
	}
	if _, err := repo.Remote("origin"); errors.Is(err, gogit.ErrRemoteNotFound) {
		return localBase(repo)
	} else if err != nil {
		return "", fmt.Errorf("read origin in %s: %w", dir, err)
	}
	ref, err := repo.Reference("refs/remotes/origin/HEAD", false)
	if errors.Is(err, plumbing.ErrReferenceNotFound) {
		return advertisedBase(dir)
	}
	if err != nil {
		return "", fmt.Errorf("read origin/HEAD in %s: %w", dir, err)
	}
	if ref.Type() != plumbing.SymbolicReference || !strings.HasPrefix(ref.Target().String(), "refs/remotes/origin/") {
		return "", fmt.Errorf("invalid origin/HEAD in %s: %s", dir, ref)
	}
	return strings.TrimPrefix(ref.Target().String(), "refs/remotes/"), nil
}

func localBase(repo *gogit.Repository) (string, error) {
	for _, branch := range []string{"main", "master"} {
		if _, err := repo.Reference(plumbing.NewBranchReferenceName(branch), true); err == nil {
			return branch, nil
		} else if !errors.Is(err, plumbing.ErrReferenceNotFound) {
			return "", fmt.Errorf("read local branch %s: %w", branch, err)
		}
	}
	return "", errors.New("cannot determine default branch: no origin or local main/master branch; configure pr.base")
}

func advertisedBase(dir string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", "-C", dir, "ls-remote", "--symref", "origin", "HEAD")
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	var stderr strings.Builder
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("query origin default branch in %s: %w: %s", dir, err, strings.TrimSpace(stderr.String()))
	}
	for line := range strings.SplitSeq(string(out), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 3 && fields[0] == "ref:" && fields[2] == "HEAD" && strings.HasPrefix(fields[1], "refs/heads/") {
			return "origin/" + strings.TrimPrefix(fields[1], "refs/heads/"), nil
		}
	}
	return "", fmt.Errorf("origin in %s did not advertise a default branch; configure pr.base", dir)
}
