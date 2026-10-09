package git

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	// revisionPattern is one argv token with no whitespace or control
	// character that cannot be read as an option.
	revisionPattern = regexp.MustCompile(`^[^-\s\x00-\x1f\x7f][^\s\x00-\x1f\x7f]*$`)
	// branchNamePattern is the character set `git check-ref-format --branch`
	// allows, with a leading "-" excluded.
	branchNamePattern = regexp.MustCompile(`^[^-\x00-\x20\x7f~^:?*\[\\][^\x00-\x20\x7f~^:?*\[\\]*$`)
)

// ValidateRevision rejects a revision that git could read as something other
// than a revision before it is ever placed on a git command line. A leading "-"
// would be parsed as an option (`--upload-pack=…` runs a program), and
// whitespace or a control character means the value was not one revision to
// begin with. Revision syntax itself — HEAD~2, origin/main, abc1234^{commit} —
// is left to git, which rejects anything it cannot resolve. Callers still put
// `--end-of-options` ahead of the value where the subcommand supports it.
func ValidateRevision(rev string) error {
	switch {
	case rev == "":
		return fmt.Errorf("invalid revision: is empty")
	case strings.HasPrefix(rev, "-"):
		return fmt.Errorf("invalid revision %q: must not begin with %q, git parses that as an option", rev, "-")
	case !revisionPattern.MatchString(rev):
		return fmt.Errorf("invalid revision %q: contains whitespace or control character", rev)
	}
	return nil
}

// ValidateBranchName applies the rules of `git check-ref-format --branch` to
// name, without spawning git, so a branch name that came from a request or a
// model is refused before it reaches `git branch`, `git push` or a refspec.
func ValidateBranchName(name string) error {
	switch {
	case name == "":
		return invalidBranchName(name, "is empty")
	case strings.HasPrefix(name, "-"):
		return invalidBranchName(name, `must not begin with "-", git parses that as an option`)
	case name == "HEAD", name == "@":
		return invalidBranchName(name, fmt.Sprintf("%q is reserved", name))
	case !branchNamePattern.MatchString(name):
		return invalidBranchName(name, `must not contain a space, a control character, or any of ~^:?*[\`)
	case strings.Contains(name, ".."):
		return invalidBranchName(name, `must not contain ".."`)
	case strings.Contains(name, "@{"):
		return invalidBranchName(name, `must not contain "@{"`)
	case strings.HasPrefix(name, "/"), strings.HasSuffix(name, "/"):
		return invalidBranchName(name, `must not begin or end with "/"`)
	case strings.HasSuffix(name, "."):
		return invalidBranchName(name, `must not end with "."`)
	}
	for _, component := range strings.Split(name, "/") {
		switch {
		case component == "":
			return invalidBranchName(name, "empty path component")
		case strings.HasPrefix(component, "."):
			return invalidBranchName(name, fmt.Sprintf("component %q must not begin with %q", component, "."))
		case strings.HasSuffix(component, ".lock"):
			return invalidBranchName(name, fmt.Sprintf("component %q must not end with %q", component, ".lock"))
		}
	}
	return nil
}

// invalidBranchName is a plain function rather than a closure over name:
// CodeQL infers ValidateBranchName as a sanitizing guard only while name stays
// an uncaptured parameter.
func invalidBranchName(name, reason string) error {
	return fmt.Errorf("invalid branch name %q: %s", name, reason)
}
