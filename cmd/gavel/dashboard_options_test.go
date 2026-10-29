// Tests `gavel serve` dashboard option handling: env fallbacks for Clerk settings and
// secrets, and Validate's rules for --auth none and --auth clerk.
package main

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/flanksource/gavel/internal/logtest"
	"github.com/google/uuid"
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
		{"docker sandbox", func(o *DashboardOptions) {
			o.Sandbox, o.SandboxImage, o.SandboxServeAddr = "docker", "ghcr.io/flanksource/gavel-runner:1", "10.0.0.5:9092"
		}, nil},
		{"kubernetes sandbox", func(o *DashboardOptions) { o.Sandbox, o.SandboxImage = "kubernetes", "runner:1" }, nil},
		{"substrate sandbox", func(o *DashboardOptions) {
			o.Sandbox, o.SandboxImage, o.SubstrateGavelName, o.credentialProviderToken = "substrate", "runner@sha256:1", "beta", "provider-token"
		}, nil},
		{"substrate sandbox without an image", func(o *DashboardOptions) {
			o.Sandbox, o.SubstrateGavelName, o.credentialProviderToken = "substrate", "beta", "provider-token"
		}, []string{"--sandbox requires --sandbox-image"}},
		{"substrate sandbox without a gavel name or provider token", func(o *DashboardOptions) { o.Sandbox, o.SandboxImage = "substrate", "runner@sha256:1" }, []string{
			"--sandbox substrate requires --substrate-gavel-name, the name substrate's credential provider knows this gavel by",
			"--sandbox substrate requires the GAVEL_CREDENTIAL_PROVIDER_TOKEN environment variable, the token substrate's credential provider fetches model keys with",
		}},
		{"substrate gavel name that isn't a DNS label", func(o *DashboardOptions) {
			o.Sandbox, o.SandboxImage, o.SubstrateGavelName, o.credentialProviderToken = "substrate", "runner@sha256:1", "Beta/1", "provider-token"
		}, []string{`--substrate-gavel-name must be a DNS label like beta, got "Beta/1"`}},
		{"unknown sandbox", func(o *DashboardOptions) { o.Sandbox, o.SandboxImage = "firecracker", "runner:1" }, []string{
			`--sandbox must be docker, podman, kubernetes or substrate, got "firecracker"`,
		}},
		{"sandbox without an image", func(o *DashboardOptions) { o.Sandbox = "kubernetes" }, []string{"--sandbox requires --sandbox-image"}},
		{"model keys for sandbox jobs", func(o *DashboardOptions) {
			o.Sandbox, o.SandboxImage, o.SandboxSecretEnv = "kubernetes", "runner:1", []string{"ANTHROPIC_API_KEY", "OPENAI_API_KEY"}
		}, nil},
		{"secrets without a sandbox", func(o *DashboardOptions) { o.SandboxSecretEnv = []string{"ANTHROPIC_API_KEY"} }, []string{
			"--sandbox-secret-env copies secrets into sandbox jobs and needs --sandbox",
		}},
		{"a secret named like a job's own", func(o *DashboardOptions) {
			o.Sandbox, o.SandboxImage, o.SandboxSecretEnv = "kubernetes", "runner:1", []string{"GAVEL_RUN_TOKEN", "not a name"}
		}, []string{
			"--sandbox-secret-env GAVEL_RUN_TOKEN: a job's GAVEL_ variables are its own",
			`--sandbox-secret-env "not a name" is not an environment variable name`,
		}},
		{"container sandbox without the serve address", func(o *DashboardOptions) { o.Sandbox, o.SandboxImage = "podman", "runner:1" }, []string{
			"--sandbox podman requires --sandbox-serve-addr, the IP:port sandbox containers reach gavel serve at",
		}},
		{"sandbox without clerk", func(o *DashboardOptions) { *o = DashboardOptions{Auth: "none", Sandbox: "docker"} }, []string{
			"--sandbox runs hosted push checks and needs --auth clerk",
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

func TestSandboxSecretsCopyServesEnvironment(t *testing.T) {
	env := map[string]string{"ANTHROPIC_API_KEY": "sk-ant", "OPENAI_API_KEY": ""}
	getenv := func(name string) string { return env[name] }

	secrets, err := sandboxSecrets([]string{"ANTHROPIC_API_KEY"}, getenv)
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"ANTHROPIC_API_KEY": "sk-ant"}, secrets)

	_, err = sandboxSecrets([]string{"ANTHROPIC_API_KEY", "OPENAI_API_KEY"}, getenv)
	require.EqualError(t, err, "--sandbox-secret-env OPENAI_API_KEY is not set in gavel serve's environment")
}

func TestSubstrateSandboxRefusesSecretsInItsJobs(t *testing.T) {
	opts := DashboardOptions{Sandbox: "substrate", SandboxSecretEnv: []string{"ANTHROPIC_API_KEY"},
		SubstrateEnvAPI: "ate-env-api.ate-env.svc:7777", SubstrateControl: "api.ate-system.svc:443"}
	_, err := newSandboxDispatcher(t.Context(), opts, nil, nil, nil)
	require.EqualError(t, err, "--sandbox substrate keeps secrets out of its jobs: drop --sandbox-secret-env; organizations set their own model keys, which substrate's egress gateway injects")

	opts.SandboxSecretEnv = nil
	opts.SubstrateGavelName = "beta"
	dispatcher, err := newSandboxDispatcher(t.Context(), opts, nil, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, "substrate", dispatcher.Executor.Name())
	run := uuid.MustParse("6f1c2d3e-4a5b-6c7d-8e9f-0a1b2c3d4e5f")
	require.NotNil(t, dispatcher.CredentialURI, "substrate jobs get their organization's keys")
	assert.Equal(t, "ate-secret://gavel/beta/6f1c2d3e-4a5b-6c7d-8e9f-0a1b2c3d4e5f/anthropic", dispatcher.CredentialURI(run, "anthropic"))
}

func TestContainerAndKubernetesSandboxesWarnTheyUseServesKeys(t *testing.T) {
	for _, test := range []struct {
		opts DashboardOptions
		want string
	}{
		{DashboardOptions{Sandbox: "kubernetes", SandboxSecretEnv: []string{"ANTHROPIC_API_KEY", "OPENAI_API_KEY"}},
			"--sandbox kubernetes: every organization's runs use serve's --sandbox-secret-env keys (ANTHROPIC_API_KEY, OPENAI_API_KEY), not their own; organization model keys need --sandbox substrate"},
		{DashboardOptions{Sandbox: "docker"},
			"--sandbox docker: runs get no model keys, since --sandbox-secret-env is unset; organization model keys need --sandbox substrate"},
		{DashboardOptions{Sandbox: "substrate"}, ""},
	} {
		assert.Equal(t, test.want, sandboxKeysWarning(test.opts), test.opts.Sandbox)
	}

	logged := logtest.Capture(t)
	_, err := newSandboxDispatcher(t.Context(), DashboardOptions{
		Sandbox: "docker", SandboxImage: "runner:1", SandboxServeAddr: "10.0.0.5:9092",
		SandboxNetwork: "gavel-sandbox-missing-" + uuid.NewString()[:8],
	}, nil, nil, nil)
	require.Error(t, err, "the network doesn't exist")
	assert.Contains(t, logged.String(), "--sandbox docker: runs get no model keys", "logged at startup, before the executor is checked")
}
