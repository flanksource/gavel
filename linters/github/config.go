package githublint

import (
	"fmt"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/verify"
	"github.com/spf13/pflag"
	"time"
)

func (l *GitHub) configuration() (github.AuditOptions, github.RepositorySelection, error) {
	opts := l.AuditOptions
	opts.Client.WorkDir = l.WorkDir
	settings := verify.GitHubLintConfig{StaleAfter: "720h"}
	if opts.PR == nil {
		cfg, err := verify.LoadGavelConfig(verify.RuntimeConfigDirectory(l.WorkDir, l.ConfigDir))
		if err != nil {
			return opts, github.RepositorySelection{}, err
		}
		settings, err = verify.DecodeGitHubLintConfig(cfg.Lint.Linters["github"].Config)
		if err != nil {
			return opts, github.RepositorySelection{}, err
		}
	}
	selection := github.RepositorySelection{Orgs: settings.Orgs, Repos: settings.Repos, ExcludeRepos: settings.ExcludeRepos}
	duration, err := time.ParseDuration(settings.StaleAfter)
	if err != nil {
		return opts, selection, err
	}
	flags := pflag.NewFlagSet("github", pflag.ContinueOnError)
	flags.StringVar(&opts.Client.Repo, "repo", opts.Client.Repo, "GitHub owner/repo; overrides configured orgs and repos")
	flags.DurationVar(&opts.StaleAfter, "stale-after", duration, "PR inactivity threshold")
	if err := flags.Parse(l.ExtraArgs); err != nil {
		return opts, selection, err
	}
	if flags.NArg() > 0 || opts.StaleAfter <= 0 {
		return opts, selection, fmt.Errorf("expected only GitHub options and a positive --stale-after")
	}
	return opts, selection, nil
}
