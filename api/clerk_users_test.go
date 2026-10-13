// Tests the Backend API user lookups against the fake Clerk instance: organization
// memberships across pages, and the user's GitHub OAuth token.
package api

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/flanksource/gavel/api/clerktest"
)

func TestUserMemberships(t *testing.T) {
	clerk := clerktest.New(t, "cli", clerktest.User{ID: "user_123"})
	var want []Membership
	for i := range 250 {
		m := clerktest.Membership{OrgID: fmt.Sprintf("org_%d", i), OrgSlug: fmt.Sprintf("team-%d", i), OrgName: fmt.Sprintf("Team %d", i), Role: "org:member"}
		clerk.Memberships = append(clerk.Memberships, m)
		want = append(want, Membership{OrgID: m.OrgID, OrgSlug: m.OrgSlug, OrgName: m.OrgName, Role: m.Role})
	}
	backend := ClerkOrganizations{SecretKey: clerktest.SecretKey, HTTPClient: clerk.BackendClient()}
	ctx := context.Background()

	got, err := backend.UserMemberships(ctx, "user_123")
	if err != nil || !slices.Equal(got, want) {
		t.Fatalf("got %d memberships, %v", len(got), err)
	}
	wantRequests := []string{
		"GET /v1/users/user_123/organization_memberships?limit=100&offset=0",
		"GET /v1/users/user_123/organization_memberships?limit=100&offset=100",
		"GET /v1/users/user_123/organization_memberships?limit=100&offset=200",
	}
	if requests := clerk.BackendRequests(); !slices.Equal(requests, wantRequests) {
		t.Fatalf("requests: %v", requests)
	}

	clerk.Memberships = nil
	if got, err := backend.UserMemberships(ctx, "user_123"); err != nil || len(got) != 0 {
		t.Fatalf("no memberships: %v, %v", got, err)
	}
	if _, err := backend.UserMemberships(ctx, "user_unknown"); err == nil {
		t.Fatal("accepted an unknown user")
	}
	if _, err := backend.UserMemberships(ctx, "user_123/../x"); err == nil {
		t.Fatal("accepted an invalid user ID")
	}
	wrongKey := ClerkOrganizations{SecretKey: "sk_test_wrong", HTTPClient: clerk.BackendClient()}
	if _, err := wrongKey.UserMemberships(ctx, "user_123"); err == nil || strings.Contains(err.Error(), "sk_test_wrong") {
		t.Fatalf("unsafe error: %v", err)
	}
}

func TestGitHubToken(t *testing.T) {
	clerk := clerktest.New(t, "cli", clerktest.User{ID: "user_123"})
	backend := ClerkOrganizations{SecretKey: clerktest.SecretKey, HTTPClient: clerk.BackendClient()}
	ctx := context.Background()

	if _, err := backend.GitHubToken(ctx, "user_123"); err == nil {
		t.Fatal("returned a token for a user without a GitHub connection")
	}
	clerk.GitHubToken = "gho_clerktest"
	if token, err := backend.GitHubToken(ctx, "user_123"); err != nil || token != "gho_clerktest" {
		t.Fatalf("got %q, %v", token, err)
	}
	if requests := clerk.BackendRequests(); requests[len(requests)-1] != "GET /v1/users/user_123/oauth_access_tokens/oauth_github" {
		t.Fatalf("requests: %v", requests)
	}
	if _, err := backend.GitHubToken(ctx, "user_unknown"); err == nil {
		t.Fatal("accepted an unknown user")
	}
	if _, err := backend.GitHubToken(ctx, "user_123?x=1"); err == nil {
		t.Fatal("accepted an invalid user ID")
	}
}

func TestClerkUser(t *testing.T) {
	clerk := clerktest.New(t, "cli", clerktest.User{ID: "user_123", Name: "Ada Lovelace", Email: "ada@example.com", GitHubLogin: "ada"})
	backend := ClerkOrganizations{SecretKey: clerktest.SecretKey, HTTPClient: clerk.BackendClient()}
	ctx := context.Background()

	got, err := backend.User(ctx, "user_123")
	if err != nil || got != (ClerkUser{Name: "Ada Lovelace", Email: "ada@example.com", EmailVerified: true, GitHubLogin: "ada"}) {
		t.Fatalf("got %+v, %v", got, err)
	}
	if _, err := backend.User(ctx, "user_unknown"); err == nil || !strings.Contains(err.Error(), "HTTP 404") {
		t.Fatalf("unknown user: %v", err)
	}
	if _, err := backend.User(ctx, "user_123/x"); err == nil {
		t.Fatal("accepted an invalid user ID")
	}

	clerk.User.GitHubLogin = ""
	if got, err := backend.User(ctx, "user_123"); err != nil || got.GitHubLogin != "" {
		t.Fatalf("user without GitHub: %+v, %v", got, err)
	}
}
