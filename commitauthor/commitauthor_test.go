// Tests the commit identity store against a migrated database: the GitHub default, the user's
// own name and email over it, and refusing what git would misread.
package commitauthor_test

import (
	"testing"

	commonsdbtest "github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/commitauthor"
	"github.com/flanksource/gavel/internal/database"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func openDB(t *testing.T) *gorm.DB {
	t.Helper()
	t.Setenv(database.EnvDSN, commonsdbtest.ForT(t, commonsdbtest.Options{Name: "gavel_commitauthor"}).DSN())
	opened, err := database.Open(t.Context(), database.WithMigrations())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, opened.Close()) })
	return opened.Gorm()
}

// user inserts a user with the given gavel account and GitHub connection, empty for none.
func user(t *testing.T, db *gorm.DB, name, email, githubLogin string, githubID int64) uuid.UUID {
	t.Helper()
	id := uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO users (id, issuer, clerk_user_id, name, email, github_login, github_id)
		VALUES (?, 'https://clerk.example', ?, ?, ?, NULLIF(?, ''), NULLIF(?, 0))`,
		id, "user_"+id.String()[:8], name, email, githubLogin, githubID).Error)
	return id
}

func TestAUsersCommitsDefaultToTheirGitHubIdentity(t *testing.T) {
	db := openDB(t)
	store := commitauthor.Store{DB: db}
	ada := user(t, db, "Ada Lovelace", "ada@example.com", "ada", 42)

	identity, err := store.Get(t.Context(), ada)
	require.NoError(t, err)
	assert.Equal(t, commitauthor.Identity{Default: commitauthor.Author{Name: "ada", Email: "42+ada@users.noreply.github.com"}}, identity)
	author, ok := identity.Author()
	assert.True(t, ok)
	assert.Equal(t, commitauthor.Author{Name: "ada", Email: "42+ada@users.noreply.github.com"}, author, "GitHub links the commits to their profile")
}

func TestAGitHubLoginWithoutItsIDDefaultsToTheLoginsNoreplyAddress(t *testing.T) {
	db := openDB(t)
	identity, err := commitauthor.Store{DB: db}.Get(t.Context(), user(t, db, "Grace", "grace@example.com", "grace", 0))
	require.NoError(t, err)
	assert.Equal(t, commitauthor.Author{Name: "grace", Email: "grace@users.noreply.github.com"}, identity.Default)
}

func TestWithoutGitHubCommitsDefaultToTheGavelAccount(t *testing.T) {
	db := openDB(t)
	identity, err := commitauthor.Store{DB: db}.Get(t.Context(), user(t, db, "Alan Turing", "alan@example.com", "", 0))
	require.NoError(t, err)
	assert.Equal(t, commitauthor.Author{Name: "Alan Turing", Email: "alan@example.com"}, identity.Default)
}

func TestAUserWithNoEmailAnywhereHasNoAuthor(t *testing.T) {
	db := openDB(t)
	identity, err := commitauthor.Store{DB: db}.Get(t.Context(), user(t, db, "Nobody", "", "", 0))
	require.NoError(t, err)
	_, ok := identity.Author()
	assert.False(t, ok, "gavel is the author instead")
}

func TestTheUsersOwnNameAndEmailOverrideEachDefault(t *testing.T) {
	db := openDB(t)
	store := commitauthor.Store{DB: db}
	ada := user(t, db, "Ada Lovelace", "ada@example.com", "ada", 42)

	identity, err := store.Set(t.Context(), ada, "  Ada Lovelace ", "")
	require.NoError(t, err)
	assert.Equal(t, "Ada Lovelace", identity.Name)
	author, _ := identity.Author()
	assert.Equal(t, commitauthor.Author{Name: "Ada Lovelace", Email: "42+ada@users.noreply.github.com"}, author, "an empty field keeps its default")

	_, err = store.Set(t.Context(), ada, "Ada Lovelace", "ada@work.example")
	require.NoError(t, err)
	identity, err = store.Get(t.Context(), ada)
	require.NoError(t, err)
	author, _ = identity.Author()
	assert.Equal(t, commitauthor.Author{Name: "Ada Lovelace", Email: "ada@work.example"}, author)

	identity, err = store.Set(t.Context(), ada, "", "")
	require.NoError(t, err)
	assert.Equal(t, commitauthor.Identity{Default: identity.Default}, identity, "clearing both goes back to GitHub's")
}

func TestRefusesANameOrEmailGitWouldMisread(t *testing.T) {
	db := openDB(t)
	store := commitauthor.Store{DB: db}
	ada := user(t, db, "Ada Lovelace", "ada@example.com", "ada", 42)

	for _, invalid := range []struct{ name, email, why string }{
		{"Ada <ada@example.com>", "", `the name can't contain "<" or ">"`},
		{"Ada\nLovelace", "", "the name must be one line"},
		{"Ada", "ada", "the email must be one address, like ada@example.com"},
		{"Ada", "ada lovelace@example.com", "the email must be one address, like ada@example.com"},
		{"Ada", "<ada@example.com>", "the email must be one address, like ada@example.com"},
	} {
		_, err := store.Set(t.Context(), ada, invalid.name, invalid.email)
		assert.ErrorIs(t, err, commitauthor.ErrInvalid, "%q <%q>", invalid.name, invalid.email)
		assert.ErrorContains(t, err, invalid.why)
	}
}
