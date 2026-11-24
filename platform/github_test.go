// Tests the platform dashboard's view of GitHub App installations against a migrated
// database: which organization each is linked to, and how many repositories it brought.
package platform_test

import (
	"testing"

	"github.com/flanksource/gavel/platform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLinkedInstallations(t *testing.T) {
	db := openDB(t)
	alpha := addOrg(t, db, issuer, "org_alpha", "alpha", "Alpha")
	other := addOrg(t, db, "https://other-clerk.example", "org_other", "other", "Other instance")
	require.NoError(t, db.Exec(`INSERT INTO github_installations (id, org_id, account_login, account_type) VALUES (7, ?, 'alpha-gh', 'Organization'), (8, ?, 'other-gh', 'User')`, alpha, other).Error)
	require.NoError(t, db.Exec(`INSERT INTO repos (org_id, name, source, github_repo_id, installation_id) VALUES (?, 'alpha-gh/api', 'github', 1, 7), (?, 'alpha-gh/web', 'github', 2, 7)`, alpha, alpha).Error)

	linked, err := platform.Store{DB: db, Issuer: issuer}.LinkedInstallations(t.Context())
	require.NoError(t, err)
	require.Len(t, linked, 1, "only the Clerk instance's organizations' installations: %+v", linked)
	assert.Equal(t, int64(7), linked[0].ID)
	assert.Equal(t, alpha, linked[0].OrgID)
	assert.Equal(t, "alpha", linked[0].OrgSlug)
	assert.Equal(t, "Alpha", linked[0].OrgName)
	assert.Equal(t, "alpha-gh", linked[0].AccountLogin)
	assert.Equal(t, 2, linked[0].Repos)
	assert.False(t, linked[0].LinkedAt.IsZero())
}
