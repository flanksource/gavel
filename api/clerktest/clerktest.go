// Package clerktest runs an in-process Clerk instance for tests: OAuth discovery, the device and
// token endpoints, JWKS, and a slice of the Backend API, signing with a generated RSA key.
package clerktest

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	// KeyID is the kid of the server's signing key.
	KeyID = "clerktest"
	// DeviceCode and RefreshToken are the only grants the token endpoint accepts.
	DeviceCode   = "device-code"
	RefreshToken = "refresh"
	// SecretKey is the only Backend API key the server accepts.
	SecretKey = "sk_test_clerktest"
)

// Membership is one organization the server's user belongs to, as the Backend API lists it.
type Membership struct {
	OrgID, OrgSlug, OrgName, Role string
}

// User is the identity the server signs into ID tokens.
type User struct {
	ID      string
	Name    string
	Email   string
	OrgID   string
	OrgSlug string
	// OrgRole is the v1 form, e.g. "org:admin"; session tokens carry it without the prefix.
	OrgRole string
}

// Server is a TLS issuer. Use Client() as the HTTP client; its certificate is self-signed.
type Server struct {
	*httptest.Server
	ClientID string
	User     User
	// DenyLogin makes the device and token endpoints answer access_denied, echoing the
	// device code (submitted, or DeviceCode) in the description the way a verbose provider might.
	DenyLogin bool
	// Memberships and GitHubToken back the Backend API's user endpoints.
	Memberships []Membership
	GitHubToken string

	key       *rsa.PrivateKey
	t         testing.TB
	refreshes atomic.Int64
	jwksReads atomic.Int64
	mu        sync.Mutex
	forms     map[string][]url.Values
	published map[string]*rsa.PublicKey
	issued    atomic.Int64
	revoked   map[string]bool
	backend   []string
}

// New starts an issuer for the public OAuth client clientID and closes it when the test ends.
func New(t testing.TB, clientID string, user User) *Server {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	s := &Server{ClientID: clientID, User: user, key: key, t: t, forms: map[string][]url.Values{}, published: map[string]*rsa.PublicKey{KeyID: &key.PublicKey}, revoked: map[string]bool{}}
	s.Server = httptest.NewTLSServer(http.HandlerFunc(s.serve))
	t.Cleanup(s.Close)
	return s
}

// IDToken returns an ID token for the server's user that expires at expiry.
func (s *Server) IDToken(expiry time.Time) string {
	claims := jwt.MapClaims{
		"sub": s.User.ID, "iss": s.URL, "aud": s.ClientID,
		"exp": expiry.Unix(), "iat": time.Now().Unix(),
		"name": s.User.Name, "email": s.User.Email, "email_verified": true,
	}
	if s.User.OrgID != "" {
		claims["org_id"] = s.User.OrgID
		claims["org_slug"] = s.User.OrgSlug
	}
	return s.Sign(claims)
}

// SessionToken returns a v2 browser session token for the server's user, minted for origin azp.
func (s *Server) SessionToken(azp string, expiry time.Time) string {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub": s.User.ID, "iss": s.URL, "azp": azp, "sid": "sess_clerktest", "sts": "active", "v": 2,
		"exp": expiry.Unix(), "iat": now.Unix(), "nbf": now.Add(-10 * time.Second).Unix(),
	}
	if s.User.OrgID != "" {
		claims["o"] = map[string]string{"id": s.User.OrgID, "slg": s.User.OrgSlug, "rol": strings.TrimPrefix(s.User.OrgRole, "org:")}
	}
	return s.Sign(claims)
}

// AccessToken returns an OAuth access token (typ at+jwt) issued to the server's client for its user.
func (s *Server) AccessToken(expiry time.Time) string {
	now := time.Now()
	return s.SignWithType("at+jwt", jwt.MapClaims{
		"sub": s.User.ID, "iss": s.URL, "client_id": s.ClientID, "jti": fmt.Sprintf("oat_%032x", s.issued.Add(1)),
		"exp": expiry.Unix(), "iat": now.Unix(), "nbf": now.Unix(), "scope": "openid profile email offline_access",
	})
}

// Sign signs arbitrary claims with the server's key, for tests that need malformed or foreign tokens.
func (s *Server) Sign(claims jwt.MapClaims) string {
	return s.SignWithKeyID(KeyID, claims)
}

// SignWithType signs claims under the header typ, e.g. "at+jwt".
func (s *Server) SignWithType(typ string, claims jwt.MapClaims) string {
	return s.sign(KeyID, typ, claims)
}

// SignWithKeyID signs with the server's key under any kid, e.g. one the JWKS doesn't publish.
func (s *Server) SignWithKeyID(kid string, claims jwt.MapClaims) string {
	return s.sign(kid, "JWT", claims)
}

