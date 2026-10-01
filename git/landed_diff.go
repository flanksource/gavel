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

// LandedDiffStat is the footprint of the n commits ending at tip, tip~n..tip:
// the work a landing carried onto a branch, measured without the run's own
// (possibly deleted and garbage-collected) branch. A tip whose object is not in
// the repository is ErrCommitNotFound; any other git failure, including a tip
// with fewer than n ancestors, is surfaced as is.
func LandedDiffStat(path, tip string, n int) (DiffStat, error) {
	if !IsValidCommitHash(tip) {
		return DiffStat{}, fmt.Errorf("invalid commit hash %q", tip)
	}
	if n < 1 {
		return DiffStat{}, fmt.Errorf("landed diff of %s needs at least one commit, got %d", tip, n)
	}
	// `git cat-file -e` exits 1 for a missing object and 128 for anything else.
	if _, err := gitOutput(path, "cat-file", "-e", tip); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
			return DiffStat{}, fmt.Errorf("%w: %s in %s", ErrCommitNotFound, tip, path)
		}
		return DiffStat{}, err
	}
	from, err := gitOutput(path, "rev-parse", "--verify", "--end-of-options", tip+"~"+strconv.Itoa(n)+"^{commit}")
	if err != nil {
		return DiffStat{}, err
	}
	return RangeDiffStat(path, strings.TrimSpace(string(from)), tip)
}
