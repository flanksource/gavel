// Tests that a prefixed store passes the conformance suite, keeps its objects under its prefix
// in the underlying store, and can't see a sibling prefix's objects.
package blob

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrefixConformance(t *testing.T) {
	testStore(t, WithPrefix(Local(t.TempDir()), "orgs/a/"))
}

func TestPrefixIsolatesStores(t *testing.T) {
	ctx := context.Background()
	base := Local(t.TempDir())
	a, b := WithPrefix(base, "orgs/a/"), WithPrefix(base, "orgs/b/")
	require.NoError(t, a.Write(ctx, "attachments/x.png", strings.NewReader("a's")))

	reader, err := base.Read(ctx, "orgs/a/attachments/x.png")
	require.NoError(t, err, "the object sits under the prefix in the underlying store")
	data, err := io.ReadAll(reader)
	require.NoError(t, err)
	require.NoError(t, reader.Close())
	assert.Equal(t, "a's", string(data))

	_, err = b.Read(ctx, "attachments/x.png")
	require.ErrorIs(t, err, ErrNotFound, "another prefix can't read it")
	assert.NotContains(t, err.Error(), "orgs/a", "errors name the caller's key, not the prefix")
	listed, err := b.List(ctx, "")
	require.NoError(t, err)
	assert.Empty(t, listed, "another prefix can't list it")
}

func TestWithPrefixRejectsInvalidPrefixes(t *testing.T) {
	for _, prefix := range []string{"", "orgs/a", "/orgs/a/", "orgs/../a/"} {
		assert.Panics(t, func() { WithPrefix(Local(t.TempDir()), prefix) }, prefix)
	}
}

func TestWorkspacePrefix(t *testing.T) {
	workDir := t.TempDir()
	prefix, err := WorkspacePrefix(workDir)
	require.NoError(t, err)
	sum := sha256.Sum256([]byte(workDir))
	assert.Equal(t, "workspaces/"+hex.EncodeToString(sum[:])[:16]+"/", prefix)

	other, err := WorkspacePrefix(t.TempDir())
	require.NoError(t, err)
	assert.NotEqual(t, prefix, other)
	assert.NotPanics(t, func() { WithPrefix(Local(t.TempDir()), prefix) }, "usable as a prefix")
}