func (s *Server) sign(kid, typ string, claims jwt.MapClaims) string {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = kid
	token.Header["typ"] = typ
	raw, err := token.SignedString(s.key)
	if err != nil {
		s.t.Fatal(err)
	}
	return raw
}

// Refreshes counts refresh_token grants served.
func (s *Server) Refreshes() int64 { return s.refreshes.Load() }

// JWKSReads counts JWKS documents served.
func (s *Server) JWKSReads() int64 { return s.jwksReads.Load() }

// PublicKey is the public half of the server's signing key.
func (s *Server) PublicKey() *rsa.PublicKey { return &s.key.PublicKey }

// RevokeAccessToken makes the Backend API report raw as revoked.
func (s *Server) RevokeAccessToken(raw string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.revoked[raw] = true
}

// BackendRequests lists the Backend API requests served, as "METHOD path?query", in order.
func (s *Server) BackendRequests() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]string(nil), s.backend...)
}

// BackendClient is an HTTP client that sends requests for api.clerk.com to this server.
func (s *Server) BackendClient() *http.Client {
	client := s.Client()
	target, err := url.Parse(s.URL)
	if err != nil {
		s.t.Fatal(err)
	}
	transport := client.Transport
	client.Transport = roundTripper(func(r *http.Request) (*http.Response, error) {
		if r.URL.Host != "api.clerk.com" {
			return nil, fmt.Errorf("clerktest: unexpected host %s", r.URL.Host)
		}
		r = r.Clone(r.Context())
		r.URL.Scheme, r.URL.Host = target.Scheme, target.Host
		return transport.RoundTrip(r)
	})
	return client
}

type roundTripper func(*http.Request) (*http.Response, error)

func (f roundTripper) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// PublishKey adds key to the JWKS under kid, e.g. to simulate a key rotation.
func (s *Server) PublishKey(kid string, key *rsa.PublicKey) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.published[kid] = key
}

// Forms returns the form bodies posted to path, in order.
func (s *Server) Forms(path string) []url.Values {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]url.Values(nil), s.forms[path]...)
}

func (s *Server) record(r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.forms[r.URL.Path] = append(s.forms[r.URL.Path], r.PostForm)
	return nil
}

func (s *Server) serve(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.URL.Path {
	case "/.well-known/oauth-authorization-server":
		writeJSON(w, map[string]any{
			"issuer":                        s.URL,
			"device_authorization_endpoint": s.URL + "/oauth/device",
			"token_endpoint":                s.URL + "/oauth/token",
			"jwks_uri":                      s.URL + "/.well-known/jwks.json",
			"scopes_supported":              []string{"openid", "profile", "email", "user:org:read"},
		})
	case "/.well-known/jwks.json":
		s.jwksReads.Add(1)
		writeJSON(w, map[string]any{"keys": s.jwks()})
	case "/oauth/device", "/oauth/token":
		if err := s.record(r); err != nil || r.PostForm.Get("client_id") != s.ClientID {
			writeError(w, http.StatusUnauthorized, "invalid_client", "")
			return
		}
		if s.DenyLogin {
			code := r.PostForm.Get("device_code")
			if code == "" {
				code = DeviceCode
			}
			writeError(w, http.StatusBadRequest, "access_denied", "denied device code "+code)
			return
		}
		if r.URL.Path == "/oauth/device" {
			writeJSON(w, map[string]any{"device_code": DeviceCode, "user_code": "ABCD-EFGH", "verification_uri": s.URL + "/verify", "expires_in": 60, "interval": 1})
			return
		}
		switch r.PostForm.Get("grant_type") {
		case "urn:ietf:params:oauth:grant-type:device_code":
			if r.PostForm.Get("device_code") != DeviceCode {
				writeError(w, http.StatusBadRequest, "invalid_grant", "")
				return
			}
		case "refresh_token":
			if r.PostForm.Get("refresh_token") != RefreshToken {
				writeError(w, http.StatusBadRequest, "invalid_grant", "")
				return
			}
			s.refreshes.Add(1)
		default:
			writeError(w, http.StatusBadRequest, "unsupported_grant_type", "")
			return
		}
		writeJSON(w, map[string]any{"access_token": "access", "refresh_token": RefreshToken, "token_type": "Bearer", "expires_in": 3600, "id_token": s.IDToken(time.Now().Add(time.Hour))})
	default:
		if strings.HasPrefix(r.URL.Path, "/v1/") {
			s.serveBackend(w, r)
			return
		}
		http.NotFound(w, r)
	}
}

