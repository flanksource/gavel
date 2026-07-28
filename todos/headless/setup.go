package headless

import (
	captainai "github.com/flanksource/captain/pkg/ai"
	dbcontext "github.com/flanksource/commons-db/context"
	"github.com/flanksource/commons-db/shell"
	todopkg "github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
)

func (e *Executor) prepareRequest(ctx *todopkg.ExecutorContext, todosInGroup []*types.TODO, req captainai.Request) (captainai.Request, func() error, error) {
	workDir := groupWorkDir(e.config.WorkDir, todosInGroup)
	setup := shell.Setup{}
	if req.Setup != nil {
		setup = *req.Setup
	}
	resolved, err := setup.Resolve(workDir)
	if err != nil {
		return captainai.Request{}, nil, err
	}
	runEnv := append([]string(nil), resolved.Env...)
	prepared, err := shell.Prepare(dbcontext.NewContext(ctx), &resolved)
	if err != nil {
		return captainai.Request{}, nil, err
	}
	resolved.Cwd = prepared.Cwd
	resolved.Env = append(prepared.Env, runEnv...)
	req.Setup = &resolved
	return req, prepared.Cleanup, nil
}
