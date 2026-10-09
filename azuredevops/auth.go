package azuredevops

import (
	"context"
	"encoding/base64"
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

type authenticator struct {
	mu     sync.Mutex
	pat    string
	fetch  func(context.Context) (accessToken, error)
	now    func() time.Time
	cached accessToken
}

func (a *authenticator) authorization(ctx context.Context) (string, error) {
	if a.pat != "" {
		return "Basic " + base64.StdEncoding.EncodeToString([]byte(":"+a.pat)), nil
	}
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
	return "Bearer " + a.cached.Token, nil
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
