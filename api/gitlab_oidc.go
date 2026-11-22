// Verifies GitLab CI ID tokens, which CI jobs present to upload results, against the GitLab that
// issued them, and extracts the project and pipeline they were issued for.
package api

import (
	"context"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// GitLabJWKSPath is where a GitLab publishes the keys that sign its ID tokens, under its URL.
const GitLabJWKSPath = "/oauth/discovery/keys"

// GitLabOIDCVerifier accepts GitLab CI ID tokens minted for Audience by an issuer the caller trusts.
type GitLabOIDCVerifier struct {
	Audience string
	// JWKS is the key set of the GitLab at issuer, the URL its tokens name.
	JWKS func(issuer string) *JWKSCache
}

// GitLabOIDCClaims identify the project, commit and pipeline job a token was issued to. GitLab
// sends the IDs as strings.
type GitLabOIDCClaims struct {
	ProjectID, ProjectPath string
	SHA, Ref               string
	PipelineID, JobID      string
}

type gitlabTokenClaims struct {
	jwt.RegisteredClaims
	ProjectID   string `json:"project_id"`
	ProjectPath string `json:"project_path"`
	SHA         string `json:"sha"`
	Ref         string `json:"ref"`
	PipelineID  string `json:"pipeline_id"`
	JobID       string `json:"job_id"`
}

// Verify checks raw was issued by the GitLab at issuer, which the caller trusts, and returns its
// claims.
func (v GitLabOIDCVerifier) Verify(ctx context.Context, issuer, raw string) (GitLabOIDCClaims, error) {
	claims := &gitlabTokenClaims{}
	if _, err := parseSigned(ctx, v.JWKS(issuer), raw, claims, jwt.WithIssuer(issuer), jwt.WithAudience(v.Audience)); err != nil {
		return GitLabOIDCClaims{}, err
	}
	if id, err := strconv.ParseInt(claims.ProjectID, 10, 64); err != nil || id <= 0 {
		return GitLabOIDCClaims{}, invalidToken("GitLab ID token has a non-numeric project_id %q", claims.ProjectID)
	}
	if claims.SHA == "" {
		return GitLabOIDCClaims{}, invalidToken("GitLab ID token has no sha")
	}
	return GitLabOIDCClaims{ProjectID: claims.ProjectID, ProjectPath: claims.ProjectPath, SHA: claims.SHA, Ref: claims.Ref,
		PipelineID: claims.PipelineID, JobID: claims.JobID}, nil
}

// UnverifiedIssuer is the issuer raw names, read without verifying it, to choose the keys that
// verify it.
func UnverifiedIssuer(raw string) (string, error) {
	claims := jwt.RegisteredClaims{}
	if _, _, err := jwt.NewParser().ParseUnverified(raw, &claims); err != nil {
		return "", invalidToken("%v", err)
	}
	if claims.Issuer == "" {
		return "", invalidToken("the token names no issuer")
	}
	return claims.Issuer, nil
}

// JWKSCaches keeps one JWKSCache per key set URL, each read with the same client and TTL. It is
// safe for concurrent use.
type JWKSCaches struct {
	client *http.Client
	ttl    time.Duration

	mu     sync.Mutex
	caches map[string]*JWKSCache
}

// NewJWKSCaches reads each key set with client, keeping its keys for ttl.
func NewJWKSCaches(client *http.Client, ttl time.Duration) *JWKSCaches {
	return &JWKSCaches{client: client, ttl: ttl, caches: map[string]*JWKSCache{}}
}

// For is the cache of the key set at url.
func (c *JWKSCaches) For(url string) *JWKSCache {
	c.mu.Lock()
	defer c.mu.Unlock()
	cache, ok := c.caches[url]
	if !ok {
		cache = NewJWKSCache(url, c.client, c.ttl)
		c.caches[url] = cache
	}
	return cache
}
