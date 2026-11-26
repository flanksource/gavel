// Dispatches sandbox runs: records each in sandbox_runs, submits its job with a run token, streams
// its log to the organization's blob store and the caller, and records how it ended.
package sandbox

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/flanksource/captain/pkg/substrate"
	clickytask "github.com/flanksource/clicky/task"
	commonscontext "github.com/flanksource/commons/context"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/blob"
	"github.com/flanksource/gavel/modelkeys"
	"github.com/flanksource/gavel/tenant"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// runKinds are the kinds sandbox_runs accepts.
var runKinds = []string{"test", "lint", "push-check", "agent"}

// RunSpec is what a sandbox job runs; the run API serves it to the job.
type RunSpec struct {
	Kind   string     `json:"kind"` // test, lint, push-check or agent
	Args   []string   `json:"args,omitempty"`
	RepoID uuid.UUID  `json:"repoId"`
	Ref    string     `json:"ref,omitempty"`
	SHA    string     `json:"sha,omitempty"`
	TodoID *uuid.UUID `json:"todoId,omitempty"`
	Step   string     `json:"step,omitempty"`
	// PR is the URL of the PR an agent run works on, the one PR serve acts on for it.
	PR string `json:"pr,omitempty"`
}

// Dispatcher starts sandbox runs on an Executor.
type Dispatcher struct {
	DB       *gorm.DB
	Executor Executor
	// Logs is an organization's part of the blob store; run logs go to runs/<id>/log.txt in it.
	Logs  func(org uuid.UUID) blob.Store
	Image string
	// Server is gavel serve's public URL, which the job calls back with its run token.
	Server   string
	TokenKey []byte
	Limits   Limits
	// SecretEnv is added to every job's secrets: the model API keys agent runs use.
	SecretEnv map[string]string
	// Keys are organizations' model keys. With CredentialURI set, each job is given its
	// organization's keys as credentials the gateway injects, which CredentialURI names for the
	// run, and an agent run without one is refused.
	Keys          modelkeys.Store
	CredentialURI func(run uuid.UUID, provider string) string
	// Sources, when set, gives each job the repository its commit is in, for an executor that
	// brings the source into the job, and has the job's runner clone it from the checkout the job
	// starts in.
	Sources *Sources
	// Templates, when set, creates each run's environment from its organization's substrate
	// template, whose name and size the run records.
	Templates *OrgTemplates
	// Admission, when set with Templates, admits each run within its organization's budgets: a
	// run is refused, or waits its turn, as they say, and Enforce stops running runs.
	Admission *Admission

	// active are the running runs Enforce can stop.
	activeMu sync.Mutex
	active   map[uuid.UUID]*activeRun
}

// ErrNoModelKey refuses an agent run whose organization has no model key.
var ErrNoModelKey = errors.New("this organization has no model key")

// credentials are the injected credentials of run, one per model key of org; none unless
// CredentialURI is set. An agent run needs at least one.
func (d *Dispatcher) credentials(ctx context.Context, run, org uuid.UUID, kind string) ([]substrate.Credential, error) {
	if d.CredentialURI == nil {
		return nil, nil
	}
	keys, err := d.Keys.List(ctx, org)
	if err != nil {
		return nil, err
	}
	if kind == "agent" && len(keys) == 0 {
		return nil, fmt.Errorf("%w: an organization admin adds one under Model keys in the organization's settings", ErrNoModelKey)
	}
	credentials := make([]substrate.Credential, 0, len(keys))
	for _, key := range keys {
		k, ok := modelkeys.KindNamed(key.Kind)
		if !ok {
			return nil, fmt.Errorf("organization %s's %s key is of kind %q, which gavel no longer knows; an organization admin must set it again", org, key.Provider, key.Kind)
		}
		credentials = append(credentials, substrate.Credential{
			Host: k.Host, Header: k.Header, Prefix: k.Prefix, Env: k.Env, URI: d.CredentialURI(run, k.Provider),
		})
	}
	return credentials, nil
}

// RunTimeout is the longest a run may work.
func (d *Dispatcher) RunTimeout() time.Duration {
	return d.Limits.withDefaults().Timeout
}

