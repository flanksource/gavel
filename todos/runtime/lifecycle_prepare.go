package runtime

import (
	"context"
	"fmt"
	"strings"

	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/lifecycle"
	"github.com/flanksource/gavel/todos/native"
	"github.com/flanksource/gavel/todos/types"
	"github.com/google/uuid"
)

// promptNameOrDefault resolves the prompt name a run is dispatched under,
// defaulting to the one that shares the behaviour class's name.
func promptNameOrDefault(name string, mode types.RunMode) string {
	if trimmed := strings.TrimSpace(name); trimmed != "" {
		return trimmed
	}
	return string(mode)
}

// PrepareRun creates Captain's authoritative session and prompt-run records,
// links that exact run to one issue, and activates it before an external agent
// can be dispatched. The issue version is part of the deterministic admission
// identity, so an exact retry resolves the same launch while a later attempt
// receives a new identity.
func (p *Provider) PrepareRun(ctx context.Context, todo *types.TODO, preparation todos.RunPreparation) (todos.RunPreparationResult, error) {
	issueID, err := p.todoID(todo)
	if err != nil {
		return todos.RunPreparationResult{}, err
	}
	mode := preparation.Mode
	if mode == "" {
		mode = types.ModeRun
	}
	promptName := promptNameOrDefault(preparation.Prompt, mode)
	// The prompt name is the lifecycle step the run is dispatched as: it seeds
	// the dispatch identity (lifecycle.Seed), so two steps of one behaviour class
	// against one issue version never collide. step is the kind the link row, the
	// ordinal sequence and a resume match against.
	step, err := stepFor(promptName)
	if err != nil {
		return todos.RunPreparationResult{}, err
	}
	issue, err := p.repository.GetIssue(ctx, issueID)
	if err != nil {
		return todos.RunPreparationResult{}, err
	}
	if issue.WorkspaceID != p.workspace.ID {
		return todos.RunPreparationResult{}, fmt.Errorf("%w: TODO %s belongs to workspace %s, provider owns %s", native.ErrCrossWorkspace, issue.ID, issue.WorkspaceID, p.workspace.ID)
	}
	owner, err := native.LocalOwner()
	if err != nil {
		return todos.RunPreparationResult{}, err
	}
	seed := lifecycle.Seed(issue.ID, promptName, todo.Version)
	// An unfinished run whose dispatcher is gone would block this TODO forever,
	// so settle it before allocating any identity: reclaiming it changes which
	// run is active, and a live incumbent changes what identity this run needs.
	// The seed's own run id goes along so a contender replaying this exact
	// dispatch is recognised as such rather than as a second run.
	ownIdentity, err := p.resolveActiveRunConflict(ctx, issue, preparation, lifecycle.IdentityFor(seed).PromptRunID)
	if err != nil {
		return todos.RunPreparationResult{}, err
	}
	if issue, err = p.repository.GetIssue(ctx, issueID); err != nil {
		return todos.RunPreparationResult{}, err
	}
	ordinal, err := p.nextPromptOrdinal(ctx, issue.ID, step)
	if err != nil {
		return todos.RunPreparationResult{}, err
	}
	if ownIdentity {
		// The seed above is deterministic per (issue, step, issue version), which
		// is what makes an ordinary redispatch idempotent — and what would
		// otherwise resolve this dispatch onto a run that is already using that
		// identity. The step ordinal discriminates it, and keeps a retry of this
		// same dispatch idempotent in turn.
		seed = fmt.Sprintf("%s:attempt:%d", seed, ordinal)
	}
	identity := lifecycle.IdentityFor(seed)
	sessionID, promptRunID := identity.SessionID, identity.PromptRunID
	if todo.Version != issue.Version {
		// A contender may read after the winning admission has already advanced
		// the issue version. Recognize that exact deterministic launch as an owned
		// dispatch instead of misclassifying it as an unrelated stale mutation.
		if issue.ActivePromptRunID != nil && *issue.ActivePromptRunID == promptRunID {
			links, linkErr := p.repository.ListPromptRuns(ctx, issue.ID)
			if linkErr != nil {
				return todos.RunPreparationResult{}, linkErr
			}
			for _, link := range links {
				if link.PromptRunID == promptRunID && link.StepKind == step {
					return todos.RunPreparationResult{}, fmt.Errorf(
						"%w: Captain prompt run %s for issue %s already has an external dispatcher",
						todos.ErrRunDispatchAlreadyClaimed, promptRunID, issue.ID,
					)
				}
			}
		}
		return todos.RunPreparationResult{}, fmt.Errorf("%w: issue %s expected version %d, current version %d", native.ErrVersionConflict, issue.ID, todo.Version, issue.Version)
	}

	executorName := strings.TrimSpace(preparation.ExecutorName)
	if executorName == "" {
		executorName = "unknown"
	}
	promptMarkdown := preparation.Spec.Prompt.User
	if strings.TrimSpace(promptMarkdown) == "" && mode != types.ModeVerify {
		promptMarkdown = issue.Body
	}
	cwd := strings.TrimSpace(todo.CWD)
	if cwd == "" {
		cwd = p.workDir
	}
	sessionInput := p.todoOperationSessionInput(issue, todoOperationSessionOptions{
		ID: sessionID, Operation: string(step), Provider: executorName, CWD: cwd, Prompt: promptMarkdown,
	})
	if preparation.Resume && issue.ActivePromptRunID != nil {
		previousRun, runErr := p.captain.GetPromptRun(ctx, *issue.ActivePromptRunID)
		if runErr != nil {
			return todos.RunPreparationResult{}, runErr
		}
		previousSession, sessionErr := p.captain.GetSession(ctx, previousRun.SessionID)
		if sessionErr != nil {
			return todos.RunPreparationResult{}, sessionErr
		}
		if !terminalPromptRun(previousRun.State) {
			links, linkErr := p.repository.ListPromptRuns(ctx, issue.ID)
			if linkErr != nil {
				return todos.RunPreparationResult{}, linkErr
			}
			activeStep := native.StepKind("")
			for _, link := range links {
				if link.PromptRunID == previousRun.ID {
					activeStep = link.StepKind
					break
				}
			}
			if activeStep == "" {
				return todos.RunPreparationResult{}, fmt.Errorf("%w: active prompt run %s is not linked to issue %s", native.ErrLinkConflict, previousRun.ID, issue.ID)
			}
			if activeStep != step {
				return todos.RunPreparationResult{}, fmt.Errorf(
					"%w: active Captain prompt run %s is step %q and cannot resume as %q",
					todos.ErrRunResumeModeMismatch, previousRun.ID, activeStep, step,
				)
			}
			// A waiting/running prompt run is one interactive operation. Resume it
			// in place rather than manufacturing a second active root operation.
			p.markPrepared(issue.ID, previousRun.ID)
			// This process now drives the run, so the claim has to name it — the
			// previous dispatcher's heartbeat stops where it stopped.
			if err := p.claimRun(ctx, previousRun.ID, owner); err != nil {
				return todos.RunPreparationResult{}, err
			}
			if err := p.replaceTODO(ctx, todo, issue, cwd); err != nil {
				return todos.RunPreparationResult{}, err
			}
			return todos.RunPreparationResult{
				SessionID: previousSession.ID.String(), PromptRunID: previousRun.ID,
			}, nil
		}
		// A terminal operation gets a new prompt run, but continuation keeps the
		// authoritative Captain/provider session identity.
		sessionInput = captaindb.CreateSessionInput{
			ID: previousSession.ID, ProviderSessionID: previousSession.ProviderSessionID,
			Source: previousSession.Source, Provider: previousSession.Provider,
			HostID: previousSession.HostID, ParentSessionID: previousSession.ParentSessionID,
			RootSessionID: previousSession.RootSessionID, Path: previousSession.Path,
			Project: previousSession.Project, CWD: previousSession.CWD,
			Title: previousSession.Title, InitialPrompt: previousSession.InitialPrompt,
			Slug: previousSession.Slug, AgentType: previousSession.AgentType,
			Description: previousSession.Description, CLIVersion: previousSession.CLIVersion,
		}
	}

	rendered, err := renderedSpec(preparation.Spec, issue.Verification)
	if err != nil {
		return todos.RunPreparationResult{}, err
	}
	promptInput := captaindb.CreatePromptRunInput{
		ID:     promptRunID,
		Origin: "gavel.todos",
		// SpecProfile records WHICH prompt ran, where Runtime.Mode records how it
		// behaved. For the built-ins the two are the same string, so stored rows are
		// unchanged; for any other prompt this is the only place its name survives.
		SpecProfile:  promptName,
		AdmissionKey: identity.AdmissionKey,
		RenderedSpec: rendered,
		Runtime: captaindb.PromptRunRuntime{
			Mode: string(mode), Driver: executorName, Requested: preparation.Requested,
		},
		PromptMarkdown:       promptMarkdown,
		VerificationMarkdown: issue.Verification,
	}
	if err := p.attachInputPlan(ctx, issue, mode, &promptInput); err != nil {
		return todos.RunPreparationResult{}, err
	}
	launch, err := p.coordinator.LaunchPromptRun(ctx, native.PromptRunLaunchInput{
		RootSession: p.todoRootSessionInput(native.CreateIssueInput{
			ID: issue.ID, Title: issue.Title, Body: issue.Body,
		}),
		Session:   sessionInput,
		PromptRun: promptInput,
		Attachment: native.PromptRunLaunchAttachment{
			IssueID:              issue.ID,
			StepKind:             step,
			Ordinal:              ordinal,
			ExpectedIssueVersion: issue.Version,
			Actor:                mutationActor,
			Owner:                &owner,
		},
	})
	if err != nil {
		return todos.RunPreparationResult{}, err
	}
	// This dispatch is a new phase run; drop the memo so a read through this
	// provider sees it rather than the index loaded before the launch.
	p.dropPhaseRuns(issue.WorkspaceID)
	p.dropExecutionIndex(issue.WorkspaceID)
	if !launch.DispatchOwned {
		return todos.RunPreparationResult{}, fmt.Errorf(
			"%w: Captain prompt run %s for issue %s already has an external dispatcher",
			todos.ErrRunDispatchAlreadyClaimed, launch.PromptRun.ID, issue.ID,
		)
	}
	p.markPrepared(issue.ID, launch.PromptRun.ID)
	// The claim was written by the launch transaction; start refreshing it so
	// another process can tell this dispatcher apart from one that has exited.
	p.ownership.start(p.repository, launch.PromptRun.ID, owner)
	if err := p.replaceTODO(ctx, todo, launch.Issue, cwd); err != nil {
		return todos.RunPreparationResult{}, err
	}
	return todos.RunPreparationResult{
		SessionID: launch.Session.ID.String(), PromptRunID: launch.PromptRun.ID,
	}, nil
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
