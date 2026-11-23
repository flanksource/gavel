package ui

import (
	"context"
	"fmt"
	"io"
	"sync"

	cexec "github.com/flanksource/clicky/exec"
	clickytask "github.com/flanksource/clicky/task"
	commonscontext "github.com/flanksource/commons/context"
	"github.com/flanksource/gavel/commit"
)

// ProjectCommitRunner runs one `gavel commit` invocation, streaming its output
// to the writer and returning its result (nil when it produced none).
type ProjectCommitRunner func(context.Context, []string, io.Writer) (*commit.Result, error)

// projectCommitRunDetails is what a commit task reports beyond its log: the PR
// its --push opened or updated, so the dashboard can link to it.
type projectCommitRunDetails struct {
	PR *commit.PushedPR `json:"pr,omitempty"`
}

func (s *Server) SetProjectCommitRunner(run ProjectCommitRunner) {
	s.projectCommitRunner = run
}

var executeProjectCommit = (*Server).runProjectCommit

func (s *Server) runProjectCommit(ctx context.Context, _ string, args []string, group *clickytask.Group, opts ...clickytask.Option) clickytask.TypedTask[cexec.ExecResult] {
	ensurePromptManager()
	opts = append([]clickytask.Option{
		clickytask.WithGroup(group), clickytask.WithContext(ctx),
		clickytask.WithSlots(0), clickytask.WithCancellationDrain(),
	}, opts...)
	task := clickytask.StartTask("Run gavel commit", func(ctx commonscontext.Context, task *clickytask.Task) (cexec.ExecResult, error) {
		if s.projectCommitRunner == nil {
			return cexec.ExecResult{ExitCode: 1}, fmt.Errorf("project commit runner is not configured")
		}
		var mu sync.Mutex
		var pushed *commit.PushedPR
		task.SetDetailsProvider(func() any {
			mu.Lock()
			defer mu.Unlock()
			if pushed == nil {
				return nil
			}
			return projectCommitRunDetails{PR: pushed}
		})
		result, err := s.projectCommitRunner(ctx, args, taskCommitWriter{task})
		// A PR can exist even when the run then failed (e.g. enabling auto-merge).
		if result != nil && result.PR != nil {
			mu.Lock()
			pushed = result.PR
			mu.Unlock()
		}
		if err != nil {
			return cexec.ExecResult{ExitCode: 1}, err
		}
		return cexec.ExecResult{}, nil
	}, opts...)
	task.SetController(&projectCommitController{task: task.Task})
	return task
}

type taskCommitWriter struct{ task *clickytask.Task }

func (w taskCommitWriter) Write(p []byte) (int, error) {
	w.task.Infof("%s", p)
	return len(p), nil
}

type projectCommitController struct{ task *clickytask.Task }

func (c *projectCommitController) Actions() []clickytask.ControlAction {
	if status := c.task.Status(); status == clickytask.StatusRunning || status == clickytask.StatusPending {
		return []clickytask.ControlAction{clickytask.ControlStop}
	}
	return nil
}

func (c *projectCommitController) Control(_ context.Context, action clickytask.ControlAction) error {
	if action != clickytask.ControlStop {
		return fmt.Errorf("project commit does not support %q", action)
	}
	c.task.Cancel()
	return nil
}
