// Caches an issuer's RSA signing keys from its JWKS endpoint for token verification.
// Keys live for a TTL; an unknown kid refetches the set, at most once per 30 seconds.
package api

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// jwksRefetchInterval bounds how often an unknown kid may refetch the key set.
const jwksRefetchInterval = 30 * time.Second

// errUnknownSigningKey reports a kid the issuer doesn't publish, as opposed to a failed fetch.
var errUnknownSigningKey = errors.New("unknown signing key")

// ErrInvalidToken rejects a bearer token: bad signature, wrong claims, or expired.
// Verifiers wrap it with the reason, so check it with errors.Is.
var ErrInvalidToken = errors.New("invalid token")

// invalidToken wraps ErrInvalidToken with why the token was rejected.
func invalidToken(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrInvalidToken, fmt.Sprintf(format, args...))
}

// tokenLeeway absorbs clock skew between the issuer and this server.
const tokenLeeway = 5 * time.Second

// JWKSCache serves RS256 verification keys by kid. It is safe for concurrent use.
type JWKSCache struct {
	url  string
	http *http.Client
	ttl  time.Duration
	now  func() time.Time

	mu      sync.Mutex
	keys    map[string]*rsa.PublicKey
	fetched time.Time
}

// NewJWKSCache reads keys from url with client, never following redirects.
func NewJWKSCache(url string, client *http.Client, ttl time.Duration) *JWKSCache {
	return &JWKSCache{url: url, http: withoutRedirects(client), ttl: ttl, now: time.Now}
}

// Key returns the signing key published under kid.
func (c *JWKSCache) Key(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := c.now()
	stale := c.keys == nil || now.Sub(c.fetched) >= c.ttl
	if key, ok := c.keys[kid]; ok && !stale {
		return key, nil
	}
	if stale || now.Sub(c.fetched) >= jwksRefetchInterval {
		if err := c.refresh(ctx, now); err != nil {
			return nil, err
		}
	}
	if key, ok := c.keys[kid]; ok {
		return key, nil
	}
	return nil, errUnknownSigningKey
}

func (c *JWKSCache) refresh(ctx context.Context, now time.Time) error {
	var set struct {
		Keys []struct {
			Kty string `json:"kty"`
			Kid string `json:"kid"`
			Use string `json:"use"`
			Alg string `json:"alg"`
			N   string `json:"n"`
			E   string `json:"e"`
		} `json:"keys"`
	}
	if err := getJSON(ctx, c.http, c.url, &set); err != nil {
		return err
	}
	keys := map[string]*rsa.PublicKey{}
	for _, key := range set.Keys {
		if key.Kid == "" || key.Kty != "RSA" || (key.Use != "" && key.Use != "sig") || (key.Alg != "" && key.Alg != "RS256") {
			continue
		}
		n, nerr := base64.RawURLEncoding.DecodeString(key.N)
		e, eerr := base64.RawURLEncoding.DecodeString(key.E)
		exponent := new(big.Int).SetBytes(e)
		if nerr != nil || eerr != nil || len(n) < 256 || !exponent.IsInt64() || exponent.Int64() < 3 || exponent.Int64() > 2147483647 {
			continue
		}
		keys[key.Kid] = &rsa.PublicKey{N: new(big.Int).SetBytes(n), E: int(exponent.Int64())}
	}
	c.keys, c.fetched = keys, now
	return nil
}

// parseSigned verifies raw as an RS256 JWT signed by a key in jwks, with an expiry
// required and checked with tokenLeeway, and decodes it into claims. It returns
// ErrInvalidToken for a bad token, and the fetch error when the keys can't be read.
func parseSigned(ctx context.Context, jwks *JWKSCache, raw string, claims jwt.Claims, options ...jwt.ParserOption) (*jwt.Token, error) {
	var keyErr error
	options = append([]jwt.ParserOption{jwt.WithValidMethods([]string{"RS256"}), jwt.WithLeeway(tokenLeeway), jwt.WithExpirationRequired()}, options...)
	token, err := jwt.ParseWithClaims(raw, claims, func(token *jwt.Token) (any, error) {
		kid, _ := token.Header["kid"].(string)
		if kid == "" {
			return nil, errors.New("missing signing key ID")
		}
		key, err := jwks.Key(ctx, kid)
		keyErr = err
		return key, err
	}, options...)
	if keyErr != nil && !errors.Is(keyErr, errUnknownSigningKey) {
		return nil, keyErr
	}
	if err != nil {
		return nil, invalidToken("%v", err)
	}
	return token, nil
}
