// Tests the model key store against a migrated database: encryption bound to its org and
// provider, one key per provider, validation, and the error a changed GAVEL_SECRET_KEY gives.
package modelkeys_test

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"

	commonsdbtest "github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/commons-db/secret"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/modelkeys"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

const (
	apiKey     = "sk-ant-api03-abcdefghijklmnop-WXYZ"
	oauthToken = "sk-ant-oat01-qrstuvwxyz012345-9876"
)

type fixture struct {
	db    *gorm.DB
	store modelkeys.Store
	acme  uuid.UUID
	other uuid.UUID
	ada   uuid.UUID
}

func newFixture(t *testing.T) fixture {
	t.Helper()
	dsn := commonsdbtest.ForT(t, commonsdbtest.Options{Name: "gavel_modelkeys"}).DSN()
	t.Setenv(database.EnvDSN, dsn)
	opened, err := database.Open(t.Context(), database.WithMigrations())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, opened.Close()) })
	db := opened.Gorm()
	f := fixture{db: db, acme: uuid.New(), other: uuid.New(), ada: uuid.New()}
	for _, org := range []uuid.UUID{f.acme, f.other} {
		require.NoError(t, db.Exec(`INSERT INTO orgs (id, issuer, clerk_org_id, name) VALUES (?, 'https://clerk.example', ?, 'Org')`, org, org.String()).Error)
	}
	require.NoError(t, db.Exec(`INSERT INTO users (id, issuer, clerk_user_id) VALUES (?, 'https://clerk.example', 'user_ada')`, f.ada).Error)
	useKeeper(t, modelkeys.KeeperConnection([]byte("a secret of at least thirty-two bytes")))
	f.store = modelkeys.Store{DB: db}
	return f
}

// useKeeper points commons-db's secret keeper at connection for the rest of t.
func useKeeper(t *testing.T, connection string) {
	t.Helper()
	previous := secret.KMSConnection
	secret.KMSConnection = connection
	secret.InvalidateKeeper()
	t.Cleanup(func() {
		secret.KMSConnection = previous
		secret.InvalidateKeeper()
	})
}

func TestSetKeyIsReadBackAndListedWithoutItsValue(t *testing.T) {
	f := newFixture(t)
	require.NoError(t, f.store.Set(t.Context(), f.acme, &f.ada, "anthropic-api-key", apiKey))

	value, err := f.store.Value(t.Context(), f.acme, "anthropic")
	require.NoError(t, err)
	assert.Equal(t, apiKey, value)

	keys, err := f.store.List(t.Context(), f.acme)
	require.NoError(t, err)
	require.Len(t, keys, 1)
	assert.Equal(t, "anthropic", keys[0].Provider)
	assert.Equal(t, "anthropic-api-key", keys[0].Kind)
	assert.Equal(t, "WXYZ", keys[0].Hint)
	assert.Equal(t, &f.ada, keys[0].UpdatedBy)
	assert.False(t, keys[0].UpdatedAt.IsZero())
	listed, err := json.Marshal(keys)
	require.NoError(t, err)
	assert.NotContains(t, string(listed), "abcdefghijklmnop")

	var stored struct{ Ciphertext []byte }
	require.NoError(t, f.db.Raw(`SELECT ciphertext FROM org_model_keys WHERE org_id = ?`, f.acme).Scan(&stored).Error)
	require.NotEmpty(t, stored.Ciphertext)
	assert.NotContains(t, string(stored.Ciphertext), "abcdefghijklmnop")

	others, err := f.store.List(t.Context(), f.other)
	require.NoError(t, err)
	assert.Empty(t, others)
	_, err = f.store.Value(t.Context(), f.other, "anthropic")
	assert.ErrorIs(t, err, modelkeys.ErrNotFound)
}

func TestSetAnotherKindOfTheSameProviderReplacesTheKey(t *testing.T) {
	f := newFixture(t)
	require.NoError(t, f.store.Set(t.Context(), f.acme, &f.ada, "anthropic-api-key", apiKey))
	require.NoError(t, f.store.Set(t.Context(), f.acme, nil, "claude-code-oauth-token", oauthToken))

	keys, err := f.store.List(t.Context(), f.acme)
	require.NoError(t, err)
	require.Len(t, keys, 1)
	assert.Equal(t, "claude-code-oauth-token", keys[0].Kind)
	assert.Equal(t, "9876", keys[0].Hint)
	assert.Nil(t, keys[0].UpdatedBy)
	value, err := f.store.Value(t.Context(), f.acme, "anthropic")
	require.NoError(t, err)
	assert.Equal(t, oauthToken, value)
}

func TestValueRefusesAKeyCopiedToAnotherOrg(t *testing.T) {
	f := newFixture(t)
	require.NoError(t, f.store.Set(t.Context(), f.acme, nil, "anthropic-api-key", apiKey))
	require.NoError(t, f.db.Exec(`INSERT INTO org_model_keys (org_id, provider, kind, ciphertext, hint)
		SELECT ?, provider, kind, ciphertext, hint FROM org_model_keys WHERE org_id = ?`, f.other, f.acme).Error)

	_, err := f.store.Value(t.Context(), f.other, "anthropic")
	require.Error(t, err)
	assert.Equal(t, "organization "+f.other.String()+"'s anthropic key was encrypted for another organization or provider; an organization admin must set it again", err.Error())
	assert.NotContains(t, err.Error(), "abcdefghijklmnop")
}

