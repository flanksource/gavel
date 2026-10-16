// Authenticates as a GitHub App: mints App JWTs, caches installation access tokens,
// and reads an installation's account and repositories.
package github

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// installationTokenMargin is how long before expiry a cached installation token is replaced.
const installationTokenMargin = 5 * time.Minute

// App mints GitHub App JWTs and caches installation access tokens. It is safe for concurrent use.
type App struct {
	ID   int64
	Slug string
	Key  *rsa.PrivateKey
	HTTP *http.Client
	now  func() time.Time

	mu     sync.Mutex
	tokens map[int64]cachedToken
}

type cachedToken struct {
	token   string
	expires time.Time
}

// Installation is a GitHub App installation and the account it is installed on.
type Installation struct {
	ID      int64               `json:"id"`
	Account InstallationAccount `json:"account"`
}

// InstallationAccount is the user or organization an installation belongs to.
type InstallationAccount struct {
	Login string `json:"login"`
	ID    int64  `json:"id"`
	Type  string `json:"type"`
}

// InstallationRepo is a repository an installation can access.
type InstallationRepo struct {
	ID            int64  `json:"id"`
	FullName      string `json:"full_name"`
	Private       bool   `json:"private"`
	DefaultBranch string `json:"default_branch"`
}

// LoadApp reads the App's PEM private key (PKCS#1 or PKCS#8) from keyPath.
func LoadApp(id int64, slug, keyPath string) (*App, error) {
	if id <= 0 {
		return nil, fmt.Errorf("the GitHub App ID must be a positive number, got %d", id)
	}
	data, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("read the GitHub App private key: %w", err)
	}
	key, err := jwt.ParseRSAPrivateKeyFromPEM(data)
	if err != nil {
		return nil, fmt.Errorf("parse the GitHub App private key %s: %w", keyPath, err)
	}
	return &App{ID: id, Slug: slug, Key: key, HTTP: &http.Client{Timeout: 20 * time.Second}, now: time.Now, tokens: map[int64]cachedToken{}}, nil
}

// InstallationToken returns an access token for installationID, minting a new one when
// the cached token is within installationTokenMargin of expiring.
func (a *App) InstallationToken(ctx context.Context, installationID int64) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if cached, ok := a.tokens[installationID]; ok && a.now().Before(cached.expires.Add(-installationTokenMargin)) {
		return cached.token, nil
	}
	appJWT, err := a.jwt()
	if err != nil {
		return "", err
	}
	var minted struct {
		Token     string    `json:"token"`
		ExpiresAt time.Time `json:"expires_at"`
	}
	path := fmt.Sprintf("/app/installations/%d/access_tokens", installationID)
	if err := a.call(ctx, http.MethodPost, path, appJWT, &minted); err != nil {
		return "", fmt.Errorf("mint a token for GitHub App installation %d: %w", installationID, err)
	}
	if minted.Token == "" {
		return "", fmt.Errorf("mint a token for GitHub App installation %d: GitHub returned no token", installationID)
	}
	a.tokens[installationID] = cachedToken{token: minted.Token, expires: minted.ExpiresAt}
	return minted.Token, nil
}

// Installation reads installationID's account.
func (a *App) Installation(ctx context.Context, installationID int64) (Installation, error) {
	appJWT, err := a.jwt()
	if err != nil {
		return Installation{}, err
	}
	var installation Installation
	if err := a.call(ctx, http.MethodGet, fmt.Sprintf("/app/installations/%d", installationID), appJWT, &installation); err != nil {
		return Installation{}, fmt.Errorf("read GitHub App installation %d: %w", installationID, err)
	}
	return installation, nil
}

// InstallationRepos lists every repository installationID can access, page by page.
func (a *App) InstallationRepos(ctx context.Context, installationID int64) ([]InstallationRepo, error) {
	token, err := a.InstallationToken(ctx, installationID)
	if err != nil {
		return nil, err
	}
	var repos []InstallationRepo
	for page := 1; ; page++ {
		var body struct {
			TotalCount   int                `json:"total_count"`
			Repositories []InstallationRepo `json:"repositories"`
		}
		path := "/installation/repositories?per_page=100&page=" + strconv.Itoa(page)
		if err := a.call(ctx, http.MethodGet, path, token, &body); err != nil {
			return nil, fmt.Errorf("list the repositories of GitHub App installation %d: %w", installationID, err)
		}
		repos = append(repos, body.Repositories...)
		if len(body.Repositories) == 0 || len(repos) >= body.TotalCount {
			return repos, nil
		}
	}
}

// UserInstallations lists this App's installations that userToken (a GitHub App user
// token, e.g. from Clerk) can access, page by page.
func (a *App) UserInstallations(ctx context.Context, userToken string) ([]Installation, error) {
	var installations []Installation
	seen := 0
	for page := 1; ; page++ {
		var body struct {
			TotalCount    int `json:"total_count"`
			Installations []struct {
				Installation
				AppID int64 `json:"app_id"`
			} `json:"installations"`
		}
		if err := a.call(ctx, http.MethodGet, "/user/installations?per_page=100&page="+strconv.Itoa(page), userToken, &body); err != nil {
			return nil, fmt.Errorf("list the user's GitHub App installations: %w", err)
		}
		for _, installation := range body.Installations {
			if installation.AppID == a.ID {
				installations = append(installations, installation.Installation)
			}
		}
		seen += len(body.Installations)
		if len(body.Installations) == 0 || seen >= body.TotalCount {
			return installations, nil
		}
	}
}

// jwt mints an App JWT: issued a minute in the past for clock skew, valid nine minutes.
func (a *App) jwt() (string, error) {
	now := a.now()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.RegisteredClaims{
		Issuer:    strconv.FormatInt(a.ID, 10),
		IssuedAt:  jwt.NewNumericDate(now.Add(-time.Minute)),
		ExpiresAt: jwt.NewNumericDate(now.Add(9 * time.Minute)),
	})
	signed, err := token.SignedString(a.Key)
	if err != nil {
		return "", fmt.Errorf("sign the GitHub App JWT: %w", err)
	}
	return signed, nil
}

// call sends an authenticated request to the GitHub API and decodes a 2xx response.
// Errors carry GitHub's message, never the credential.
func (a *App) call(ctx context.Context, method, path, bearer string, output any) error {
	req, err := http.NewRequestWithContext(ctx, method, githubAPIBase()+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+bearer)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	resp, err := a.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("cannot reach GitHub: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return fmt.Errorf("read GitHub's response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		var failure struct {
			Message string `json:"message"`
		}
		if json.Unmarshal(body, &failure) == nil && failure.Message != "" {
			return fmt.Errorf("HTTP %d: %s", resp.StatusCode, failure.Message)
		}
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	if err := json.Unmarshal(body, output); err != nil {
		return errors.New("GitHub returned an invalid response")
	}
	return nil
}
