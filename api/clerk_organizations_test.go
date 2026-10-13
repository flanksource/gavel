package api

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

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
