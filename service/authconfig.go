package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// AuthConfig persists a GitHub token to ~/.config/gavel/auth.json so the
// background pr-ui daemon can authenticate GitHub API calls without the user
// propagating GITHUB_TOKEN/GH_TOKEN into the launchd/systemd service
// environment (neither inherits the user's shell env).
//
// The file is written with 0600 perms — tokens are credentials.
type AuthConfig struct {
	Token string `json:"token,omitempty"`
}

const authConfigFile = "auth.json"

// AuthConfigPath returns the absolute path of the persisted auth config.
func AuthConfigPath() (string, error) {
	dir, err := StateDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, authConfigFile), nil
}

// LoadAuthConfig reads auth.json. A missing file yields a zero-value
// AuthConfig (no error) so the token resolver can fall through to other
// sources cleanly.
func LoadAuthConfig() (AuthConfig, error) {
	path, err := AuthConfigPath()
	if err != nil {
		return AuthConfig{}, err
	}
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return AuthConfig{}, nil
		}
		return AuthConfig{}, fmt.Errorf("read %s: %w", path, err)
	}
	var cfg AuthConfig
	if err := json.Unmarshal(b, &cfg); err != nil {
		return AuthConfig{}, fmt.Errorf("parse %s: %w", path, err)
	}
	return cfg, nil
}

// SaveAuthConfig persists cfg to auth.json with 0600 perms.
func SaveAuthConfig(cfg AuthConfig) error {
	if cfg.Token == "" {
		return errors.New("refusing to save empty token")
	}
	path, err := AuthConfigPath()
	if err != nil {
		return err
	}
	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal auth config: %w", err)
	}
	if err := os.WriteFile(path, b, 0o600); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

// TokenSource names where DiscoverGitHubToken found a token. Exposed so the
// install command can log "Discovered token from GITHUB_TOKEN" etc.
type TokenSource string

const (
	TokenSourceEnvGitHub TokenSource = "GITHUB_TOKEN"
	TokenSourceEnvGH     TokenSource = "GH_TOKEN"
	TokenSourceGHCLI     TokenSource = "gh auth token"
)

// DiscoverGitHubToken checks the usual sources in order and returns the
// first match along with its source. Returns ("", "", nil) when no source
// yielded a token — callers treat that as "user must supply one".
//
// Order: GITHUB_TOKEN env → GH_TOKEN env → `gh auth token` CLI. The gh CLI
// is used as a last resort because it spawns a subprocess — but it's the
// most common way developers authenticate today, so worth the overhead at
// install time.
func DiscoverGitHubToken() (string, TokenSource, error) {
	if t := strings.TrimSpace(os.Getenv("GITHUB_TOKEN")); t != "" {
		return t, TokenSourceEnvGitHub, nil
	}
	if t := strings.TrimSpace(os.Getenv("GH_TOKEN")); t != "" {
		return t, TokenSourceEnvGH, nil
	}
	if t, err := tokenFromGHCLI(); err != nil {
		return "", "", fmt.Errorf("probe gh auth token: %w", err)
	} else if t != "" {
		return t, TokenSourceGHCLI, nil
	}
	return "", "", nil
}

// tokenFromGHCLI runs `gh auth token` if the gh CLI is on PATH. Not having
// gh or not being authenticated are both valid "no token" states — only
// unexpected failures (exec error other than NotFound / ExitError) bubble up.
func tokenFromGHCLI() (string, error) {
	path, err := exec.LookPath("gh")
	if err != nil {
		return "", nil
	}
	out, err := exec.Command(path, "auth", "token").Output()
	if err != nil {
		// ExitError means gh ran but returned non-zero (typically "not
		// logged in") — not a real error for our purposes.
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			return "", nil
		}
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
