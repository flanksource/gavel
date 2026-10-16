package lifecycle

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/commons-db/shell"
	"github.com/flanksource/gavel/pr/model"
	"github.com/flanksource/gavel/pr/provider"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
)

type prExecution struct {
	client   provider.Client
	number   int
	snapshot *model.StatusSnapshot
	target   *provider.RunTarget
	workDir  string
	write    bool
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
	schema, err := api.SchemaJSON(&types.PRRunEnvelope{})
	if err != nil {
		return err
	}
	spec.Prompt.Schema, spec.Prompt.SchemaJSON = nil, schema
	spec.Prompt.User += "\n\nThis TODO is linked to the PR below. Work on its source branch. Do not commit or push: Gavel commits each completed turn and pushes it to this PR before fixture verification. Never merge, approve, force-push, or resolve unrelated feedback. Emit addressedThreadIds containing only native thread IDs whose feedback your completed turn addressed. Gavel resolves those threads after a successful push. CI, merge readiness, and unresolved feedback from the entire PR must pass at the pushed revision; failures from the fixture are feedback for your next turn. After a fixture failure, refresh gavel pr status for this URL with --json --logs to inspect new reviews and full failure details. Ask when human decisions or credentials are required."
	return setPRStatusContext(&spec.Prompt, p.snapshot)
}

const prStatusContextMarker = "\n\nCurrent PR status and native review thread IDs:\n"

func setPRStatusContext(prompt *api.Prompt, snapshot *model.StatusSnapshot) error {
	status, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return err
	}
	prefix, _, _ := strings.Cut(prompt.User, prStatusContextMarker)
	prompt.User = prefix + prStatusContextMarker + "```json\n" + string(status) + "\n```"
	return nil
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
	return &prExecution{client: client, number: number, snapshot: snapshot, target: target, workDir: workDir, write: class == types.ModeRun}, nil
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
