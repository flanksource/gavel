// Tests CLI access token verification: local JWT checks, the Clerk revocation check,
// and the fixed-expiry cache that bounds how long a revoked token keeps working.
package api

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/flanksource/gavel/api/clerktest"
	"github.com/golang-jwt/jwt/v5"
)

const verifyPath = "POST /v1/oauth_applications/access_tokens/verify"

func newAccessTokenFixture(t *testing.T, secret string, ttl time.Duration) (*clerktest.Server, *AccessTokenVerifier, func() int) {
	clerk := clerktest.New(t, "cli", clerktest.User{ID: "user_123"})
	jwks := NewJWKSCache(clerk.URL+"/.well-known/jwks.json", clerk.Client(), 10*time.Minute)
	verifier := NewAccessTokenVerifier(clerk.URL, "cli", jwks, ClerkOrganizations{SecretKey: secret, HTTPClient: clerk.BackendClient()}, ttl)
	checks := func() int {
		count := 0
		for _, request := range clerk.BackendRequests() {
			if request == verifyPath {
				count++
			}
		}
		return count
	}
	return clerk, verifier, checks
}

func TestAccessTokenVerifier(t *testing.T) {
	ctx := context.Background()
	hour := time.Now().Add(time.Hour)

	t.Run("caches a valid token until the TTL, then re-checks Clerk", func(t *testing.T) {
		clerk, verifier, checks := newAccessTokenFixture(t, clerktest.SecretKey, time.Second)
		token := clerk.AccessToken(hour)
		for range 2 {
			if user, err := verifier.Verify(ctx, token); err != nil || user != "user_123" {
				t.Fatalf("got %q, %v", user, err)
			}
		}
		if checks() != 1 {
			t.Fatalf("Clerk checks: %d, want 1", checks())
		}
		clerk.RevokeAccessToken(token)
		if _, err := verifier.Verify(ctx, token); err != nil {
			t.Fatalf("cached token rejected before its entry expired: %v", err)
		}
		time.Sleep(1200 * time.Millisecond)
		_, err := verifier.Verify(ctx, token)
		expectRejected(t, err, "Clerk reports the access token was revoked", nil)
		if checks() != 2 {
			t.Fatalf("Clerk checks: %d, want 2", checks())
		}
	})

	t.Run("rejects a revoked token on a cache miss and caches no failure", func(t *testing.T) {
		clerk, verifier, checks := newAccessTokenFixture(t, clerktest.SecretKey, 10*time.Minute)
		token := clerk.AccessToken(hour)
		clerk.RevokeAccessToken(token)
		for range 2 {
			_, err := verifier.Verify(ctx, token)
			expectRejected(t, err, "Clerk reports the access token was revoked", nil)
		}
		if checks() != 2 {
			t.Fatalf("Clerk checks: %d, want 2", checks())
		}
	})

	t.Run("never keeps an entry past the token's expiry", func(t *testing.T) {
		clerk, verifier, checks := newAccessTokenFixture(t, clerktest.SecretKey, 10*time.Minute)
		expiry := time.Unix(time.Now().Add(2*time.Second).Unix(), 0)
		token := clerk.AccessToken(expiry)
		if _, err := verifier.Verify(ctx, token); err != nil {
			t.Fatal(err)
		}
		// Past exp but inside the 5 second leeway, so only Clerk can reject it.
		time.Sleep(time.Until(expiry.Add(time.Second)))
		_, err := verifier.Verify(ctx, token)
		expectRejected(t, err, "Clerk reports the access token is no longer active", nil)
		if checks() != 2 {
			t.Fatalf("Clerk checks: %d, want 2", checks())
		}
	})

	t.Run("reports an unavailable Backend API as an error, not an invalid token", func(t *testing.T) {
		clerk, verifier, checks := newAccessTokenFixture(t, "sk_test_wrong", 10*time.Minute)
		token := clerk.AccessToken(hour)
		for range 2 {
			_, err := verifier.Verify(ctx, token)
			if err == nil || errors.Is(err, ErrInvalidToken) || strings.Contains(err.Error(), "sk_test_wrong") ||
				err.Error() != "Clerk Backend API returned HTTP 401: check the Clerk secret key" {
				t.Fatalf("got %v", err)
			}
		}
		if checks() != 2 {
			t.Fatalf("Clerk checks: %d, want 2", checks())
		}
	})

	clerk, verifier, checks := newAccessTokenFixture(t, clerktest.SecretKey, 10*time.Minute)
	claims := func(extra jwt.MapClaims) jwt.MapClaims {
		base := jwt.MapClaims{"sub": "user_123", "iss": clerk.URL, "client_id": "cli", "exp": hour.Unix(), "iat": time.Now().Unix()}
		for key, value := range extra {
			if value == nil {
				delete(base, key)
				continue
			}
			base[key] = value
		}
		return base
	}
	for name, test := range map[string]struct{ token, want string }{
		"ID token":          {clerk.IDToken(hour), `not an OAuth access token (typ "JWT")`},
		"session token":     {clerk.SessionToken("https://gavel.example", hour), `not an OAuth access token (typ "JWT")`},
		"JWT type":          {clerk.Sign(claims(nil)), `not an OAuth access token (typ "JWT")`},
		"other client":      {clerk.SignWithType("at+jwt", claims(jwt.MapClaims{"client_id": "other"})), `issued to OAuth client "other", not "cli"`},
		"no client":         {clerk.SignWithType("at+jwt", claims(jwt.MapClaims{"client_id": nil})), `issued to OAuth client "", not "cli"`},
		"wrong issuer":      {clerk.SignWithType("at+jwt", claims(jwt.MapClaims{"iss": "https://other.example"})), "token has invalid issuer"},
		"no subject":        {clerk.SignWithType("at+jwt", claims(jwt.MapClaims{"sub": nil})), "access token has no subject"},
		"expired":           {clerk.AccessToken(time.Now().Add(-time.Minute)), "token is expired"},
		"no expiry":         {clerk.SignWithType("at+jwt", claims(jwt.MapClaims{"exp": nil})), "exp claim is required"},
		"malformed":         {"not-a-token", "token is malformed"},
		"unsigned (none)":   {unsignedToken(t, claims(nil)), "signing method none is invalid"},
		"unpublished key":   {clerk.SignWithKeyID("unknown", claims(nil)), "unknown signing key"},
		"future not-before": {clerk.SignWithType("at+jwt", claims(jwt.MapClaims{"nbf": time.Now().Add(time.Minute).Unix()})), "token is not valid yet"},
	} {
		t.Run(name, func(t *testing.T) {
			user, err := verifier.Verify(ctx, test.token)
			expectRejected(t, err, test.want, user)
		})
	}
	if checks() != 0 {
		t.Fatalf("asked Clerk about %d tokens that failed local checks", checks())
	}
}
