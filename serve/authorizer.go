// DBAuthorizer decides hosted pushes from the database: a key identifies a user, the
// path names an org they belong to in Clerk, and each pushed repository is recorded.
package serve

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/api"
	"github.com/flanksource/gavel/sshkeys"
	"github.com/gliderlabs/ssh"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DBAuthorizer authorizes pushes to <org-slug>/<repo> for the org's members. Any role may
// push. Bare repositories live under DataDir/repos/<org id>/<repo id>.git.
type DBAuthorizer struct {
	DB      *gorm.DB
	Members *api.Members
	// Issuer is the Clerk instance whose users push.
	Issuer  string
	DataDir string
}

func (a DBAuthorizer) UserForKey(ctx context.Context, key ssh.PublicKey) (uuid.UUID, bool) {
	user, ok, err := sshkeys.Store{DB: a.DB}.UserForKey(ctx, key)
	if err != nil {
		logger.Warnf("look up a pushing SSH key: %v", err)
		return uuid.Nil, false
	}
	return user, ok
}

func (a DBAuthorizer) AuthorizePush(ctx context.Context, user uuid.UUID, repoPath string) (Target, error) {
	slug, name, ok := strings.Cut(repoPath, "/")
	if !ok {
		return Target{}, fmt.Errorf("push to <org>/<repo>, not %q", repoPath)
	}
	var clerkUser struct{ ClerkUserID string }
	if err := a.DB.WithContext(ctx).Raw(`SELECT clerk_user_id FROM users WHERE id = ? AND issuer = ?`, user, a.Issuer).
		Scan(&clerkUser).Error; err != nil {
		return Target{}, fmt.Errorf("read the pushing user: %w", err)
	}
	if clerkUser.ClerkUserID == "" {
		return Target{}, fmt.Errorf("unknown user %s", user)
	}
	scope, err := a.Members.ForAccessToken(ctx, clerkUser.ClerkUserID, slug)
	if err != nil {
		return Target{}, err
	}
	var repo struct{ ID uuid.UUID }
	if err := a.DB.WithContext(ctx).Raw(`
		INSERT INTO repos (org_id, name, source) VALUES (?, ?, 'push')
		ON CONFLICT (org_id, source, name) DO UPDATE SET updated_at = now()
		RETURNING id`, scope.OrgID, name).Scan(&repo).Error; err != nil {
		return Target{}, fmt.Errorf("record the pushed repository: %w", err)
	}
	return Target{
		OrgID:   scope.OrgID,
		RepoID:  repo.ID,
		BareDir: filepath.Join(a.DataDir, "repos", scope.OrgID.String(), repo.ID.String()+".git"),
	}, nil
}
