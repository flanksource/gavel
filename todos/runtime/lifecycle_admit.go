package runtime

import (
	"context"
	"fmt"
	"strings"

	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/captain/pkg/promptrun"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/native"
	"github.com/flanksource/gavel/todos/types"
	"github.com/google/uuid"
)

// admitLaunch is the Recording a new prompt run is admitted with. Captain
// creates the placement's sessions and the run; Link attaches that exact run to
// the issue and activates it — gavel's rows only — in the same transaction, so
// a refused attachment leaves no Captain record behind.
func (p *Provider) admitLaunch(ctx context.Context, launch *launchPlan, placement promptrun.Placement) (todos.RunAdmission, error) {
	issue := launch.issue
	record := &promptrun.Recording{
		DB: p.captain, Placement: placement,
		PromptRunID: launch.identity.PromptRunID, AdmissionKey: launch.identity.AdmissionKey,
		Origin: runOrigin,
		// SpecProfile records WHICH prompt ran, where Runtime.Mode records how it
		// behaved. For the built-ins the two are the same string; for any other
		// prompt this is the only place its name survives.
		SpecProfile:          launch.promptName,
		PromptMarkdown:       launch.promptMarkdown(),
		VerificationMarkdown: issue.Verification,
		Runtime: captaindb.PromptRunRuntime{
			Mode: string(launch.mode), Driver: launch.executorName(), Requested: launch.preparation.Requested,
		},
		Link: p.attachLaunch(launch),
	}
	if err := p.attachInputPlan(ctx, issue, launch.mode, record); err != nil {
		return todos.RunAdmission{}, err
	}
	sessionID := launch.identity.SessionID
	if placement.SessionID != uuid.Nil {
		sessionID = placement.SessionID
	}
	return todos.RunAdmission{
		RunPreparationResult: todos.RunPreparationResult{SessionID: sessionID.String(), PromptRunID: launch.identity.PromptRunID},
		Record:               record,
	}, nil
}

// attachLaunch links the admitted run to its issue, claims its dispatch for
// this process and makes it the issue's active run. A replay that finds the
// run already attached owns no dispatch: another caller is driving it.
func (p *Provider) attachLaunch(launch *launchPlan) func(context.Context, *captaindb.DB, *captaindb.PromptRun) error {
	return func(ctx context.Context, tx *captaindb.DB, run *captaindb.PromptRun) error {
		owner := launch.owner
		attached, err := p.coordinator.AttachPromptRun(ctx, tx, run, native.PromptRunAttachment{
			IssueID: launch.issue.ID, PromptRunID: run.ID, StepKind: launch.step, Ordinal: launch.ordinal,
			ExpectedIssueVersion: launch.issue.Version, Actor: mutationActor, Owner: &owner,
		})
		if err != nil {
			return err
		}
		if !attached.DispatchOwned {
			return fmt.Errorf("%w: Captain prompt run %s for issue %s already has an external dispatcher",
				todos.ErrRunDispatchAlreadyClaimed, run.ID, launch.issue.ID)
		}
		return nil
	}
}

