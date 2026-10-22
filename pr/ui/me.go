// Serves /api/auth/config, which tells the SPA how to sign in, and /api/me, which
// reports who the caller is and which organizations they act in.
package ui

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/flanksource/gavel/api"
	"github.com/flanksource/gavel/tenant"
	"github.com/google/uuid"
)

// AuthConfigResponse is how the SPA signs in: mode "none" needs nothing, mode
// "clerk" names the Clerk instance, its publishable key, and the CLI's OAuth client.
type AuthConfigResponse struct {
	Mode           string `json:"mode"`
	PublishableKey string `json:"publishableKey,omitempty"`
	Issuer         string `json:"issuer,omitempty"`
	ClientID       string `json:"clientId,omitempty"`
}

// SetAuthConfig sets what /api/auth/config reports. Without it the mode is "none".
func (s *Server) SetAuthConfig(cfg AuthConfigResponse) {
	s.authConfig = cfg
}

func (s *Server) handleAuthConfig(w http.ResponseWriter, _ *http.Request) {
	cfg := s.authConfig
	if cfg.Mode == "" {
		cfg.Mode = "none"
	}
	writeJSON(w, cfg)
}

// MeUser is the caller. Its ID is gavel's user ID, or the Clerk user ID of a local login.
type MeUser struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	GitHubLogin string `json:"githubLogin"`
}

// MeOrg is an organization the caller acts in, with their role in it.
type MeOrg struct {
	ID   string `json:"id"`
	Slug string `json:"slug"`
	Name string `json:"name"`
	Role string `json:"role"`
}

// MeResponse answers /api/me. User is null when no one is signed in locally.
type MeResponse struct {
	User *MeUser `json:"user"`
	Org  MeOrg   `json:"org"`
	Orgs []MeOrg `json:"orgs"`
}

var localMeOrg = MeOrg{ID: tenant.LocalOrgID.String(), Slug: "local", Name: "Local"}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	scope := tenant.FromContext(r.Context())
	if scope.IsLocal() {
		writeJSON(w, MeResponse{User: localLogin(), Org: localMeOrg, Orgs: []MeOrg{localMeOrg}})
		return
	}
	me, err := s.orgMember(r, scope)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, me)
}

// localLogin is the CLI's saved login, read without a network call, or nil without one.
func localLogin() *MeUser {
	path, err := api.CredentialPath()
	if err != nil {
		return nil
	}
	session, err := api.LoadSession(path)
	if err != nil {
		return nil
	}
	identity, err := session.StoredIdentity()
	if err != nil {
		return nil
	}
	return &MeUser{ID: identity.ID, Name: identity.Name, Email: identity.Email}
}

// orgMember reads the caller and their organizations from the org's database.
func (s *Server) orgMember(r *http.Request, scope tenant.Scope) (MeResponse, error) {
	if s.orgDB == nil || scope.UserID == nil {
		return MeResponse{}, fmt.Errorf("the organization's dashboard has no signed-in user")
	}
	db := s.orgDB.WithContext(r.Context())
	var user struct {
		ID                 uuid.UUID
		Name, Email, Login string
	}
	if err := db.Raw(`SELECT id, name, email, COALESCE(github_login, '') AS login FROM users WHERE id = ?`, *scope.UserID).
		Scan(&user).Error; err != nil {
		return MeResponse{}, fmt.Errorf("read the signed-in user: %w", err)
	}
	var orgs []struct {
		ID               uuid.UUID
		Slug, Name, Role string
	}
	if err := db.Raw(`
		SELECT o.id, COALESCE(o.slug, '') AS slug, o.name, m.role
		FROM org_memberships m JOIN orgs o ON o.id = m.org_id
		WHERE m.user_id = ? ORDER BY o.slug, o.name`, *scope.UserID).Scan(&orgs).Error; err != nil {
		return MeResponse{}, fmt.Errorf("read the user's organizations: %w", err)
	}
	me := MeResponse{
		User: &MeUser{ID: user.ID.String(), Name: user.Name, Email: user.Email, GitHubLogin: user.Login},
		Orgs: make([]MeOrg, 0, len(orgs)),
	}
	for _, org := range orgs {
		entry := MeOrg{ID: org.ID.String(), Slug: org.Slug, Name: org.Name, Role: org.Role}
		me.Orgs = append(me.Orgs, entry)
		if org.ID == scope.OrgID {
			me.Org = entry
			// The request's role is authoritative: a session's role can be newer.
			me.Org.Role = scope.Role
		}
	}
	return me, nil
}

func writeJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(value) //nolint:errcheck
}
