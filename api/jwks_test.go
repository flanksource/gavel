// Tests the shared JWKS cache: keys are served from memory until the TTL lapses,
// an unknown kid refetches at most once per 30 seconds, and weak keys are refused.
package api

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/flanksource/gavel/api/clerktest"
)

func TestJWKSCache(t *testing.T) {
	clerk := clerktest.New(t, "cli", clerktest.User{ID: "user_123"})
	ctx := context.Background()
	now := time.Now()
	cache := NewJWKSCache(clerk.URL+"/.well-known/jwks.json", clerk.Client(), 10*time.Minute)
	cache.now = func() time.Time { return now }
	key := func(kid string) (*rsa.PublicKey, error) {
		t.Helper()
		return cache.Key(ctx, kid)
	}
	expectReads := func(want int64) {
		t.Helper()
		if got := clerk.JWKSReads(); got != want {
			t.Fatalf("JWKS reads: got %d, want %d", got, want)
		}
	}

	found, err := key(clerktest.KeyID)
	if err != nil || !found.Equal(clerk.PublicKey()) {
		t.Fatalf("first lookup: %v", err)
	}
	expectReads(1)
	if _, err := key(clerktest.KeyID); err != nil {
		t.Fatal(err)
	}
	expectReads(1)

	// An unknown kid refetches, but never more than once per 30 seconds.
	if _, err := key("rotated"); err == nil {
		t.Fatal("accepted an unpublished key")
	}
	expectReads(1)
	now = now.Add(31 * time.Second)
	if _, err := key("rotated"); err == nil {
		t.Fatal("accepted an unpublished key")
	}
	expectReads(2)
	clerk.PublishKey("rotated", clerk.PublicKey())
	if _, err := key("rotated"); err == nil {
		t.Fatal("refetched within 30 seconds")
	}
	expectReads(2)
	now = now.Add(31 * time.Second)
	if _, err := key("rotated"); err != nil {
		t.Fatalf("rotated key: %v", err)
	}
	expectReads(3)

	// A known kid is refetched once the TTL lapses.
	now = now.Add(10*time.Minute + time.Second)
	if _, err := key(clerktest.KeyID); err != nil {
		t.Fatal(err)
	}
	expectReads(4)

	weak, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatal(err)
	}
	clerk.PublishKey("weak", &weak.PublicKey)
	now = now.Add(31 * time.Second)
	if _, err := key("weak"); err == nil {
		t.Fatal("accepted a 1024-bit key")
	}
}

func TestJWKSCacheUnreachable(t *testing.T) {
	clerk := clerktest.New(t, "cli", clerktest.User{ID: "user_123"})
	cache := NewJWKSCache(clerk.URL+"/missing", clerk.Client(), 10*time.Minute)
	if _, err := cache.Key(context.Background(), clerktest.KeyID); err == nil {
		t.Fatal("accepted a JWKS endpoint that answers 404")
	}
}

func TestVerifyIdentityReportsUnreachableJWKS(t *testing.T) {
	clerk := clerktest.New(t, "cli", clerktest.User{ID: "user_123"})
	ctx := context.Background()
	client, err := NewAuthClient(ctx, AuthConfig{Issuer: clerk.URL, ClientID: "cli"}, clerk.Client())
	if err != nil {
		t.Fatal(err)
	}
	client.jwks = NewJWKSCache(clerk.URL+"/missing", clerk.Client(), 10*time.Minute)
	if _, err := client.verifyIdentity(ctx, clerk.IDToken(time.Now().Add(time.Hour))); err == nil || err.Error() != "the Clerk authentication endpoint returned HTTP 404" {
		t.Fatalf("want the JWKS fetch error, got %v", err)
	}
}

// expectRejected fails unless err is an ErrInvalidToken whose message contains want.
func expectRejected(t *testing.T, err error, want string, got any) {
	t.Helper()
	if !errors.Is(err, ErrInvalidToken) || !strings.Contains(err.Error(), want) {
		t.Fatalf("got %+v, %v; want an invalid token error containing %q", got, err, want)
	}
}
