// Package content resolves the free-text fields of a TODO — body, plan,
// verification — from whatever a caller passed: literal markdown, a path, or an
// @path reference.
//
// It lives outside the CLI because every surface that writes those fields needs
// the same resolution. When it lived in package main, the entity's create and
// edit actions could not reach it, which is why they did not exist: the chat
// window could comment on a TODO but not write one.
package content

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/todos"
)

// Options names one field's raw value. Flag is used only in error messages, so
// a failure says which input the caller has to fix.
type Options struct {
	WorkDir string
	Flag    string
	Value   string
}

// Resolve expands a field value to its markdown.
//
// The precedence is deliberate: an explicit @path or a resolved file wins, an
// escaped \@ or anything containing a newline is literal text, and a bare value
// that happens to name a readable regular file is read as one. That last case
// is what makes `--body ./notes.md` work without the sigil, and it is checked
// last so a literal string is never shadowed by a same-named file.
func Resolve(opts Options) (string, error) {
	ref, err := fixtures.ResolveFileRef(opts.WorkDir, opts.Value)
	if err != nil {
		return "", fmt.Errorf("resolve %s: %w", opts.Flag, err)
	}
	if ref.IsFile {
		return strings.TrimSpace(ref.Contents), nil
	}
	if strings.HasPrefix(opts.Value, `\@`) || strings.ContainsAny(opts.Value, "\r\n\x00") {
		return strings.TrimSpace(ref.Raw), nil
	}
	path := opts.Value
	if path == "" {
		return "", nil
	}
	if !filepath.IsAbs(path) {
		path = filepath.Join(opts.WorkDir, path)
	}
	info, err := os.Stat(path)
	if err == nil && info.Mode().IsRegular() {
		contents, err := os.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("resolve %s: read existing file %q: %w", opts.Flag, opts.Value, err)
		}
		return strings.TrimSpace(string(contents)), nil
	}
	return strings.TrimSpace(ref.Raw), nil
}

// CreateOptions are the three content fields of a new TODO. Each carries its
// own "was it set" flag rather than relying on emptiness, because an explicitly
// empty --plan is a rejection and an unset one is not.
type CreateOptions struct {
	Body            string
	BodySet         bool
	Plan            string
	PlanSet         bool
	Verification    string
	VerificationSet bool
}

// Create is the resolved content of a new TODO.
type Create struct {
	Body         string
	Plan         string
	Verification string
}

// ResolveCreate expands every set field and folds an inline verification
// fixture out of the body, so `gavel todos create --body ./issue.md` picks up a
// definition of done written in the same document.
func ResolveCreate(workDir string, opts CreateOptions) (Create, error) {
	var content Create
	var err error
	if opts.BodySet {
		content.Body, err = Resolve(Options{WorkDir: workDir, Flag: "--body", Value: opts.Body})
		if err != nil {
			return Create{}, err
		}
	}
	if opts.PlanSet {
		content.Plan, err = Resolve(Options{WorkDir: workDir, Flag: "--plan", Value: opts.Plan})
		if err != nil {
			return Create{}, err
		}
		if content.Plan == "" {
			return Create{}, fmt.Errorf("--plan cannot be empty")
		}
	}
	if opts.VerificationSet {
		content.Verification, err = Resolve(Options{WorkDir: workDir, Flag: "--verification", Value: opts.Verification})
		if err != nil {
			return Create{}, err
		}
		if content.Verification == "" {
			return Create{}, fmt.Errorf("--verification cannot be empty")
		}
	}
	var bodyVerification string
	content.Body, bodyVerification, _ = todos.SplitVerificationFixture(content.Body)
	content.Verification = todos.CombineVerificationFixtures(content.Verification, bodyVerification)
	return content, nil
}
