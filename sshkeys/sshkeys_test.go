// Tests the SSH key store against a migrated database: accepted key types, duplicates,
// owner-only deletion, and resolving a key to its user for the push server.
package sshkeys_test

import (
	"context"
	"crypto/dsa"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"strings"
	"testing"

	commonsdbtest "github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/sshkeys"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gossh "golang.org/x/crypto/ssh"
)

func authorizedKey(t *testing.T, public any) string {
	t.Helper()
	key, err := gossh.NewPublicKey(public)
	require.NoError(t, err)
	return strings.TrimSpace(string(gossh.MarshalAuthorizedKey(key))) + " ada@laptop"
}

func ed25519Key(t *testing.T) (string, gossh.PublicKey) {
	t.Helper()
	public, _, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	key, err := gossh.NewPublicKey(public)
	require.NoError(t, err)
	return authorizedKey(t, public), key
}

func newStore(t *testing.T) (sshkeys.Store, uuid.UUID, uuid.UUID) {
	t.Helper()
	dsn := commonsdbtest.ForT(t, commonsdbtest.Options{Name: "gavel_sshkeys"}).DSN()
	t.Setenv(database.EnvDSN, dsn)
	opened, err := database.Open(t.Context(), database.WithMigrations())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, opened.Close()) })
	ada, bob := uuid.New(), uuid.New()
	for _, user := range []struct {
		id    uuid.UUID
		clerk string
	}{{ada, "user_ada"}, {bob, "user_bob"}} {
		require.NoError(t, opened.Gorm().Exec(`INSERT INTO users (id, issuer, clerk_user_id) VALUES (?, 'https://clerk.example', ?)`, user.id, user.clerk).Error)
	}
	return sshkeys.Store{DB: opened.Gorm()}, ada, bob
}

func TestParseAcceptsOnlyStrongKeys(t *testing.T) {
	ecdsaKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	rsa3072, err := rsa.GenerateKey(rand.Reader, 3072)
	require.NoError(t, err)
	rsa2048, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	var dsaKey dsa.PrivateKey
	require.NoError(t, dsa.GenerateParameters(&dsaKey.Parameters, rand.Reader, dsa.L1024N160))
	require.NoError(t, dsa.GenerateKey(&dsaKey, rand.Reader))
	edLine, _ := ed25519Key(t)

	for name, line := range map[string]string{"ed25519": edLine, "ecdsa": authorizedKey(t, &ecdsaKey.PublicKey), "rsa 3072": authorizedKey(t, &rsa3072.PublicKey)} {
		_, err := sshkeys.Parse(line)
		assert.NoError(t, err, name)
	}
	for name, test := range map[string]struct{ line, want string }{
		"rsa 2048":  {authorizedKey(t, &rsa2048.PublicKey), "ssh-rsa keys must be at least 3072 bits, this one is 2048"},
		"dsa":       {authorizedKey(t, &dsaKey.PublicKey), `ssh-dss keys aren't accepted; use ssh-ed25519`},
		"not a key": {"hello", "not an SSH public key"},
		"empty":     {"  ", "not an SSH public key"},
	} {
		_, err := sshkeys.Parse(test.line)
		require.ErrorIs(t, err, sshkeys.ErrUnsupportedKey, name)
		assert.Contains(t, err.Error(), test.want, name)
	}
}

func TestStore(t *testing.T) {
	store, ada, bob := newStore(t)
	ctx := context.Background()
	line, public := ed25519Key(t)

	added, err := store.Add(ctx, ada, "laptop", line)
	require.NoError(t, err)
	assert.Equal(t, "laptop", added.Name)
	assert.Equal(t, gossh.FingerprintSHA256(public), added.Fingerprint)
	assert.Nil(t, added.LastUsedAt)

	_, err = store.Add(ctx, bob, "stolen", line)
	require.ErrorIs(t, err, sshkeys.ErrDuplicate, "a key identifies one user")

	keys, err := store.List(ctx, ada)
	require.NoError(t, err)
	require.Len(t, keys, 1)
	assert.Equal(t, added.ID, keys[0].ID)
	keys, err = store.List(ctx, bob)
	require.NoError(t, err)
	assert.Empty(t, keys)

	user, ok, err := store.UserForKey(ctx, public)
	require.NoError(t, err)
	require.True(t, ok)
	assert.Equal(t, ada, user)
	keys, err = store.List(ctx, ada)
	require.NoError(t, err)
	assert.NotNil(t, keys[0].LastUsedAt, "a push records when the key was used")
	_, unknown := ed25519Key(t)
	_, ok, err = store.UserForKey(ctx, unknown)
	require.NoError(t, err)
	assert.False(t, ok)

	require.ErrorIs(t, store.Delete(ctx, bob, added.ID), sshkeys.ErrNotFound, "only the owner deletes a key")
	require.NoError(t, store.Delete(ctx, ada, added.ID))
	require.ErrorIs(t, store.Delete(ctx, ada, added.ID), sshkeys.ErrNotFound)
	_, ok, err = store.UserForKey(ctx, public)
	require.NoError(t, err)
	assert.False(t, ok)
}
