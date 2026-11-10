package api

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/flanksource/gavel/api/clerktest"
)

func TestOrganizations(t *testing.T) {
	clerk := clerktest.New(t, "cli", clerktest.User{ID: "user_123"})
	created := time.UnixMilli(1760000000123).UTC()
	var want []ClerkOrganization
	for i := range 150 {
		o := clerktest.Organization{ID: fmt.Sprintf("org_%d", i), Slug: fmt.Sprintf("team-%d", i), Name: fmt.Sprintf("Team %d", i), Members: i, CreatedAt: created}
		clerk.Organizations = append(clerk.Organizations, o)
		want = append(want, ClerkOrganization{ID: o.ID, Slug: o.Slug, Name: o.Name, Members: i, CreatedAt: created})
	}
	backend := ClerkOrganizations{SecretKey: clerktest.SecretKey, HTTPClient: clerk.BackendClient()}
	ctx := context.Background()

	got, err := backend.Organizations(ctx)
	if err != nil || !slices.Equal(got, want) {
		t.Fatalf("got %d organizations, %v", len(got), err)
	}
	wantRequests := []string{
		"GET /v1/organizations?include_members_count=true&limit=100&offset=0",
		"GET /v1/organizations?include_members_count=true&limit=100&offset=100",
	}
	if requests := clerk.BackendRequests(); !slices.Equal(requests, wantRequests) {
		t.Fatalf("requests: %v", requests)
	}

	clerk.Organizations = nil
	if got, err := backend.Organizations(ctx); err != nil || len(got) != 0 {
		t.Fatalf("no organizations: %v, %v", got, err)
	}
	wrongKey := ClerkOrganizations{SecretKey: "sk_test_wrong", HTTPClient: clerk.BackendClient()}
	if _, err := wrongKey.Organizations(ctx); err == nil || !strings.Contains(err.Error(), "check the Clerk secret key") || strings.Contains(err.Error(), "sk_test_wrong") {
		t.Fatalf("wrong key: %v", err)
	}
}

type organizationTransport func(*http.Request) (*http.Response, error)

func (f organizationTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestClerkOrganizationResolver(t *testing.T) {
	for _, test := range []struct {
		name      string
		status    int
		body      string
		wantError bool
	}{
		{"metadata", 200, `{"id":"org_1","slug":"my-team","name":"My Team"}`, false},
		{"different org", 200, `{"id":"org_2","slug":"wrong"}`, true},
		{"malformed", 200, `not-json`, true},
		{"denied", 403, `server-secret should not be logged`, true},
		{"redirect", 302, `server-secret`, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			client := ClerkOrganizations{SecretKey: "server-secret", HTTPClient: &http.Client{Transport: organizationTransport(func(r *http.Request) (*http.Response, error) {
				if r.URL.String() != "https://api.clerk.com/v1/organizations/org_1" || r.Header.Get("Authorization") != "Bearer server-secret" {
					t.Fatal("wrong backend request")
				}
				return &http.Response{StatusCode: test.status, Body: io.NopCloser(strings.NewReader(test.body)), Header: make(http.Header), Request: r}, nil
			})}}
			org, err := client.ResolveOrganization(context.Background(), "org_1")
			if test.wantError {
				if err == nil || strings.Contains(err.Error(), "server-secret") {
					t.Fatalf("unsafe error: %v", err)
				}
			} else if err != nil || org.Slug != "my-team" || org.Name != "My Team" {
				t.Fatalf("wrong metadata: %+v %v", org, err)
			}
		})
	}
	if _, err := (ClerkOrganizations{}).ResolveOrganization(context.Background(), "org_1"); err == nil {
		t.Fatal("accepted missing backend key")
	}
	if _, err := (ClerkOrganizations{SecretKey: "secret"}).ResolveOrganization(context.Background(), "org_1/other"); err == nil {
		t.Fatal("accepted invalid organization ID")
	}
}

func TestIdentityStoreRejectsMissingDatabase(t *testing.T) {
	if _, _, err := (IdentityStore{}).Sync(context.Background(), Identity{ID: "user_1", Issuer: "https://clerk.example"}); err == nil {
		t.Fatal("accepted missing database")
	}
}
