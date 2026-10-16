// Verifies the CLI's Clerk OAuth access tokens: RS256 JWTs of type at+jwt, issued to
// gavel's public OAuth client and signed with the same keys as session and ID tokens.
package api

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"time"

	"github.com/eko/gocache/lib/v4/cache"
	"github.com/eko/gocache/lib/v4/store"
	commonscache "github.com/flanksource/commons-db/cache"
	"github.com/golang-jwt/jwt/v5"
)

// accessTokenType is the JWT header typ of a Clerk OAuth access token.
const accessTokenType = "at+jwt"

// AccessTokenVerifier accepts access tokens issued to one OAuth client. After the local
// checks, it asks Clerk whether the token is still valid and caches the answer for at
// most ttl, never past the token's exp, and never longer for being used.
type AccessTokenVerifier struct {
	issuer   string
	clientID string
	jwks     *JWKSCache
	clerk    ClerkOrganizations
	ttl      time.Duration
	checked  cache.CacheInterface[string]
}

// NewAccessTokenVerifier verifies tokens from issuer for clientID, re-checking each
// with clerk at most once per ttl.
func NewAccessTokenVerifier(issuer, clientID string, jwks *JWKSCache, clerk ClerkOrganizations, ttl time.Duration) *AccessTokenVerifier {
	return &AccessTokenVerifier{
		issuer: issuer, clientID: clientID, jwks: jwks, clerk: clerk, ttl: ttl,
		checked: commonscache.NewCache[string]("clerk-access-tokens", ttl),
	}
}

type accessTokenClaims struct {
	jwt.RegisteredClaims
	ClientID string `json:"client_id"`
}

// Verify returns the Clerk user ID of a valid, unrevoked token. A bad token is
// ErrInvalidToken; any other error means the token couldn't be checked.
func (v *AccessTokenVerifier) Verify(ctx context.Context, raw string) (string, error) {
	claims := &accessTokenClaims{}
	token, err := parseSigned(ctx, v.jwks, raw, claims, jwt.WithIssuer(v.issuer))
	if err != nil {
		return "", err
	}
	if typ, _ := token.Header["typ"].(string); typ != accessTokenType {
		return "", invalidToken("not an OAuth access token (typ %q)", typ)
	}
	if claims.ClientID != v.clientID {
		return "", invalidToken("access token was issued to OAuth client %q, not %q", claims.ClientID, v.clientID)
	}
	if claims.Subject == "" {
		return "", invalidToken("access token has no subject")
	}
	sum := sha256.Sum256([]byte(raw))
	key := hex.EncodeToString(sum[:])
	if user, err := v.checked.Get(ctx, key); err == nil {
		return user, nil
	}
	checkedAt := time.Now()
	if err := v.clerk.checkAccessToken(ctx, raw, claims.Subject, v.clientID); err != nil {
		return "", err
	}
	lifetime := min(v.ttl, claims.ExpiresAt.Sub(checkedAt))
	// A zero expiry would mean the cache's default TTL, not "already expired".
	if lifetime > 0 {
		if err := v.checked.Set(ctx, key, claims.Subject, store.WithExpiration(lifetime)); err != nil {
			return "", err
		}
	}
	return claims.Subject, nil
}

// checkAccessToken asks Clerk whether raw is an active, unrevoked access token for
// subject and clientID. Rejections wrap ErrInvalidToken.
func (c ClerkOrganizations) checkAccessToken(ctx context.Context, raw, subject, clientID string) error {
	var status struct {
		Object           string `json:"object"`
		Subject          string `json:"subject"`
		ClientID         string `json:"client_id"`
		Revoked          bool   `json:"revoked"`
		RevocationReason string `json:"revocation_reason"`
		Expired          bool   `json:"expired"`
	}
	err := c.call(ctx, http.MethodPost, "/oauth_applications/access_tokens/verify", map[string]string{"access_token": raw}, &status)
	var statusErr backendStatusError
	if errors.As(err, &statusErr) && statusErr.status == http.StatusNotFound {
		return invalidToken("Clerk doesn't recognise the access token")
	}
	if err != nil {
		return err
	}
	switch {
	case status.Revoked && status.RevocationReason != "":
		return invalidToken("Clerk reports the access token was revoked: %s", status.RevocationReason)
	case status.Revoked:
		return invalidToken("Clerk reports the access token was revoked")
	// {"active": false} decodes with no object.
	case status.Expired || status.Object != "clerk_idp_oauth_access_token":
		return invalidToken("Clerk reports the access token is no longer active")
	case status.Subject != subject || status.ClientID != clientID:
		return invalidToken("Clerk's record of the access token doesn't match its claims")
	}
	return nil
}
