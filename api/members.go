// Resolves which gavel user and organization a verified Clerk user acts as, keeping
// gavel's users, orgs and memberships in step with Clerk's Backend API.
package api

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/flanksource/gavel/tenant"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Errors for a request whose organization can't be resolved. Callers answer them
// with 409 and 403; other errors mean gavel or Clerk couldn't answer.
var (
	ErrNoActiveOrg = errors.New("no active organization")
	ErrNotMember   = errors.New("not a member of the organization")
)

// Members turns verified Clerk users into tenant scopes. It is safe for concurrent use.
type Members struct {
	db     *gorm.DB
	clerk  ClerkOrganizations
	issuer string
	maxAge time.Duration

	mu     sync.Mutex
	synced map[uuid.UUID]time.Time
}

// NewMembers resolves users of the Clerk instance issuer, re-reading a user's profile
// and memberships from clerk once gavel's copy is older than maxAge.
func NewMembers(db *gorm.DB, clerk ClerkOrganizations, issuer string, maxAge time.Duration) *Members {
	return &Members{db: db, clerk: clerk, issuer: issuer, maxAge: maxAge, synced: map[uuid.UUID]time.Time{}}
}

// ForSession scopes a browser session to its active organization. The session's org
// and role are authoritative, since Clerk signed them.
func (m *Members) ForSession(ctx context.Context, claims SessionClaims) (tenant.Scope, error) {
	user, err := m.user(ctx, claims.UserID)
	if err != nil {
		return tenant.Scope{}, err
	}
	if claims.OrgID == "" {
		return tenant.Scope{}, fmt.Errorf("%w: choose an organization in the dashboard", ErrNoActiveOrg)
	}
	var orgID uuid.UUID
	err = m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var known struct {
			ID   uuid.UUID
			Slug *string
		}
		if err := tx.Raw(`SELECT id, slug FROM orgs WHERE issuer = ? AND clerk_org_id = ?`, m.issuer, claims.OrgID).Scan(&known).Error; err != nil {
			return err
		}
		orgID = known.ID
		if known.ID == uuid.Nil || (claims.OrgSlug != "" && (known.Slug == nil || *known.Slug != claims.OrgSlug)) {
			org, err := upsertOrg(tx, m.issuer, claims.OrgID, claims.OrgSlug, "")
			if err != nil {
				return err
			}
			orgID = org.ID
		}
		return tx.Exec(`
			INSERT INTO org_memberships (user_id, org_id, role, synced_at) VALUES (?, ?, ?, now())
			ON CONFLICT (user_id, org_id) DO UPDATE SET role = EXCLUDED.role, synced_at = now()
			WHERE org_memberships.role <> EXCLUDED.role`,
			user.id, orgID, claims.OrgRole).Error
	})
	if err != nil {
		return tenant.Scope{}, fmt.Errorf("record the session's organization: %w", err)
	}
	return tenant.Scope{OrgID: orgID, UserID: &user.id, Role: claims.OrgRole, GitHubLogin: user.githubLogin}, nil
}

// MemberOrg is an organization a user belongs to, with their role in it.
type MemberOrg struct {
	ID         uuid.UUID `json:"id"`
	ClerkOrgID string    `json:"clerkOrgId"`
	Slug       string    `json:"slug"`
	Name       string    `json:"name"`
	Role       string    `json:"role"`
}

// selector is how the CLI names the org: its slug, or its Clerk ID without one.
func (o MemberOrg) selector() string {
	if o.Slug != "" {
		return o.Slug
	}
	return o.ClerkOrgID
}

// Orgs lists every organization clerkUserID belongs to, without picking one.
func (m *Members) Orgs(ctx context.Context, clerkUserID string) ([]MemberOrg, error) {
	user, err := m.user(ctx, clerkUserID)
	if err != nil {
		return nil, err
	}
	return m.orgs(ctx, user.id)
}

func (m *Members) orgs(ctx context.Context, userID uuid.UUID) ([]MemberOrg, error) {
	orgs := []MemberOrg{}
	if err := m.db.WithContext(ctx).Raw(`
		SELECT o.id, o.clerk_org_id, COALESCE(o.slug, '') AS slug, o.name, m.role
		FROM org_memberships m JOIN orgs o ON o.id = m.org_id
		WHERE m.user_id = ? ORDER BY o.slug, o.clerk_org_id`, userID).Scan(&orgs).Error; err != nil {
		return nil, fmt.Errorf("read the user's organizations: %w", err)
	}
	return orgs, nil
}

