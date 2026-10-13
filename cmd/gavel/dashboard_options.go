// Options for `gavel serve`: the PR dashboard flags plus authentication. --auth none
// serves one local user; --auth clerk serves many organizations behind Clerk sign-in.
package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
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
	} {
		if *setting.value == "" {
			*setting.value = getenv(setting.env)
		}
	}
	o.clerkSecretKey = getenv("CLERK_SECRET_KEY")
	o.secretKey = getenv("GAVEL_SECRET_KEY")
	return o
}

// Validate reports every problem with the options at once. Messages never include secrets.
func (o DashboardOptions) Validate() error {
	var problems []error
	add := func(format string, args ...any) { problems = append(problems, fmt.Errorf(format, args...)) }
	switch o.Auth {
	case authNone:
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
