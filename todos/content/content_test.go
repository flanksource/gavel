// Tests resolving a TODO's content fields: literal content is never read from a file, whatever
// it looks like, so a caller that may not read this host's files can still pass text.
package content

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLiteralContentNamesNoFile(t *testing.T) {
	dir := t.TempDir()
	secret := filepath.Join(dir, "secret.md")
	require.NoError(t, os.WriteFile(secret, []byte("host secret"), 0o600))

	for _, value := range []string{"@secret.md", "secret.md", "@" + secret, secret} {
		text, err := Resolve(Options{WorkDir: dir, Flag: "--body", Value: value, Literal: true})
		require.NoError(t, err)
		assert.Equal(t, value, text, "%q is text, not a file", value)

		created, err := ResolveCreate(dir, CreateOptions{Body: value, BodySet: true, Plan: value, PlanSet: true, Literal: true})
		require.NoError(t, err)
		assert.Equal(t, value, created.Body)
		assert.Equal(t, value, created.Plan)
	}

	read, err := Resolve(Options{WorkDir: dir, Flag: "--body", Value: "@secret.md"})
	require.NoError(t, err)
	assert.Equal(t, "host secret", read, "without Literal an @path is read")
}
