package git

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

const (
	// TrailerIssueID is the git trailer key that ties a commit to the gavel todo
	// issue it implements; consumers read it to link commits back to their issue.
	TrailerIssueID = "Gavel-Issue-Id"
	// TrailerSessionID is the git trailer key recording the agent session that
	// produced a commit.
	TrailerSessionID = "Claude-Session-Id"
)

var (
	commitHashPattern = regexp.MustCompile(`^[0-9a-fA-F]{4,64}$`)
	diffPathPattern   = regexp.MustCompile(`^[^\x00-\x1f\x7f]+$`)
)

// IsValidCommitHash reports whether s, exactly as given, is a syntactically
// valid abbreviated or full git object hash, so callers can reject untrusted
// input before shelling out to git. Surrounding whitespace is not trimmed: the
// value accepted is the value git receives.
func IsValidCommitHash(s string) bool {
	return commitHashPattern.MatchString(s)
}

// validateDiffPath rejects an untrusted path that git would reinterpret as
// something other than a literal file or directory pathspec inside the
// repository. `--` already stops
// option parsing, but git still reads pathspec magic (`:(exclude)`, `:/`) after
// it, and a control character would corrupt the argument for any consumer that
// re-splits the command line, so the shape is checked before the value is ever
// handed to git.
func validateDiffPath(file string) error {
	switch {
	case strings.HasPrefix(file, "-"):
		return fmt.Errorf("invalid diff path %q: must not begin with %q, git parses that as an option", file, "-")
	case strings.HasPrefix(file, ":"):
		return fmt.Errorf("invalid diff path %q: must not begin with %q, git parses that as pathspec magic", file, ":")
	case strings.HasPrefix(file, "/"):
		return fmt.Errorf("invalid diff path %q: must be relative to the repository root, not absolute", file)
	case !diffPathPattern.MatchString(file):
		return fmt.Errorf("invalid diff path %q: contains a control character", file)
	}
	for _, segment := range strings.Split(file, "/") {
		if segment == ".." {
			return fmt.Errorf("invalid diff path %q: %q segments escape the repository", file, "..")
		}
	}
	return nil
}

// RemoteWebURL returns the https web base for the repository's origin remote
// (e.g. https://github.com/owner/repo), or "" when there is no origin or it is
// not a recognizable host/owner/repo URL. Callers append "/commit/<hash>" to
// build a commit link; a local-only repo simply has no link.
func RemoteWebURL(path string) (string, error) {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	cmd.Dir = path
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git remote get-url origin in %s: %w", path, err)
	}
	return remoteToWebURL(strings.TrimSpace(string(out))), nil
}

// remoteToWebURL converts a git remote URL into its https web base, handling the
// scp-like ssh form (git@host:owner/repo.git) and scheme URLs
// (ssh://, https://, http://). Returns "" for anything it cannot parse into a
// host + owner/repo path.
func remoteToWebURL(remote string) string {
	remote = strings.TrimSpace(remote)
	if remote == "" {
		return ""
	}
	remote = strings.TrimSuffix(remote, ".git")

	var host, repoPath string
	switch {
	case strings.HasPrefix(remote, "git@"):
		host, repoPath, _ = strings.Cut(strings.TrimPrefix(remote, "git@"), ":")
	case strings.Contains(remote, "://"):
		_, rest, _ := strings.Cut(remote, "://")
		// Strip userinfo (user@) that precedes the host.
		if at := strings.Index(rest, "@"); at >= 0 {
			if slash := strings.Index(rest, "/"); slash < 0 || at < slash {
				rest = rest[at+1:]
			}
		}
		host, repoPath, _ = strings.Cut(rest, "/")
	default:
		return ""
	}

	host = strings.TrimSpace(host)
	repoPath = strings.Trim(strings.TrimSpace(repoPath), "/")
	if host == "" || repoPath == "" {
		return ""
	}
	return "https://" + host + "/" + repoPath
}
