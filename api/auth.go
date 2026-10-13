// Package api implements Gavel's hosted authentication client.
package api

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
)

// AuthConfig identifies a Clerk public OAuth application. No client secret belongs in a CLI.
type AuthConfig struct {
	Issuer   string `json:"issuer"`
	ClientID string `json:"client_id"`
}

// Identity contains verified OIDC claims, never credentials.
type Identity struct {
	ID               string `json:"id" yaml:"id"`
	Issuer           string `json:"issuer" yaml:"issuer"`
	Name             string `json:"name" yaml:"name"`
	Email            string `json:"email" yaml:"email"`
	OrganizationID   string `json:"organization_id" yaml:"organization_id"`
	OrganizationSlug string `json:"organization_slug" yaml:"organization_slug"`
	EmailVerified    bool   `json:"-" yaml:"-"`
}

// Session is sensitive and must only be persisted by a credential store.
type Session struct {
	Config  AuthConfig    `json:"config"`
	Token   *oauth2.Token `json:"token"`
	IDToken string        `json:"id_token"`
}

type metadata struct {
	Issuer                string   `json:"issuer"`
	AuthorizationEndpoint string   `json:"authorization_endpoint"`
	DeviceEndpoint        string   `json:"device_authorization_endpoint"`
	TokenEndpoint         string   `json:"token_endpoint"`
	JWKSURI               string   `json:"jwks_uri"`
	ScopesSupported       []string `json:"scopes_supported"`
}

// AuthClient uses Clerk's OAuth metadata and JWKS, not an unverified JWT decode.
type AuthClient struct {
	config   AuthConfig
	metadata metadata
	http     *http.Client
	oauth    oauth2.Config
}

func secureURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || u.Host == "" || u.User != nil || u.Fragment != "" || u.RawQuery != "" {
		return errors.New("authentication endpoints must be absolute HTTPS URLs without credentials, query or fragment")
	}
	return nil
}

// NewAuthClient discovers the configured Clerk issuer. The transport is injectable for tests.
func NewAuthClient(ctx context.Context, config AuthConfig, client *http.Client) (*AuthClient, error) {
	config.Issuer = strings.TrimRight(strings.TrimSpace(config.Issuer), "/")
	config.ClientID = strings.TrimSpace(config.ClientID)
	if config.ClientID == "" {
		return nil, errors.New("set GAVEL_CLERK_CLIENT_ID or --client-id to a Clerk public OAuth application ID")
	}
	if err := secureURL(config.Issuer); err != nil {
		return nil, fmt.Errorf("set GAVEL_CLERK_ISSUER or --issuer to your Clerk Frontend API URL: %w", err)
	}
	if client == nil {
		client = &http.Client{Timeout: 20 * time.Second}
	}
	// Never follow redirects with tokens or discovery requests.
	copyClient := *client
	copyClient.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	a := &AuthClient{config: config, http: &copyClient}
	if err := a.getJSON(ctx, config.Issuer+"/.well-known/oauth-authorization-server", &a.metadata); err != nil {
		return nil, err
	}
	if a.metadata.Issuer != config.Issuer {
		return nil, errors.New("OAuth discovery issuer mismatch")
	}
	for _, endpoint := range []string{a.metadata.DeviceEndpoint, a.metadata.TokenEndpoint, a.metadata.JWKSURI} {
		if err := secureURL(endpoint); err != nil {
			return nil, fmt.Errorf("invalid discovery endpoint: %w", err)
		}
		// Clerk publishes these endpoints on its Frontend API origin.
		issuer, _ := url.Parse(config.Issuer)
		target, _ := url.Parse(endpoint)
		if issuer.Host != target.Host {
			return nil, errors.New("OAuth endpoints must belong to the configured Clerk origin")
		}
	}
	a.oauth = oauth2.Config{ClientID: config.ClientID, Scopes: []string{"openid", "profile", "email", "offline_access"}, Endpoint: oauth2.Endpoint{
		AuthURL: a.metadata.AuthorizationEndpoint, TokenURL: a.metadata.TokenEndpoint, DeviceAuthURL: a.metadata.DeviceEndpoint, AuthStyle: oauth2.AuthStyleInParams,
	}}
	// Clerk advertises this scope only when Organizations is enabled. Request
	// it so the device consent screen can select an organization and issue org_id.
	for _, scope := range a.metadata.ScopesSupported {
		if scope == "user:org:read" {
			a.oauth.Scopes = append(a.oauth.Scopes, scope)
			break
		}
	}
	return a, nil
}

func (a *AuthClient) oauthContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, oauth2.HTTPClient, a.http)
}

// BeginLogin requests a short-lived device code. Never log DeviceCode.
func (a *AuthClient) BeginLogin(ctx context.Context) (*oauth2.DeviceAuthResponse, error) {
	response, err := a.oauth.DeviceAuth(a.oauthContext(ctx))
	if err != nil {
		return nil, errors.New("Clerk device authorization failed; check the public OAuth application and Device Authorization Grant settings")
	}
	if response.DeviceCode == "" || response.UserCode == "" || response.Expiry.IsZero() {
		return nil, errors.New("Clerk returned an incomplete device authorization response")
	}
	for _, raw := range []string{response.VerificationURI, response.VerificationURIComplete} {
		if raw == "" {
			continue
		}
		u, err := url.Parse(raw)
		if err != nil || u.Scheme != "https" || u.Host == "" || u.User != nil {
			return nil, errors.New("Clerk returned an invalid verification URL")
		}
	}
	if response.VerificationURI == "" {
		return nil, errors.New("Clerk returned no verification URL")
	}
	return response, nil
}

