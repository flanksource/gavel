package github

import (
	"context"
	"encoding/json"
	"fmt"
	nethttp "net/http"
	"time"
)

// Org is a lightweight view of a GitHub organization the authenticated user
// belongs to. Surfaced by /api/orgs so the PR UI's org chooser can list
// memberships with avatars.
type Org struct {
	Login     string `json:"login"`
	AvatarURL string `json:"avatarUrl"`
}

// FetchUserOrgs calls GET /user/orgs and returns the authenticated user's
// org memberships. Returns an empty slice (not an error) when the user
// belongs to no orgs — that's a valid state, not a failure.
//
// This is not paginated: users with more than 30 org memberships get the
// first page. If we ever hit that in practice we can add page follow-up.
func FetchUserOrgs(opts Options) ([]Org, error) {
	return fetchUserOrgs(opts, githubAPIBase())
}

// FetchUserLogin calls GET /user and returns the authenticated user's
// login. Used as a fallback org when the token has no org memberships —
// `SearchPRs` can scope `org:<login>` for solo developers whose PRs live
// under their personal account.
func FetchUserLogin(opts Options) (string, error) {
	return fetchUserLogin(opts, githubAPIBase())
}

// ResolveDefaultOrg returns the org that `--all` should default to when
// no explicit `--org=...` is supplied. It asks GitHub — never the local
// git remote — so gavel works from any directory:
//
//  1. `/user/orgs`: if the authenticated user has org memberships, return
//     the first one NOT in `ignored`. Good enough for most multi-org
//     setups; users override via --org or the header chooser.
//  2. `/user.login`: solo developers (or users whose only orgs are all
//     ignored) fall through to their own login; their PRs search-scope
//     as `org:<username>`.
//
// Errors bubble up only when BOTH probes fail — the caller sees a real
// authentication/network problem, not a "no orgs here" false positive.
func ResolveDefaultOrg(opts Options, ignored []string) (string, error) {
	return resolveDefaultOrg(opts, githubAPIBase(), ignored)
}

func resolveDefaultOrg(opts Options, baseURL string, ignored []string) (string, error) {
	skip := make(map[string]bool, len(ignored))
	for _, o := range ignored {
		skip[o] = true
	}

	orgs, orgErr := fetchUserOrgs(opts, baseURL)
	if orgErr == nil {
		for _, o := range orgs {
			if !skip[o.Login] {
				return o.Login, nil
			}
		}
		// All orgs ignored — fall through to /user.login below. This is
		// intentional: it lets a user with only "noise" orgs still scope
		// their search to their own PRs.
	}
	login, userErr := fetchUserLogin(opts, baseURL)
	if userErr == nil && login != "" {
		return login, nil
	}
	if orgErr != nil {
		return "", fmt.Errorf("resolve default org: %w", orgErr)
	}
	if userErr != nil {
		return "", fmt.Errorf("resolve default org from /user: %w", userErr)
	}
	return "", fmt.Errorf("resolve default org: GitHub returned no orgs and no user login")
}

func fetchUserLogin(opts Options, baseURL string) (string, error) {
	token, err := opts.token()
	if err != nil {
		return "", err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, err := nethttp.NewRequestWithContext(ctx, "GET", baseURL+"/user", nil)
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := nethttp.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("GET /user: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != nethttp.StatusOK {
		return "", fmt.Errorf("GET /user: status %d", resp.StatusCode)
	}
	var body struct {
		Login string `json:"login"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", fmt.Errorf("decode /user: %w", err)
	}
	return body.Login, nil
}

func fetchUserOrgs(opts Options, baseURL string) ([]Org, error) {
	token, err := opts.token()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, err := nethttp.NewRequestWithContext(ctx, "GET", baseURL+"/user/orgs?per_page=100", nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := nethttp.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GET /user/orgs: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != nethttp.StatusOK {
		return nil, fmt.Errorf("GET /user/orgs: status %d", resp.StatusCode)
	}
	var raw []struct {
		Login     string `json:"login"`
		AvatarURL string `json:"avatar_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decode /user/orgs: %w", err)
	}
	orgs := make([]Org, 0, len(raw))
	for _, o := range raw {
		orgs = append(orgs, Org{Login: o.Login, AvatarURL: o.AvatarURL})
	}
	return orgs, nil
}
