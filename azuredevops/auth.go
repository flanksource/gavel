// The az CLI login a local gavel reaches Azure DevOps with when no git connection or
// AZURE_DEVOPS_EXT_PAT does: its access token, fetched again once it nears expiry.
package azuredevops

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type accessToken struct {
	Token   string
	Expires time.Time
}

// authenticator caches the az CLI login's access token until it nears expiry.
type authenticator struct {
	mu     sync.Mutex
	fetch  func(context.Context) (accessToken, error)
	now    func() time.Time
	cached accessToken
}

// cliToken is the az CLI login's access token, fetched again once it nears expiry.
func (a *authenticator) cliToken(ctx context.Context) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.fetch == nil || a.now == nil {
		return "", fmt.Errorf("azure authentication resolver is not configured")
	}
	if a.cached.Token == "" || !a.cached.Expires.After(a.now().Add(time.Minute)) {
		token, err := a.fetch(ctx)
		if err != nil {
			return "", err
		}
		if token.Token == "" || !token.Expires.After(a.now()) {
			return "", fmt.Errorf("azure CLI returned an empty or expired access token")
		}
		a.cached = token
	}
	return a.cached.Token, nil
}

func cliAccessToken(ctx context.Context) (accessToken, error) {
	out, err := exec.CommandContext(ctx, "az", "account", "get-access-token", "--resource", "499b84ac-1321-427f-aa17-267ca6975798", "--output", "json").Output()
	if err != nil {
		return accessToken{}, fmt.Errorf("azure authentication: set AZURE_DEVOPS_EXT_PAT or run az login: %w", err)
	}
	var response struct {
		Token        string `json:"accessToken"`
		ExpiresUnix  int64  `json:"expires_on"`
		ExpiresLocal string `json:"expiresOn"`
	}
	if err := json.Unmarshal(out, &response); err != nil {
		return accessToken{}, fmt.Errorf("decode Azure CLI credentials: %w", err)
	}
	expires := time.Unix(response.ExpiresUnix, 0)
	if response.ExpiresUnix == 0 {
		expires, err = time.ParseInLocation("2006-01-02 15:04:05", strings.TrimSpace(response.ExpiresLocal), time.Local)
		if err != nil {
			return accessToken{}, fmt.Errorf("decode Azure CLI token expiration: %w", err)
		}
	}
	return accessToken{Token: response.Token, Expires: expires}, nil
}
