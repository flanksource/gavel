// Reads per-user data from the Clerk Backend API: a user's profile, and the GitHub
// OAuth token they granted when signing in with GitHub.
package api

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
)

// ClerkUser is the profile gavel keeps for a Clerk user.
type ClerkUser struct {
	Name          string
	Email         string
	EmailVerified bool
	// GitHubLogin is the username of the user's GitHub connection, empty without one.
	GitHubLogin string
}

// User reads clerkUserID's profile: name, primary email, and GitHub username.
func (c ClerkOrganizations) User(ctx context.Context, clerkUserID string) (ClerkUser, error) {
	if err := validUserID(clerkUserID); err != nil {
		return ClerkUser{}, err
	}
	var user struct {
		FirstName      string `json:"first_name"`
		LastName       string `json:"last_name"`
		Username       string `json:"username"`
		PrimaryEmailID string `json:"primary_email_address_id"`
		EmailAddresses []struct {
			ID           string `json:"id"`
			EmailAddress string `json:"email_address"`
			Verification struct {
				Status string `json:"status"`
			} `json:"verification"`
		} `json:"email_addresses"`
		ExternalAccounts []struct {
			Provider string `json:"provider"`
			Username string `json:"username"`
		} `json:"external_accounts"`
	}
	if err := c.call(ctx, http.MethodGet, "/users/"+url.PathEscape(clerkUserID), nil, &user); err != nil {
		return ClerkUser{}, err
	}
	profile := ClerkUser{Name: strings.TrimSpace(user.FirstName + " " + user.LastName)}
	if profile.Name == "" {
		profile.Name = user.Username
	}
	for _, email := range user.EmailAddresses {
		if email.ID == user.PrimaryEmailID {
			profile.Email, profile.EmailVerified = email.EmailAddress, email.Verification.Status == "verified"
		}
	}
	for _, account := range user.ExternalAccounts {
		if account.Provider == "oauth_github" {
			profile.GitHubLogin = account.Username
		}
	}
	return profile, nil
}

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
