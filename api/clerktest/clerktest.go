// Package clerktest runs an in-process Clerk issuer for tests: OAuth discovery, the device and
// token endpoints, and JWKS, all signing with a generated RSA key that the test controls.
package clerktest

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
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
)

// User is the identity the server signs into ID tokens.
type User struct {
	ID      string
	Name    string
	Email   string
	OrgID   string
	OrgSlug string
}

// Server is a TLS issuer. Use Client() as the HTTP client; its certificate is self-signed.
type Server struct {
	*httptest.Server
	ClientID string
	User     User
	// DenyLogin makes the device and token endpoints answer access_denied, echoing the
	// device code (submitted, or DeviceCode) in the description the way a verbose provider might.
	DenyLogin bool
	key       *rsa.PrivateKey
	t         testing.TB
	refreshes atomic.Int64
	mu        sync.Mutex
	forms     map[string][]url.Values
}

// New starts an issuer for the public OAuth client clientID and closes it when the test ends.
func New(t testing.TB, clientID string, user User) *Server {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	s := &Server{ClientID: clientID, User: user, key: key, t: t, forms: map[string][]url.Values{}}
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

// Sign signs arbitrary claims with the server's key, for tests that need malformed or foreign tokens.
func (s *Server) Sign(claims jwt.MapClaims) string {
	return s.SignWithKeyID(KeyID, claims)
}

// SignWithKeyID signs with the server's key under any kid, e.g. one the JWKS doesn't publish.
func (s *Server) SignWithKeyID(kid string, claims jwt.MapClaims) string {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = kid
	raw, err := token.SignedString(s.key)
	if err != nil {
		s.t.Fatal(err)
	}
	return raw
}

// Refreshes counts refresh_token grants served.
func (s *Server) Refreshes() int64 { return s.refreshes.Load() }

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
		writeJSON(w, map[string]any{"keys": []map[string]string{{
			"kty": "RSA", "kid": KeyID, "alg": "RS256", "use": "sig",
			"n": base64.RawURLEncoding.EncodeToString(s.key.N.Bytes()),
			"e": base64.RawURLEncoding.EncodeToString(big.NewInt(int64(s.key.E)).Bytes()),
		}}})
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
		http.NotFound(w, r)
	}
}

func writeError(w http.ResponseWriter, status int, code, description string) {
	w.WriteHeader(status)
	writeJSON(w, map[string]string{"error": code, "error_description": description})
}

func writeJSON(w http.ResponseWriter, value any) {
	_ = json.NewEncoder(w).Encode(value)
}
