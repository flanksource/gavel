package verify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/bmatcuk/doublestar/v4"
	"regexp"
	"strings"
	"time"

	"github.com/flanksource/gavel/internal/githubselector"
)

type GitHubLintConfig struct {
	Orgs         []string `json:"orgs,omitempty" yaml:"orgs,omitempty"`
	Repos        []string `json:"repos,omitempty" yaml:"repos,omitempty"`
	ExcludeRepos []string `json:"excludeRepos,omitempty" yaml:"excludeRepos,omitempty"`
	StaleAfter   string   `json:"staleAfter,omitempty" yaml:"staleAfter,omitempty"`
}

var githubOrgName = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9-]*$`)
var githubRepoName = regexp.MustCompile(`^[A-Za-z0-9_.-]+$`)

func DecodeGitHubLintConfig(raw map[string]any) (GitHubLintConfig, error) {
	cfg := GitHubLintConfig{StaleAfter: "720h"}
	if raw == nil {
		return cfg, nil
	}
	data, err := json.Marshal(raw)
	if err != nil {
		return cfg, err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&cfg); err != nil {
		return cfg, err
	}
	for _, org := range cfg.Orgs {
		if !githubOrgName.MatchString(org) {
			return cfg, fmt.Errorf("invalid GitHub organization %q", org)
		}
	}
	for _, repo := range cfg.Repos {
		if githubselector.IsQuery(repo) {
			continue
		}
		parts := strings.Split(repo, "/")
		if len(parts) != 2 || !githubOrgName.MatchString(parts[0]) || !githubRepoName.MatchString(parts[1]) || parts[1] == "." || parts[1] == ".." {
			return cfg, fmt.Errorf("invalid GitHub repository %q: expected owner/repo or a qualified repository search query", repo)
		}
	}
	for _, pattern := range cfg.ExcludeRepos {
		if !strings.Contains(pattern, "/") || strings.HasPrefix(pattern, "!") || !doublestar.ValidatePattern(pattern) {
			return cfg, fmt.Errorf("invalid repository exclusion %q: expected an owner/repo glob", pattern)
		}
	}
	duration, err := time.ParseDuration(cfg.StaleAfter)
	if err != nil || duration <= 0 {
		return cfg, fmt.Errorf("staleAfter must be a positive Go duration, got %q", cfg.StaleAfter)
	}
	return cfg, nil
}
