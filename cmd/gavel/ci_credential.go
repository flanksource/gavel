// The credential a CI job uploads results to a hosted gavel with: GitHub Actions' OIDC token,
// GitLab CI's ID token, or Azure Pipelines' job access token with the build it proves.
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/flanksource/gavel/pr/ui"
)

// gitlabIDTokenEnv names the variable a GitLab job's id_tokens declares for gavel.
const gitlabIDTokenEnv = "GAVEL_ID_TOKEN"

// errActionsTokenMissing is why a job outside GitLab and Azure Pipelines has no credential.
var errActionsTokenMissing = errors.New("ACTIONS_ID_TOKEN_REQUEST_URL and ACTIONS_ID_TOKEN_REQUEST_TOKEN are not set: run in GitHub Actions with permissions: id-token: write")

// ciCredential is a CI job's bearer token for a hosted gavel, with the headers that name the
// build it proves.
type ciCredential struct {
	token   string
	headers map[string]string
}

// expiry is when the credential's token lapses, read without verifying it, or zero when it names
// no expiry or isn't a JWT.
func (c ciCredential) expiry() time.Time { return oidcTokenExpiry(c.token) }

// inGitLab and inAzurePipelines tell which CI getenv's job runs in.
func inGitLab(getenv func(string) string) bool { return getenv("GITLAB_CI") == "true" }

func inAzurePipelines(getenv func(string) string) bool {
	return strings.EqualFold(getenv("TF_BUILD"), "true")
}

// ciCredentialMissing says why getenv's job has no credential for audience, before any request,
// or is nil when it has one.
func ciCredentialMissing(getenv func(string) string, audience string) error {
	switch {
	case inGitLab(getenv):
		if getenv(gitlabIDTokenEnv) == "" {
			return fmt.Errorf("GITLAB_CI is set but %s is not: declare it in the job's id_tokens with aud: %s", gitlabIDTokenEnv, audience)
		}
	case inAzurePipelines(getenv):
		if getenv("SYSTEM_ACCESSTOKEN") == "" {
			return fmt.Errorf("TF_BUILD is set but SYSTEM_ACCESSTOKEN is not: map it into the step's env as SYSTEM_ACCESSTOKEN: $(System.AccessToken)")
		}
		if getenv("BUILD_BUILDID") == "" || getenv("BUILD_REPOSITORY_ID") == "" {
			return fmt.Errorf("TF_BUILD is set but BUILD_BUILDID or BUILD_REPOSITORY_ID is not: run in an Azure Pipelines build of an Azure Repos repository")
		}
	default:
		if getenv("ACTIONS_ID_TOKEN_REQUEST_URL") == "" || getenv("ACTIONS_ID_TOKEN_REQUEST_TOKEN") == "" {
			return errActionsTokenMissing
		}
	}
	return nil
}

// requestCICredential is the credential of getenv's job for audience, the hosted gavel's URL: a
// GitLab job's ID token, an Azure Pipelines job's access token with its build, else the OIDC
// token GitHub Actions issues the job.
func requestCICredential(ctx context.Context, client *http.Client, getenv func(string) string, audience string) (ciCredential, error) {
	if err := ciCredentialMissing(getenv, audience); err != nil {
		return ciCredential{}, err
	}
	switch {
	case inGitLab(getenv):
		return ciCredential{token: getenv(gitlabIDTokenEnv)}, nil
	case inAzurePipelines(getenv):
		return ciCredential{token: getenv("SYSTEM_ACCESSTOKEN"), headers: map[string]string{
			ui.CIAzureBuildHeader: getenv("BUILD_BUILDID"), ui.CIAzureRepositoryHeader: getenv("BUILD_REPOSITORY_ID")}}, nil
	}
	token, err := requestActionsOIDCToken(ctx, client, getenv, audience)
	return ciCredential{token: token}, err
}
