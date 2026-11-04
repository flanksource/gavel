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

// RunLifetime is how long a run's token, and so the run, may last: its timeout plus 15 minutes
// to report how it ended.
func (d *Dispatcher) RunLifetime() time.Duration {
	return d.Limits.withDefaults().Timeout + 15*time.Minute
}

// Run is a dispatched run.
type Run struct {
	ID     uuid.UUID
	done   chan struct{}
	handle Handle

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

// Cancel stops the run's job.
func (r *Run) Cancel(ctx context.Context) error {
	if r.handle == nil {
		return nil
	}
	return r.handle.Cancel(ctx)
}

// Start records a queued run, submits its job and follows it in the background. output, when
// set, receives the run's log as it arrives. The returned Run is set even when submitting fails,
// so the caller can name the failed record.
func (d *Dispatcher) Start(ctx context.Context, scope tenant.Scope, spec RunSpec, output io.Writer) (*Run, error) {
	return d.StartRun(ctx, uuid.New(), scope, spec, output)
}

// StartRun is Start for a run whose ID the caller chose, so it can be ready for the run's
// messages before its job starts.
func (d *Dispatcher) StartRun(ctx context.Context, id uuid.UUID, scope tenant.Scope, spec RunSpec, output io.Writer) (*Run, error) {
	credentials, err := d.record(ctx, id, scope, spec)
	if err != nil {
		return nil, err
	}
	run := &Run{ID: id, done: make(chan struct{})}

	limits := d.Limits.withDefaults()
	job := Job{
		RunID:       run.ID,
		Scope:       scope,
		Image:       d.Image,
		Command:     []string{"gavel", "runner", "exec", "--run", run.ID.String()},
		Env:         map[string]string{"GAVEL_SERVER": d.Server, "GAVEL_RUN_ID": run.ID.String(), "GAVEL_DB": "off", "CI": "1"},
		SecretEnv:   map[string]string{"GAVEL_RUN_TOKEN": d.runToken(run.ID, scope.OrgID)},
		Credentials: credentials,
		Limits:      limits,
	}
	for key, value := range d.SecretEnv {
		if key != "GAVEL_RUN_TOKEN" {
			job.SecretEnv[key] = value
		}
	}
	handle, err := d.Executor.Submit(ctx, job)
	if err != nil {
		d.finish(context.WithoutCancel(ctx), run.ID, "failed", nil)
		return run, fmt.Errorf("start sandbox run %s: %w", run.ID, err)
	}
	run.handle = handle
	logKey := "runs/" + run.ID.String() + "/log.txt"
	if err := d.DB.WithContext(ctx).Exec(`UPDATE sandbox_runs SET status = 'running', backend_ref = ?, log_key = ?, started_at = now() WHERE id = ?`,
		handle.Ref(), logKey, run.ID).Error; err != nil {
		logger.Warnf("record sandbox run %s as running: %v", run.ID, err)
	}

	follow := func(taskCtx commonscontext.Context, task *clickytask.Task) (Result, error) {
		return d.follow(context.WithoutCancel(ctx), run.ID, scope.OrgID, handle, logKey, output, task)
	}
	go func() {
		defer close(run.done)
		typed := clickytask.StartTask(fmt.Sprintf("Sandbox %s run %s", spec.Kind, run.ID.String()[:8]), follow)
		run.result, run.err = typed.GetResult()
	}()
	return run, nil
}

// record files run id as queued, with its organization's credentials, which it returns.
func (d *Dispatcher) record(ctx context.Context, id uuid.UUID, scope tenant.Scope, spec RunSpec) ([]substrate.Credential, error) {
	if !contains(runKinds, spec.Kind) {
		return nil, fmt.Errorf("sandbox run kind %q is not one of %s", spec.Kind, strings.Join(runKinds, ", "))
	}
	credentials, err := d.credentials(ctx, id, scope.OrgID, spec.Kind)
	if err != nil {
		return nil, err
	}
	encoded, err := json.Marshal(spec)
	if err != nil {
		return nil, fmt.Errorf("encode the run spec: %w", err)
	}
	var repo *uuid.UUID
	if spec.RepoID != uuid.Nil {
		repo = &spec.RepoID
	}
	if err := d.DB.WithContext(ctx).Exec(`INSERT INTO sandbox_runs (id, org_id, user_id, repo_id, kind, ref, sha, todo_id, status, backend, spec)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'queued', ?, ?)`,
		id, scope.OrgID, scope.UserID, repo, spec.Kind, spec.Ref, spec.SHA, spec.TodoID, d.Executor.Name(), string(encoded)).Error; err != nil {
		return nil, fmt.Errorf("record sandbox run: %w", err)
	}
	return credentials, nil
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
	// Log takes the run's output, which is stored as its log and shown on its task.
	Log    io.Writer
	finish func(succeeded bool)
}

// Finish stores the run's log and records whether it succeeded.
func (r *RecordedRun) Finish(succeeded bool) { r.finish(succeeded) }

// RecordRun files run id as running, for a run that executes elsewhere: it signs the run's token,
// gives it its organization's credentials, and stores what is written to its Log as its log.
func (d *Dispatcher) RecordRun(ctx context.Context, id uuid.UUID, scope tenant.Scope, spec RunSpec) (*RecordedRun, error) {
	credentials, err := d.record(ctx, id, scope, spec)
	if err != nil {
		return nil, err
	}
	logKey := "runs/" + id.String() + "/log.txt"
	if err := d.DB.WithContext(ctx).Exec(`UPDATE sandbox_runs SET status = 'running', backend_ref = ?, log_key = ?, started_at = now() WHERE id = ?`,
		JobName(id), logKey, id).Error; err != nil {
		logger.Warnf("record sandbox run %s as running: %v", id, err)
	}
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
	return &RecordedRun{ID: id, Token: d.runToken(id, scope.OrgID), Credentials: credentials, Log: writer, finish: func(succeeded bool) {
		_ = writer.Close()
		<-stored
		status := "succeeded"
		if !succeeded {
			status = "failed"
		}
		d.finish(context.WithoutCancel(ctx), id, status, nil)
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
		d.finish(ctx, id, "failed", nil)
		return Result{}, fmt.Errorf("sandbox run %s: %w", id, err)
	}
	status := "succeeded"
	if result.ExitCode != 0 {
		status = "failed"
	}
	d.finish(ctx, id, status, &result.ExitCode)
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

func (d *Dispatcher) finish(ctx context.Context, id uuid.UUID, status string, exitCode *int) {
	if err := d.DB.WithContext(ctx).Exec(`UPDATE sandbox_runs SET status = ?, exit_code = ?, finished_at = now() WHERE id = ?`,
		status, exitCode, id).Error; err != nil {
		logger.Warnf("record sandbox run %s as %s: %v", id, status, err)
	}
}

// RunRecord is what the run API needs of a run: its organization, the user it runs for (nil
// for a system run) and its spec.
type RunRecord struct {
	ID   uuid.UUID
	Org  uuid.UUID
	User *uuid.UUID
	// Status is queued, running, succeeded or failed.
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
