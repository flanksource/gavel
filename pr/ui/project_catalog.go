package ui

import (
	"context"
	"slices"

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

func createProject(ctx context.Context, p projects.Project) (projects.Project, error) {
	catalog, err := projects.Shared(ctx)
	if err != nil {
		return projects.Project{}, err
	}
	return catalog.Create(ctx, p)
}

func updateProject(ctx context.Context, name string, p projects.Project) (projects.Project, error) {
	catalog, err := projects.Shared(ctx)
	if err != nil {
		return projects.Project{}, err
	}
	return catalog.Update(ctx, name, p)
}

func deleteProject(ctx context.Context, name string) error {
	catalog, err := projects.Shared(ctx)
	if err != nil {
		return err
	}
	return catalog.Delete(ctx, name)
}

// loadProjects lists the process catalog's directory projects, sorted by name:
// the dashboard's workspaces. A web project (gavel launch's pages) has no
// checkout to track.
func loadProjects(ctx context.Context) ([]projects.Project, error) {
	catalog, err := projects.Shared(ctx)
	if err != nil {
		return nil, err
	}
	all, err := catalog.List(ctx)
	if err != nil {
		return nil, err
	}
	return slices.DeleteFunc(all, func(p projects.Project) bool { return p.Kind == projects.KindWeb }), nil
}
