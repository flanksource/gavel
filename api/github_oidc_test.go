// Tests GitHub Actions OIDC token verification: issuer, audience, expiry, and parsing
// of the repository claims GitHub sends as strings.
package api

import (
	"context"
	"testing"
	"time"

	"github.com/flanksource/gavel/api/clerktest"
	"github.com/golang-jwt/jwt/v5"
)

func TestGitHubOIDCVerifier(t *testing.T) {
	keys := clerktest.New(t, "cli", clerktest.User{ID: "user_123"})
	verifier := GitHubOIDCVerifier{
		Audience: "https://gavel.example",
		JWKS:     NewJWKSCache(keys.URL+"/.well-known/jwks.json", keys.Client(), 10*time.Minute),
	}
	ctx := context.Background()
	claims := func(extra jwt.MapClaims) jwt.MapClaims {
		base := jwt.MapClaims{
			"iss": GitHubOIDCIssuer, "aud": "https://gavel.example", "sub": "repo:acme/app:ref:refs/heads/main",
			"exp": time.Now().Add(5 * time.Minute).Unix(), "iat": time.Now().Unix(), "nbf": time.Now().Unix(),
			"repository": "acme/app", "repository_id": "123456", "repository_owner": "acme", "repository_owner_id": "789",
			"sha": "abc123", "ref": "refs/heads/main", "run_id": "42", "workflow": "CI",
		}
		for key, value := range extra {
			if value == nil {
				delete(base, key)
				continue
			}
			base[key] = value
		}
		return base
	}

	got, err := verifier.Verify(ctx, keys.Sign(claims(nil)))
	want := GitHubOIDCClaims{Repository: "acme/app", RepositoryID: 123456, RepositoryOwnerID: 789, SHA: "abc123", Ref: "refs/heads/main", RunID: "42", Workflow: "CI"}
	if err != nil || got != want {
		t.Fatalf("got %+v, %v; want %+v", got, err, want)
	}
	if _, err := verifier.Verify(ctx, keys.Sign(claims(jwt.MapClaims{"aud": []string{"other", "https://gavel.example"}}))); err != nil {
		t.Fatalf("audience list containing ours: %v", err)
	}

	for name, test := range map[string]struct{ token, want string }{
		"wrong audience":          {keys.Sign(claims(jwt.MapClaims{"aud": "https://other.example"})), "token has invalid audience"},
		"no audience":             {keys.Sign(claims(jwt.MapClaims{"aud": nil})), "aud claim is required"},
		"wrong issuer":            {keys.Sign(claims(jwt.MapClaims{"iss": keys.URL})), "token has invalid issuer"},
		"expired":                 {keys.Sign(claims(jwt.MapClaims{"exp": time.Now().Add(-time.Minute).Unix()})), "token is expired"},
		"no repository":           {keys.Sign(claims(jwt.MapClaims{"repository": nil})), "GitHub OIDC token has no repository"},
		"non-numeric repo id":     {keys.Sign(claims(jwt.MapClaims{"repository_id": "abc"})), `GitHub OIDC token has a non-numeric repository_id "abc"`},
		"no repository id":        {keys.Sign(claims(jwt.MapClaims{"repository_id": nil})), `GitHub OIDC token has a non-numeric repository_id ""`},
		"non-numeric owner id":    {keys.Sign(claims(jwt.MapClaims{"repository_owner_id": "x"})), `GitHub OIDC token has a non-numeric repository_owner_id "x"`},
		"unpublished signing key": {keys.SignWithKeyID("unknown", claims(nil)), "unknown signing key"},
		"malformed":               {"not-a-token", "token is malformed"},
	} {
		t.Run(name, func(t *testing.T) {
			got, err := verifier.Verify(ctx, test.token)
			expectRejected(t, err, test.want, got)
		})
	}
}
