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
	"github.com/flanksource/commons-db/shell"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/commitauthor"
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
	// Author is who the run's commits are by, the user who started it; nil makes gavel their
	// author.
	Author *commitauthor.Author
}

// Run records run as an agent run of the dispatcher and runs it through captain's substrate
// sandbox: in environment JobName(run), from the todo's commit in serve's git directory of the
// repository, with gavel's program rebuilding its hooks there and reaching the step's PR through
// serve's run API over the run's channel. The run's output is its log, its transcripts are kept, and its commits come back
// on its own branch, which becomes the repository's, or, for a step that continues on one of the
// todo's run branches, on that branch, which moves forward.
func (r *SubstrateRunner) Run(ctx context.Context, run lifecycle.RemoteRun) (result promptrun.Result, err error) {
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
	// The run's record says why it failed; its events say how far it got.
	defer func() { recorded.Finish(err) }()
	gitDir, err := r.Sources.GitDir(ctx, r.Repo)
	if err != nil {
		return promptrun.Result{}, err
	}
	branch, from := sandbox.RunBranch(todo, id), ""
	if continued := continuedRunBranch(run.Input.Resolved.Spec, todo); continued != "" {
		if from, err = r.Sources.RunBranchTip(ctx, r.Repo, continued); err != nil {
			return promptrun.Result{}, err
		}
		branch = continued
	}

	in := run.Input
	if r.Author != nil {
		hooks := run.Hooks
		hooks.Author = r.Author
		if in.HostData, err = json.Marshal(hooks); err != nil {
			return promptrun.Result{}, fmt.Errorf("encode the step's hooks with its author: %w", err)
		}
	}
	in.Config.SandboxSelection = &api.SandboxConfig{Kind: api.SandboxSubstrate, Name: "gavel-hosted", Options: r.runOptions(recorded, spec.SHA, branch)}
	in.Resolved.Spec.Sandbox = &api.SandboxRef{Mode: api.SandboxSubstrate}
	in.Repo = gitDir
	in.EnvironmentOutput = recorded.Log
	record := sandbox.RunRecord{ID: id, Org: r.Scope.OrgID, User: r.Scope.UserID, Spec: spec}
	in.HostCalls = map[string]func(context.Context, json.RawMessage) (json.RawMessage, error){
		runclient.HostCallName: runclient.AnswerHostCall(r.RunAPI, id.String(), recorded.Token, runclient.PRReach),
	}
	in.OnTranscript = func(source, name string, data []byte) {
		if err := r.Transcripts(context.WithoutCancel(ctx), record, source, name, strings.NewReader(string(data))); err != nil {
			logger.Warnf("keep transcript %s of agent run %s: %v", name, id, err)
		}
		// The transcript's model calls count towards the organization's budgets at once.
		if err := r.Dispatcher.Enforce(context.WithoutCancel(ctx), r.Scope.OrgID); err != nil {
			logger.Warnf("agent run %s: %v", id, err)
		}
	}
	// The run's spend stops at what the organization's model cost budget leaves, or its step's
	// own ceiling when that is lower.
	if remaining, capped, err := r.Dispatcher.RemainingModelCost(ctx, r.Scope.OrgID); err != nil {
		return promptrun.Result{}, err
	} else if capped && (in.Resolved.Spec.Budget.Cost == 0 || remaining < in.Resolved.Spec.Budget.Cost) {
		in.Resolved.Spec.Budget.Cost = remaining
	}
	sandbox.WatchStart(&in, recorded.Started)
	// The dispatcher's timeout bounds every run, whatever its step's own budget, so the run ends
	// within its environment's lifetime; the dispatcher stops it sooner once its budgets are used up.
	runCtx, cancel := context.WithTimeout(recorded.Context(), r.Dispatcher.RunTimeout())
	defer cancel()
	result, err = promptrun.Run(runCtx, in)
	if stopped := context.Cause(recorded.Context()); stopped != nil && err != nil {
		return result, fmt.Errorf("agent run %s stopped: %w", id, stopped)
	}
	if err != nil {
		return result, err
	}
	if worktree := runWorktree(result); worktree != nil && worktree.Branch == branch {
		if err := r.publish(ctx, branch, from, worktree.Head); err != nil {
			return result, fmt.Errorf("agent run %s: %w", id, err)
		}
	}
	return result, nil
}

// publish makes branch, at head, the repository's: a branch of the run's own is created, and a
// branch the run continued on from from moves forward.
func (r *SubstrateRunner) publish(ctx context.Context, branch, from, head string) error {
	if from == "" {
		return r.Sources.PublishRunBranch(ctx, r.Repo, branch, head)
	}
	return r.Sources.ContinueRunBranch(ctx, r.Repo, branch, from, head)
}

// continuedRunBranch is the run branch of todo the step's setup checks out to continue on, or ""
// when the step starts a branch of its own.
func continuedRunBranch(spec api.Spec, todo uuid.UUID) string {
	if spec.Setup == nil || spec.Setup.Checkout == nil || spec.Setup.Checkout.Worktree == nil {
		return ""
	}
	worktree := spec.Setup.Checkout.Worktree
	if worktree.Mode != shell.WorktreeBranch || !sandbox.RunBranchOf(todo, worktree.Branch) {
		return ""
	}
	return worktree.Branch
}

// runOptions are the substrate backend's options with the run's own: its environment, the commit
// it starts from and the branch its commits come back on, gavel's program, the organization's
// model keys and template, and the run's lifetime.
func (r *SubstrateRunner) runOptions(recorded *sandbox.RecordedRun, sha, branch string) map[string]any {
	options := maps.Clone(r.Backend)
	options["environment"] = sandbox.JobName(recorded.ID)
	options["ref"] = sha
	options["branch"] = branch
	options["command"] = []string{"gavel", "runner", "run-task"}
	// The environment holds no run token: the run's program asks serve, its caller, for what it
	// can't do itself over the run's own channel.
	options["env"] = map[string]string{"GAVEL_DB": "off", "CI": "1"}
	options["credentials"] = sandbox.SubstrateCredentialOptions(recorded.Credentials)
	options["maxLifetime"] = r.Dispatcher.RunLifetime().String()
	if recorded.Template != "" {
		options["template"] = recorded.Template
	}
	return options
}

// runWorktree is the worktree a run's commits came back on, when they did.
func runWorktree(result promptrun.Result) *api.WorktreeState {
	if result.Response == nil || result.Response.Workspace == nil {
		return nil
	}
	return result.Response.Workspace.Worktree
}
