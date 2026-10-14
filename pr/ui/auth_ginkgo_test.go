// Tests the hosted auth middleware against a migrated database and the fake Clerk:
// public paths, token sources, org resolution, CSRF, and the scope handlers receive.
package ui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/api"
	"github.com/flanksource/gavel/api/clerktest"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/tenant"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/gorm"
)

const authPublicURL = "https://gavel.example.com"

type authFixture struct {
	clerk   *clerktest.Server
	db      *gorm.DB
	handler http.Handler
	scopes  chan tenant.Scope
}

func newAuthFixture() authFixture {
	GinkgoHelper()
	handle := dbtest.ForGinkgo(dbtest.Options{Name: "gavel_prui_auth"})
	GinkgoT().Setenv(database.EnvDSN, handle.DSN())
	opened, err := database.Open(GinkgoT().Context(), database.WithMigrations())
	Expect(err).NotTo(HaveOccurred())
	DeferCleanup(func() { Expect(opened.Close()).To(Succeed()) })

	clerk := clerktest.New(GinkgoT(), "cli", clerktest.User{ID: "user_ada", Name: "Ada Lovelace", Email: "ada@example.com",
		OrgID: "org_a", OrgSlug: "alpha", OrgRole: "org:admin", GitHubLogin: "ada"})
	clerk.Memberships = []clerktest.Membership{
		{OrgID: "org_a", OrgSlug: "alpha", OrgName: "Alpha", Role: "org:admin"},
		{OrgID: "org_b", OrgSlug: "beta", OrgName: "Beta", Role: "org:member"},
	}
	backend := api.ClerkOrganizations{SecretKey: clerktest.SecretKey, HTTPClient: clerk.BackendClient()}
	jwks := api.NewJWKSCache(clerk.URL+"/.well-known/jwks.json", clerk.Client(), 10*time.Minute)
	auth := &Authenticator{
		Session:   api.SessionVerifier{Issuer: clerk.URL, AuthorizedParties: []string{authPublicURL}, JWKS: jwks},
		Access:    api.NewAccessTokenVerifier(clerk.URL, "cli", jwks, backend, 10*time.Minute),
		Members:   api.NewMembers(opened.Gorm(), backend, clerk.URL, 5*time.Minute),
		PublicURL: authPublicURL,
	}
	scopes := make(chan tenant.Scope, 1)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		scopes <- tenant.FromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})
	public := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusTeapot) })
	return authFixture{clerk: clerk, db: opened.Gorm(), handler: auth.Middleware(next, public), scopes: scopes}
}

func (f authFixture) serve(req *http.Request) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	f.handler.ServeHTTP(rec, req)
	return rec
}

func (f authFixture) orgID(clerkOrgID string) uuid.UUID {
	GinkgoHelper()
	var org struct{ ID uuid.UUID }
	Expect(f.db.Raw(`SELECT id FROM orgs WHERE clerk_org_id = ?`, clerkOrgID).Scan(&org).Error).To(Succeed())
	return org.ID
}

func authError(rec *httptest.ResponseRecorder) (code, message string) {
	GinkgoHelper()
	var body struct{ Error, Message string }
	Expect(json.Unmarshal(rec.Body.Bytes(), &body)).To(Succeed(), rec.Body.String())
	return body.Error, body.Message
}

func withSessionCookie(req *http.Request, token string) *http.Request {
	req.AddCookie(&http.Cookie{Name: "__session", Value: token})
	return req
}