// reportTime is how long a run has, past its timeout, to report how it ended.
const reportTime = 15 * time.Minute

// RunLifetime is how long a run's token, and so the run, may last: its timeout plus reportTime.
func (d *Dispatcher) RunLifetime() time.Duration {
	return d.RunTimeout() + reportTime
}

// Run is a dispatched run.
type Run struct {
	ID   uuid.UUID
	done chan struct{}

	mu sync.Mutex
	// handle is the run's job once it is submitted; stopWaiting takes it out of its
	// organization's line while it waits for room.
	handle      Handle
	stopWaiting context.CancelFunc

	result Result
	err    error
}

// Wait blocks until the run ends and returns its result.
func (r *Run) Wait(ctx context.Context) (Result, error) {
	select {
	case <-r.done:
		return r.result, r.err
	case <-ctx.Done():
		return Result{}, ctx.Err()
	}
}

// Cancel stops the run's job, or, while the run waits for room, takes it out of line.
func (r *Run) Cancel(ctx context.Context) error {
	r.mu.Lock()
	handle, stopWaiting := r.handle, r.stopWaiting
	r.mu.Unlock()
	if handle != nil {
		return handle.Cancel(ctx)
	}
	if stopWaiting != nil {
		stopWaiting()
	}
	return nil
}

// Start records a queued run, submits its job and follows it in the background. output, when
// set, receives the run's log as it arrives. The returned Run is set even when submitting fails,
// so the caller can name the failed record.
func (d *Dispatcher) Start(ctx context.Context, scope tenant.Scope, spec RunSpec, output io.Writer) (*Run, error) {
	return d.StartRun(ctx, uuid.New(), scope, spec, output)
}

// StartRun is Start for a run whose ID the caller chose, so it can be ready for the run's
// messages before its job starts. A run its organization's budgets have no room for is refused;
// one that must wait for room is returned queued, its job submitted once it has room, and output
// is told why it waits.
func (d *Dispatcher) StartRun(ctx context.Context, id uuid.UUID, scope tenant.Scope, spec RunSpec, output io.Writer) (*Run, error) {
	recorded, err := d.record(ctx, id, scope, spec)
	if err != nil {
		return nil, err
	}
	run := &Run{ID: id, done: make(chan struct{})}
	ticket, err := d.admit(ctx, id, scope.OrgID, Demand{Size: recorded.template.Size, ModelCalls: spec.Kind == "agent"})
	if err != nil {
		return run, err
	}
	if queued, reason := ticket.Queued(); queued {
		d.setStatusMessage(context.WithoutCancel(ctx), id, "waiting", reason)
		waitCtx, stopWaiting := context.WithCancel(context.WithoutCancel(ctx))
		run.stopWaiting = stopWaiting
		go func() {
			defer stopWaiting()
			if err := d.waitForRoom(waitCtx, id, ticket, output); err != nil {
				run.err = err
				close(run.done)
				return
			}
			if err := d.launch(ctx, run, scope, spec, recorded, ticket, output); err != nil {
				run.err = err
				close(run.done)
			}
		}()
		return run, nil
	}
	return run, d.launch(ctx, run, scope, spec, recorded, ticket, output)
}

// admit admits run of org within the organization's budgets, recording why when it is refused.
// Without Admission or Templates every run is admitted, holding nothing.
func (d *Dispatcher) admit(ctx context.Context, run, org uuid.UUID, demand Demand) (*Ticket, error) {
	if d.Admission == nil || d.Templates == nil {
		return nil, nil
	}
	ticket, err := d.Admission.Enter(ctx, org, demand)
	if err != nil {
		_ = d.finish(context.WithoutCancel(ctx), run, "failed", nil, err.Error())
		return nil, fmt.Errorf("start sandbox run %s: %w", run, err)
	}
	return ticket, nil
}

// waitForRoom waits for ticket's run to have room, telling output what it waits for, and records
// why when it never gets it.
func (d *Dispatcher) waitForRoom(ctx context.Context, run uuid.UUID, ticket *Ticket, output io.Writer) error {
	err := ticket.Wait(ctx, func(report string) {
		if output != nil {
			fmt.Fprintln(output, report)
		}
	})
	switch {
	case errors.Is(err, context.Canceled):
		_ = d.finish(context.WithoutCancel(ctx), run, "cancelled", nil, "cancelled while it waited for room")
	case err != nil:
		_ = d.finish(context.WithoutCancel(ctx), run, "failed", nil, err.Error())
	}
	return err
}

