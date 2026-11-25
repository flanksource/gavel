package azuredevops

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os/exec"
	"strings"
)

func blameGit(ctx context.Context, dir string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(string(out)))
	}
	return strings.TrimSpace(string(out)), nil
}

func (c *Client) blameRemote(ctx context.Context, dir string) (string, error) {
	remotes, err := blameGit(ctx, dir, "remote")
	if err != nil {
		return "", err
	}
	for _, name := range strings.Fields(remotes) {
		remote, err := blameGit(ctx, dir, "remote", "get-url", name)
		if err != nil {
			return "", err
		}
		parsed, err := ParseRepository(remote)
		if err != nil {
			continue
		}
		if !strings.EqualFold(parsed.Organization, c.repo.Organization) {
			continue
		}
		if !strings.EqualFold(parsed.Name, c.repo.Name) && !strings.EqualFold(parsed.Name, c.info.ID) {
			continue
		}
		if parsed.Project != "" && !strings.EqualFold(parsed.Project, c.repo.Project) && !strings.EqualFold(parsed.Project, c.info.Project.ID) {
			continue
		}
		if parsed.Project == "" {
			data, _, err := c.request(ctx, http.MethodGet, c.projectPath("", "git/repositories/"+url.PathEscape(parsed.Name)), nil, nil)
			if err != nil {
				return "", err
			}
			var resolved repositoryInfo
			if err := json.Unmarshal(data, &resolved); err != nil {
				return "", fmt.Errorf("resolve Azure remote repository: %w", err)
			}
			if resolved.ID == "" {
				return "", fmt.Errorf("Azure remote repository %s has no ID", parsed.Name)
			}
			if !strings.EqualFold(resolved.ID, c.info.ID) {
				continue
			}
		}
		return name, nil
	}
	return "", fmt.Errorf("no git remote points at %s; run from its local checkout", c.repo.URL())
}

type blameCheckout struct {
	Dir, Remote string
	PR          pullRequest
}

func (c *Client) blameCommits(ctx context.Context, checkout blameCheckout) error {
	source, err := c.sourceRepository(ctx, checkout.PR)
	if err != nil {
		return err
	}
	sourceURL := (Repository{Organization: c.repo.Organization, Project: source.Project.Name, Name: source.Name}).URL()
	for _, revision := range []struct{ remote, ref, sha string }{
		{checkout.Remote, checkout.PR.Target, checkout.PR.TargetCommit.ID},
		{sourceURL, checkout.PR.Source, checkout.PR.SourceCommit.ID},
	} {
		if _, err := blameGit(ctx, checkout.Dir, "cat-file", "-e", revision.sha+"^{commit}"); err == nil {
			continue
		}
		if _, err := blameGit(ctx, checkout.Dir, "fetch", "--no-tags", "--no-write-fetch-head", "--refmap=", "--quiet", revision.remote, revision.ref); err != nil {
			return err
		}
		if _, err := blameGit(ctx, checkout.Dir, "cat-file", "-e", revision.sha+"^{commit}"); err != nil {
			return fmt.Errorf("Azure PR commit %s is missing after fetching; refresh the PR: %w", revision.sha, err)
		}
	}
	return nil
}
