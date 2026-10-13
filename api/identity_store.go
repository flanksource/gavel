package api

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/flanksource/gavel/tenant"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// User and Org use internal UUIDs for references, never Clerk IDs as local FKs.
type User struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey"`
	Issuer        string
	ClerkUserID   string
	Name          string
	Email         string
	EmailVerified bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (User) TableName() string { return "users" }

type Org struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey"`
	Issuer     string
	ClerkOrgID string
	Slug       *string
	Name       string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (Org) TableName() string { return "orgs" }

// Ownership carries local FK values for authenticated writes. It is metadata,
// not authorization: callers must independently check permissions/membership.
// An identity with no active organization acts for the local organization.
type Ownership struct {
	UserID uuid.UUID
	OrgID  uuid.UUID
}

// OrganizationResolver is an optional trusted, backend-only metadata provider.
// A CLI must not receive the Clerk backend secret to resolve an org slug.
type OrganizationResolver interface {
	ResolveOrganization(context.Context, string) (Organization, error)
}

type Organization struct {
	ID   string `json:"id"`
	Slug string `json:"slug"`
	Name string `json:"name"`
}

// IdentityStore persists verified identities, with no tokens in the database.
type IdentityStore struct {
	DB                   *gorm.DB
	OrganizationResolver OrganizationResolver
}

// Ready reports whether the identity tables exist, so commands that don't migrate
// can skip recording identities on a database that hasn't been migrated yet.
func (s IdentityStore) Ready(ctx context.Context) (bool, error) {
	if s.DB == nil {
		return false, errors.New("identity database is unavailable")
	}
	var exists bool
	err := s.DB.WithContext(ctx).Raw(`SELECT to_regclass('public.users') IS NOT NULL`).Scan(&exists).Error
	return exists, err
}

// Sync must only receive an identity from a successful authentication verifier.
// It does not claim existing resources or infer a membership/role from a token.
func (s IdentityStore) Sync(ctx context.Context, identity Identity) (Identity, Ownership, error) {
	if s.DB == nil {
		return identity, Ownership{}, errors.New("identity database is unavailable")
	}
	if identity.ID == "" || secureURL(identity.Issuer) != nil ||
		(identity.OrganizationID == "" && identity.OrganizationSlug != "") {
		return identity, Ownership{}, errors.New("invalid verified identity")
	}
	var orgName string
	if identity.OrganizationID != "" && s.OrganizationResolver != nil {
		org, err := s.OrganizationResolver.ResolveOrganization(ctx, identity.OrganizationID)
		if err != nil {
			return identity, Ownership{}, err
		}
		if org.ID != identity.OrganizationID {
			return identity, Ownership{}, errors.New("organization resolver returned a different organization")
		}
		identity.OrganizationSlug = strings.TrimSpace(org.Slug)
		orgName = org.Name
	}
	var ownership Ownership
	err := s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		user := User{ID: uuid.New(), Issuer: identity.Issuer, ClerkUserID: identity.ID, Name: identity.Name, Email: identity.Email, EmailVerified: identity.EmailVerified}
		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "issuer"}, {Name: "clerk_user_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"name", "email", "email_verified", "updated_at"}),
		}).Create(&user).Error; err != nil {
			return err
		}
		user = User{}
		if err := tx.Where("issuer = ? AND clerk_user_id = ?", identity.Issuer, identity.ID).First(&user).Error; err != nil {
			return err
		}
		ownership.UserID = user.ID
		ownership.OrgID = tenant.LocalOrgID
		if identity.OrganizationID == "" {
			return nil
		}
		org, err := upsertOrg(tx, identity.Issuer, identity.OrganizationID, identity.OrganizationSlug, orgName)
		if err != nil {
			return err
		}
		ownership.OrgID = org.ID
		if org.Slug != nil {
			identity.OrganizationSlug = *org.Slug
		}
		return nil
	})
	if err != nil {
		return identity, Ownership{}, err
	}
	return identity, ownership, nil
}

// upsertOrg creates or refreshes an organization by its issuer and Clerk ID.
// An empty slug or name leaves the stored value untouched.
func upsertOrg(tx *gorm.DB, issuer, clerkOrgID, slug, name string) (Org, error) {
	org := Org{ID: uuid.New(), Issuer: issuer, ClerkOrgID: clerkOrgID, Name: name}
	updates := []string{"updated_at"}
	if slug = strings.TrimSpace(slug); slug != "" {
		org.Slug = &slug
		updates = append(updates, "slug")
	}
	if name != "" {
		updates = append(updates, "name")
	}
	if err := tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "issuer"}, {Name: "clerk_org_id"}},
		DoUpdates: clause.AssignmentColumns(updates),
	}).Create(&org).Error; err != nil {
		return Org{}, err
	}
	org = Org{}
	err := tx.Where("issuer = ? AND clerk_org_id = ?", issuer, clerkOrgID).First(&org).Error
	return org, err
}

// SyncMemberships replaces a user's organization memberships with memberships,
// creating or refreshing each organization under the user's issuer. Dropped
// memberships are removed; their organizations are kept.
func (s IdentityStore) SyncMemberships(ctx context.Context, userID uuid.UUID, memberships []Membership) error {
	if s.DB == nil {
		return errors.New("identity database is unavailable")
	}
	for _, membership := range memberships {
		if membership.OrgID == "" || membership.Role == "" {
			return errors.New("invalid organization membership")
		}
	}
	return s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var issuer string
		result := tx.Raw(`SELECT issuer FROM users WHERE id = ?`, userID).Scan(&issuer)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return fmt.Errorf("sync memberships: unknown user %s", userID)
		}
		if err := tx.Exec(`DELETE FROM org_memberships WHERE user_id = ?`, userID).Error; err != nil {
			return err
		}
		for _, membership := range memberships {
			org, err := upsertOrg(tx, issuer, membership.OrgID, membership.OrgSlug, membership.OrgName)
			if err != nil {
				return err
			}
			if err := tx.Exec(`
				INSERT INTO org_memberships (user_id, org_id, role, synced_at) VALUES (?, ?, ?, now())`,
				userID, org.ID, membership.Role,
			).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