// launch submits run's job, signing its token now, and follows it in the background, giving
// back its ticket's room when it ends.
func (d *Dispatcher) launch(ctx context.Context, run *Run, scope tenant.Scope, spec RunSpec, recorded recordedRun, ticket *Ticket, output io.Writer) error {
	limits := d.Limits.withDefaults()
	job := Job{
		RunID:       run.ID,
		Scope:       scope,
		Image:       d.Image,
		Command:     []string{"gavel", "runner", "exec", "--run", run.ID.String()},
		Env:         map[string]string{"GAVEL_SERVER": d.Server, "GAVEL_RUN_ID": run.ID.String(), "GAVEL_DB": "off", "CI": "1"},
		SecretEnv:   map[string]string{"GAVEL_RUN_TOKEN": d.runToken(run.ID, scope.OrgID)},
		Credentials: recorded.credentials,
		Limits:      limits,
		Template:    recorded.template.Name,
	}
	for key, value := range d.SecretEnv {
		if key != "GAVEL_RUN_TOKEN" {
			job.SecretEnv[key] = value
		}
	}
	if d.Sources != nil {
		source, err := d.jobSource(ctx, scope, spec)
		if err != nil {
			ticket.Release()
			// The run is not tracked until it is submitted, so nothing can have stopped it yet.
			_ = d.finish(context.WithoutCancel(ctx), run.ID, "failed", nil, "could not find the commit to run: "+err.Error())
			return fmt.Errorf("start sandbox run %s: %w", run.ID, err)
		}
		job.Source = source
		job.Command = append(job.Command, "--source", ".")
	}
	handle, err := d.Executor.Submit(ctx, job)
	if err != nil {
		ticket.Release()
		// The run is not tracked until it is submitted, so nothing can have stopped it yet.
		_ = d.finish(context.WithoutCancel(ctx), run.ID, "failed", nil, fmt.Sprintf("%s could not start the job: %v", d.Executor.Name(), err))
		return fmt.Errorf("start sandbox run %s: %w", run.ID, err)
	}
	run.mu.Lock()
	run.handle = handle
	run.mu.Unlock()
	d.track(run.ID, scope.OrgID, spec.Kind == "agent", func(reason error) {
		d.setStatusMessage(context.WithoutCancel(ctx), run.ID, "stopping", "stopped: "+reason.Error())
		if err := handle.Cancel(context.WithoutCancel(ctx)); err != nil {
			logger.Warnf("stop sandbox run %s: %v", run.ID, err)
		}
	})
	logKey := "runs/" + run.ID.String() + "/log.txt"
	provisioned, waits := handle.(Provisioned)
	d.submitted(ctx, run.ID, handle.Ref(), logKey, waits)
	if waits {
		watchCtx, stopWatching := context.WithCancel(context.WithoutCancel(ctx))
		go d.watchActor(watchCtx, run.ID)
		go func() {
			defer stopWatching()
			select {
			case <-provisioned.Started():
				d.started(context.WithoutCancel(ctx), run.ID)
			case <-run.done:
			}
		}()
	}

	follow := func(taskCtx commonscontext.Context, task *clickytask.Task) (Result, error) {
		defer ticket.Release()
		return d.follow(context.WithoutCancel(ctx), run.ID, scope.OrgID, handle, logKey, output, task)
	}
	go func() {
		defer close(run.done)
		// A job ends once: following it again would wait on the ended job, finding its run already
		// untracked and losing why the dispatcher stopped it. Clicky's default policy retries any
		// error whose text holds a pattern such as "429", which a run's ID can.
		typed := clickytask.StartTask(fmt.Sprintf("Sandbox %s run %s", spec.Kind, run.ID.String()[:8]), follow,
			clickytask.WithRetryConfig(clickytask.RetryConfig{Retryable: func(error) bool { return false }}))
		run.result, run.err = typed.GetResult()
	}()
	return nil
}

