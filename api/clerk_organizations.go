package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// ClerkOrganizations resolves canonical org metadata on a trusted backend.
// SecretKey must remain on that backend; never ship/persist it in CLI sessions.
// Configure one resolver per issuer/Clerk instance.
type ClerkOrganizations struct {
	SecretKey  string
	HTTPClient *http.Client
}

func (c ClerkOrganizations) ResolveOrganization(ctx context.Context, id string) (Organization, error) {
	if c.SecretKey == "" {
		return Organization{}, errors.New("backend Clerk secret is not configured")
	}
	if id == "" || strings.ContainsAny(id, "/?#") {
		return Organization{}, errors.New("invalid Clerk organization ID")
	}
	client := c.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 20 * time.Second}
	}
	copyClient := *client
	copyClient.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.clerk.com/v1/organizations/"+url.PathEscape(id), nil)
	if err != nil {
		return Organization{}, err
	}
	req.Header.Set("Authorization", "Bearer "+c.SecretKey)
	response, err := copyClient.Do(req)
	if err != nil {
		return Organization{}, errors.New("cannot reach Clerk organization service")
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return Organization{}, fmt.Errorf("Clerk organization service returned HTTP %d", response.StatusCode)
	}
	var org Organization
	if err := json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&org); err != nil {
		return Organization{}, errors.New("invalid Clerk organization response")
	}
	if org.ID != id {
		return Organization{}, errors.New("Clerk returned a different organization")
	}
	return org, nil
}

// Membership is one Clerk organization a user belongs to, with their role in it.
type Membership struct {
	OrgID   string `json:"orgId"`
	OrgSlug string `json:"orgSlug"`
	OrgName string `json:"orgName"`
	Role    string `json:"role"`
}
