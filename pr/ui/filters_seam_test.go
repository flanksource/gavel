package ui

import "context"

// Package tests have no process database: a Server without a filter store hides
// nothing. Specs exercising the hidden-projects filter give their Server a store
// (see newTrackedGitServer), which reads PostgreSQL as production does.
func init() {
	fromStore := hiddenProjects
	hiddenProjects = func(ctx context.Context, s *Server) (workspaceSet, error) {
		if s == nil || s.filters == nil {
			return workspaceSet{}, nil
		}
		return fromStore(ctx, s)
	}
}
