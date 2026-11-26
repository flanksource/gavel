// Keeps a configured path inside the directory it must stay in: resolving it, and reading the
// file it names without following a symlink out.
package utils

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ErrPathEscapesBase reports a path that resolves outside the directory it is
// required to stay in. Callers wrap it so the message also names the setting or
// file that supplied the value.
var ErrPathEscapesBase = errors.New("path escapes its base directory")

// ResolveWithin resolves name against base and returns the absolute result only
// when it stays inside base.
//
// The paths this guards are configuration, not filesystem capabilities: a
// `todos.lifecycle.file`, a prompt override's `file:`, a snapshot pointer. Each
// names something under a directory gavel already decided to read, so a `..`
// segment — or an absolute path pointing somewhere else on disk — means the
// value is not what the caller asked for. It is rejected with an error naming
// both the base and the offending value rather than silently clamped to
// whatever the traversal happens to land on.
//
// A relative name is joined onto base; an absolute name is accepted only when
// it already lives under base, base itself included. Containment is lexical —
// symlinks inside base are not followed — so it constrains what a configured
// path may name, not what the filesystem underneath base may point at.
func ResolveWithin(base, name string) (string, error) {
	root, rel, err := resolveWithin(base, name)
	if err != nil {
		return "", err
	}
	return filepath.Join(root, rel), nil
}

// ReadFileWithin reads the file name names under base, as ResolveWithin resolves it, through an
// os.Root at base, so a symlink under base that leads outside it is refused too. Use it for a
// path a repository's own content supplies: the repository can check in such a symlink.
func ReadFileWithin(base, name string) ([]byte, error) {
	root, rel, err := resolveWithin(base, name)
	if err != nil {
		return nil, err
	}
	dir, err := os.OpenRoot(root)
	if err != nil {
		return nil, fmt.Errorf("read %q under %s: %w", name, root, err)
	}
	defer dir.Close() //nolint:errcheck
	data, err := dir.ReadFile(rel)
	if err != nil {
		return nil, fmt.Errorf("read %q under %s: %w", name, root, err)
	}
	return data, nil
}

// resolveWithin returns base as a clean absolute root and name as a path relative to it, or an
// error when name resolves outside root.
func resolveWithin(base, name string) (string, string, error) {
	if strings.TrimSpace(base) == "" {
		return "", "", fmt.Errorf("resolve %q: a base directory is required", name)
	}
	if strings.TrimSpace(name) == "" {
		return "", "", fmt.Errorf("resolve a path under %s: a path is required", base)
	}

	root, err := filepath.Abs(base)
	if err != nil {
		return "", "", fmt.Errorf("resolve base directory %q: %w", base, err)
	}
	root = filepath.Clean(root)

	rel := filepath.Clean(name)
	if filepath.IsAbs(rel) {
		if rel, err = filepath.Rel(root, rel); err != nil {
			return "", "", fmt.Errorf("%w: %q cannot be expressed relative to %s", ErrPathEscapesBase, name, root)
		}
	}
	// filepath.IsLocal rejects ".", but base is trivially within base — a
	// workspace naming its own root is resolving to itself, not escaping.
	if rel != "." && !filepath.IsLocal(rel) {
		return "", "", fmt.Errorf("%w: %q resolves outside %s", ErrPathEscapesBase, name, root)
	}
	return root, rel, nil
}