// jobSource is where on serve the commit spec runs is.
func (d *Dispatcher) jobSource(ctx context.Context, scope tenant.Scope, spec RunSpec) (*JobSource, error) {
	repo, err := LoadRepo(ctx, d.DB, scope.OrgID, spec.RepoID)
	if err != nil {
		return nil, err
	}
	dir, err := d.Sources.GitDir(ctx, repo)
	if err != nil {
		return nil, err
	}
	commit := spec.SHA
	if commit == "" {
		commit = spec.Ref
	}
	return &JobSource{GitDir: dir, Commit: commit}, nil
}

// recordedRun is what a run is given when it is recorded: its organization's credentials, and
// the template its environment is created from, which is empty without org templates.
type recordedRun struct {
	credentials []substrate.Credential
	template    OrgTemplate
}

// record files run id as queued, with its organization's credentials and template.
func (d *Dispatcher) record(ctx context.Context, id uuid.UUID, scope tenant.Scope, spec RunSpec) (recordedRun, error) {
	if !contains(runKinds, spec.Kind) {
		return recordedRun{}, fmt.Errorf("sandbox run kind %q is not one of %s", spec.Kind, strings.Join(runKinds, ", "))
	}
	credentials, err := d.credentials(ctx, id, scope.OrgID, spec.Kind)
	if err != nil {
		return recordedRun{}, err
	}
	recorded := recordedRun{credentials: credentials}
	var cpu, memory *int64
	if d.Templates != nil {
		if recorded.template, err = d.Templates.Resolve(ctx, scope.OrgID); err != nil {
			return recordedRun{}, fmt.Errorf("start sandbox run %s: %w", id, err)
		}
		cpu, memory = &recorded.template.Size.CPUMillicores, &recorded.template.Size.MemoryBytes
	}
	encoded, err := json.Marshal(spec)
	if err != nil {
		return recordedRun{}, fmt.Errorf("encode the run spec: %w", err)
	}
	var repo *uuid.UUID
	if spec.RepoID != uuid.Nil {
		repo = &spec.RepoID
	}
	if err := d.DB.WithContext(ctx).Exec(withEvent(`INSERT INTO sandbox_runs (id, org_id, user_id, repo_id, kind, ref, sha, todo_id, status, backend, spec, template, cpu_millicores, memory_bytes)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'queued', ?, ?, ?, ?, ?)`),
		id, scope.OrgID, scope.UserID, repo, spec.Kind, spec.Ref, spec.SHA, spec.TodoID, d.Executor.Name(), string(encoded),
		recorded.template.Name, cpu, memory, "queued", "created").Error; err != nil {
		return recordedRun{}, fmt.Errorf("record sandbox run: %w", err)
	}
	return recorded, nil
}

// runToken is the token run in org calls the run API with, good for the run's lifetime.
func (d *Dispatcher) runToken(run, org uuid.UUID) string {
	return SignRunToken(d.TokenKey, run, org, time.Now().Add(d.RunLifetime()))
}

// RecordedRun is a run recorded here that executes elsewhere than in a job of the dispatcher's
// executor, as an agent run does through captain's substrate sandbox.
type RecordedRun struct {
	ID uuid.UUID
	// Token is the run's token, which its program in the environment calls the run API with.
	Token string
	// Credentials are its organization's model keys, which substrate's egress gateway injects.
	Credentials []substrate.Credential
	// Template is the substrate template its environment is created from, its organization's;
	// empty is the backend's own.
	Template string
	// Log takes the run's output, which is stored as its log and shown on its task.
	Log     io.Writer
	ctx     context.Context
	started func()
	finish  func(err error)
}

// Context ends when the dispatcher stops the run, with why as its cause; the run executes within it.
func (r *RecordedRun) Context() context.Context { return r.ctx }

// Started records the run as running once it runs in its environment, which until then is being
// created and started; WatchStart tells when.
func (r *RecordedRun) Started() { r.started() }

// Finish stores the run's log and records how it ended: succeeded when err is nil, otherwise
// failed with err as why.
func (r *RecordedRun) Finish(err error) { r.finish(err) }

