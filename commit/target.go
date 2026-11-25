package commit

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/pr/provider"
)

type commitTarget struct {
	remote string
	base   string
	head   string
	repo   string
}

func parseCommitTarget(workDir, value string) (commitTarget, error) {
	base, head, hasHead := strings.Cut(value, ":")
	target := commitTarget{remote: "origin", base: base, head: head}
	if remote, branch, ok := strings.Cut(base, "/"); ok {
		target.remote, target.base = remote, branch
	}
	if target.remote == "" || strings.HasPrefix(target.remote, "-") || (hasHead && head == "") {
		return commitTarget{}, fmt.Errorf("--target must be [remote/]base[:topic], got %q", value)
	}
	branches := []string{target.base}
	if hasHead {
		branches = append(branches, target.head)
	}
	for _, branch := range branches {
		cmd := exec.Command("git", "check-ref-format", "--branch", branch)
		cmd.Dir = workDir
		if out, err := cmd.CombinedOutput(); err != nil || strings.TrimSpace(string(out)) != branch {
			return commitTarget{}, fmt.Errorf("--target %q contains invalid branch %q", value, branch)
		}
	}
	if target.head == target.base {
		return commitTarget{}, fmt.Errorf("--target topic must differ from base %q", target.base)
	}
	cmd := exec.Command("git", "remote", "get-url", target.remote)
	cmd.Dir = workDir
	out, err := cmd.Output()
	if err != nil {
		return commitTarget{}, fmt.Errorf("--target remote %q: %w", target.remote, err)
	}
	target.repo = strings.TrimSpace(string(out))
	return target, nil
}

func (o *Options) resolveTarget() error {
	if o.Target == "" {
		return nil
	}
	if !o.Push {
		return fmt.Errorf("--target requires --push (-p)")
	}
	if o.PushBranch != "" {
		return fmt.Errorf("--target cannot be combined with PushBranch")
	}
	if o.target.base != "" {
		return nil
	}
	var err error
	o.target, err = parseCommitTarget(o.WorkDir, o.Target)
	return err
}

func (t commitTarget) remoteName() string {
	if t.remote != "" {
		return t.remote
	}
	return "origin"
}

func (o Options) pushGithubOptions() (github.Options, error) {
	opts := github.Options{WorkDir: o.WorkDir}
	if o.target.repo != "" && !provider.IsAzure(o.target.repo) {
		var err error
		opts.Repo, err = github.ParseRepoURL(o.target.repo)
		if err != nil {
			return opts, fmt.Errorf("--target repository: %w", err)
		}
	}
	return opts, nil
}

func (o Options) pushBase(deps pushDeps, ghOpts github.Options) (string, error) {
	if o.target.base != "" {
		return o.target.base, nil
	}
	return deps.defaultBranch(ghOpts)
}
