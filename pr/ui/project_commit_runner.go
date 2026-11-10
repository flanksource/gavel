package ui

import (
	"context"
	"fmt"
	"io"

	cexec "github.com/flanksource/clicky/exec"
	clickytask "github.com/flanksource/clicky/task"
	commonscontext "github.com/flanksource/commons/context"
)

type ProjectCommitRunner func(context.Context, []string, io.Writer) error

func (s *Server) SetProjectCommitRunner(run ProjectCommitRunner) {
	s.projectCommitRunner = run
}

var executeProjectCommit = (*Server).runProjectCommit

func (s *Server) runProjectCommit(ctx context.Context, _ string, args []string, group *clickytask.Group, opts ...clickytask.Option) clickytask.TypedTask[cexec.ExecResult] {
	ensurePromptManager()
	task := clickytask.StartTask("Run gavel commit", func(ctx commonscontext.Context, task *clickytask.Task) (cexec.ExecResult, error) {
		if s.projectCommitRunner == nil {
			return cexec.ExecResult{ExitCode: 1}, fmt.Errorf("project commit runner is not configured")
		}
		if err := s.projectCommitRunner(ctx, args, taskCommitWriter{task}); err != nil {
			return cexec.ExecResult{ExitCode: 1}, err
		}
		return cexec.ExecResult{}, nil
	}, append([]clickytask.Option{clickytask.WithGroup(group), clickytask.WithContext(ctx)}, opts...)...)
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
