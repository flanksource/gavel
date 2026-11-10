// Package tenant carries the organization and user a request or run acts for.
// The local organization owns all data when gavel runs without authentication.
package tenant

import (
	"context"

	"github.com/google/uuid"
)

// LocalOrgID is the built-in organization used by --auth none.
var LocalOrgID = uuid.MustParse("00000000-0000-0000-0000-000000000001")

// Scope is the organization and user a request or run acts for.
type Scope struct {
	OrgID       uuid.UUID
	UserID      *uuid.UUID // nil for local mode and for system jobs
	Role        string     // Clerk org role, e.g. "org:admin", "org:member"; "" in local mode
	GitHubLogin string     // viewer login for the @me PR filter; "" when unknown
}

type scopeKey struct{}

// Local returns the scope of the local organization with no user.
func Local() Scope { return Scope{OrgID: LocalOrgID} }

// WithScope returns a context carrying s.
func WithScope(ctx context.Context, s Scope) context.Context {
	return context.WithValue(ctx, scopeKey{}, s)
}

// FromContext returns the scope on ctx, or Local() when there is none.
func FromContext(ctx context.Context) Scope {
	if s, ok := ctx.Value(scopeKey{}).(Scope); ok {
		return s
	}
	return Local()
}

// IsLocal reports whether s acts for the local organization.
func (s Scope) IsLocal() bool { return s.OrgID == LocalOrgID }

// IsAdmin reports whether s acts as an admin of its organization.
func (s Scope) IsAdmin() bool { return s.Role == "org:admin" }