// CompleteLogin respects the provider's polling interval and device-code expiry.
func (a *AuthClient) CompleteLogin(ctx context.Context, response *oauth2.DeviceAuthResponse) (*Session, Identity, error) {
	token, err := a.oauth.DeviceAccessToken(a.oauthContext(ctx), response)
	if err != nil {
		return nil, Identity{}, errors.New("login was denied, expired, or interrupted; run gavel login again")
	}
	raw, _ := token.Extra("id_token").(string)
	identity, err := a.verifyIdentity(ctx, raw)
	if err != nil {
		return nil, Identity{}, err
	}
	return &Session{Config: a.config, Token: token, IDToken: raw}, identity, nil
}

// WhoAmI refreshes expired access tokens and validates the signed identity.
// The caller must save the returned session to preserve rotated refresh tokens.
func (a *AuthClient) WhoAmI(ctx context.Context, session *Session) (*Session, Identity, error) {
	if session == nil || session.Token == nil || session.Config != a.config {
		return nil, Identity{}, errors.New("no matching login; run gavel login")
	}
	updated := *session
	// ID and access tokens may have different lifetimes. Unverified parsing is
	// only used to decide whether to refresh; identity is always verified below.
	var expiryClaims jwt.RegisteredClaims
	_, _, parseErr := jwt.NewParser().ParseUnverified(session.IDToken, &expiryClaims)
	identityExpired := parseErr == nil && expiryClaims.ExpiresAt != nil && !expiryClaims.ExpiresAt.After(time.Now())
	if !session.Token.Valid() || identityExpired {
		if session.Token.RefreshToken == "" {
			return nil, Identity{}, errors.New("session expired; run gavel login again")
		}
		refreshToken := *session.Token
		refreshToken.Expiry = time.Now().Add(-time.Minute)
		token, err := a.oauth.TokenSource(a.oauthContext(ctx), &refreshToken).Token()
		if err != nil {
			return nil, Identity{}, errors.New("session expired or revoked; run gavel login again")
		}
		updated.Token = token
		if raw, ok := token.Extra("id_token").(string); ok && raw != "" {
			updated.IDToken = raw
		}
	}
	identity, err := a.verifyIdentity(ctx, updated.IDToken)
	return &updated, identity, err
}

type identityClaims struct {
	jwt.RegisteredClaims
	Name             string `json:"name"`
	Email            string `json:"email"`
	EmailVerified    bool   `json:"email_verified"`
	OrganizationID   string `json:"org_id"`
	OrganizationSlug string `json:"org_slug"`
	AuthorizedParty  string `json:"azp"`
}

func (a *AuthClient) verifyIdentity(ctx context.Context, raw string) (Identity, error) {
	var keys struct {
		Keys []struct {
			Kty string `json:"kty"`
			Kid string `json:"kid"`
			Use string `json:"use"`
			Alg string `json:"alg"`
			N   string `json:"n"`
			E   string `json:"e"`
		} `json:"keys"`
	}
	if raw == "" {
		return Identity{}, errors.New("Clerk returned no ID token; enable the openid scope")
	}
	if err := a.getJSON(ctx, a.metadata.JWKSURI, &keys); err != nil {
		return Identity{}, err
	}
	claims := &identityClaims{}
	_, err := jwt.ParseWithClaims(raw, claims, func(token *jwt.Token) (any, error) {
		kid, _ := token.Header["kid"].(string)
		if kid == "" {
			return nil, errors.New("missing signing key ID")
		}
		for _, key := range keys.Keys {
			if key.Kid != kid || key.Kty != "RSA" || (key.Use != "" && key.Use != "sig") || (key.Alg != "" && key.Alg != "RS256") {
				continue
			}
			n, nerr := base64.RawURLEncoding.DecodeString(key.N)
			e, eerr := base64.RawURLEncoding.DecodeString(key.E)
			exponent := new(big.Int).SetBytes(e)
			if nerr != nil || eerr != nil || len(n) < 256 || !exponent.IsInt64() || exponent.Int64() < 3 || exponent.Int64() > 2147483647 {
				return nil, errors.New("invalid RSA signing key")
			}
			return &rsa.PublicKey{N: new(big.Int).SetBytes(n), E: int(exponent.Int64())}, nil
		}
		return nil, errors.New("unknown signing key")
	}, jwt.WithValidMethods([]string{"RS256"}), jwt.WithIssuer(a.config.Issuer), jwt.WithAudience(a.config.ClientID), jwt.WithExpirationRequired(), jwt.WithIssuedAt())
	if err != nil || claims.Subject == "" || claims.IssuedAt == nil ||
		(claims.OrganizationSlug != "" && claims.OrganizationID == "") ||
		(len(claims.Audience) > 1 && claims.AuthorizedParty == "") ||
		(claims.AuthorizedParty != "" && claims.AuthorizedParty != a.config.ClientID) {
		return Identity{}, errors.New("identity token is invalid or expired; run gavel login again")
	}
	return Identity{ID: claims.Subject, Issuer: claims.Issuer, Name: claims.Name, Email: claims.Email, EmailVerified: claims.EmailVerified, OrganizationID: claims.OrganizationID, OrganizationSlug: claims.OrganizationSlug}, nil
}

func (a *AuthClient) getJSON(ctx context.Context, endpoint string, output any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	response, err := a.http.Do(req)
	if err != nil {
		return errors.New("cannot reach Clerk authentication service")
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("Clerk authentication endpoint returned HTTP %d", response.StatusCode)
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(output); err != nil {
		return errors.New("invalid Clerk authentication response")
	}
	return nil
}