func TestValueAfterTheKeeperChangedSaysSo(t *testing.T) {
	f := newFixture(t)
	require.NoError(t, f.store.Set(t.Context(), f.acme, nil, "anthropic-api-key", apiKey))

	useKeeper(t, modelkeys.KeeperConnection([]byte("another secret of at least thirty-two bytes")))
	_, err := f.store.Value(t.Context(), f.acme, "anthropic")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "can't decrypt organization "+f.acme.String()+"'s anthropic key; GAVEL_SECRET_KEY changed since it was set, so an organization admin must set it again: ")
}

func TestSetRejectsKeysTheGatewayCantInject(t *testing.T) {
	f := newFixture(t)
	for name, test := range map[string]struct{ kind, value, want string }{
		"unknown kind":     {"openai-api-key", "sk-proj-1", `invalid model key: unknown kind "openai-api-key"; kinds are anthropic-api-key, claude-code-oauth-token`},
		"empty":            {"anthropic-api-key", "", "invalid model key: the key is empty"},
		"too long":         {"anthropic-api-key", "sk-ant-api" + strings.Repeat("a", 4096), "invalid model key: the key is longer than 4096 bytes"},
		"inner whitespace": {"anthropic-api-key", "sk-ant-api03 abc", "invalid model key: the key contains whitespace or a control character"},
		"control":          {"anthropic-api-key", "sk-ant-api03\x00abc", "invalid model key: the key contains whitespace or a control character"},
		"wrong kind":       {"anthropic-api-key", oauthToken, `invalid model key: an Anthropic API key starts with "sk-ant-api"`},
		"wrong kind oauth": {"claude-code-oauth-token", apiKey, `invalid model key: a Claude Code OAuth token starts with "sk-ant-oat"`},
	} {
		err := f.store.Set(t.Context(), f.acme, nil, test.kind, test.value)
		require.ErrorIs(t, err, modelkeys.ErrInvalidKey, name)
		assert.Equal(t, test.want, err.Error(), name)
		if test.value != "" {
			assert.NotContains(t, err.Error(), test.value, name)
		}
	}
	keys, err := f.store.List(t.Context(), f.acme)
	require.NoError(t, err)
	assert.Empty(t, keys)
}

func TestSetTrimsSurroundingWhitespace(t *testing.T) {
	f := newFixture(t)
	require.NoError(t, f.store.Set(t.Context(), f.acme, nil, "anthropic-api-key", "  "+apiKey+"\n"))
	value, err := f.store.Value(t.Context(), f.acme, "anthropic")
	require.NoError(t, err)
	assert.Equal(t, apiKey, value)
}

func TestDeleteRemovesTheProvidersKey(t *testing.T) {
	f := newFixture(t)
	require.NoError(t, f.store.Set(t.Context(), f.acme, nil, "anthropic-api-key", apiKey))
	require.NoError(t, f.store.Delete(t.Context(), f.acme, "anthropic"))

	_, err := f.store.Value(t.Context(), f.acme, "anthropic")
	assert.ErrorIs(t, err, modelkeys.ErrNotFound)
	assert.ErrorIs(t, f.store.Delete(t.Context(), f.acme, "anthropic"), modelkeys.ErrNotFound)
}

func TestKeeperConnectionIsALocalKeyDerivedPerSecret(t *testing.T) {
	first := modelkeys.KeeperConnection([]byte("a secret of at least thirty-two bytes"))
	encoded, ok := strings.CutPrefix(first, "local_kms://")
	require.True(t, ok, first)
	key, err := base64.URLEncoding.DecodeString(encoded)
	require.NoError(t, err)
	assert.Len(t, key, 32)
	assert.NotContains(t, first, base64.URLEncoding.EncodeToString([]byte("a secret of at least thirty-two bytes")))
	assert.Equal(t, first, modelkeys.KeeperConnection([]byte("a secret of at least thirty-two bytes")))
	assert.NotEqual(t, first, modelkeys.KeeperConnection([]byte("another secret of at least thirty-two bytes")))
}

func TestKindsDescribeHowTheGatewayInjectsEachKey(t *testing.T) {
	apiKind, ok := modelkeys.KindNamed("anthropic-api-key")
	require.True(t, ok)
	assert.Equal(t, modelkeys.Kind{
		Name: "anthropic-api-key", Provider: "anthropic", Label: "Anthropic API key",
		Host: "api.anthropic.com", Header: "x-api-key", Env: "ANTHROPIC_API_KEY", KeyPrefix: "sk-ant-api",
	}, apiKind)
	oauthKind, ok := modelkeys.KindNamed("claude-code-oauth-token")
	require.True(t, ok)
	assert.Equal(t, modelkeys.Kind{
		Name: "claude-code-oauth-token", Provider: "anthropic", Label: "Claude Code OAuth token",
		Host: "api.anthropic.com", Header: "Authorization", Prefix: "Bearer ", Env: "CLAUDE_CODE_OAUTH_TOKEN", KeyPrefix: "sk-ant-oat",
	}, oauthKind)
	_, ok = modelkeys.KindNamed("openai-api-key")
	assert.False(t, ok)
}
