// Verifies Clerk browser session tokens and extracts the user and active organization.
// Supports both Clerk session token versions: v2's "o" claim and v1's org_* claims.
package api

import (
	"context"
	"slices"

	"github.com/golang-jwt/jwt/v5"
)

// SessionVerifier accepts session tokens from one Clerk instance. The JWKS URL is
// Issuer + "/.well-known/jwks.json". AuthorizedParties lists the allowed origins.
type SessionVerifier struct {
	Issuer            string
	AuthorizedParties []string
	JWKS              *JWKSCache
}

// SessionClaims identify the signed-in user and their active organization, if any.
// OrgRole always has Clerk's "org:" prefix, e.g. "org:admin".
type SessionClaims struct {
	UserID  string
	OrgID   string
	OrgSlug string
	OrgRole string
}

type sessionTokenClaims struct {
	jwt.RegisteredClaims
	AuthorizedParty string `json:"azp"`
	ClientID        string `json:"client_id"`
	Status          string `json:"sts"`
	OrgID           string `json:"org_id"`
	OrgSlug         string `json:"org_slug"`
	OrgRole         string `json:"org_role"`
	Org             *struct {
		ID   string `json:"id"`
		Slug string `json:"slg"`
		Role string `json:"rol"`
	} `json:"o"`
}

// Verify checks raw and returns its claims. OAuth access tokens are rejected, and
// so are pending sessions, which Clerk treats as signed out.
func (v SessionVerifier) Verify(ctx context.Context, raw string) (SessionClaims, error) {
	claims := &sessionTokenClaims{}
	token, err := parseSigned(ctx, v.JWKS, raw, claims, jwt.WithIssuer(v.Issuer))
	if err != nil {
		return SessionClaims{}, err
	}
	if typ, _ := token.Header["typ"].(string); typ == accessTokenType || claims.ClientID != "" {
		return SessionClaims{}, invalidToken("an OAuth access token is not a browser session token")
	}
	if claims.Subject == "" {
		return SessionClaims{}, invalidToken("session token has no subject")
	}
	if claims.Status == "pending" {
		return SessionClaims{}, invalidToken("the Clerk session is pending: finish signing in, e.g. choose an organization")
	}
	if claims.AuthorizedParty != "" && !slices.Contains(v.AuthorizedParties, claims.AuthorizedParty) {
		return SessionClaims{}, invalidToken("session token was issued for origin %q, which is not an authorized party", claims.AuthorizedParty)
	}
	session := SessionClaims{UserID: claims.Subject, OrgID: claims.OrgID, OrgSlug: claims.OrgSlug, OrgRole: claims.OrgRole}
	if claims.Org != nil {
		session.OrgID, session.OrgSlug, session.OrgRole = claims.Org.ID, claims.Org.Slug, ""
		if claims.Org.Role != "" {
			session.OrgRole = "org:" + claims.Org.Role
		}
	}
	if session.OrgID == "" && (session.OrgSlug != "" || session.OrgRole != "") {
		return SessionClaims{}, invalidToken("session token has an organization slug or role without an organization ID")
	}
	return session, nil
}
