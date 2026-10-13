package main

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/flanksource/gavel/api"
	"github.com/flanksource/gavel/api/clerktest"
	"github.com/flanksource/gavel/internal/database"
	"golang.org/x/oauth2"
)

func TestAuthCommandsRegistered(t *testing.T) {
	for _, name := range []string{"login", "whoami", "logout"} {
		command, _, err := rootCmd.Find([]string{name})
		if err != nil || command.Name() != name {
			t.Fatalf("missing %s command: %v", name, err)
		}
		if err := command.Args(command, []string{"unexpected"}); err == nil {
			t.Fatalf("%s accepted positional arguments", name)
		}
	}
	login, _, _ := rootCmd.Find([]string{"login"})
	for _, flag := range []string{"issuer", "client-id", "no-browser"} {
		if login.Flags().Lookup(flag) == nil {
			t.Fatalf("missing login flag %s", flag)
		}
	}
}

func TestWhoAmIRequiresLogin(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("HOME", t.TempDir())
	t.Setenv("AppData", t.TempDir())
	command, _, err := rootCmd.Find([]string{"whoami"})
	if err != nil {
		t.Fatal(err)
	}
	command.SetContext(context.Background())
	if err := command.RunE(command, nil); err == nil || !strings.Contains(err.Error(), "gavel login") {
		t.Fatalf("expected login guidance, got %v", err)
	}
}

func TestLogoutWithoutLogin(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("HOME", t.TempDir())
	t.Setenv("AppData", t.TempDir())
	command, _, _ := rootCmd.Find([]string{"logout"})
	var output bytes.Buffer
	command.SetOut(&output)
	defer command.SetOut(nil)
	if err := command.RunE(command, nil); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "Logged out") {
		t.Fatalf("unexpected output: %s", output.String())
	}
}

func TestPrintAuthIdentityRequestsMigrationsOnlyWhenAsked(t *testing.T) {
	t.Setenv(database.EnvDisable, "off")
	for _, tc := range []struct {
		migrate bool
		options int
	}{{migrate: false, options: 0}, {migrate: true, options: 1}} {
		var received []database.Option
		original := openAuthDatabase
		openAuthDatabase = func(ctx context.Context, options ...database.Option) (*database.DB, error) {
			received = options
			return original(ctx, options...)
		}
		command, _, _ := rootCmd.Find([]string{"whoami"})
		command.SetContext(context.Background())
		var output bytes.Buffer
		command.SetOut(&output)
		err := printAuthIdentity(command, api.Identity{ID: "user_1", Name: "User"}, tc.migrate)
		command.SetOut(nil)
		openAuthDatabase = original
		if err != nil {
			t.Fatalf("migrate=%v: %v", tc.migrate, err)
		}
		if len(received) != tc.options {
			t.Fatalf("migrate=%v: got %d database options, want %d", tc.migrate, len(received), tc.options)
		}
		if !strings.Contains(output.String(), "user_1") {
			t.Fatalf("migrate=%v: identity not printed: %s", tc.migrate, output.String())
		}
	}
}

func TestWhoAmIVerifiesSessionWithoutMigrations(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("HOME", t.TempDir())
	t.Setenv("AppData", t.TempDir())
	t.Setenv(database.EnvDisable, "off")
	clerk := clerktest.New(t, "cli", clerktest.User{ID: "user_123", Name: "Test User", Email: "user@example.com", OrgID: "org_123", OrgSlug: "team"})
	path, err := api.CredentialPath()
	if err != nil {
		t.Fatal(err)
	}
	session := &api.Session{
		Config:  api.AuthConfig{Issuer: clerk.URL, ClientID: "cli"},
		Token:   &oauth2.Token{AccessToken: "access", RefreshToken: "refresh", Expiry: time.Now().Add(time.Hour)},
		IDToken: clerk.IDToken(time.Now().Add(time.Hour)),
	}
	if err := api.SaveSession(path, session); err != nil {
		t.Fatal(err)
	}
	originalClient, originalOpen := authHTTPClient, openAuthDatabase
	authHTTPClient = clerk.Client()
	var opened bool
	var received []database.Option
	openAuthDatabase = func(ctx context.Context, options ...database.Option) (*database.DB, error) {
		opened, received = true, options
		return originalOpen(ctx, options...)
	}
	t.Cleanup(func() { authHTTPClient, openAuthDatabase = originalClient, originalOpen })

	command, _, _ := rootCmd.Find([]string{"whoami"})
	command.SetContext(context.Background())
	var output bytes.Buffer
	command.SetOut(&output)
	defer command.SetOut(nil)
	if err := command.RunE(command, nil); err != nil {
		t.Fatal(err)
	}
	if !opened || len(received) != 0 {
		t.Fatalf("whoami must open the database without migrations: opened=%v options=%d", opened, len(received))
	}
	for _, want := range []string{"user_123", "Test User", "team"} {
		if !strings.Contains(output.String(), want) {
			t.Fatalf("output missing %q: %s", want, output.String())
		}
	}
	if clerk.Refreshes() != 0 {
		t.Fatalf("a valid session must not refresh, got %d refreshes", clerk.Refreshes())
	}
}
