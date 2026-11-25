package lint

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/flanksource/gavel/comments"
	"github.com/flanksource/gavel/linters"
)

type RunOptions struct {
	Service   *comments.Service
	Target    comments.Target
	RequestID string
	Selection Selection
}

// RunRecorded retains prompt selection, native execution references and findings.
// Captain remains the source of execution and approval state.
func RunRecorded(ctx context.Context, opts RunOptions) (comments.LintRun, error) {
	if opts.Service == nil {
		return comments.LintRun{}, comments.ErrInvalid
	}
	selection, err := json.Marshal(opts.Selection)
	if err != nil {
		return comments.LintRun{}, err
	}
	run, created, err := opts.Service.BeginLint(ctx, comments.LintRequest{Target: opts.Target, RequestID: opts.RequestID, PromptID: opts.Selection.Prompt, PromptVersion: opts.Selection.Version, Selection: selection})
	if err != nil {
		return run, err
	}
	if !created {
		return run, fmt.Errorf("%w: lint request already admitted; inspect its execution references", comments.ErrConflict)
	}
	prepared, err := Prepare(ctx, opts.Selection)
	if err != nil {
		return run, err
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	recorder := &runRecorder{ctx: ctx, cancel: cancel, service: opts.Service, id: run.ID, refs: map[string]comments.ExecutionRef{}}
	results := prepared.Run(ctx, recorder)
	var executionErr error
	for _, result := range results {
		recorder.UpsertLintResult(result)
		if result.Error != "" {
			executionErr = fmt.Errorf("lint %s: %s", result.Linter, result.Error)
		}
	}
	if recorder.err != nil {
		return run, recorder.err
	}
	if executionErr != nil {
		return run, executionErr
	}
	return opts.Service.LintRun(ctx, run.ID)
}

type runRecorder struct {
	mu      sync.Mutex
	ctx     context.Context
	cancel  context.CancelFunc
	service *comments.Service
	id      string
	err     error
	refs    map[string]comments.ExecutionRef
}

func (r *runRecorder) UpsertLintResult(result *linters.LinterResult) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.err != nil || result.RunID == "" {
		return
	}
	findings, err := json.Marshal(result.Violations)
	if err == nil {
		ref := comments.ExecutionRef{RunID: result.RunID, SessionID: result.SessionID}
		_, err = r.service.RecordLint(r.ctx, r.id, ref, result.Linter, findings)
		if err == nil {
			r.refs[result.Linter] = ref
		}
	}
	if err != nil {
		r.err = err
		r.cancel()
	}
}
func (r *runRecorder) RemoveLintResult(linter, _ string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	ref, ok := r.refs[linter]
	if !ok || r.err != nil {
		return
	}
	_, r.err = r.service.RecordLint(r.ctx, r.id, ref, linter, json.RawMessage(`[]`))
	if r.err != nil {
		r.cancel()
	}
}
