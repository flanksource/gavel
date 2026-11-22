// The CI job an upload of results comes from, identified by its own credential: a GitHub Actions
// OIDC token, a GitLab CI ID token, or an Azure Pipelines job's access token for its running build.
package ui

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/flanksource/commons-db/git"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/api"
	"github.com/flanksource/gavel/forge"
	"github.com/flanksource/gavel/tenant"
	"github.com/google/uuid"
)

const (
	// CIAzureBuildHeader names the running Azure Pipelines build (BUILD_BUILDID) whose job uploads
	// results with its access token.
	CIAzureBuildHeader = "X-Gavel-Azure-Build"
	// CIAzureRepositoryHeader names the repository the build builds (BUILD_REPOSITORY_ID).
	CIAzureRepositoryHeader = "X-Gavel-Azure-Repository"
)

// remoteHost is the lowercased host of a repos row's remote URL, without its user or port, taking
// remoteHostPattern as its parameter: gorm would read the pattern's question marks as parameters.
const remoteHost = `lower(substring(remote_url from ?))`

const remoteHostPattern = `^https?://(?:[^@/]*@)?([^/:]+)`

// ciJob is the CI job an upload came from: the repositories its results are stored in, its
// commit, and the job as messages name it.
type ciJob struct {
	holders []resultsRepo
	sha     string
	what    string
}

// ciRefusal is an upload refused with an HTTP status and the reason.
type ciRefusal struct {
	status  int
	message string
}

func (r ciRefusal) Error() string { return r.message }

func refuse(status int, format string, args ...any) error {
	return ciRefusal{status: status, message: fmt.Sprintf(format, args...)}
}

// identify is the job r's credential raw belongs to: an Azure Pipelines build when r names one,
// else the job of the GitHub Actions or GitLab token raw is.
func (c *CIResults) identify(r *http.Request, raw string) (ciJob, error) {
	if build := r.Header.Get(CIAzureBuildHeader); build != "" {
		return c.azureJob(r.Context(), raw, build, r.Header.Get(CIAzureRepositoryHeader))
	}
	issuer, err := api.UnverifiedIssuer(raw)
	if err != nil {
		return ciJob{}, refuse(http.StatusUnauthorized, "%v", err)
	}
	if issuer == api.GitHubOIDCIssuer {
		return c.githubJob(r.Context(), raw)
	}
	return c.gitlabJob(r.Context(), issuer, raw)
}

// githubJob is the GitHub Actions job of OIDC token raw, its results stored in every organization
// holding its repository.
func (c *CIResults) githubJob(ctx context.Context, raw string) (ciJob, error) {
	claims, err := c.OIDC.Verify(ctx, raw)
	if err != nil {
		return ciJob{}, refuse(http.StatusUnauthorized, "%v", err)
	}
	var holders []resultsRepo
	if err := c.DB.WithContext(ctx).Raw(`SELECT id, org_id FROM repos WHERE github_repo_id = ? ORDER BY created_at, id`, claims.RepositoryID).
		Scan(&holders).Error; err != nil {
		logger.Warnf("look up repository %d for CI results: %v", claims.RepositoryID, err)
		return ciJob{}, refuse(http.StatusInternalServerError, "could not look up the repository")
	}
	if len(holders) == 0 {
		return ciJob{}, refuse(http.StatusForbidden, "%s (%d) is not linked to a gavel organization; install the gavel GitHub App on it", claims.Repository, claims.RepositoryID)
	}
	return ciJob{holders: holders, sha: claims.SHA, what: claims.Repository}, nil
}

// gitlabJob is the GitLab CI job of ID token raw from issuer, its results stored in every
// organization whose git connection syncs its project from that GitLab.
func (c *CIResults) gitlabJob(ctx context.Context, issuer, raw string) (ciJob, error) {
	host, err := c.trustedGitLab(ctx, issuer)
	if err != nil {
		return ciJob{}, err
	}
	claims, err := c.GitLab.Verify(ctx, issuer, raw)
	if err != nil {
		return ciJob{}, refuse(http.StatusUnauthorized, "%v", err)
	}
	what := fmt.Sprintf("GitLab project %s (%s) on %s", claims.ProjectPath, claims.ProjectID, host)
	var holders []resultsRepo
	if err := c.DB.WithContext(ctx).Raw(`SELECT id, org_id FROM repos WHERE source = 'gitlab' AND connection_id IS NOT NULL
		AND provider_repo_id = ? AND `+remoteHost+` = ? ORDER BY created_at, id`, claims.ProjectID, remoteHostPattern, host).Scan(&holders).Error; err != nil {
		logger.Warnf("look up %s for CI results: %v", what, err)
		return ciJob{}, refuse(http.StatusInternalServerError, "could not look up the project")
	}
	if len(holders) == 0 {
		return ciJob{}, refuse(http.StatusForbidden, "%s is not synced by a gavel organization; add a gitlab connection that brings it", what)
	}
	return ciJob{holders: holders, sha: claims.SHA, what: what}, nil
}