// resumeInPlace continues a waiting or running prompt run: one interactive
// operation, not a second active root operation. The run is replayed by its
// own identity, so Captain resolves it rather than admitting another; this
// process claims it first, because the previous dispatcher's heartbeat stops
// where it stopped.
func (p *Provider) resumeInPlace(ctx context.Context, todo *types.TODO, launch *launchPlan, previousRun *captaindb.PromptRun) (todos.RunAdmission, error) {
	issue := launch.issue
	links, err := p.repository.ListPromptRuns(ctx, issue.ID)
	if err != nil {
		return todos.RunAdmission{}, err
	}
	activeStep := native.StepKind("")
	for _, link := range links {
		if link.PromptRunID == previousRun.ID {
			activeStep = link.StepKind
			break
		}
	}
	if activeStep == "" {
		return todos.RunAdmission{}, fmt.Errorf("%w: active prompt run %s is not linked to issue %s", native.ErrLinkConflict, previousRun.ID, issue.ID)
	}
	if activeStep != launch.step {
		return todos.RunAdmission{}, fmt.Errorf("%w: active Captain prompt run %s is step %q and cannot resume as %q",
			todos.ErrRunResumeModeMismatch, previousRun.ID, activeStep, launch.step)
	}
	if issue, err = p.recordResumedAnswer(ctx, issue, previousRun, launch.preparation.PromptMarkdown); err != nil {
		return todos.RunAdmission{}, err
	}
	p.markPrepared(issue.ID, previousRun.ID)
	if err := p.claimRun(ctx, previousRun.ID, launch.owner); err != nil {
		return todos.RunAdmission{}, err
	}
	if err := p.replaceTODO(ctx, todo, issue, launch.cwd); err != nil {
		return todos.RunAdmission{}, err
	}
	return todos.RunAdmission{
		RunPreparationResult: todos.RunPreparationResult{SessionID: previousRun.SessionID.String(), PromptRunID: previousRun.ID},
		Record: &promptrun.Recording{
			DB: p.captain, Placement: promptrun.Placement{SessionID: previousRun.SessionID},
			PromptRunID: previousRun.ID, AdmissionKey: previousRun.AdmissionKey,
			Origin: previousRun.Origin, SpecProfile: previousRun.SpecProfile, Runtime: previousRun.Runtime,
			PromptMarkdown: previousRun.PromptMarkdown, VerificationMarkdown: previousRun.VerificationMarkdown,
			Link: func(_ context.Context, _ *captaindb.DB, run *captaindb.PromptRun) error {
				if run.ID != previousRun.ID {
					return fmt.Errorf("%w: resuming prompt run %s admitted run %s instead", native.ErrLinkConflict, previousRun.ID, run.ID)
				}
				return nil
			},
		},
	}, nil
}

// RunAdmitted takes up a run Captain has admitted: this process drives it from
// here, so its ownership claim is kept alive, and the todo is re-read as the
// admission left it — the activation advanced the issue version.
func (p *Provider) RunAdmitted(ctx context.Context, todo *types.TODO, admission todos.RunPreparationResult) error {
	issueID, err := p.todoID(todo)
	if err != nil {
		return err
	}
	if admission.PromptRunID == uuid.Nil {
		return fmt.Errorf("%w: admitted run of issue %s has no prompt run id", native.ErrInvalidInput, issueID)
	}
	owner, err := native.LocalOwner()
	if err != nil {
		return err
	}
	// This dispatch is a new phase run; drop the memo so a read through this
	// provider sees it rather than the index loaded before the launch.
	p.dropPhaseRuns(p.workspace.ID)
	p.dropExecutionIndex(p.workspace.ID)
	p.markPrepared(issueID, admission.PromptRunID)
	p.ownership.start(p.repository, admission.PromptRunID, owner)
	return p.reloadTODO(ctx, todo, firstNonBlank(todo.CWD, p.workDir))
}

// stepFor is the step kind a run is RECORDED under: what the link row stores,
// what a resume must match, and what a backlog groups by.
//
// It is the lifecycle step's own name, unmapped. There is no longer a second
// vocabulary to translate into: a project that declares a `shape-it` step gets
// runs recorded as `shape-it`, and the storage CHECK accepts any lower-case
// name. Whether the name is a step of the loaded lifecycle is the host's
// question, asked before the run is ever prepared.
func stepFor(stepName string) (native.StepKind, error) {
	step := native.StepKind(strings.TrimSpace(stepName))
	if step == "" {
		return "", fmt.Errorf("a TODO run must name the lifecycle step it is dispatched as")
	}
	return step, nil
}

func (p *Provider) todoID(todo *types.TODO) (uuid.UUID, error) {
	if todo == nil {
		return uuid.Nil, fmt.Errorf("native TODO is nil")
	}
	id, err := uuid.Parse(strings.TrimSpace(todo.ID))
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid native TODO ID %q: %w", todo.ID, err)
	}
	if todo.WorkspaceID != "" {
		workspaceID, err := uuid.Parse(todo.WorkspaceID)
		if err != nil {
			return uuid.Nil, fmt.Errorf("invalid native TODO workspace ID %q: %w", todo.WorkspaceID, err)
		}
		if workspaceID != p.workspace.ID {
			return uuid.Nil, fmt.Errorf("%w: TODO %s belongs to workspace %s, provider owns %s", native.ErrCrossWorkspace, id, workspaceID, p.workspace.ID)
		}
	}
	return id, nil
}
