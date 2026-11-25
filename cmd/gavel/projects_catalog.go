package main

import (
	"context"

	"github.com/flanksource/gavel/projects"
)

// getProject reads one project from the process's catalog.
func getProject(ctx context.Context, name string) (projects.Project, error) {
	catalog, err := projects.Shared(ctx)
	if err != nil {
		return projects.Project{}, err
	}
	return catalog.Get(ctx, name)
}

// projectForDir is the process catalog's most specific project containing dir.
func projectForDir(ctx context.Context, dir string) (projects.Project, error) {
	catalog, err := projects.Shared(ctx)
	if err != nil {
		return projects.Project{}, err
	}
	return catalog.ForDir(ctx, dir)
}
