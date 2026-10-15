// Options for `gavel serve`: the PR dashboard flags plus authentication. --auth none
// serves one local user; --auth clerk serves many organizations behind Clerk sign-in.
package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
)

const (
	authNone  = "none"
	authClerk = "clerk"
)

// DashboardOptions configures `gavel serve`. Secrets are read from the environment
// only (withEnv), so they never appear in flags, shell history or process lists.
type DashboardOptions struct {
	PRListOptions
	Auth                string `flag:"auth" help:"Authentication: none (local, single user) or clerk (hosted, one dashboard per Clerk organization)" default:"none"`
	PublicURL           string `flag:"public-url" help:"Origin users reach the dashboard at, e.g. https://gavel.example.com (required with --auth clerk)"`
	ClerkIssuer         string `flag:"clerk-issuer" help:"Clerk Frontend API URL (or GAVEL_CLERK_ISSUER)"`
	ClerkPublishableKey string `flag:"clerk-publishable-key" help:"Clerk publishable key for the browser sign-in (or GAVEL_CLERK_PUBLISHABLE_KEY)"`
	ClerkClientID       string `flag:"clerk-client-id" help:"Clerk public OAuth client ID that gavel login uses (or GAVEL_CLERK_CLIENT_ID)"`
	SSHAddr             string `flag:"ssh-addr" help:"Address for the hosted git push server, e.g. :2222 (--auth clerk only; gavel ssh serve is the local one)"`
	SSHHostKey          string `flag:"ssh-host-key" help:"Push server host key file; created if missing (default <data-dir>/ssh_host_key)"`
	DataDir             string `flag:"data-dir" help:"Directory for hosted repositories" default:"/var/lib/gavel"`
	GitHubAppID         string `flag:"github-app-id" help:"ID of the GitHub App organizations install to show their pull requests (--auth clerk)"`
	GitHubAppSlug       string `flag:"github-app-slug" help:"URL name of the GitHub App, as in github.com/apps/<slug>"`
	GitHubAppPrivateKey string `flag:"github-app-private-key" help:"Path to the GitHub App's private key PEM (or GAVEL_GITHUB_APP_PRIVATE_KEY_FILE)"`
	PropertiesFile      string `flag:"properties-file" help:"Properties file with retention.* windows such as retention.job_logs=60d (default /etc/gavel/gavel.properties when it exists)"`

	clerkSecretKey string
	secretKey      string
}

// withEnv fills settings left empty by flags from getenv, and reads the secrets.
func (o DashboardOptions) withEnv(getenv func(string) string) DashboardOptions {
	for _, setting := range []struct {
		value *string
		env   string
	}{
		{&o.ClerkIssuer, "GAVEL_CLERK_ISSUER"},
		{&o.ClerkPublishableKey, "GAVEL_CLERK_PUBLISHABLE_KEY"},
		{&o.ClerkClientID, "GAVEL_CLERK_CLIENT_ID"},
		{&o.GitHubAppPrivateKey, "GAVEL_GITHUB_APP_PRIVATE_KEY_FILE"},
	} {
		if *setting.value == "" {
			*setting.value = getenv(setting.env)
		}
	}
	o.clerkSecretKey = getenv("CLERK_SECRET_KEY")
	o.secretKey = getenv("GAVEL_SECRET_KEY")
	return o
}

// gitHubAppID is --github-app-id as GitHub's numeric App ID.
func (o DashboardOptions) gitHubAppID() (int64, error) {
	id, err := strconv.ParseInt(o.GitHubAppID, 10, 64)
	if err == nil && id <= 0 {
		err = errors.New("not positive")
	}
	return id, err
}

// Validate reports every problem with the options at once. Messages never include secrets.
func (o DashboardOptions) Validate() error {
	var problems []error
	add := func(format string, args ...any) { problems = append(problems, fmt.Errorf(format, args...)) }
	if o.SSHAddr != "" {
		if _, port, err := net.SplitHostPort(o.SSHAddr); err != nil || port == "" {
			add("--ssh-addr must be host:port like :2222, got %q", o.SSHAddr)
		}
	}
	githubApp := o.GitHubAppID != "" || o.GitHubAppSlug != "" || o.GitHubAppPrivateKey != ""
	switch o.Auth {
	case authNone:
		if o.SSHAddr != "" {
			add("--ssh-addr serves hosted pushes and needs --auth clerk; use gavel ssh serve locally")
		}
		if githubApp {
			add("--github-app-* flags serve hosted orgs and need --auth clerk")
		}
	case authClerk:
		if o.PublicURL == "" {
			add("--auth clerk requires --public-url")
		} else if u, err := url.Parse(o.PublicURL); err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" ||
			strings.TrimSuffix(u.Path, "/") != "" || u.RawQuery != "" || u.Fragment != "" || u.User != nil {
			add("--public-url must be an origin like https://gavel.example.com, got %q", o.PublicURL)
		}
		if o.ClerkIssuer == "" {
			add("--auth clerk requires --clerk-issuer or GAVEL_CLERK_ISSUER")
		} else if u, err := url.Parse(o.ClerkIssuer); err != nil || u.Scheme != "https" || u.Host == "" {
			add("--clerk-issuer must be an https URL, got %q", o.ClerkIssuer)
		}
		if o.ClerkPublishableKey == "" {
			add("--auth clerk requires --clerk-publishable-key or GAVEL_CLERK_PUBLISHABLE_KEY")
		} else if !strings.HasPrefix(o.ClerkPublishableKey, "pk_") {
			add("--clerk-publishable-key must start with pk_")
		}
		if githubApp {
			var missing []string
			if o.GitHubAppID == "" {
				missing = append(missing, "--github-app-id")
			} else if _, err := o.gitHubAppID(); err != nil {
				add("--github-app-id must be the App's numeric ID (from its settings page), got %q", o.GitHubAppID)
			}
			if o.GitHubAppSlug == "" {
				missing = append(missing, "--github-app-slug")
			}
			if o.GitHubAppPrivateKey == "" {
				missing = append(missing, "--github-app-private-key")
			}
			if len(missing) > 0 {
				add("--github-app-id, --github-app-slug and --github-app-private-key (or GAVEL_GITHUB_APP_PRIVATE_KEY_FILE) go together; missing %s", strings.Join(missing, ", "))
			}
		}
		if o.ClerkClientID == "" {
			add("--auth clerk requires --clerk-client-id or GAVEL_CLERK_CLIENT_ID")
		}
		for _, local := range []struct {
			flag string
			set  bool
		}{{"--dev", o.Dev}, {"--menu-bar", o.MenuBar}} {
			if local.set {
				add("%s is local-only and can't be used with --auth clerk", local.flag)
			}
		}
		if o.clerkSecretKey == "" {
			add("--auth clerk requires the CLERK_SECRET_KEY environment variable")
		} else if !strings.HasPrefix(o.clerkSecretKey, "sk_") {
			add("CLERK_SECRET_KEY must start with sk_")
		}
		if o.secretKey == "" {
			add("--auth clerk requires the GAVEL_SECRET_KEY environment variable")
		} else if key, err := base64.StdEncoding.DecodeString(o.secretKey); err != nil {
			add("GAVEL_SECRET_KEY must be at least 32 random bytes, base64-encoded (it isn't valid base64)")
		} else if len(key) < 32 {
			add("GAVEL_SECRET_KEY must be at least 32 random bytes, base64-encoded (got %d bytes)", len(key))
		}
	default:
		add("--auth must be %q or %q, got %q", authNone, authClerk, o.Auth)
	}
	return errors.Join(problems...)
}