// RecordRun files run id as provisioning, until Started, for a run that executes elsewhere: it signs the run's token,
// gives it its organization's credentials, and stores what is written to its Log as its log. A
// run its organization's budgets have no room for is refused; one that must wait for room waits
// here, saying why in its log.
func (d *Dispatcher) RecordRun(ctx context.Context, id uuid.UUID, scope tenant.Scope, spec RunSpec) (*RecordedRun, error) {
	recorded, err := d.record(ctx, id, scope, spec)
	if err != nil {
		return nil, err
	}
	ticket, err := d.admit(ctx, id, scope.OrgID, Demand{Size: recorded.template.Size, ModelCalls: spec.Kind == "agent"})
	if err != nil {
		return nil, err
	}
	logKey := "runs/" + id.String() + "/log.txt"
	reader, writer := io.Pipe()
	stored := make(chan struct{})
	go func() {
		defer close(stored)
		typed := clickytask.StartTask(fmt.Sprintf("Sandbox %s run %s", spec.Kind, id.String()[:8]), func(_ commonscontext.Context, task *clickytask.Task) (Result, error) {
			d.storeLog(context.WithoutCancel(ctx), id, scope.OrgID, logKey, reader, task, nil)
			return Result{}, nil
		})
		_, _ = typed.GetResult()
	}()
	if queued, reason := ticket.Queued(); queued {
		d.setStatusMessage(context.WithoutCancel(ctx), id, "waiting", reason)
		if err := d.waitForRoom(ctx, id, ticket, writer); err != nil {
			_ = writer.Close()
			<-stored
			return nil, fmt.Errorf("start sandbox run %s: %w", id, err)
		}
	}
	d.submitted(ctx, id, JobName(id), logKey, true)
	watchCtx, stopWatching := context.WithCancel(context.WithoutCancel(ctx))
	go d.watchActor(watchCtx, id)
	runCtx, stop := context.WithCancelCause(ctx)
	d.track(id, scope.OrgID, spec.Kind == "agent", func(reason error) {
		d.setStatusMessage(context.WithoutCancel(ctx), id, "stopping", "stopped: "+reason.Error())
		stop(reason)
	})
	var startOnce sync.Once
	began := make(chan struct{})
	return &RecordedRun{ID: id, Token: d.runToken(id, scope.OrgID), Credentials: recorded.credentials, Template: recorded.template.Name, Log: writer, ctx: runCtx,
		started: func() {
			startOnce.Do(func() {
				stopWatching()
				close(began)
				d.started(context.WithoutCancel(ctx), id)
			})
		},
		finish: func(err error) {
			stopWatching()
			_ = writer.Close()
			<-stored
			ticket.Release()
			defer stop(nil)
			status, message := "succeeded", ""
			if err != nil {
				status, message = "failed", failureMessage(err, began)
			}
			// A stop has already recorded its reason on the run and cancelled the run's context with it.
			_ = d.finish(context.WithoutCancel(ctx), id, status, nil, message)
		}}, nil
}

// follow stores the job's log, waits for it and records how it ended.
func (d *Dispatcher) follow(ctx context.Context, id, org uuid.UUID, handle Handle, logKey string, output io.Writer, task *clickytask.Task) (Result, error) {
	logs, err := handle.Logs(ctx)
	if err != nil {
		logger.Warnf("follow sandbox run %s: %v", id, err)
	} else {
		d.storeLog(ctx, id, org, logKey, logs, task, output)
		_ = logs.Close()
	}
	result, err := handle.Wait(ctx)
	if err != nil {
		var started <-chan struct{}
		if provisioned, ok := handle.(Provisioned); ok {
			started = provisioned.Started()
		}
		if stopped := d.finish(ctx, id, "failed", nil, failureMessage(err, started)); stopped != nil {
			return Result{}, fmt.Errorf("sandbox run %s stopped: %w", id, stopped)
		}
		return Result{}, fmt.Errorf("sandbox run %s: %w", id, err)
	}
	status := "succeeded"
	if result.ExitCode != 0 {
		status = "failed"
	}
	if stopped := d.finish(ctx, id, status, &result.ExitCode, ""); stopped != nil {
		return result, fmt.Errorf("sandbox run %s stopped: %w", id, stopped)
	}
	return result, nil
}

