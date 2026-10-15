// Conformance suite every blob.Store passes: write, read, exists, list and delete, including
// the not-found cases, plus the key validation all stores share.
package blob

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testStore runs the conformance suite against store, which must start empty.
func testStore(t *testing.T, store Store) {
	t.Helper()
	ctx := context.Background()
	write := func(key, body string) {
		t.Helper()
		require.NoError(t, store.Write(ctx, key, strings.NewReader(body)), "write %s", key)
	}
	read := func(key string) string {
		t.Helper()
		reader, err := store.Read(ctx, key)
		require.NoError(t, err, "read %s", key)
		defer reader.Close()
		data, err := io.ReadAll(reader)
		require.NoError(t, err, "read %s", key)
		return string(data)
	}
	keys := func(prefix string) []string {
		t.Helper()
		objects, err := store.List(ctx, prefix)
		require.NoError(t, err, "list %q", prefix)
		out := []string{}
		for _, object := range objects {
			out = append(out, object.Key)
		}
		return out
	}

	t.Run("write and read", func(t *testing.T) {
		write("runs/run-1.json", `{"run":1}`)
		assert.Equal(t, `{"run":1}`, read("runs/run-1.json"))
		write("runs/run-1.json", `{"run":"replaced"}`)
		assert.Equal(t, `{"run":"replaced"}`, read("runs/run-1.json"), "a write replaces the object")
	})

	t.Run("read a missing key", func(t *testing.T) {
		_, err := store.Read(ctx, "runs/missing.json")
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrNotFound), "want ErrNotFound, got %v", err)
		assert.Contains(t, err.Error(), "runs/missing.json")
	})

	t.Run("exists", func(t *testing.T) {
		exists, err := store.Exists(ctx, "runs/run-1.json")
		require.NoError(t, err)
		assert.True(t, exists)
		exists, err = store.Exists(ctx, "runs/missing.json")
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("list", func(t *testing.T) {
		write("runs/run-2.json", "22")
		write("runs/nested/deep.json", "deep")
		write("recordings/a.har", "har")
		assert.Equal(t, []string{"recordings/a.har", "runs/nested/deep.json", "runs/run-1.json", "runs/run-2.json"}, keys(""))
		assert.Equal(t, []string{"runs/nested/deep.json", "runs/run-1.json", "runs/run-2.json"}, keys("runs/"))
		assert.Equal(t, []string{"runs/run-1.json", "runs/run-2.json"}, keys("runs/run-"), "a prefix need not end at a slash")
		assert.Empty(t, keys("missing/"), "listing a prefix with nothing under it")

		objects, err := store.List(ctx, "runs/run-2.json")
		require.NoError(t, err)
		require.Len(t, objects, 1)
		assert.EqualValues(t, 2, objects[0].Size)
		assert.False(t, objects[0].ModTime.IsZero(), "objects carry a modification time")
	})

	t.Run("delete", func(t *testing.T) {
		require.NoError(t, store.Delete(ctx, "runs/run-2.json"))
		exists, err := store.Exists(ctx, "runs/run-2.json")
		require.NoError(t, err)
		assert.False(t, exists)
		require.NoError(t, store.Delete(ctx, "runs/run-2.json"), "deleting a missing key is not an error")
	})

	t.Run("invalid keys", func(t *testing.T) {
		for _, key := range []string{"", "/abs", "a/../b", `a\b`} {
			require.ErrorIs(t, store.Write(ctx, key, strings.NewReader("x")), ErrInvalidKey, "write %q", key)
			_, err := store.Read(ctx, key)
			require.ErrorIs(t, err, ErrInvalidKey, "read %q", key)
			_, err = store.Exists(ctx, key)
			require.ErrorIs(t, err, ErrInvalidKey, "exists %q", key)
			require.ErrorIs(t, store.Delete(ctx, key), ErrInvalidKey, "delete %q", key)
		}
		_, err := store.List(ctx, "../")
		require.ErrorIs(t, err, ErrInvalidKey, "list ../")
	})
}

func TestLocalConformance(t *testing.T) {
	testStore(t, Local(t.TempDir()))
}

func TestValidateKey(t *testing.T) {
	for key, want := range map[string]string{
		"":                "the key is empty",
		"/runs/a.json":    `key "/runs/a.json" is absolute`,
		"runs/../a.json":  `key "runs/../a.json" has a ".." segment`,
		"..":              `key ".." has a ".." segment`,
		`runs\a.json`:     `key "runs\\a.json" has a backslash`,
		"runs//a.json":    `key "runs//a.json" has an empty segment`,
		"runs/./a.json":   `key "runs/./a.json" has a "." segment`,
		"runs/run-1.json": "",
		"a":               "",
	} {
		err := validateKey(key)
		if want == "" {
			assert.NoError(t, err, key)
			continue
		}
		require.ErrorIs(t, err, ErrInvalidKey, key)
		assert.EqualError(t, err, "invalid blob key: "+want, key)
	}
}
