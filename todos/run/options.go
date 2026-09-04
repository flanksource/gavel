package run

import (
	"context"
	"fmt"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/commons-db/shell"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/lifecycle"
	"github.com/flanksource/gavel/todos/types"
	"github.com/google/uuid"
)

// StartResult is what a caller learns the moment a run is admitted: whether it
// started, the session it started under, and the message to show. The run
// itself continues in the background.
type StartResult struct {
	Status    string
	SessionID string
	Message   string
	// Done receives the run's final error (nil on success) once the background
	// run has applied its outcome, then closes. A caller that must block on the
	// run — the CLI — reads it; an HTTP handler ignores it. Nil means there is
	// nothing left to wait for.
	Done <-chan error
	// Stop cancels the in-flight run with a cause, for a caller that owns a
	// terminal and has to honour an interrupt. Nil when the run is not live.
	Stop context.CancelCauseFunc
}

// Request is one run: the TODO to execute, where, and what the caller asked
// for. Registry is the process's in-flight run map, so a TODO cannot be run
// twice at once and an in-flight run stays cancellable.
type Request struct {
	Provider todos.Provider
	Registry *Registry
	// Todo is the issue this run works. Grouped execution does not exist on the
	// PostgreSQL runtime — one issue, one lifecycle, one run.
	Todo *types.TODO
	Dir  string
	// Options is what the caller decided; everything else comes from the
	// lifecycle definition.
	Options Options
	// Broker answers the run's tool-permission requests. Only a caller that
	// serves an approval surface supplies one — the CLI leaves it nil, because a
	// run that asked it for a decision would block until its timeout.
	Broker todos.ApprovalBroker
	// Prepared is the resolution a caller already performed as its pre-flight.
	// Start dispatches exactly that fold instead of folding again, so the step
	// and session it reported cannot differ from the ones that run: a todo that
	// changed in between would otherwise run a different step from the one the
	// caller was told about, and a cmux run would be watched on a session id
	// nothing writes to. Nil means Start resolves for itself.
	Prepared *Prepared
}

// Options is what a caller decides about one step run. Everything else — which
// prompt renders, which spec layers fold, how long the run may take, and which
// status its result lands the todo in — is the lifecycle definition's, resolved
// by the host.
//
// It is deliberately NOT a resolved run any more. Resolution belongs to exactly
// one place, and a second pre-resolved copy travelling alongside it is a second
// answer to "what is this run", which is one more than can be right.
type Options struct {
	// Step names the lifecycle step to run. Empty runs the step the lifecycle
	// picks next for this todo.
	Step string
	// Request is the caller's explicit spec — parsed CLI flags or the dashboard
	// payload — folded as the TOP layer. A knob the caller did not set must
	// arrive zero, or it beats the configuration it claims to defer to.
	Request api.Spec
	// Prior are the layers a continuation inherits from the run it continues.
	// See lifecycle.LayerInput.Prior.
	Prior []api.SpecLayer
	// Resume continues the todo's recorded agent session instead of opening one.
	Resume bool
	// Message is the user turn a resumed session continues with — the answer to
	// an ask, a reviewer's feedback on a plan. It replaces the rendered prompt:
	// the session already holds the todo, and re-sending the instructions it has
	// acted on would ask for the work twice. It requires Resume.
	Message string
	// Concurrent dispatches this run alongside one that is already live and
	// owned by a running process, instead of refusing. It is never a default:
	// the caller sets it after confirming (--force, or the dashboard dialog).
	Concurrent bool
	// Host is the entrypoint; it decides the permission posture the host itself
	// contributes. See lifecycle.HostKind.
	Host lifecycle.HostKind
}

// IsStoppable reports whether an in-flight run can be cancelled. A headless
// runtime owns the agent process and honours context cancellation; a cmux run
// is driven through a detached surface that outlives the request.
func IsStoppable(spec api.Spec) bool {
	return spec.Mode != api.ModeCmux
}

// Commits returns the run's commit policies (nil when the spec asks for none).
// Each entry names the lifecycle phase it fires at, so a run can commit per
// turn, once the agent loop ends, or once at the end.
func Commits(spec api.Spec) []api.Commit {
	if spec.Workflow == nil {
		return nil
	}
	return spec.Workflow.Commits
}

// Commit reports whether the run auto-commits at all.
func Commit(spec api.Spec) bool {
	return len(Commits(spec)) > 0
}

// DryRun reports whether the run is a dry run: the agent runs normally but
// every commit is reported rather than cut. A spec that mixes dry and live
// stanzas is not a dry run — only the stanzas marked so are suppressed, and
// this reports the whole-run view a dashboard badge shows.
func DryRun(spec api.Spec) bool {
	commits := Commits(spec)
	for _, c := range commits {
		if !c.DryRun {
			return false
		}
	}
	return len(commits) > 0
}

// Dirty reports whether the run's checkout carries the working tree's
// uncommitted changes across. It reads the worktree's clone mode rather than
// the pointer: uncommitted work is only ever carried into a worktree, so a
// checkout without one has nothing to carry — the run already happens in the
// dirty tree.
func Dirty(spec api.Spec) bool {
	if spec.Setup == nil || spec.Setup.Checkout == nil || spec.Setup.Checkout.Worktree == nil {
		return false
	}
	return spec.Setup.Checkout.Worktree.Uncommitted == shell.CloneClone
}

// Refs are the TODO references a run covers, for reporting.
func Refs(todoList []*types.TODO) []string {
	refs := make([]string, len(todoList))
	for i, todo := range todoList {
		refs[i] = todos.TODOReference(todo)
	}
	return refs
}

func StartedMessage(count int) string {
	if count > 1 {
		return fmt.Sprintf("Started run for %d todos", count)
	}
	return "Todo run started"
}

// Label names a run for a log line.
func Label(todo *types.TODO) string {
	return todos.TODOReference(todo)
}

// SessionIDFor is the agent session id a resolved run will use, so a caller can
// follow the session log from the moment it starts. A resume reuses the todo's
// prior session; a fresh cmux run mints a new id (claude is launched with it);
// every other runtime reuses the todo's known session when it has one and
// otherwise lets the provider establish one.
func SessionIDFor(spec api.Spec, todo *types.TODO, resume bool) string {
	prior := PriorSessionID(todo)
	if resume && prior != "" {
		return prior
	}
	if spec.Mode == api.ModeCmux {
		return uuid.NewString()
	}
	return prior
}

// PriorSessionID is the todo's recorded agent session.
func PriorSessionID(todo *types.TODO) string {
	if todo != nil && todo.LLM != nil {
		return todo.LLM.SessionId
	}
	return ""
}