// ForAccessToken scopes a CLI request to the organization named by selector (a slug
// or Clerk org ID), or to the user's only organization when selector is empty.
func (m *Members) ForAccessToken(ctx context.Context, clerkUserID, selector string) (tenant.Scope, error) {
	user, err := m.user(ctx, clerkUserID)
	if err != nil {
		return tenant.Scope{}, err
	}
	memberships, err := m.orgs(ctx, user.id)
	if err != nil {
		return tenant.Scope{}, err
	}
	names := make([]string, 0, len(memberships))
	for _, membership := range memberships {
		names = append(names, membership.selector())
		if selector != "" && (selector == membership.selector() || selector == membership.ClerkOrgID) ||
			selector == "" && len(memberships) == 1 {
			return tenant.Scope{OrgID: membership.ID, UserID: &user.id, Role: membership.Role, GitHubLogin: user.githubLogin}, nil
		}
	}
	switch {
	case selector != "":
		return tenant.Scope{}, fmt.Errorf("%w %q", ErrNotMember, selector)
	case len(memberships) == 0:
		return tenant.Scope{}, fmt.Errorf("%w: the user belongs to no organization in Clerk", ErrNoActiveOrg)
	default:
		return tenant.Scope{}, fmt.Errorf("%w: set X-Gavel-Org to one of: %s", ErrNoActiveOrg, strings.Join(names, ", "))
	}
}

type member struct {
	id          uuid.UUID
	githubLogin string
}

// user returns clerkUserID's gavel user, re-reading its profile and memberships from
// Clerk when gavel has none or they are older than maxAge.
func (m *Members) user(ctx context.Context, clerkUserID string) (member, error) {
	var row struct {
		ID          uuid.UUID
		GitHubLogin *string `gorm:"column:github_login"`
	}
	if err := m.db.WithContext(ctx).Raw(`SELECT id, github_login FROM users WHERE issuer = ? AND clerk_user_id = ?`,
		m.issuer, clerkUserID).Scan(&row).Error; err != nil {
		return member{}, fmt.Errorf("read the user: %w", err)
	}
	if row.ID != uuid.Nil && !m.stale(row.ID) {
		user := member{id: row.ID}
		if row.GitHubLogin != nil {
			user.githubLogin = *row.GitHubLogin
		}
		return user, nil
	}
	return m.refresh(ctx, clerkUserID)
}

func (m *Members) stale(userID uuid.UUID) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	synced, ok := m.synced[userID]
	return !ok || time.Since(synced) > m.maxAge
}

// refresh copies clerkUserID's profile and memberships from Clerk into gavel.
func (m *Members) refresh(ctx context.Context, clerkUserID string) (member, error) {
	profile, err := m.clerk.User(ctx, clerkUserID)
	if err != nil {
		return member{}, fmt.Errorf("read the user from Clerk: %w", err)
	}
	memberships, err := m.clerk.UserMemberships(ctx, clerkUserID)
	if err != nil {
		return member{}, fmt.Errorf("read the user's organizations from Clerk: %w", err)
	}
	var user struct{ ID uuid.UUID }
	if err := m.db.WithContext(ctx).Raw(`
		INSERT INTO users (issuer, clerk_user_id, name, email, email_verified, github_login)
		VALUES (?, ?, ?, ?, ?, NULLIF(?, ''))
		ON CONFLICT (issuer, clerk_user_id) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email,
			email_verified = EXCLUDED.email_verified, github_login = EXCLUDED.github_login, updated_at = now()
		RETURNING id`,
		m.issuer, clerkUserID, profile.Name, profile.Email, profile.EmailVerified, profile.GitHubLogin).Scan(&user).Error; err != nil {
		return member{}, fmt.Errorf("record the user: %w", err)
	}
	if err := (IdentityStore{DB: m.db}).SyncMemberships(ctx, user.ID, memberships); err != nil {
		return member{}, fmt.Errorf("record the user's organizations: %w", err)
	}
	m.mu.Lock()
	m.synced[user.ID] = time.Now()
	m.mu.Unlock()
	return member{id: user.ID, githubLogin: profile.GitHubLogin}, nil
}
