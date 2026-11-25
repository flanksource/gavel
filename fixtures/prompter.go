package fixtures

import (
	"context"

	"github.com/flanksource/clicky/prompt"
)

// Prompter asks a person to fill a JSON-schema form and blocks until they
// answer, cancel, or ctx ends. The gavel UI backs it with a prompt.Manager,
// a TTY with a terminal form, an isolated worker with a relay to its parent;
// hosts (oipa-cli serve) inject their own through RunOptions.Prompter.
type Prompter interface {
	Ask(ctx context.Context, p prompt.Prompt) (prompt.Answer, error)
}