// storeLog stores logs as run id's log under logKey in org's blob store, writing each line to task
// and everything to output, when there is one, as it goes. It reads logs to the end, even when the
// store can't take them, so their writer never blocks.
func (d *Dispatcher) storeLog(ctx context.Context, id, org uuid.UUID, logKey string, logs io.Reader, task *clickytask.Task, output io.Writer) {
	writers := []io.Writer{taskLog{task}}
	if output != nil {
		writers = append(writers, output)
	}
	store := d.Logs(org)
	if store == nil {
		logger.Warnf("sandbox run %s: no blob store for its log", id)
		_, _ = io.Copy(io.MultiWriter(writers...), logs)
		return
	}
	if err := store.Write(ctx, logKey, io.TeeReader(logs, io.MultiWriter(writers...))); err != nil {
		logger.Warnf("store the log of sandbox run %s: %v", id, err)
		_, _ = io.Copy(io.Discard, logs)
	}
}

// withEvent makes statement, an UPDATE or INSERT of sandbox_runs, also record a gavel event for
// each run it touches, of kind with message, the two arguments after statement's own. A statement
// that touches no run records none.
func withEvent(statement string) string {
	return `WITH run AS (` + statement + ` RETURNING id)
		INSERT INTO sandbox_run_events (run_id, source, kind, message) SELECT id, 'gavel', ?, ? FROM run`
}

// submitted records run id's job as submitted as ref, its log kept under logKey: provisioning
// while it waits for its environment, running otherwise.
func (d *Dispatcher) submitted(ctx context.Context, id uuid.UUID, ref, logKey string, waits bool) {
	status, message := "running", fmt.Sprintf("submitted to %s as %s", d.Executor.Name(), ref)
	if waits {
		status, message = "provisioning", fmt.Sprintf("creating sandbox %s on %s and waiting for it to start", ref, d.Executor.Name())
	}
	if err := d.DB.WithContext(ctx).Exec(withEvent(`UPDATE sandbox_runs SET status = ?, status_message = '', backend_ref = ?, log_key = ?, started_at = now() WHERE id = ?`),
		status, ref, logKey, id, status, message).Error; err != nil {
		logger.Warnf("record sandbox run %s as %s: %v", id, status, err)
	}
}

// started records provisioning run id as running: its sandbox started and its job runs there.
func (d *Dispatcher) started(ctx context.Context, id uuid.UUID) {
	if err := d.DB.WithContext(ctx).Exec(withEvent(`UPDATE sandbox_runs SET status = 'running' WHERE id = ? AND status = 'provisioning'`),
		id, "running", "the sandbox started and the job is running in it").Error; err != nil {
		logger.Warnf("record sandbox run %s as running: %v", id, err)
	}
}

// failureMessage is why a job that failed with err failed, saying so when its sandbox never started
// so the job never ran: a platform failure, not the job's. started is closed once the job ran; nil
// when its backend doesn't say.
func failureMessage(err error, started <-chan struct{}) string {
	if started != nil {
		select {
		case <-started:
		default:
			return "the sandbox never started, so the job never ran: " + err.Error()
		}
	}
	return err.Error()
}

// finish records how run id ended, and why when message says, as cancelled when the dispatcher
// stopped it, and returns why it was stopped: nil when it wasn't. A stopped run keeps the reason
// its stop recorded.
func (d *Dispatcher) finish(ctx context.Context, id uuid.UUID, status string, exitCode *int, message string) (stopped error) {
	stopped = d.untrack(id)
	event := message
	if stopped != nil {
		status, message, event = "cancelled", "", "stopped: "+stopped.Error()
	}
	if event == "" && exitCode != nil {
		event = fmt.Sprintf("the job exited with code %d", *exitCode)
	}
	if event == "" {
		event = "the run " + status
	}
	if err := d.DB.WithContext(ctx).Exec(withEvent(`UPDATE sandbox_runs SET status = ?, status_message = CASE WHEN ?::text = '' THEN status_message ELSE ?::text END,
		exit_code = ?, finished_at = now() WHERE id = ?`),
		status, message, message, exitCode, id, status, event).Error; err != nil {
		logger.Warnf("record sandbox run %s as %s: %v", id, status, err)
	}
	return stopped
}

