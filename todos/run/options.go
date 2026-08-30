package run

import (
	"fmt"
	"time"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/commons-db/shell"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/drivers"
	todoprompt "github.com/flanksource/gavel/todos/prompt"
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
}

// Request is one run: the TODOs to execute, where, and under which resolved
// options. Registry is the process's in-flight run map, so a TODO cannot be run
// twice at once and an in-flight run stays cancellable.
type Request struct {
	Provider todos.Provider
	Registry *Registry
	// Todos are executed together in a single agent session (multi-select run);
	// a single-element slice is the ordinary one-todo run.
	Todos   []*types.TODO
	Dir     string
	Backend string
	Options Options
}

// Options is a fully resolved run — the output of the resolution fold, never
// raw request input. Everything here has already been validated and layered.
type Options struct {
	// Spec carries the model/backend/effort/budget/prompt/permissions/session
	// knobs plus the run's Setup (dirty/checkout) and Workflow (verify/commits).
	// dirty/checks/commit/dryRun are read from Spec.Setup/Spec.Workflow (see
	// Dirty/Commit/DryRun below), not sibling flags.
	//
	// Named, not embedded: api.Spec's promoted MarshalJSON/MarshalYAML would emit
	// only the spec and drop Driver/RunMode/Resume.
	Spec    api.Spec
	Driver  string
	RunMode types.RunMode
	// Prompt is the resolved prompt name and Envelope the structured result it
	// returns. RunMode says how the run behaves; these say what it runs.
	Prompt   string
	Envelope todoprompt.EnvelopeKind
	Resume   bool
	// Concurrent dispatches this run alongside one that is already live and
	// owned by a running process, instead of refusing. It is never a default:
	// the caller sets it after confirming (--force, or the dashboard dialog).
	Concurrent bool
	// Template is the .gavel.yaml prompt override source resolved alongside Spec;
	// empty renders the mode's embedded default. It travels with the options so
	// the preview and the executor cannot resolve different overrides.
	Template string
	// Approvals gates Bash behind human approval. Whether it is *enabled* is a
	// .gavel.yaml decision (todos.approvals), not an entrypoint constant; whether
	// it can be *answered* is up to the caller that sets it.
	Approvals bool
	// Timeout is Spec.Budget.Timeout already parsed by the resolution seam, so no
	// consumer re-parses a string it cannot fail on.
	Timeout time.Duration
}

// IsStoppable reports whether an in-flight run can be cancelled. Headless
// drivers own the agent process and honour context cancellation; a cmux run is
// driven through a detached surface that outlives the request.
func IsStoppable(opts Options) bool {
	kind, err := drivers.Parse(opts.Driver)
	return err == nil && kind != drivers.Cmux
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

// Label names a run for a log line: the single TODO's reference, or a count.
func Label(todoList []*types.TODO) string {
	if len(todoList) == 1 {
		return todos.TODOReference(todoList[0])
	}
	return fmt.Sprintf("%d todos", len(todoList))
}

// ResolveSessionID determines the claude session id a run will use, so the
// caller knows it up front. A resume run reuses the todo's prior session; a
// fresh cmux run mints a new id (claude is launched with it, so a dashboard can
// follow the log immediately); other backends reuse a single todo's known
// session when available and otherwise let the provider establish one.
func ResolveSessionID(opts Options, todoList []*types.TODO) string {
	if opts.Resume {
		if sid := FirstSessionID(todoList); sid != "" {
			return sid
		}
	}
	kind, err := drivers.Parse(opts.Driver)
	if err != nil {
		panic(fmt.Sprintf("resolved TODO run has invalid driver %q: %v", opts.Driver, err))
	}
	switch kind {
	case drivers.Cmux:
		return uuid.NewString()
	case drivers.Api, drivers.Agent, drivers.Cli:
		if len(todoList) == 1 {
			return FirstSessionID(todoList)
		}
	default:
		panic(fmt.Sprintf("resolved TODO run has unsupported driver %q", opts.Driver))
	}
	return ""
}

func FirstSessionID(todoList []*types.TODO) string {
	for _, todo := range todoList {
		if todo != nil && todo.LLM != nil && todo.LLM.SessionId != "" {
			return todo.LLM.SessionId
		}
	}
	return ""
}
