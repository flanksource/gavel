package git

import (
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// ErrCommitNotFound reports a well-formed commit hash whose object is not in
// the local repository — e.g. a pull request's topic head that only lives on
// the remote.
var ErrCommitNotFound = errors.New("commit not found in the local repository")

// LandedBase is tip~n, the commit below the n commits a landing carried onto a
// branch, as a full hash: LandedBase..tip is the landed work, measured without
// the run's own (possibly deleted and garbage-collected) branch. Both commits
// are immutable, so the result is too. A tip whose object is not in the
// repository is ErrCommitNotFound; any other git failure, including a tip with
// fewer than n ancestors, is surfaced as is.
func LandedBase(path, tip string, n int) (string, error) {
	if !IsValidCommitHash(tip) {
		return "", fmt.Errorf("invalid commit hash %q", tip)
	}
	if n < 1 {
		return "", fmt.Errorf("landed diff of %s needs at least one commit, got %d", tip, n)
	}
	// `git cat-file -e` exits 1 for a missing object and 128 for anything else.
	if _, err := gitOutput(path, "cat-file", "-e", tip); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
			return "", fmt.Errorf("%w: %s in %s", ErrCommitNotFound, tip, path)
		}
		return "", err
	}
	from, err := gitOutput(path, "rev-parse", "--verify", "--end-of-options", tip+"~"+strconv.Itoa(n)+"^{commit}")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(from)), nil
}