var _ = Describe("hosted auth middleware", func() {
	var f authFixture
	BeforeEach(func() { f = newAuthFixture() })

	It("serves page loads, assets and the auth config without a token", func() {
		for _, path := range []string{"/", "/prs", "/todos/abc", "/_assets/app.js", "/favicon.svg", "/api/auth/config"} {
			Expect(f.serve(httptest.NewRequest(http.MethodGet, path, nil)).Code).To(Equal(http.StatusTeapot), path)
		}
		Expect(f.scopes).To(BeEmpty())
	})

	It("answers an API call without a token with 401", func() {
		rec := f.serve(httptest.NewRequest(http.MethodGet, "/api/config", nil))
		Expect(rec.Code).To(Equal(http.StatusUnauthorized))
		code, message := authError(rec)
		Expect(code).To(Equal("unauthenticated"))
		Expect(message).To(ContainSubstring("sign in"))
	})

	It("answers a bad token with 401 and the reason", func() {
		req := httptest.NewRequest(http.MethodGet, "/api/config", nil)
		req.Header.Set("Authorization", "Bearer "+f.clerk.SessionToken(authPublicURL, time.Now().Add(-time.Minute)))
		rec := f.serve(req)
		Expect(rec.Code).To(Equal(http.StatusUnauthorized))
		code, message := authError(rec)
		Expect(code).To(Equal("unauthenticated"))
		Expect(message).To(ContainSubstring("token is expired"))
	})

	It("scopes a browser session to its active org", func() {
		req := withSessionCookie(httptest.NewRequest(http.MethodGet, "/api/config", nil), f.clerk.SessionToken(authPublicURL, time.Now().Add(time.Hour)))
		Expect(f.serve(req).Code).To(Equal(http.StatusOK))
		scope := <-f.scopes
		Expect(scope.OrgID).To(Equal(f.orgID("org_a")))
		Expect(scope.UserID).NotTo(BeNil())
		Expect(scope.Role).To(Equal("org:admin"))
		Expect(scope.GitHubLogin).To(Equal("ada"))
	})

	It("answers a session without an active org with 409", func() {
		token := f.clerk.Sign(jwt.MapClaims{"sub": "user_ada", "iss": f.clerk.URL, "azp": authPublicURL,
			"exp": time.Now().Add(time.Hour).Unix(), "iat": time.Now().Unix()})
		rec := f.serve(withSessionCookie(httptest.NewRequest(http.MethodGet, "/api/config", nil), token))
		Expect(rec.Code).To(Equal(http.StatusConflict))
		code, message := authError(rec)
		Expect(code).To(Equal("no_active_org"))
		Expect(message).To(ContainSubstring("choose an organization"))
	})

	It("scopes a CLI access token to the org it names, and refuses another", func() {
		req := httptest.NewRequest(http.MethodGet, "/api/config", nil)
		req.Header.Set("Authorization", "Bearer "+f.clerk.AccessToken(time.Now().Add(time.Hour)))
		req.Header.Set("X-Gavel-Org", "beta")
		Expect(f.serve(req).Code).To(Equal(http.StatusOK))
		Expect((<-f.scopes).OrgID).To(Equal(f.orgID("org_b")))

		req.Header.Set("X-Gavel-Org", "gamma")
		rec := f.serve(req)
		Expect(rec.Code).To(Equal(http.StatusForbidden))
		code, message := authError(rec)
		Expect(code).To(Equal("not_a_member"))
		Expect(message).To(ContainSubstring(`"gamma"`))

		req.Header.Del("X-Gavel-Org")
		rec = f.serve(req)
		Expect(rec.Code).To(Equal(http.StatusConflict))
		_, message = authError(rec)
		Expect(message).To(ContainSubstring("alpha, beta"))
	})

	It("lists the caller's organizations without needing one picked", func() {
		req := httptest.NewRequest(http.MethodGet, "/api/auth/orgs", nil)
		req.Header.Set("Authorization", "Bearer "+f.clerk.AccessToken(time.Now().Add(time.Hour)))
		rec := f.serve(req)
		Expect(rec.Code).To(Equal(http.StatusOK), rec.Body.String())
		var orgs []api.MemberOrg
		Expect(json.Unmarshal(rec.Body.Bytes(), &orgs)).To(Succeed())
		Expect(orgs).To(HaveLen(2))
		Expect(orgs[0].Slug).To(Equal("alpha"))
		Expect(orgs[1].Slug).To(Equal("beta"))
		Expect(f.scopes).To(BeEmpty(), "answered by the middleware, not an org's dashboard")

		rec = f.serve(httptest.NewRequest(http.MethodGet, "/api/auth/orgs", nil))
		Expect(rec.Code).To(Equal(http.StatusUnauthorized))
	})

	It("refuses a cross-site write made with the session cookie", func() {
		token := f.clerk.SessionToken(authPublicURL, time.Now().Add(time.Hour))
		post := func(header http.Header) *httptest.ResponseRecorder {
			req := withSessionCookie(httptest.NewRequest(http.MethodPost, "/api/config", nil), token)
			for name, values := range header {
				req.Header[name] = values
			}
			return f.serve(req)
		}

		rec := post(http.Header{"Origin": {"https://evil.example"}})
		Expect(rec.Code).To(Equal(http.StatusForbidden))
		code, message := authError(rec)
		Expect(code).To(Equal("cross_site_request"))
		Expect(message).To(ContainSubstring("https://evil.example"))
		Expect(post(http.Header{}).Code).To(Equal(http.StatusForbidden), "no Origin and no Sec-Fetch-Site")
		Expect(post(http.Header{"Origin": {authPublicURL}}).Code).To(Equal(http.StatusOK))
		<-f.scopes
		Expect(post(http.Header{"Sec-Fetch-Site": {"same-origin"}}).Code).To(Equal(http.StatusOK))
		<-f.scopes

		bearer := httptest.NewRequest(http.MethodPost, "/api/config", nil)
		bearer.Header.Set("Authorization", "Bearer "+token)
		bearer.Header.Set("Origin", "https://evil.example")
		Expect(f.serve(bearer).Code).To(Equal(http.StatusOK), "a bearer token isn't sent by the browser on its own")
		<-f.scopes
	})

	It("answers 503 when Clerk can't be asked", func() {
		token := f.clerk.SessionToken(authPublicURL, time.Now().Add(time.Hour))
		f.clerk.Close()
		req := httptest.NewRequest(http.MethodGet, "/api/config", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := f.serve(req)
		Expect(rec.Code).To(Equal(http.StatusServiceUnavailable))
		code, message := authError(rec)
		Expect(code).To(Equal("auth_unavailable"))
		Expect(message).NotTo(BeEmpty())
	})
})
