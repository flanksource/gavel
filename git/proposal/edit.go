package proposal

import (
	"context"
	"fmt"
	"io/fs"
	"strings"
)

// Edit is one Edit-tool-style change: an exact OldString→NewString
// replacement that must match once unless ReplaceAll, a whole-file Content
// write that creates a missing file, or a Delete. Exactly one kind is set.
type Edit struct {
	Path       string
	OldString  string
	NewString  string
	ReplaceAll bool
	Content    *string
	Delete     bool
}

const newFileMode = "100644"

// file is a path's evolving state while a set of edits applies.
type file struct {
	mode    string
	content string
	exists  bool
}

func applyPatch(ctx context.Context, git call, base string, patch []byte) error {
	git.stdin = patch
	if _, err := git.run(ctx, "apply", "--cached", "-"); err != nil {
		return fmt.Errorf("proposal: patch does not apply to %s: %w", base, err)
	}
	return nil
}

// applyEdits applies edits in order on base's files, then writes each edited
// path's final state into git's index.
func applyEdits(ctx context.Context, git call, base string, edits []Edit) error {
	files := map[string]*file{}
	var order []string
	for i, e := range edits {
		if err := checkEdit(e); err != nil {
			return fmt.Errorf("proposal: edit %d: %w", i+1, err)
		}
		f, ok := files[e.Path]
		if !ok {
			var err error
			if f, err = readFile(ctx, git, base, e.Path); err != nil {
				return err
			}
			files[e.Path], order = f, append(order, e.Path)
		}
		if err := f.apply(e); err != nil {
			return fmt.Errorf("proposal: edit %d: %w", i+1, err)
		}
	}
	for _, path := range order {
		if err := writeFile(ctx, git, path, files[path]); err != nil {
			return err
		}
	}
	return nil
}

// checkEdit accepts one kind of edit on a clean repository-relative path
// outside .git.
func checkEdit(e Edit) error {
	invalid := !fs.ValidPath(e.Path) || e.Path == "." || strings.ContainsAny(e.Path, "\\\x00")
	for _, part := range strings.Split(e.Path, "/") {
		invalid = invalid || strings.EqualFold(part, ".git")
	}
	if invalid {
		return fmt.Errorf("invalid path %q: want a clean path relative to the repository root", e.Path)
	}
	kinds := 0
	for _, set := range []bool{e.Delete, e.Content != nil, e.OldString != "" || e.NewString != "" || e.ReplaceAll} {
		if set {
			kinds++
		}
	}
	if kinds != 1 {
		return fmt.Errorf("%s: set exactly one of oldString/newString, content or delete", e.Path)
	}
	if e.Content == nil && !e.Delete && e.OldString == "" {
		return fmt.Errorf("%s: oldString is required for a replacement; use content to write the whole file", e.Path)
	}
	return nil
}

// readFile is path's state in base: a regular file, or a file that does not exist.
func readFile(ctx context.Context, git call, base, path string) (*file, error) {
	entry, err := git.trimmed(ctx, "ls-tree", "-z", base, "--", path)
	if err != nil {
		return nil, fmt.Errorf("proposal: list %s in %s: %w", path, base, err)
	}
	if entry = strings.TrimSuffix(entry, "\x00"); entry == "" {
		return &file{mode: newFileMode}, nil
	}
	meta, _, _ := strings.Cut(entry, "\t")
	fields := strings.Fields(meta)
	if len(fields) != 3 || fields[1] != "blob" || (fields[0] != "100644" && fields[0] != "100755") {
		return nil, fmt.Errorf("proposal: %s is not a regular file in %s (%s)", path, base, meta)
	}
	content, err := git.run(ctx, "cat-file", "blob", fields[2])
	if err != nil {
		return nil, fmt.Errorf("proposal: read %s in %s: %w", path, base, err)
	}
	return &file{mode: fields[0], content: content, exists: true}, nil
}

func (f *file) apply(e Edit) error {
	switch {
	case e.Content != nil:
		f.content, f.exists = *e.Content, true
		return nil
	case !f.exists:
		return fmt.Errorf("%s: no such file", e.Path)
	case e.Delete:
		f.content, f.exists = "", false
		return nil
	}
	if e.OldString == e.NewString {
		return fmt.Errorf("%s: oldString and newString are the same", e.Path)
	}
	n := strings.Count(f.content, e.OldString)
	if n == 0 || (n > 1 && !e.ReplaceAll) {
		return fmt.Errorf("%s: oldString matches %d times; it must match exactly once, so add surrounding context or set replaceAll", e.Path, n)
	}
	f.content = strings.ReplaceAll(f.content, e.OldString, e.NewString)
	return nil
}

func writeFile(ctx context.Context, git call, path string, f *file) error {
	if !f.exists {
		if _, err := git.run(ctx, "update-index", "--force-remove", "--", path); err != nil {
			return fmt.Errorf("proposal: remove %s: %w", path, err)
		}
		return nil
	}
	hasher := git
	hasher.stdin = []byte(f.content)
	sha, err := hasher.trimmed(ctx, "hash-object", "-w", "--no-filters", "--stdin")
	if err != nil {
		return fmt.Errorf("proposal: store %s: %w", path, err)
	}
	if _, err := git.run(ctx, "update-index", "--add", "--cacheinfo", f.mode+","+sha+","+path); err != nil {
		return fmt.Errorf("proposal: stage %s: %w", path, err)
	}
	return nil
}
