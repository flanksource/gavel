// The connection type of a repository whose pull requests a git connection reaches: Azure DevOps by
// its host or _git path, GitLab on gitlab.com or a host a gitlab connection names, over HTTPS or SSH.
package provider

import (
	"context"
	"net/url"
	"regexp"
	"strings"

	"github.com/flanksource/commons-db/git"
	"github.com/flanksource/commons-db/models"
	"github.com/flanksource/gavel/forge"
	"github.com/flanksource/gavel/gitlab"
)

// scpRepo is an SSH repository reference in git's scp-like form, user@host:path.
var scpRepo = regexp.MustCompile(`^[^@/:]+@([^:/]+):(.+)$`)

// ConnectionKind is the connection type of the repository at repo when a git connection reaches its
// pull requests, or "" for a GitHub one and for a reference that isn't a URL, such as owner/repo.
func ConnectionKind(ctx context.Context, repo string) (string, error) {
	host, path, ok := repoLocation(repo)
	if !ok {
		return "", nil
	}
	if IsAzure(host) || strings.Contains(path, "/_git/") {
		return models.ConnectionTypeAzureDevops, nil
	}
	switch strings.ToLower(host) {
	case "github.com":
		return "", nil
	case "gitlab.com":
		return models.ConnectionTypeGitlab, nil
	}
	target, err := git.ParseRepoURL(repo, models.ConnectionTypeGitlab)
	if err != nil {
		return "", nil
	}
	owner, err := forge.OwningConnection(ctx, gitlab.FindConnection, target, "repository "+git.Redact(repo))
	if err != nil {
		return "", err
	}
	if owner != nil && owner.Type == models.ConnectionTypeGitlab {
		return models.ConnectionTypeGitlab, nil
	}
	return "", nil
}

// repoLocation is the host and path of a repository URL or scp-like SSH reference; ok is false for
// anything else.
func repoLocation(repo string) (host, path string, ok bool) {
	if parsed, err := url.Parse(repo); err == nil && parsed.Scheme != "" && parsed.Host != "" {
		return parsed.Hostname(), parsed.Path, true
	}
	if match := scpRepo.FindStringSubmatch(repo); match != nil {
		return match[1], "/" + match[2], true
	}
	return "", "", false
}