// setStatusMessage records why run id is as it is, such as why it waits, as an event of kind.
func (d *Dispatcher) setStatusMessage(ctx context.Context, id uuid.UUID, kind, message string) {
	if err := d.DB.WithContext(ctx).Exec(withEvent(`UPDATE sandbox_runs SET status_message = ? WHERE id = ?`), message, id, kind, message).Error; err != nil {
		logger.Warnf("record why sandbox run %s is %s: %v", id, kind, err)
	}
}

// FailQueuedRuns fails the runs left waiting for room when serve stopped: their callers are gone
// with it. It returns how many it failed.
func (d *Dispatcher) FailQueuedRuns(ctx context.Context) (int64, error) {
	const message = "serve restarted while the run waited for room"
	result := d.DB.WithContext(ctx).Exec(withEvent(`UPDATE sandbox_runs SET status = 'failed', status_message = ?, finished_at = now()
		WHERE status = 'queued'`), message, "failed", message)
	if result.Error != nil {
		return 0, fmt.Errorf("fail the sandbox runs left queued: %w", result.Error)
	}
	return result.RowsAffected, nil
}

// FailExpiredRuns fails the runs still recorded as provisioning or running past RunLifetime. No run lives that
// long: its token has expired and the sweeper deletes its environment, so serve lost track of it,
// by restarting while it ran or by its follow never returning. It returns how many it failed.
func (d *Dispatcher) FailExpiredRuns(ctx context.Context) (int64, error) {
	lifetime := d.RunLifetime().Seconds()
	const message = "serve lost track of the run: it did not end within its lifetime"
	result := d.DB.WithContext(ctx).Exec(withEvent(`UPDATE sandbox_runs SET status = 'failed', status_message = ?,
		finished_at = started_at + ? * interval '1 second'
		WHERE status IN ('provisioning', 'running') AND started_at < now() - ? * interval '1 second'`), message, lifetime, lifetime, "failed", message)
	if result.Error != nil {
		return 0, fmt.Errorf("fail the sandbox runs past their lifetime: %w", result.Error)
	}
	return result.RowsAffected, nil
}

// RunRecord is what the run API needs of a run: its organization, the user it runs for (nil
// for a system run) and its spec.
type RunRecord struct {
	ID   uuid.UUID
	Org  uuid.UUID
	User *uuid.UUID
	// Status is queued, provisioning, running, succeeded, failed or cancelled.
	Status    string
	CreatedAt time.Time
	Spec      RunSpec
}

// LoadRun reads run id's organization, status and spec.
func LoadRun(ctx context.Context, db *gorm.DB, id uuid.UUID) (RunRecord, error) {
	var row struct {
		OrgID     uuid.UUID
		UserID    *uuid.UUID
		Status    string
		CreatedAt time.Time
		Spec      string
	}
	result := db.WithContext(ctx).Raw(`SELECT org_id, user_id, status, created_at, spec::text AS spec FROM sandbox_runs WHERE id = ?`, id).Scan(&row)
	if result.Error != nil {
		return RunRecord{}, fmt.Errorf("read sandbox run %s: %w", id, result.Error)
	}
	if result.RowsAffected == 0 {
		return RunRecord{}, fmt.Errorf("%w: sandbox run %s", ErrRunNotFound, id)
	}
	var spec RunSpec
	if err := json.Unmarshal([]byte(row.Spec), &spec); err != nil {
		return RunRecord{}, fmt.Errorf("decode the spec of sandbox run %s: %w", id, err)
	}
	return RunRecord{ID: id, Org: row.OrgID, User: row.UserID, Status: row.Status, CreatedAt: row.CreatedAt, Spec: spec}, nil
}

// ErrRunNotFound reports a run ID with no sandbox_runs record.
var ErrRunNotFound = errors.New("sandbox run not found")

// taskLog writes each line it receives to a clicky task's log.
type taskLog struct{ task *clickytask.Task }

func (l taskLog) Write(p []byte) (int, error) {
	if l.task != nil {
		for _, line := range strings.Split(strings.TrimRight(string(p), "\n"), "\n") {
			l.task.Infof("%s", line)
		}
	}
	return len(p), nil
}

func contains(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}
