package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/flanksource/gavel/api/clerktest"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
)

func TestClerkDeviceLoginAndWhoAmI(t *testing.T) {
	clerk := clerktest.New(t, "cli", clerktest.User{ID: "user_123", Email: "user@example.com", OrgID: "org_123", OrgSlug: "example-team"})
	sign := func(issuer, audience, kid string, expiry time.Time) string {
		return clerk.SignWithKeyID(kid, jwt.MapClaims{"sub": "user_123", "iss": issuer, "aud": audience, "exp": expiry.Unix(), "iat": time.Now().Unix(), "email": "user@example.com", "email_verified": true, "org_id": "org_123", "org_slug": "example-team"})
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := NewAuthClient(ctx, AuthConfig{Issuer: clerk.URL, ClientID: "cli"}, clerk.Client())
	if err != nil {
		t.Fatal(err)
	}
	device, err := client.BeginLogin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	deviceForm := clerk.Forms("/oauth/device")[0]
	if deviceForm.Get("client_id") != "cli" || !strings.Contains(deviceForm.Get("scope"), "openid") || !strings.Contains(deviceForm.Get("scope"), "user:org:read") || deviceForm.Has("client_secret") {
		t.Fatalf("wrong device request: %v", deviceForm)
	}
	session, identity, err := client.CompleteLogin(ctx, device)
	if err != nil {
		t.Fatal(err)
	}
	if identity.ID != "user_123" || identity.Email != "user@example.com" || identity.OrganizationID != "org_123" || identity.OrganizationSlug != "example-team" || !identity.EmailVerified {
		t.Fatalf("wrong identity: %+v", identity)
	}
	_, identity, err = client.WhoAmI(ctx, session)
	if err != nil || identity.ID != "user_123" || identity.OrganizationID != "org_123" {
		t.Fatalf("whoami: %+v %v", identity, err)
	}
	session.Token.Expiry = time.Now().Add(-time.Minute)
	updated, _, err := client.WhoAmI(ctx, session)
	if err != nil || clerk.Refreshes() != 1 || !updated.Token.Valid() {
		t.Fatalf("refresh failed: %v", err)
	}
	// Expired ID tokens trigger refresh even while the access token is valid.
	session = updated
	session.IDToken = sign(clerk.URL, "cli", clerktest.KeyID, time.Now().Add(-time.Minute))
	if _, _, err := client.WhoAmI(ctx, session); err != nil || clerk.Refreshes() != 2 {
		t.Fatalf("ID refresh failed: %v", err)
	}
	for _, form := range clerk.Forms("/oauth/token") {
		if form.Has("client_secret") {
			t.Fatalf("public client sent a secret: %v", form)
		}
	}
	for name, token := range map[string]string{
		"wrong issuer":   sign("https://other.example", "cli", clerktest.KeyID, time.Now().Add(time.Hour)),
		"wrong audience": sign(clerk.URL, "other", clerktest.KeyID, time.Now().Add(time.Hour)),
		"unknown key":    sign(clerk.URL, "cli", "unknown", time.Now().Add(time.Hour)),
		"expired":        sign(clerk.URL, "cli", clerktest.KeyID, time.Now().Add(-time.Hour)),
		"malformed":      "not-a-token",
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := client.verifyIdentity(ctx, token); err == nil {
				t.Fatal("accepted invalid token")
			}
		})
	}
}

func TestCredentialStore(t *testing.T) {
	path := filepath.Join(t.TempDir(), "gavel", "session.json")
	if _, err := LoadSession(path); err == nil {
		t.Fatal("missing credentials accepted")
	}
	session := &Session{Config: AuthConfig{Issuer: "https://clerk.example", ClientID: "cli"}, Token: &oauth2.Token{AccessToken: "access", RefreshToken: "refresh"}, IDToken: "signed"}
	if err := SaveSession(path, session); err != nil {
		t.Fatal(err)
	}
	info, _ := os.Stat(path)
	if info.Mode().Perm() != 0600 {
		t.Fatalf("permissions: %v", info.Mode())
	}
	loaded, err := LoadSession(path)
	if err != nil || loaded.Token.RefreshToken != "refresh" {
		t.Fatalf("roundtrip: %v", err)
	}
	if err := os.Chmod(path, 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadSession(path); err == nil {
		t.Fatal("insecure credentials accepted")
	}
	if err := Logout(path); err != nil {
		t.Fatal(err)
	}
	if err := Logout(path); err != nil {
		t.Fatal(err)
	}
}

func TestRejectInvalidConfiguration(t *testing.T) {
	for _, config := range []AuthConfig{{}, {Issuer: "http://clerk.example", ClientID: "cli"}, {Issuer: "https://user:secret@clerk.example", ClientID: "cli"}} {
		if _, err := NewAuthClient(context.Background(), config, nil); err == nil {
			t.Fatal("accepted invalid config")
		}
	}
}

func TestRejectDiscoveryAndDeviceURLs(t *testing.T) {
	var server *httptest.Server
	var mode string
	server = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/device" {
			json.NewEncoder(w).Encode(map[string]any{"device_code": "secret", "user_code": "ABCD", "verification_uri": "http://unsafe.example", "expires_in": 60})
			return
		}
		m := metadata{Issuer: server.URL, DeviceEndpoint: server.URL + "/device", TokenEndpoint: server.URL + "/token", JWKSURI: server.URL + "/jwks"}
		if mode == "issuer" {
			m.Issuer = "https://other.example"
		}
		if mode == "endpoint" {
			m.TokenEndpoint = "https://other.example/token"
		}
		json.NewEncoder(w).Encode(m)
	}))
	defer server.Close()
	for _, value := range []string{"issuer", "endpoint"} {
		mode = value
		if _, err := NewAuthClient(context.Background(), AuthConfig{Issuer: server.URL, ClientID: "cli"}, server.Client()); err == nil {
			t.Fatal("accepted unsafe metadata")
		}
	}
	mode = ""
	client, err := NewAuthClient(context.Background(), AuthConfig{Issuer: server.URL, ClientID: "cli"}, server.Client())
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(strings.Join(client.oauth.Scopes, " "), "user:org:read") {
		t.Fatal("requested organization scope without provider support")
	}
	if _, err := client.BeginLogin(context.Background()); err == nil {
		t.Fatal("accepted unsafe verification URL")
	}
}

func TestStoredIdentityNeedsNoNetwork(t *testing.T) {
	clerk := clerktest.New(t, "cli", clerktest.User{ID: "user_123", Name: "Ada", Email: "ada@example.com", OrgID: "org_123", OrgSlug: "team"})
	session := &Session{Config: AuthConfig{Issuer: clerk.URL, ClientID: "cli"}, IDToken: clerk.IDToken(time.Now().Add(-time.Hour))}
	clerk.Close()

	identity, err := session.StoredIdentity()
	if err != nil {
		t.Fatal(err)
	}
	want := Identity{ID: "user_123", Issuer: clerk.URL, Name: "Ada", Email: "ada@example.com", OrganizationID: "org_123", OrganizationSlug: "team", EmailVerified: true}
	if identity != want {
		t.Fatalf("got %+v, want %+v", identity, want)
	}
	for name, raw := range map[string]string{"empty": "", "malformed": "not-a-token"} {
		if _, err := (&Session{IDToken: raw}).StoredIdentity(); err == nil {
			t.Fatalf("%s ID token accepted", name)
		}
	}
}
