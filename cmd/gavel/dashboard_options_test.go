// Tests `gavel serve` dashboard option handling: env fallbacks for Clerk settings and
// secrets, and Validate's rules for --auth none and --auth clerk.
package main

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testSecretKey = base64.StdEncoding.EncodeToString([]byte(strings.Repeat("k", 32)))

func clerkDashboardOptions() DashboardOptions {
	return DashboardOptions{
		Auth:                "clerk",
		PublicURL:           "https://gavel.example.com",
		ClerkIssuer:         "https://clerk.example.com",
		ClerkPublishableKey: "pk_test_abc",
		ClerkClientID:       "cli",
		clerkSecretKey:      "sk_test_abc",
		secretKey:           testSecretKey,
	}
}

func TestDashboardOptionsValidate(t *testing.T) {
	for _, test := range []struct {
		name   string
		modify func(*DashboardOptions)
		want   []string
	}{
		{"none by default", func(o *DashboardOptions) { *o = DashboardOptions{Auth: "none"} }, nil},
		{"clerk fully configured", func(*DashboardOptions) {}, nil},
		{"clerk with an origin-only public URL and port", func(o *DashboardOptions) { o.PublicURL = "http://localhost:9092" }, nil},
		{"unknown auth mode", func(o *DashboardOptions) { o.Auth = "oauth" }, []string{`--auth must be "none" or "clerk", got "oauth"`}},
		{"clerk without anything", func(o *DashboardOptions) { *o = DashboardOptions{Auth: "clerk"} }, []string{
			"--auth clerk requires --public-url",
			"--auth clerk requires --clerk-issuer or GAVEL_CLERK_ISSUER",
			"--auth clerk requires --clerk-publishable-key or GAVEL_CLERK_PUBLISHABLE_KEY",
			"--auth clerk requires --clerk-client-id or GAVEL_CLERK_CLIENT_ID",
			"--auth clerk requires the CLERK_SECRET_KEY environment variable",
			"--auth clerk requires the GAVEL_SECRET_KEY environment variable",
		}},
		{"menu bar with clerk", func(o *DashboardOptions) { o.MenuBar = true }, []string{"--menu-bar is local-only and can't be used with --auth clerk"}},
		{"dev mode with clerk", func(o *DashboardOptions) { o.Dev = true }, []string{"--dev is local-only and can't be used with --auth clerk"}},
		{"push server with clerk", func(o *DashboardOptions) { o.SSHAddr = ":2222" }, nil},
		{"push server without clerk", func(o *DashboardOptions) { *o = DashboardOptions{Auth: "none", SSHAddr: ":2222"} }, []string{
			"--ssh-addr serves hosted pushes and needs --auth clerk; use gavel ssh serve locally",
		}},
		{"push server address without a port", func(o *DashboardOptions) { o.SSHAddr = "localhost" }, []string{
			`--ssh-addr must be host:port like :2222, got "localhost"`,
		}},
		{"GitHub App with clerk", func(o *DashboardOptions) {
			o.GitHubAppID, o.GitHubAppSlug, o.GitHubAppPrivateKey = "123", "gavel-dev", "/keys/app.pem"
		}, nil},
		{"GitHub App without clerk", func(o *DashboardOptions) { *o = DashboardOptions{Auth: "none", GitHubAppID: "123"} }, []string{
			"--github-app-* flags serve hosted orgs and need --auth clerk",
		}},
		{"GitHub App ID that isn't a number", func(o *DashboardOptions) {
			o.GitHubAppID, o.GitHubAppSlug, o.GitHubAppPrivateKey = "gavel-dev", "gavel-dev", "/keys/app.pem"
		}, []string{`--github-app-id must be the App's numeric ID (from its settings page), got "gavel-dev"`}},
		{"partial GitHub App", func(o *DashboardOptions) { o.GitHubAppID = "123" }, []string{
			"--github-app-id, --github-app-slug and --github-app-private-key (or GAVEL_GITHUB_APP_PRIVATE_KEY_FILE) go together; missing --github-app-slug, --github-app-private-key",
		}},
		{"public URL with a path", func(o *DashboardOptions) { o.PublicURL = "https://gavel.example.com/app" }, []string{
			`--public-url must be an origin like https://gavel.example.com, got "https://gavel.example.com/app"`,
		}},
		{"public URL without a scheme", func(o *DashboardOptions) { o.PublicURL = "gavel.example.com" }, []string{
			`--public-url must be an origin like https://gavel.example.com, got "gavel.example.com"`,
		}},
		{"plain HTTP issuer", func(o *DashboardOptions) { o.ClerkIssuer = "http://clerk.example.com" }, []string{
			`--clerk-issuer must be an https URL, got "http://clerk.example.com"`,
		}},
		{"secret key as publishable key", func(o *DashboardOptions) { o.ClerkPublishableKey = "sk_test_abc" }, []string{
			"--clerk-publishable-key must start with pk_",
		}},
		{"publishable key as secret key", func(o *DashboardOptions) { o.clerkSecretKey = "pk_test_abc" }, []string{
			"CLERK_SECRET_KEY must start with sk_",
		}},
		{"short GAVEL_SECRET_KEY", func(o *DashboardOptions) { o.secretKey = base64.StdEncoding.EncodeToString([]byte("short")) }, []string{
			"GAVEL_SECRET_KEY must be at least 32 random bytes, base64-encoded (got 5 bytes)",
		}},
		{"GAVEL_SECRET_KEY not base64", func(o *DashboardOptions) { o.secretKey = "not base64!" }, []string{
			"GAVEL_SECRET_KEY must be at least 32 random bytes, base64-encoded (it isn't valid base64)",
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			opts := clerkDashboardOptions()
			test.modify(&opts)
			err := opts.Validate()
			if len(test.want) == 0 {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			for _, want := range test.want {
				assert.Contains(t, err.Error(), want)
			}
			assert.NotContains(t, err.Error(), "sk_test_abc", "errors must not echo secrets")
		})
	}
}

func TestDashboardOptionsFromEnv(t *testing.T) {
	env := map[string]string{
		"GAVEL_CLERK_ISSUER":                "https://env-clerk.example.com",
		"GAVEL_CLERK_PUBLISHABLE_KEY":       "pk_env",
		"GAVEL_CLERK_CLIENT_ID":             "env-client",
		"CLERK_SECRET_KEY":                  "sk_env",
		"GAVEL_SECRET_KEY":                  testSecretKey,
		"GAVEL_GITHUB_APP_PRIVATE_KEY_FILE": "/keys/app.pem",
	}
	got := DashboardOptions{ClerkIssuer: "https://flag-clerk.example.com"}.withEnv(func(key string) string { return env[key] })
	assert.Equal(t, "https://flag-clerk.example.com", got.ClerkIssuer, "a flag wins over its env var")
	assert.Equal(t, "pk_env", got.ClerkPublishableKey)
	assert.Equal(t, "env-client", got.ClerkClientID)
	assert.Equal(t, "sk_env", got.clerkSecretKey)
	assert.Equal(t, testSecretKey, got.secretKey)
	assert.Equal(t, "/keys/app.pem", got.GitHubAppPrivateKey)
}

func TestServeCommandHasAuthFlags(t *testing.T) {
	serve := findCommand(rootCmd, "serve")
	require.NotNil(t, serve)
	for _, name := range []string{"auth", "public-url", "clerk-issuer", "clerk-publishable-key", "clerk-client-id", "ssh-addr", "ssh-host-key", "data-dir", "github-app-id", "github-app-slug", "github-app-private-key", "dev", "port"} {
		assert.NotNilf(t, serve.Flags().Lookup(name), "serve should expose --%s", name)
	}
	assert.Nil(t, serve.Flags().Lookup("trusted-proxy"), "nothing reads forwarded headers: hosted origins come from --public-url")
	assert.Equal(t, "none", serve.Flags().Lookup("auth").DefValue)
	for _, name := range []string{"clerk-secret-key", "secret-key"} {
		assert.Nilf(t, serve.Flags().Lookup(name), "secrets come from env only, not --%s", name)
	}
}
