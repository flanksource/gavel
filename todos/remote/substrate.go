// Serve's side of a hosted step on substrate: SubstrateRunner runs the step's whole prompt run in
// a substrate environment through captain's substrate sandbox, as an agent run of serve's own.
package remote

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"net/http"
	"strings"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/captain/pkg/promptrun"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/sandbox"
	"github.com/flanksource/gavel/sandbox/runclient"
	"github.com/flanksource/gavel/tenant"
	"github.com/flanksource/gavel/todos/lifecycle"
	"github.com/google/uuid"
)

// SubstrateRunner runs one todo's steps on substrate. It is a lifecycle.RemoteRunner.
type SubstrateRunner struct {
	Dispatcher *sandbox.Dispatcher
	Sources    *sandbox.Sources
	// Backend are the options of the substrate backend runs execute on, as captain's substrate
	// sandbox takes them: how to reach the substrate, the image, its template, egress. Each run
	// adds its own.
	Backend map[string]any
	Scope   tenant.Scope
	Repo    sandbox.Repo
	// SHA is the commit of the repository's default branch the run was resolved from.
	SHA    string
	TodoID uuid.UUID
	// Transcripts keeps a run's transcript files.
	Transcripts func(ctx context.Context, run sandbox.RunRecord, source, name string, data io.Reader) error
	// RunAPI is serve's run API, which answers what a run's program asks of its PR, as the run.
	RunAPI http.Handler
}

// Run records run as an agent run of the dispatcher and runs it through captain's substrate
// sandbox: in environment JobName(run), from the todo's commit in serve's git directory of the
// repository, with gavel's program rebuilding its hooks there and reaching the step's PR through
// serve's run API over the run's channel. The run's output is its log, its transcripts are kept, and its commits come back
// on its own branch, which becomes the repository's.
func (r *SubstrateRunner) Run(ctx context.Context, run lifecycle.RemoteRun) (promptrun.Result, error) {
	id := uuid.New()
	todo := r.TodoID
	spec := sandbox.RunSpec{Kind: "agent", RepoID: r.Repo.ID, Ref: "HEAD", SHA: r.SHA, TodoID: &todo}
	if pr := run.Hooks.PR; pr != nil {
		// A step on a linked PR starts from the PR's head, which the base repository has under
		// the PR's own ref, and serve acts on that PR alone for it.
		spec.Ref, spec.SHA, spec.PR = fmt.Sprintf("refs/pull/%d/head", pr.Number), pr.Target.HeadSHA, pr.URL
	}
	recorded, err := r.Dispatcher.RecordRun(ctx, id, r.Scope, spec)
	if err != nil {
		return promptrun.Result{}, err
	}
	succeeded := false
	defer func() { recorded.Finish(succeeded) }()
	gitDir, err := r.Sources.GitDir(ctx, r.Repo)
	if err != nil {
		return promptrun.Result{}, err
	}
	branch := sandbox.RunBranch(todo, id)

	in := run.Input
	in.Config.SandboxSelection = &api.SandboxConfig{Kind: api.SandboxSubstrate, Name: "gavel-hosted", Options: r.runOptions(recorded, spec.SHA, branch)}
	in.Resolved.Spec.Sandbox = &api.SandboxRef{Mode: api.SandboxSubstrate}
	in.Repo = gitDir
	in.EnvironmentOutput = recorded.Log
	record := sandbox.RunRecord{ID: id, Org: r.Scope.OrgID, User: r.Scope.UserID, Spec: spec}
	in.HostCalls = map[string]func(context.Context, json.RawMessage) (json.RawMessage, error){
		runclient.HostCallName: runclient.AnswerHostCall(r.RunAPI, id.String(), recorded.Token),
	}
	in.OnTranscript = func(source, name string, data []byte) {
		if err := r.Transcripts(context.WithoutCancel(ctx), record, source, name, strings.NewReader(string(data))); err != nil {
			logger.Warnf("keep transcript %s of agent run %s: %v", name, id, err)
		}
	}
	result, err := promptrun.Run(ctx, in)
	if err != nil {
		return result, err
	}
	if worktree := runWorktree(result); worktree != nil && worktree.Branch == branch {
		if err := r.Sources.PublishRunBranch(ctx, r.Repo, branch, worktree.Head); err != nil {
			return result, fmt.Errorf("agent run %s: %w", id, err)
		}
	}
	succeeded = true
	return result, nil
}

// runOptions are the substrate backend's options with the run's own: its environment, the commit
// it starts from and the branch its commits come back on, gavel's program, the organization's
// model keys, and the run's lifetime.
func (r *SubstrateRunner) runOptions(recorded *sandbox.RecordedRun, sha, branch string) map[string]any {
	options := maps.Clone(r.Backend)
	credentials := make([]any, len(recorded.Credentials))
	for i, credential := range recorded.Credentials {
		credentials[i] = map[string]any{
			"host": credential.Host, "header": credential.Header, "prefix": credential.Prefix, "env": credential.Env, "uri": credential.URI,
		}
	}
	options["environment"] = sandbox.JobName(recorded.ID)
	options["ref"] = sha
	options["branch"] = branch
	options["command"] = []string{"gavel", "runner", "run-task"}
	// The environment holds no run token: the run's program asks serve, its caller, for what it
	// can't do itself over the run's own channel.
	options["env"] = map[string]string{"GAVEL_DB": "off", "CI": "1"}
	options["credentials"] = credentials
	options["maxLifetime"] = r.Dispatcher.RunLifetime().String()
	return options
}

// runWorktree is the worktree a run's commits came back on, when they did.
func runWorktree(result promptrun.Result) *api.WorktreeState {
	if result.Response == nil || result.Response.Workspace == nil {
		return nil
	}
	return result.Response.Workspace.Worktree
}
