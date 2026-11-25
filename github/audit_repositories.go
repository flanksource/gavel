package github

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bmatcuk/doublestar/v4"
	"net/url"
	"sort"
	"strings"

	"github.com/flanksource/gavel/internal/githubselector"
)

type RepositorySelection struct {
	Orgs, Repos, ExcludeRepos []string
}

func (o Options) AuditRepositories(ctx context.Context, selection RepositorySelection) ([]string, error) {
	repos := append([]string(nil), selection.Repos...)
	var failures []error
	if o.Repo != "" {
		repos = []string{o.Repo}
	} else if len(selection.Orgs)+len(repos) == 0 {
		repo, err := o.resolveRepo()
		if err != nil {
			return nil, err
		}
		repos = append(repos, repo)
	} else {
		repos = nil
		queries := map[string]bool{}
		for _, selector := range selection.Repos {
			if !githubselector.IsQuery(selector) {
				repos = append(repos, selector)
				continue
			}
			selector = strings.TrimSpace(selector)
			if queries[selector] {
				continue
			}
			queries[selector] = true
			found, err := o.searchAuditRepositories(ctx, selector)
			repos = append(repos, found...)
			if err != nil {
				failures = append(failures, fmt.Errorf("repository query %q: %w", selector, err))
			}
		}
		seen := map[string]bool{}
		for _, org := range selection.Orgs {
			org = strings.ToLower(org)
			if seen[org] {
				continue
			}
			seen[org] = true
			found, err := o.organizationRepositories(ctx, org)
			repos = append(repos, found...)
			if err != nil {
				failures = append(failures, err)
			}
		}
	}
	unique := map[string]bool{}
	for _, repo := range repos {
		repo = strings.ToLower(repo)
		excluded := false
		for _, pattern := range selection.ExcludeRepos {
			matched, err := doublestar.Match(strings.ToLower(pattern), repo)
			if err != nil {
				return nil, fmt.Errorf("repository exclusion %q: %w", pattern, err)
			}
			excluded = excluded || matched
		}
		if !excluded {
			unique[repo] = true
		}
	}
	selected := make([]string, 0, len(unique))
	for repo := range unique {
		selected = append(selected, repo)
	}
	sort.Strings(selected)
	return selected, errors.Join(failures...)
}

func (o Options) organizationRepositories(ctx context.Context, org string) ([]string, error) {
	token, err := o.token()
	if err != nil {
		return nil, err
	}
	path := "/orgs/" + url.PathEscape(org) + "/repos?type=all&per_page=100"
	seen := map[string]bool{}
	var repos []string
	for path != "" {
		if seen[path] {
			return repos, fmt.Errorf("organization %s pagination did not advance", org)
		}
		seen[path] = true
		response, err := newClient(token).R(ctx).Get(githubAPIBase() + path)
		if err != nil {
			return repos, err
		}
		body, err := response.AsString()
		if err != nil {
			return repos, err
		}
		if !response.IsOK() {
			return repos, fmt.Errorf("organization %s repositories HTTP %d: %s", org, response.StatusCode, body)
		}
		var rows []struct {
			FullName string `json:"full_name"`
		}
		if err := json.Unmarshal([]byte(body), &rows); err != nil {
			return repos, err
		}
		if rows == nil {
			return repos, fmt.Errorf("organization %s repositories returned null", org)
		}
		for _, row := range rows {
			if row.FullName == "" {
				return repos, fmt.Errorf("organization %s returned a repository without full_name", org)
			}
			repos = append(repos, row.FullName)
		}
		path, err = auditNextPage(response.Header.Get("Link"))
		if err != nil {
			return repos, err
		}
	}
	return repos, nil
}
