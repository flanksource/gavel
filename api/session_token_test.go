// Tests browser session token verification: issuer, authorized party, expiry leeway,
// both Clerk org claim versions, and refusal of OAuth access tokens and pending sessions.
package api

import (
	"context"
	"testing"
	"time"

	"github.com/flanksource/gavel/api/clerktest"
	"github.com/golang-jwt/jwt/v5"
)

func TestSessionVerifier(t *testing.T) {
	clerk := clerktest.New(t, "cli", clerktest.User{ID: "user_123", OrgID: "org_123", OrgSlug: "team", OrgRole: "org:admin"})
	const origin = "https://gavel.example"
	verifier := SessionVerifier{
		Issuer:            clerk.URL,
		AuthorizedParties: []string{origin},
		JWKS:              NewJWKSCache(clerk.URL+"/.well-known/jwks.json", clerk.Client(), 10*time.Minute),
	}
	ctx := context.Background()
	hour := time.Now().Add(time.Hour)
	claims := func(extra jwt.MapClaims) jwt.MapClaims {
		base := jwt.MapClaims{"sub": "user_123", "iss": clerk.URL, "azp": origin, "exp": hour.Unix(), "iat": time.Now().Unix()}
		for key, value := range extra {
			if value == nil {
				delete(base, key)
				continue
			}
			base[key] = value
		}
		return base
	}

	t.Run("v2 org claim", func(t *testing.T) {
		got, err := verifier.Verify(ctx, clerk.SessionToken(origin, hour))
		want := SessionClaims{UserID: "user_123", OrgID: "org_123", OrgSlug: "team", OrgRole: "org:admin"}
		if err != nil || got != want {
			t.Fatalf("got %+v, %v; want %+v", got, err, want)
		}
	})
	t.Run("v1 org claims", func(t *testing.T) {
		got, err := verifier.Verify(ctx, clerk.Sign(claims(jwt.MapClaims{"org_id": "org_9", "org_slug": "nine", "org_role": "org:member"})))
		want := SessionClaims{UserID: "user_123", OrgID: "org_9", OrgSlug: "nine", OrgRole: "org:member"}
		if err != nil || got != want {
			t.Fatalf("got %+v, %v; want %+v", got, err, want)
		}
	})
	t.Run("no active org", func(t *testing.T) {
		got, err := verifier.Verify(ctx, clerk.Sign(claims(nil)))
		if err != nil || got != (SessionClaims{UserID: "user_123"}) {
			t.Fatalf("got %+v, %v", got, err)
		}
	})
	t.Run("no azp", func(t *testing.T) {
		if _, err := verifier.Verify(ctx, clerk.Sign(claims(jwt.MapClaims{"azp": nil}))); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("expired within leeway", func(t *testing.T) {
		if _, err := verifier.Verify(ctx, clerk.Sign(claims(jwt.MapClaims{"exp": time.Now().Add(-2 * time.Second).Unix()}))); err != nil {
			t.Fatal(err)
		}
	})

	for name, test := range map[string]struct{ token, want string }{
		"expired":           {clerk.SessionToken(origin, time.Now().Add(-time.Minute)), "token is expired"},
		"not yet valid":     {clerk.Sign(claims(jwt.MapClaims{"nbf": time.Now().Add(time.Minute).Unix()})), "token is not valid yet"},
		"no expiry":         {clerk.Sign(claims(jwt.MapClaims{"exp": nil})), "exp claim is required"},
		"wrong issuer":      {clerk.Sign(claims(jwt.MapClaims{"iss": "https://other.example"})), "token has invalid issuer"},
		"wrong azp":         {clerk.SessionToken("https://evil.example", hour), `issued for origin "https://evil.example", which is not an authorized party`},
		"no subject":        {clerk.Sign(claims(jwt.MapClaims{"sub": nil})), "session token has no subject"},
		"pending session":   {clerk.Sign(claims(jwt.MapClaims{"sts": "pending"})), "the Clerk session is pending"},
		"access token":      {clerk.AccessToken(hour), "an OAuth access token is not a browser session token"},
		"client_id claim":   {clerk.Sign(claims(jwt.MapClaims{"client_id": "cli"})), "an OAuth access token is not a browser session token"},
		"at+jwt type":       {clerk.SignWithType("at+jwt", claims(nil)), "an OAuth access token is not a browser session token"},
		"unpublished kid":   {clerk.SignWithKeyID("unknown", claims(nil)), "unknown signing key"},
		"malformed":         {"not-a-token", "token is malformed"},
		"empty":             {"", "token is malformed"},
		"unsigned (none)":   {unsignedToken(t, claims(nil)), "signing method none is invalid"},
		"slug without org":  {clerk.Sign(claims(jwt.MapClaims{"org_slug": "nine"})), "organization slug or role without an organization ID"},
		"v2 org without id": {clerk.Sign(claims(jwt.MapClaims{"o": map[string]string{"slg": "team", "rol": "admin"}})), "organization slug or role without an organization ID"},
	} {
		t.Run(name, func(t *testing.T) {
			got, err := verifier.Verify(ctx, test.token)
			expectRejected(t, err, test.want, got)
		})
	}
}

// unsignedToken builds an alg=none token, which no verifier may accept.
func unsignedToken(t *testing.T, claims jwt.MapClaims) string {
	t.Helper()
	raw, err := jwt.NewWithClaims(jwt.SigningMethodNone, claims).SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}
