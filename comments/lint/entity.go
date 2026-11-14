package lint

import (
	"context"
	"fmt"

	"github.com/flanksource/clicky/entity"
	"github.com/flanksource/gavel/comments"
	commentapi "github.com/flanksource/gavel/comments/clicky"
)

type APIOptions struct {
	Service   commentapi.Resolver
	Directory func(context.Context, comments.Target) (string, error)
}
type lintFlags struct {
	Target    comments.Target `flag:"target" help:"Authorized host target as JSON"`
	Selection Selection       `flag:"selection" help:"Prompt version, paths, variables and runtime as JSON"`
	RequestID string          `flag:"requestId" help:"Stable lint request identity"`
}

func (lintFlags) ClickyActionFlags() {}

type materializeFlags struct {
	RunID      string    `flag:"runId" help:"Native Captain run identity from the lint record"`
	WorkflowID string    `flag:"workflowId" help:"Workflow for accepted findings"`
	Findings   []Finding `flag:"findings" help:"Selected findings as JSON"`
}

func (materializeFlags) ClickyActionFlags() {}

func Register(opts APIOptions) error {
	if opts.Service == nil || opts.Directory == nil {
		return fmt.Errorf("lint API requires service and authorized directory resolvers")
	}
	b := entity.NewEntity[comments.LintRun, struct{}, comments.LintRun]("comment-lint").GetWithContext(func(ctx context.Context, id string) (comments.LintRun, error) {
		s, err := opts.Service(ctx)
		if err != nil {
			return comments.LintRun{}, err
		}
		run, err := s.LintRun(ctx, id)
		return run, commentapi.StatusError(err)
	})
	for _, action := range []string{"discover", "inspect", "preview", "run"} {
		b.WithAction(entity.TypedActionWithContext(action, lintFlags{}, func(ctx context.Context, _ string, f lintFlags) (any, error) {
			s, err := opts.Service(ctx)
			if err != nil {
				return nil, err
			}
			target, err := s.ResolveTarget(ctx, "lint", f.Target)
			if err != nil {
				return nil, commentapi.StatusError(err)
			}
			dir, err := opts.Directory(ctx, target)
			if err != nil {
				return nil, commentapi.StatusError(err)
			}
			if f.Selection.Directory != "" && f.Selection.Directory != dir {
				return nil, commentapi.StatusError(comments.ErrInvalid)
			}
			f.Selection.Directory = dir
			var result any
			switch action {
			case "discover":
				result, err = Discover(ctx, dir)
			case "inspect":
				result, err = Inspect(ctx, f.Selection)
			case "preview":
				result, err = Preview(ctx, f.Selection)
			case "run":
				result, err = RunRecorded(ctx, RunOptions{Service: s, Target: target, RequestID: f.RequestID, Selection: f.Selection})
			}
			return result, commentapi.StatusError(err)
		}).WithOptionalID().WithPath("comment-lint/" + action).WithMethod("POST"))
	}
	b.WithAction(entity.TypedActionWithContext("materialize", materializeFlags{}, func(ctx context.Context, id string, f materializeFlags) ([]comments.Comment, error) {
		s, err := opts.Service(ctx)
		if err != nil {
			return nil, err
		}
		run, err := s.LintRun(ctx, id)
		if err != nil {
			return nil, commentapi.StatusError(err)
		}
		found := false
		for _, ref := range run.Executions {
			found = found || ref.RunID == f.RunID
		}
		if !found {
			return nil, commentapi.StatusError(comments.ErrInvalid)
		}
		for _, finding := range f.Findings {
			if finding.Target.ScopeURI != run.Request.Target.ScopeURI {
				return nil, commentapi.StatusError(comments.ErrForbidden)
			}
		}
		result, err := Materialize(ctx, s, MaterializeRequest(f))
		return result, commentapi.StatusError(err)
	}).WithMethod("POST"))
	b.Register()
	return nil
}