func (s *Server) serveBackend(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	s.backend = append(s.backend, r.Method+" "+r.URL.RequestURI())
	s.mu.Unlock()
	if r.Header.Get("Authorization") != "Bearer "+SecretKey {
		writeBackendError(w, http.StatusUnauthorized)
		return
	}
	userPath := "/v1/users/" + s.User.ID + "/"
	switch {
	case r.Method == http.MethodPost && r.URL.Path == "/v1/oauth_applications/access_tokens/verify":
		s.verifyAccessToken(w, r)
	case r.Method == http.MethodGet && r.URL.Path == userPath+"organization_memberships":
		s.listMemberships(w, r)
	case r.Method == http.MethodGet && r.URL.Path == userPath+"oauth_access_tokens/oauth_github":
		tokens := []map[string]any{}
		if s.GitHubToken != "" {
			tokens = append(tokens, map[string]any{
				"object": "oauth_access_token", "external_account_id": "eac_clerktest", "provider_user_id": "1",
				"token": s.GitHubToken, "provider": "oauth_github", "public_metadata": map[string]any{}, "label": nil, "expires_at": nil,
			})
		}
		writeJSON(w, tokens)
	default:
		writeBackendError(w, http.StatusNotFound)
	}
}

// verifyAccessToken answers like Clerk: 404 for a token it didn't issue,
// {"active": false} once expired, and otherwise the token with its revoked flag.
func (s *Server) verifyAccessToken(w http.ResponseWriter, r *http.Request) {
	var body struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeBackendError(w, http.StatusBadRequest)
		return
	}
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(body.AccessToken, claims, func(*jwt.Token) (any, error) { return &s.key.PublicKey, nil }, jwt.WithoutClaimsValidation())
	if err != nil || token.Header["typ"] != "at+jwt" {
		writeBackendError(w, http.StatusNotFound)
		return
	}
	expiry, err := claims.GetExpirationTime()
	if err != nil || expiry == nil || !expiry.After(time.Now()) {
		writeJSON(w, map[string]bool{"active": false})
		return
	}
	s.mu.Lock()
	revoked := s.revoked[body.AccessToken]
	s.mu.Unlock()
	writeJSON(w, map[string]any{
		"object": "clerk_idp_oauth_access_token", "id": claims["jti"], "client_id": claims["client_id"], "subject": claims["sub"],
		"scopes": strings.Fields(fmt.Sprint(claims["scope"])), "revoked": revoked, "revocation_reason": nil, "expired": false,
		"expiration": expiry.Unix(), "created_at": claims["iat"], "updated_at": claims["iat"],
	})
}

// listMemberships pages Memberships with limit (default 10) and offset, like Clerk.
func (s *Server) listMemberships(w http.ResponseWriter, r *http.Request) {
	limit, offset := 10, 0
	if raw := r.URL.Query().Get("limit"); raw != "" {
		limit, _ = strconv.Atoi(raw)
	}
	if raw := r.URL.Query().Get("offset"); raw != "" {
		offset, _ = strconv.Atoi(raw)
	}
	if limit < 1 || limit > 500 || offset < 0 {
		writeBackendError(w, http.StatusBadRequest)
		return
	}
	data := []map[string]any{}
	for i := offset; i < len(s.Memberships) && i < offset+limit; i++ {
		m := s.Memberships[i]
		data = append(data, map[string]any{
			"object": "organization_membership", "id": fmt.Sprintf("orgmem_%d", i), "role": m.Role,
			"organization": map[string]any{"object": "organization", "id": m.OrgID, "slug": m.OrgSlug, "name": m.OrgName},
		})
	}
	writeJSON(w, map[string]any{"data": data, "total_count": len(s.Memberships)})
}

func writeBackendError(w http.ResponseWriter, status int) {
	w.WriteHeader(status)
	text := http.StatusText(status)
	writeJSON(w, map[string]any{"errors": []map[string]string{{"message": text, "long_message": text, "code": strings.ToLower(strings.ReplaceAll(text, " ", "_"))}}})
}

func (s *Server) jwks() []map[string]string {
	s.mu.Lock()
	defer s.mu.Unlock()
	keys := make([]map[string]string, 0, len(s.published))
	for kid, key := range s.published {
		keys = append(keys, map[string]string{
			"kty": "RSA", "kid": kid, "alg": "RS256", "use": "sig",
			"n": base64.RawURLEncoding.EncodeToString(key.N.Bytes()),
			"e": base64.RawURLEncoding.EncodeToString(big.NewInt(int64(key.E)).Bytes()),
		})
	}
	return keys
}

func writeError(w http.ResponseWriter, status int, code, description string) {
	w.WriteHeader(status)
	writeJSON(w, map[string]string{"error": code, "error_description": description})
}

func writeJSON(w http.ResponseWriter, value any) {
	_ = json.NewEncoder(w).Encode(value)
}
