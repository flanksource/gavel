// Records which organization and user own each Captain session gavel creates.
// Captain's tables stay untouched; ownership lives in gavel's session_owners sidecar.
package native

import (
	"context"

	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/gavel/tenant"
	"github.com/google/uuid"
)

// recordSessionOwners records the repository's organization and the request's
// user as the owner of each session and of its root, inside the transaction
// that created them. A session that already has an owner keeps it.
func (r *Repository) recordSessionOwners(ctx context.Context, tx *captaindb.DB, sessions []*captaindb.Session) error {
	userID := tenant.FromContext(ctx).UserID
	for _, session := range sessions {
		ids := []uuid.UUID{session.ID}
		if session.RootSessionID != nil {
			ids = append(ids, *session.RootSessionID)
		}
		for _, id := range ids {
			if err := tx.Gorm().WithContext(ctx).Exec(`
				INSERT INTO session_owners (session_id, org_id, user_id) VALUES (?, ?, ?)
				ON CONFLICT (session_id) DO NOTHING`, id, r.org, userID,
			).Error; err != nil {
				return err
			}
		}
	}
	return nil
}
