// Tests GitLab CI ID token verification: the issuer the caller trusts, audience, expiry, and the
// project and pipeline claims; and the issuer read from a token before it is verified.
package api

import (
	"context"
	"testing"
	"time"

	"github.com/flanksource/gavel/api/clerktest"
	"github.com/golang-jwt/jwt/v5"
)

func TestGitLabOIDCVerifier(t *testing.T) {
	keys := clerktest.New(t, "cli", clerktest.User{ID: "user_123"})
	const issuer = "https://gitlab.example.com"
	var askedFor []string
	verifier := GitLabOIDCVerifier{
		Audience: "https://gavel.example",
		JWKS: func(issuer string) *JWKSCache {
			askedFor = append(askedFor, issuer)
			return NewJWKSCache(keys.URL+"/.well-known/jwks.json", keys.Client(), 10*time.Minute)
		},
	}
	ctx := context.Background()
	claims := func(extra jwt.MapClaims) jwt.MapClaims {
		base := jwt.MapClaims{
			"iss": issuer, "aud": "https://gavel.example", "sub": "project_path:acme/api:ref_type:branch:ref:main",
			"exp": time.Now().Add(5 * time.Minute).Unix(), "iat": time.Now().Unix(), "nbf": time.Now().Unix(),
			"project_id": "42", "project_path": "acme/api", "namespace_id": "7", "sha": "abc123", "ref": "main",
			"pipeline_id": "1001", "job_id": "5005",
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

	got, err := verifier.Verify(ctx, issuer, keys.Sign(claims(nil)))
	want := GitLabOIDCClaims{ProjectID: "42", ProjectPath: "acme/api", SHA: "abc123", Ref: "main", PipelineID: "1001", JobID: "5005"}
	if err != nil || got != want {
		t.Fatalf("got %+v, %v; want %+v", got, err, want)
	}
	if len(askedFor) != 1 || askedFor[0] != issuer {
		t.Fatalf("keys asked for %v, want the trusted issuer's", askedFor)
	}

	for name, test := range map[string]struct{ token, want string }{
		"another issuer than trusted": {keys.Sign(claims(jwt.MapClaims{"iss": "https://gitlab.com"})), "token has invalid issuer"},
		"wrong audience":              {keys.Sign(claims(jwt.MapClaims{"aud": "https://other.example"})), "token has invalid audience"},
		"expired":                     {keys.Sign(claims(jwt.MapClaims{"exp": time.Now().Add(-time.Minute).Unix()})), "token is expired"},
		"no project":                  {keys.Sign(claims(jwt.MapClaims{"project_id": nil})), `GitLab ID token has a non-numeric project_id ""`},
		"non-numeric project":         {keys.Sign(claims(jwt.MapClaims{"project_id": "acme"})), `GitLab ID token has a non-numeric project_id "acme"`},
		"no commit":                   {keys.Sign(claims(jwt.MapClaims{"sha": nil})), "GitLab ID token has no sha"},
	} {
		t.Run(name, func(t *testing.T) {
			got, err := verifier.Verify(ctx, issuer, test.token)
			expectRejected(t, err, test.want, got)
		})
	}
}

func TestUnverifiedIssuer(t *testing.T) {
	keys := clerktest.New(t, "cli", clerktest.User{ID: "user_123"})
	issuer, err := UnverifiedIssuer(keys.Sign(jwt.MapClaims{"iss": "https://gitlab.example.com"}))
	if err != nil || issuer != "https://gitlab.example.com" {
		t.Fatalf("got %q, %v", issuer, err)
	}
	for name, test := range map[string]struct{ token, want string }{
		"malformed": {"not-a-token", "token is malformed"},
		"no issuer": {keys.Sign(jwt.MapClaims{"sub": "x"}), "the token names no issuer"},
	} {
		t.Run(name, func(t *testing.T) {
			got, err := UnverifiedIssuer(test.token)
			expectRejected(t, err, test.want, got)
		})
	}
}

func TestJWKSCachesKeepOnePerKeySet(t *testing.T) {
	caches := NewJWKSCaches(nil, time.Minute)
	first, again, other := caches.For("https://gitlab.com"+GitLabJWKSPath), caches.For("https://gitlab.com"+GitLabJWKSPath), caches.For("https://gitlab.example.com"+GitLabJWKSPath)
	if first != again || first == other || other.url != "https://gitlab.example.com/oauth/discovery/keys" {
		t.Fatalf("got %p, %p and %p (%s); want one cache per key set URL", first, again, other, other.url)
	}
}
