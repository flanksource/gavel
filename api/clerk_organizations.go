package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// ClerkOrganizations resolves canonical org metadata on a trusted backend.
// SecretKey must remain on that backend; never ship/persist it in CLI sessions.
// Configure one resolver per issuer/Clerk instance.
type ClerkOrganizations struct {
	SecretKey  string
	HTTPClient *http.Client
}

func (c ClerkOrganizations) ResolveOrganization(ctx context.Context, id string) (Organization, error) {
	if id == "" || strings.ContainsAny(id, "/?#") {
		return Organization{}, errors.New("invalid Clerk organization ID")
	}
	var org Organization
	if err := c.call(ctx, http.MethodGet, "/organizations/"+url.PathEscape(id), nil, &org); err != nil {
		return Organization{}, err
	}
	if org.ID != id {
		return Organization{}, errors.New("the organization from Clerk has a different ID than the one requested")
	}
	return org, nil
}

// backendStatusError is a Backend API response other than 200 OK.
type backendStatusError struct{ status int }

func (e backendStatusError) Error() string {
	if e.status == http.StatusUnauthorized || e.status == http.StatusForbidden {
		return fmt.Sprintf("Clerk Backend API returned HTTP %d: check the Clerk secret key", e.status)
	}
	return fmt.Sprintf("Clerk Backend API returned HTTP %d", e.status)
}

// call sends body (if any) as JSON to the Backend API path under /v1 and decodes a 200
// response into output. Errors never include the secret key or the response body.
func (c ClerkOrganizations) call(ctx context.Context, method, path string, body, output any) error {
	if c.SecretKey == "" {
		return errors.New("backend Clerk secret is not configured")
	}
	var payload io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return err
		}
		payload = bytes.NewReader(encoded)
	}
	req, err := http.NewRequestWithContext(ctx, method, "https://api.clerk.com/v1"+path, payload)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.SecretKey)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	response, err := withoutRedirects(c.HTTPClient).Do(req)
	if err != nil {
		return fmt.Errorf("cannot reach Clerk Backend API: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return backendStatusError{status: response.StatusCode}
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(output); err != nil {
		return errors.New("invalid Clerk Backend API response")
	}
	return nil
}

// Membership is one Clerk organization a user belongs to, with their role in it.
type Membership struct {
	OrgID   string `json:"orgId"`
	OrgSlug string `json:"orgSlug"`
	OrgName string `json:"orgName"`
	Role    string `json:"role"`
}

// membershipPageSize is how many memberships each Backend API page asks for.
const membershipPageSize = 100

// UserMemberships lists every organization clerkUserID belongs to, paging by offset.
func (c ClerkOrganizations) UserMemberships(ctx context.Context, clerkUserID string) ([]Membership, error) {
	if err := validUserID(clerkUserID); err != nil {
		return nil, err
	}
	memberships := []Membership{}
	for {
		var page struct {
			Data []struct {
				Role         string `json:"role"`
				Organization struct {
					ID   string `json:"id"`
					Slug string `json:"slug"`
					Name string `json:"name"`
				} `json:"organization"`
			} `json:"data"`
			TotalCount int `json:"total_count"`
		}
		path := fmt.Sprintf("/users/%s/organization_memberships?limit=%d&offset=%d", url.PathEscape(clerkUserID), membershipPageSize, len(memberships))
		if err := c.call(ctx, http.MethodGet, path, nil, &page); err != nil {
			return nil, err
		}
		for _, m := range page.Data {
			memberships = append(memberships, Membership{OrgID: m.Organization.ID, OrgSlug: m.Organization.Slug, OrgName: m.Organization.Name, Role: m.Role})
		}
		if len(page.Data) == 0 || len(memberships) >= page.TotalCount {
			return memberships, nil
		}
	}
}

// validUserID rejects a Clerk user ID that would change the Backend API path.
func validUserID(id string) error {
	if id == "" || strings.ContainsAny(id, "/?#") {
		return errors.New("invalid Clerk user ID")
	}
	return nil
}
