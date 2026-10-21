package commit

import (
	"errors"
	"fmt"
	"strings"
)

// ErrRebaseInProgress is returned while an agent run's tree is stopped
// mid-rebase. It is not a failure of the turn: committing now would cut the
// commit the rebase stopped on under a generated message, so the rebase stays in
// place for the agent to continue, and the run's verify commands say what is
// left.
var ErrRebaseInProgress = errors.New("rebase in progress")

func rebaseInProgressErr(unresolved []string) error {
	if len(unresolved) == 0 {
		return fmt.Errorf("%w; continue it with `git rebase --continue`", ErrRebaseInProgress)
	}
	return fmt.Errorf("%w (unresolved: %s)", ErrRebaseInProgress, strings.Join(unresolved, ", "))
}
