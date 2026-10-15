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

	clickytask "github.com/flanksource/clicky/task"
	commonscontext "github.com/flanksource/commons/context"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/blob"
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
}

// Run is a dispatched run.
type Run struct {
	ID   uuid.UUID
	done chan struct{}

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

// Start records a queued run, submits its job and follows it in the background. output, when
// set, receives the run's log as it arrives. The returned Run is set even when submitting fails,
// so the caller can name the failed record.
func (d *Dispatcher) Start(ctx context.Context, scope tenant.Scope, spec RunSpec, output io.Writer) (*Run, error) {
	if !contains(runKinds, spec.Kind) {
		return nil, fmt.Errorf("sandbox run kind %q is not one of %s", spec.Kind, strings.Join(runKinds, ", "))
	}
	encoded, err := json.Marshal(spec)
	if err != nil {
		return nil, fmt.Errorf("encode the run spec: %w", err)
	}
	run := &Run{ID: uuid.New(), done: make(chan struct{})}
	var repo *uuid.UUID
	if spec.RepoID != uuid.Nil {
		repo = &spec.RepoID
	}
	if err := d.DB.WithContext(ctx).Exec(`INSERT INTO sandbox_runs (id, org_id, user_id, repo_id, kind, ref, sha, todo_id, status, backend, spec)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'queued', ?, ?)`,
		run.ID, scope.OrgID, scope.UserID, repo, spec.Kind, spec.Ref, spec.SHA, spec.TodoID, d.Executor.Name(), string(encoded)).Error; err != nil {
		return nil, fmt.Errorf("record sandbox run: %w", err)
	}

	limits := d.Limits.withDefaults()
	job := Job{
		RunID:     run.ID,
		Scope:     scope,
		Image:     d.Image,
		Command:   []string{"gavel", "runner", "exec", "--run", run.ID.String()},
		Env:       map[string]string{"GAVEL_SERVER": d.Server, "GAVEL_RUN_ID": run.ID.String(), "GAVEL_DB": "off", "CI": "1"},
		SecretEnv: map[string]string{"GAVEL_RUN_TOKEN": SignRunToken(d.TokenKey, run.ID, scope.OrgID, time.Now().Add(limits.Timeout+15*time.Minute))},
		Limits:    limits,
	}
	handle, err := d.Executor.Submit(ctx, job)
	if err != nil {
		d.finish(context.WithoutCancel(ctx), run.ID, "failed", nil)
		return run, fmt.Errorf("start sandbox run %s: %w", run.ID, err)
	}
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

// follow stores the job's log, waits for it and records how it ended.
func (d *Dispatcher) follow(ctx context.Context, id, org uuid.UUID, handle Handle, logKey string, output io.Writer, task *clickytask.Task) (Result, error) {
	logs, err := handle.Logs(ctx)
	if err != nil {
		logger.Warnf("follow sandbox run %s: %v", id, err)
	} else {
		writers := []io.Writer{taskLog{task}}
		if output != nil {
			writers = append(writers, output)
		}
		store := d.Logs(org)
		if store == nil {
			logger.Warnf("sandbox run %s: no blob store for its log", id)
			_, _ = io.Copy(io.MultiWriter(writers...), logs)
		} else if err := store.Write(ctx, logKey, io.TeeReader(logs, io.MultiWriter(writers...))); err != nil {
			logger.Warnf("store the log of sandbox run %s: %v", id, err)
		}
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

func (d *Dispatcher) finish(ctx context.Context, id uuid.UUID, status string, exitCode *int) {
	if err := d.DB.WithContext(ctx).Exec(`UPDATE sandbox_runs SET status = ?, exit_code = ?, finished_at = now() WHERE id = ?`,
		status, exitCode, id).Error; err != nil {
		logger.Warnf("record sandbox run %s as %s: %v", id, status, err)
	}
}

// Spec is the RunSpec run id was started with.
func (d *Dispatcher) Spec(ctx context.Context, id uuid.UUID) (RunSpec, error) {
	var row struct{ Spec string }
	result := d.DB.WithContext(ctx).Raw(`SELECT spec::text AS spec FROM sandbox_runs WHERE id = ?`, id).Scan(&row)
	if result.Error != nil {
		return RunSpec{}, fmt.Errorf("read sandbox run %s: %w", id, result.Error)
	}
	if result.RowsAffected == 0 {
		return RunSpec{}, fmt.Errorf("%w: sandbox run %s", ErrRunNotFound, id)
	}
	var spec RunSpec
	if err := json.Unmarshal([]byte(row.Spec), &spec); err != nil {
		return RunSpec{}, fmt.Errorf("decode the spec of sandbox run %s: %w", id, err)
	}
	return spec, nil
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
