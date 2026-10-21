package lifecycle

import (
	"context"
	"fmt"
	"maps"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/commons-db/shell"
	"github.com/flanksource/gavel/pr/model"
	"github.com/flanksource/gavel/pr/provider"
	"github.com/flanksource/gavel/todos"
	todoprompt "github.com/flanksource/gavel/todos/prompt"
	"github.com/flanksource/gavel/todos/tools"
	"github.com/flanksource/gavel/todos/types"
)

type prExecution struct {
	client   provider.Client
	number   int
	snapshot *model.StatusSnapshot
	// openThreads are the thread IDs a turn may report addressed; see openThreads.
	openThreads map[string]bool
	target      *provider.RunTarget
	workDir     string
	write       bool
	// events is the step's envelope vocabulary, which a turn's result is held to.
	events []string
	// tools is the run's tool state; pr_thread_addressed records into it.
	tools *tools.Run
}

// forRun is the PR hook of one dispatch, reading the threads that run's tools
// record. A nil execution has no hook.
func (p *prExecution) forRun(run *tools.Run) *prExecution {
	if p == nil {
		return nil
	}
	bound := *p
	bound.tools = run
	return &bound
}

func (p *prExecution) configure(prepared *preparedStep) error {
	spec := &prepared.request
	setup := shell.Setup{}
	if spec.Setup != nil {
		setup = *spec.Setup
	}
	setup.Cwd, setup.Checkout = p.workDir, p.target.Checkout()
	spec.Setup = &setup
	if !p.write {
		if spec.Workflow != nil {
			workflow := *spec.Workflow
			workflow.Commits = nil
			spec.Workflow = &workflow
		}
		prepared.verifyTarget = &types.VerifyTarget{Source: types.VerifyTargetPR, PR: p.snapshot.PR.URL, SHA: p.target.HeadSHA, Branch: p.target.Branch}
		return nil
	}
	workflow := api.Workflow{}
	if spec.Workflow != nil {
		workflow = *spec.Workflow
	}
	workflow.Commits = []api.Commit{{On: api.CommitOnTurn, Mode: api.CommitModeCommit, Stage: api.CommitStageWorktree, Gates: api.CommitGatesFull}}
	spec.Workflow = &workflow
	spec.Prompt.Metadata = maps.Clone(spec.Prompt.Metadata)
	if spec.Prompt.Metadata == nil {
		spec.Prompt.Metadata = map[string]string{}
	}
	spec.Prompt.Metadata[todoprompt.PRWriteMetadataKey] = "true"
	return nil
}

// openThreads are the review threads a run may report addressed: unresolved and
// resolvable when the run was prepared. A thread opened mid-run — by anyone who
// can comment on the PR — is outside the TODO the run was asked to do.
func openThreads(snapshot *model.StatusSnapshot) map[string]bool {
	open := map[string]bool{}
	for _, thread := range snapshot.PR.ReviewThreads {
		if thread.ThreadID != "" && !thread.IsResolved && thread.CanResolve {
			open[thread.ThreadID] = true
		}
	}
	return open
}

func (h *Host) resolvePR(ctx context.Context, todo *types.TODO, class types.RunMode, workDir string) (*prExecution, error) {
	if todo.PR == nil || class == types.ModePlan {
		return nil, nil
	}
	if err := todo.PR.Validate(); err != nil {
		return nil, err
	}
	resolve := func(ctx context.Context, url string) (provider.Client, int, error) {
		return provider.ResolveURL(provider.Options{WorkDir: workDir, Context: ctx}, url)
	}
	if h.PRs != nil {
		resolve = h.PRs
	}
	client, number, err := resolve(ctx, todo.PR.URL)
	if err != nil {
		return nil, err
	}
	snapshot, err := client.Status(ctx, number, model.StatusOptions{Threads: true, Logs: true, TailLogs: 40})
	if err != nil {
		return nil, fmt.Errorf("linked PR status: %w", err)
	}
	if snapshot == nil || snapshot.PR == nil {
		return nil, fmt.Errorf("linked PR provider returned no status")
	}
	target, err := provider.TargetOf(snapshot.PR)
	if err != nil {
		return nil, err
	}
	return &prExecution{client: client, number: number, snapshot: snapshot, openThreads: openThreads(snapshot), target: target, workDir: workDir, write: class == types.ModeRun}, nil
}

func requirePRFixture(todo *types.TODO, spec *api.Spec) error {
	if todo.PR == nil {
		return nil
	}
	workflow := api.Workflow{}
	if spec.Workflow != nil {
		workflow = *spec.Workflow
	}
	verification := api.Verify{}
	if workflow.Verify != nil {
		verification = *workflow.Verify
	}
	fixture, err := todos.PRVerificationFixture(todo.PR, verification.Fixture)
	if err != nil {
		return err
	}
	verification.Fixture = fixture
	if verification.MaxIterations == 0 {
		verification.MaxIterations = 3
	}
	workflow.Verify = &verification
	spec.Workflow = &workflow
	return nil
}
