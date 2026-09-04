package lifecycle

import (
	"fmt"
	"strings"

	"github.com/flanksource/gavel/todos/types"
	"github.com/google/uuid"
)

// Identity is the durable Captain identity one dispatch runs under. It is
// derived, not allocated: the same todo version dispatched for the same step
// resolves to the same session and prompt run, which is what makes a concurrent
// retry recognisable as an owned dispatch rather than a stale mutation.
type Identity struct {
	SessionID    uuid.UUID
	PromptRunID  uuid.UUID
	AdmissionKey string
}

// Seed is the string every identity hashes from: `<issue>:<step>:<version>`.
// The step is the lifecycle step's name — run, plan, verify, triage, or a
// project's own — so two steps dispatched against one issue version never
// collide, and the built-in steps hash exactly as they did before the lifecycle
// named them.
func Seed(issueID uuid.UUID, step string, version int64) string {
	return fmt.Sprintf("%s:%s:%d", issueID, step, version)
}

// IdentityFor derives the identity for a seed. Any suffix a runtime appends to
// disambiguate a retry (`:attempt:N`) goes into the seed first, so the two
// UUIDs and the admission key always agree on which dispatch they name.
func IdentityFor(seed string) Identity {
	return Identity{
		SessionID:    uuid.NewSHA1(uuid.NameSpaceOID, []byte("gavel-todo-session:"+seed)),
		PromptRunID:  uuid.NewSHA1(uuid.NameSpaceOID, []byte("gavel-todo-prompt-run:"+seed)),
		AdmissionKey: "gavel-todo:" + seed,
	}
}

// Identity is the identity a step run of this todo would be admitted under.
func (h *Host) Identity(todo *types.TODO, step Step) (Identity, error) {
	if todo == nil {
		return Identity{}, fmt.Errorf("lifecycle identity: todo is nil")
	}
	issueID, err := uuid.Parse(strings.TrimSpace(todo.ID))
	if err != nil {
		return Identity{}, fmt.Errorf("lifecycle identity: todo %q has no native id: %w", todo.ID, err)
	}
	return IdentityFor(Seed(issueID, step.Name, todo.Version)), nil
}
