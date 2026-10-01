package runtime

import (
	"context"
	"fmt"
	"strings"

	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/captain/pkg/promptrun"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/lifecycle"
	"github.com/flanksource/gavel/todos/native"
	"github.com/flanksource/gavel/todos/types"
)

// runOrigin is the origin every gavel todo run is admitted under.
const runOrigin = "gavel.todos"

// promptNameOrDefault resolves the prompt name a run is dispatched under,
// defaulting to the one that shares the behaviour class's name.
func promptNameOrDefault(name string, mode types.RunMode) string {
	if trimmed := strings.TrimSpace(name); trimmed != "" {
		return trimmed
	}
	return string(mode)
}

// launchPlan is one dispatch cleared against the issue as it stands: which
// step it runs, the deterministic identity it is admitted under, and who
// drives it.
type launchPlan struct {
	issue       *native.Issue
	step        native.StepKind
	mode        types.RunMode
	promptName  string
	identity    lifecycle.Identity
	ordinal     int
	owner       native.RunOwner
	cwd         string
	preparation todos.RunPreparation
}

// PrepareRun clears a dispatch before anything is started and returns the
// Recording Captain admits it with. It writes no captain_* row: Captain creates
// the session tree and the run inside promptrun.Run, and Record.Link attaches
// that exact run to the issue in the same transaction. The issue version is
// part of the deterministic admission identity, so an exact retry resolves the
// same launch while a later attempt receives a new identity.
func (p *Provider) PrepareRun(ctx context.Context, todo *types.TODO, preparation todos.RunPreparation) (todos.RunAdmission, error) {
	launch, err := p.planLaunch(ctx, todo, preparation)
	if err != nil {
		return todos.RunAdmission{}, err
	}
	placement := promptrun.Placement{Sessions: p.launchSessions(launch)}
	if preparation.Resume && launch.issue.ActivePromptRunID != nil {
		previousRun, err := p.captain.GetPromptRun(ctx, *launch.issue.ActivePromptRunID)
		if err != nil {
			return todos.RunAdmission{}, err
		}
		if !terminalPromptRun(previousRun.State) {
			return p.resumeInPlace(ctx, todo, launch, previousRun)
		}
		// A terminal operation gets a new prompt run, but continuation keeps the
		// authoritative Captain/provider session identity.
		placement = promptrun.Placement{SessionID: previousRun.SessionID}
	}
	return p.admitLaunch(ctx, launch, placement)
}

// planLaunch resolves the step and identity a dispatch runs under, settling an
// orphaned incumbent and refusing a live one unless the caller confirmed it.
func (p *Provider) planLaunch(ctx context.Context, todo *types.TODO, preparation todos.RunPreparation) (*launchPlan, error) {
	issueID, err := p.todoID(todo)
	if err != nil {
		return nil, err
	}
	launch := &launchPlan{mode: preparation.Mode, preparation: preparation}
	if launch.mode == "" {
		launch.mode = types.ModeRun
	}
	launch.promptName = promptNameOrDefault(preparation.Prompt, launch.mode)
	// The prompt name is the step the run is dispatched as and seeds its identity.
	if launch.step, err = stepFor(launch.promptName); err != nil {
		return nil, err
	}
	issue, err := p.repository.GetIssue(ctx, issueID)
	if err != nil {
		return nil, err
	}
	if issue.WorkspaceID != p.workspace.ID {
		return nil, fmt.Errorf("%w: TODO %s belongs to workspace %s, provider owns %s", native.ErrCrossWorkspace, issue.ID, issue.WorkspaceID, p.workspace.ID)
	}
	if launch.owner, err = native.LocalOwner(); err != nil {
		return nil, err
	}
	seed := lifecycle.Seed(issue.ID, launch.promptName, todo.Version)
	// An unfinished run whose dispatcher is gone would block this TODO forever,
	// so settle it before choosing any identity: reclaiming it changes which run
	// is active, and a live incumbent changes what identity this run needs. The
	// seed's own run id goes along so a contender replaying this exact dispatch
	// is recognised as such rather than as a second run.
	ownIdentity, err := p.resolveActiveRunConflict(ctx, issue, preparation, lifecycle.IdentityFor(seed).PromptRunID)
	if err != nil {
		return nil, err
	}
	if launch.issue, err = p.repository.GetIssue(ctx, issueID); err != nil {
		return nil, err
	}
	if launch.ordinal, err = p.nextPromptOrdinal(ctx, issue.ID, launch.step); err != nil {
		return nil, err
	}
	if ownIdentity {
		// The seed is deterministic per (issue, step, issue version), which makes an
		// ordinary redispatch idempotent — and would resolve this dispatch onto the
		// run already using that identity. The ordinal discriminates it, and keeps
		// a retry of this same dispatch idempotent in turn.
		seed = fmt.Sprintf("%s:attempt:%d", seed, launch.ordinal)
	}
	launch.identity = lifecycle.IdentityFor(seed)
	launch.cwd = firstNonBlank(todo.CWD, p.workDir)
	return launch, p.checkLaunchVersion(ctx, todo, launch)
}

// checkLaunchVersion refuses a dispatch made against a stale issue. A contender
// may read after the winning admission has already advanced the issue version;
// that exact deterministic launch is recognised as an owned dispatch instead of
// being misclassified as an unrelated stale mutation.
func (p *Provider) checkLaunchVersion(ctx context.Context, todo *types.TODO, launch *launchPlan) error {
	issue := launch.issue
	if todo.Version == issue.Version {
		return nil
	}
	if issue.ActivePromptRunID != nil && *issue.ActivePromptRunID == launch.identity.PromptRunID {
		links, err := p.repository.ListPromptRuns(ctx, issue.ID)
		if err != nil {
			return err
		}
		for _, link := range links {
			if link.PromptRunID == launch.identity.PromptRunID && link.StepKind == launch.step {
				return fmt.Errorf("%w: Captain prompt run %s for issue %s already has an external dispatcher",
					todos.ErrRunDispatchAlreadyClaimed, launch.identity.PromptRunID, issue.ID)
			}
		}
	}
	return fmt.Errorf("%w: issue %s expected version %d, current version %d", native.ErrVersionConflict, issue.ID, todo.Version, issue.Version)
}

// launchSessions is the session tree a fresh run is admitted on: the TODO's
// root and this dispatch's operation session under it.
func (p *Provider) launchSessions(launch *launchPlan) []captaindb.CreateSessionInput {
	issue := launch.issue
	root := p.todoRootSessionInput(native.CreateIssueInput{ID: issue.ID, Title: issue.Title, Body: issue.Body})
	operation := p.todoOperationSessionInput(issue, todoOperationSessionOptions{
		ID: launch.identity.SessionID, Operation: string(launch.step), Provider: launch.executorName(),
		CWD: launch.cwd, Prompt: launch.promptMarkdown(),
	})
	return []captaindb.CreateSessionInput{root, operation}
}

func (launch *launchPlan) executorName() string {
	return firstNonBlank(launch.preparation.ExecutorName, "unknown")
}

// promptMarkdown is the prompt the run is filed with: the rendered prompt, or
// the issue body for a step that renders none.
func (launch *launchPlan) promptMarkdown() string {
	prompt := launch.preparation.PromptMarkdown
	if strings.TrimSpace(prompt) == "" && launch.mode != types.ModeVerify {
		return launch.issue.Body
	}
	return prompt
}
