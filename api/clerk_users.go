// Reads per-user credentials from the Clerk Backend API, such as the GitHub OAuth
// token a user granted when signing in with GitHub.
package api

import (
	"context"
	"errors"
	"net/http"
	"net/url"
)

// GitHubToken returns clerkUserID's GitHub OAuth access token. Clerk refreshes an
// expired token before returning it.
func (c ClerkOrganizations) GitHubToken(ctx context.Context, clerkUserID string) (string, error) {
	if err := validUserID(clerkUserID); err != nil {
		return "", err
	}
	var tokens []struct {
		Token string `json:"token"`
	}
	if err := c.call(ctx, http.MethodGet, "/users/"+url.PathEscape(clerkUserID)+"/oauth_access_tokens/oauth_github", nil, &tokens); err != nil {
		return "", err
	}
	if len(tokens) == 0 || tokens[0].Token == "" {
		return "", errors.New("the user has no GitHub connection in Clerk")
	}
	return tokens[0].Token, nil
}
