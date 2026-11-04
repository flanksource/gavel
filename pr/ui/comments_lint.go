package ui

import (
	"context"
	"fmt"
	"github.com/flanksource/gavel/comments"
)

func commentLintDirectory(ctx context.Context, target comments.Target) (string, error) {
	if s := entityServer(ctx); s != nil && s.hosted {
		return "", fmt.Errorf("%w: hosted lint must execute in a repository sandbox", comments.ErrForbidden)
	}
	project, err := commentProject(ctx, target.ScopeURI)
	if err != nil {
		return "", err
	}
	return project.Dir, nil
}
