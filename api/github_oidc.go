// Verifies GitHub Actions OIDC tokens, which CI jobs present to upload results,
// and extracts the repository and run they were issued for.
package api

import (
	"context"
	"strconv"

	"github.com/golang-jwt/jwt/v5"
)

const (
	// GitHubOIDCIssuer issues the ID tokens of GitHub Actions jobs.
	GitHubOIDCIssuer = "https://token.actions.githubusercontent.com"
	// GitHubOIDCJWKSURL publishes the keys that sign them.
	GitHubOIDCJWKSURL = GitHubOIDCIssuer + "/.well-known/jwks"
)

// GitHubOIDCVerifier accepts GitHub Actions tokens minted for Audience.
type GitHubOIDCVerifier struct {
	Audience string
	JWKS     *JWKSCache
}

// GitHubOIDCClaims identify the repository and workflow run a token was issued to.
type GitHubOIDCClaims struct {
	Repository        string
	RepositoryID      int64
	RepositoryOwnerID int64
	SHA, Ref, RunID   string
	Workflow          string
}

type githubTokenClaims struct {
	jwt.RegisteredClaims
	Repository        string `json:"repository"`
	RepositoryID      string `json:"repository_id"`
	RepositoryOwnerID string `json:"repository_owner_id"`
	SHA               string `json:"sha"`
	Ref               string `json:"ref"`
	RunID             string `json:"run_id"`
	Workflow          string `json:"workflow"`
}

// Verify checks raw and returns its claims. GitHub sends the repository and owner IDs as strings.
func (v GitHubOIDCVerifier) Verify(ctx context.Context, raw string) (GitHubOIDCClaims, error) {
	claims := &githubTokenClaims{}
	if _, err := parseSigned(ctx, v.JWKS, raw, claims, jwt.WithIssuer(GitHubOIDCIssuer), jwt.WithAudience(v.Audience)); err != nil {
		return GitHubOIDCClaims{}, err
	}
	if claims.Repository == "" {
		return GitHubOIDCClaims{}, invalidToken("GitHub OIDC token has no repository")
	}
	repositoryID, err := strconv.ParseInt(claims.RepositoryID, 10, 64)
	if err != nil {
		return GitHubOIDCClaims{}, invalidToken("GitHub OIDC token has a non-numeric repository_id %q", claims.RepositoryID)
	}
	ownerID, err := strconv.ParseInt(claims.RepositoryOwnerID, 10, 64)
	if err != nil {
		return GitHubOIDCClaims{}, invalidToken("GitHub OIDC token has a non-numeric repository_owner_id %q", claims.RepositoryOwnerID)
	}
	return GitHubOIDCClaims{
		Repository: claims.Repository, RepositoryID: repositoryID, RepositoryOwnerID: ownerID,
		SHA: claims.SHA, Ref: claims.Ref, RunID: claims.RunID, Workflow: claims.Workflow,
	}, nil
}
