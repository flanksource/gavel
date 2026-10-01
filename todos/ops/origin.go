package ops

import (
	"os"
	"strings"

	"github.com/flanksource/gavel/commit"
	"github.com/flanksource/gavel/todos"
)

// OriginFromEnv reports the run this process was started inside: the TODO and
// agent session a lifecycle step exports to its agent (the same variables
// `gavel commit` reads for its trailers). It is nil outside a run.
//
// It lives with the callers rather than in a provider, which takes the origin as
// an argument so that storage never depends on the process environment.
func OriginFromEnv() *todos.CreateOrigin {
	origin := todos.CreateOrigin{
		IssueID:   strings.TrimSpace(os.Getenv(commit.EnvIssueID)),
		SessionID: strings.TrimSpace(os.Getenv(commit.EnvSessionID)),
	}
	if origin == (todos.CreateOrigin{}) {
		return nil
	}
	return &origin
}
