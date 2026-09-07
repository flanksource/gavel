package utils

import (
	"errors"
	"fmt"
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
// it already lives under base. Containment is lexical — symlinks inside base
// are not followed — so it constrains what a configured path may name, not what
// the filesystem underneath base may point at.
func ResolveWithin(base, name string) (string, error) {
	if strings.TrimSpace(base) == "" {
		return "", fmt.Errorf("resolve %q: a base directory is required", name)
	}
	if strings.TrimSpace(name) == "" {
		return "", fmt.Errorf("resolve a path under %s: a path is required", base)
	}

	root, err := filepath.Abs(base)
	if err != nil {
		return "", fmt.Errorf("resolve base directory %q: %w", base, err)
	}
	root = filepath.Clean(root)

	rel := filepath.Clean(name)
	if filepath.IsAbs(rel) {
		if rel, err = filepath.Rel(root, rel); err != nil {
			return "", fmt.Errorf("%w: %q cannot be expressed relative to %s", ErrPathEscapesBase, name, root)
		}
	}
	if !filepath.IsLocal(rel) {
		return "", fmt.Errorf("%w: %q resolves outside %s", ErrPathEscapesBase, name, root)
	}
	return filepath.Join(root, rel), nil
}