// trustedGitLab is the host of issuer, an https URL, when it is gitlab.com or a GitLab a gavel
// organization's git connection syncs repositories from: only then are its keys fetched.
func (c *CIResults) trustedGitLab(ctx context.Context, issuer string) (string, error) {
	untrusted := refuse(http.StatusUnauthorized, "token issuer %s is neither GitHub Actions, gitlab.com nor a GitLab a gavel organization's git connection syncs repositories from", issuer)
	parsed, err := url.Parse(issuer)
	if err != nil || parsed.Scheme != "https" || parsed.Hostname() == "" || parsed.User != nil {
		return "", untrusted
	}
	host := strings.ToLower(parsed.Hostname())
	if host == "gitlab.com" {
		return host, nil
	}
	var synced bool
	if err := c.DB.WithContext(ctx).Raw(`SELECT EXISTS (SELECT 1 FROM repos WHERE source = 'gitlab' AND connection_id IS NOT NULL AND `+remoteHost+` = ?)`, remoteHostPattern, host).
		Scan(&synced).Error; err != nil {
		logger.Warnf("look up GitLab %s for CI results: %v", host, err)
		return "", refuse(http.StatusInternalServerError, "could not look up the GitLab")
	}
	if !synced {
		return "", untrusted
	}
	return host, nil
}

// azureHolder is a repository an Azure Pipelines build's results are stored in, with the
// connection and remote its build is read through.
type azureHolder struct {
	ID           uuid.UUID
	OrgID        uuid.UUID
	ConnectionID uuid.UUID
	RemoteURL    string
}

// azureJob is the job of running Azure Pipelines build of repository, proved by reading the build
// with raw, the job's access token, through the connection of a repository synced from it; its
// results are stored in every organization holding the repository.
func (c *CIResults) azureJob(ctx context.Context, raw, build, repository string) (ciJob, error) {
	if id, err := strconv.ParseInt(build, 10, 64); err != nil || id <= 0 {
		return ciJob{}, refuse(http.StatusBadRequest, "Azure DevOps build ID %q is not a positive number", build)
	}
	if repository == "" {
		return ciJob{}, refuse(http.StatusBadRequest, "send %s with the build's repository ID (BUILD_REPOSITORY_ID)", CIAzureRepositoryHeader)
	}
	var holders []azureHolder
	if err := c.DB.WithContext(ctx).Raw(`SELECT id, org_id, connection_id, remote_url FROM repos WHERE source = 'azure_devops'
		AND connection_id IS NOT NULL AND lower(provider_repo_id) = lower(?) ORDER BY created_at, id`, repository).Scan(&holders).Error; err != nil {
		logger.Warnf("look up Azure DevOps repository %s for CI results: %v", repository, err)
		return ciJob{}, refuse(http.StatusInternalServerError, "could not look up the repository")
	}
	if len(holders) == 0 {
		return ciJob{}, refuse(http.StatusForbidden, "Azure DevOps repository %s is not synced by a gavel organization; add an azure_devops connection that brings it", repository)
	}
	run, err := c.azureBuild(ctx, holders[0], raw, build)
	if err != nil {
		return ciJob{}, err
	}
	if !strings.EqualFold(run.RepositoryID, repository) {
		return ciJob{}, refuse(http.StatusForbidden, "Azure DevOps build %s built repository %s, not %s", build, run.RepositoryID, repository)
	}
	if run.Status == "completed" {
		return ciJob{}, refuse(http.StatusForbidden, "Azure DevOps build %s has completed; only a running build uploads its results", build)
	}
	job := ciJob{sha: run.SHA, what: "Azure DevOps build " + build}
	for _, holder := range holders {
		job.holders = append(job.holders, resultsRepo{ID: holder.ID, OrgID: holder.OrgID})
	}
	return job, nil
}

// azureBuild reads build through holder's connection, authenticated by raw alone: never the
// connection's PAT, nor a response cached for another caller.
func (c *CIResults) azureBuild(ctx context.Context, holder azureHolder, raw, build string) (git.PipelineExecution, error) {
	options := c.ClientOptions
	options.ReadCache = nil
	options.TokenSource = func(context.Context) (string, error) { return raw, nil }
	remote, err := forge.Remote(ctx, c.DB, tenant.Scope{OrgID: holder.OrgID}, holder.ConnectionID, holder.RemoteURL, options)
	if err != nil {
		logger.Warnf("reach %s to read Azure DevOps build %s: %v", holder.RemoteURL, build, err)
		return git.PipelineExecution{}, refuse(http.StatusInternalServerError, "could not reach the repository's Azure DevOps")
	}
	reader, ok := remote.Provider.(git.PipelineExecutionReader)
	if !ok {
		return git.PipelineExecution{}, refuse(http.StatusInternalServerError, "the repository's provider doesn't read builds")
	}
	run, err := reader.PipelineExecution(ctx, remote.URL, build)
	var apiErr *git.APIError
	switch {
	case err == nil:
		return run, nil
	case errors.As(err, &apiErr) && (apiErr.StatusCode == http.StatusUnauthorized || apiErr.StatusCode == http.StatusForbidden):
		return git.PipelineExecution{}, refuse(http.StatusUnauthorized, "Azure DevOps refused the job's access token for build %s: %v", build, err)
	case errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound:
		return git.PipelineExecution{}, refuse(http.StatusForbidden, "Azure DevOps has no build %s the job's access token can read", build)
	}
	return git.PipelineExecution{}, refuse(http.StatusBadGateway, "read Azure DevOps build %s: %v", build, err)
}
