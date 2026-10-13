package api

import (
	"strings"
	"testing"

	"github.com/flanksource/clicky"
)

func TestIdentityOutputOrder(t *testing.T) {
	for _, identity := range []Identity{
		{ID: "user_123", Issuer: "https://clerk.example", Name: "Example User", Email: "user@example.com", OrganizationID: "org_123", OrganizationSlug: "example-team", EmailVerified: true},
		{ID: "user_123", Issuer: "https://clerk.example"},
	} {
		for _, format := range []string{"pretty", "json", "yaml", "markdown"} {
			t.Run(format+identity.OrganizationID, func(t *testing.T) {
				output, err := clicky.Format(identity, clicky.FormatOptions{Format: format, NoColor: true})
				if err != nil {
					t.Fatal(err)
				}
				previous := -1
				for _, key := range []string{"id", "issuer", "name", "email", "organization_id", "organization_slug"} {
					marker := key + ":"
					if format == "pretty" || format == "markdown" {
						labels := map[string]string{"id": "ID", "issuer": "Issuer", "name": "Name", "email": "Email", "organization_id": "Organization ID", "organization_slug": "Organization Slug"}
						marker = labels[key] + ":"
					}
					if format == "json" {
						marker = `"` + key + `":`
					}
					position := strings.Index(output, marker)
					if position <= previous {
						t.Fatalf("missing/out-of-order %s in %s", key, output)
					}
					previous = position
				}
				if strings.Contains(output, "email_verified") {
					t.Fatalf("unexpected field in %s", output)
				}
				if identity.OrganizationID == "" && format == "json" && !strings.Contains(output, `"organization_id": ""`) {
					t.Fatalf("empty org omitted: %s", output)
				}
			})
		}
	}
}
