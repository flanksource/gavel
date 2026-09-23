package ui

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/entity"
	rpchttp "github.com/flanksource/clicky/rpc/http"
	"github.com/flanksource/gavel/todos/query"
)

// The project entity declares the registered projects as a read-only clicky
// entity beside todos. Declaring it once is what gives the assistant its
// project tools and the chat's context picker its project listing, both served
// from the same list /api/projects answers with. Create, update and delete stay
// on the hand-written /api/projects routes.

var projectEntityOnce sync.Once

// GetID and GetName make projectInfo an entity item. A project is addressed by
// its short name, which is also what a TODO list's --project takes and what a
// TODO row names its project with.
func (p projectInfo) GetID() string   { return p.Short }
func (p projectInfo) GetName() string { return p.Name }

// registerProjectEntity declares the entity exactly once; registering twice
// would duplicate every generated command and route.
func registerProjectEntity() {
	projectEntityOnce.Do(func() {
		readOnly, destructive := true, false
		clicky.NewEntity[projectInfo, struct{}, projectInfo]("project").
			Aliases("projects").
			ListWithContext(func(ctx context.Context, _ struct{}) ([]projectInfo, error) {
				return listProjectEntities(ctx)
			}).
			GetWithContext(getProjectInfo).
			ToolHints(entity.MCPToolHints{
				Icon: "folder", Group: "Projects",
				ReadOnlyHint: &readOnly, DestructiveHint: &destructive,
			}).
			Register()
	})
}

// loadProjectsTracked reads the project catalog, timed as file work.
func loadProjectsTracked(ctx context.Context) ([]Project, error) {
	stopFile := rpchttp.Track(ctx, "file")
	defer stopFile()
	return LoadProjects()
}

// listProjects resolves every registered project's wire shape. A project whose
// TODO counts cannot be loaded carries the failure on its own row.
func listProjects(ctx context.Context) ([]projectInfo, error) {
	ps, err := loadProjectsTracked(ctx)
	if err != nil {
		return nil, err
	}
	return listProjectInfos(ctx, ps), nil
}

// listProjectEntities is listProjects for the surfaces that address a project
// by its short name, which refuse a catalog where two projects share one: a
// filter by that name could not say which project it meant. The dashboard's own
// /api/projects keeps listing such a catalog, so the projects bar stays usable
// while the clash is fixed.
func listProjectEntities(ctx context.Context) ([]projectInfo, error) {
	ps, err := loadProjectsTracked(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := query.IndexByShortName(TodoWorkspaces(ps)); err != nil {
		return nil, err
	}
	return listProjectInfos(ctx, ps), nil
}

// getProjectInfo resolves one registered project by its short name. An unknown
// name is a 404 rather than a server fault.
func getProjectInfo(ctx context.Context, short string) (projectInfo, error) {
	ps, err := loadProjectsTracked(ctx)
	if err != nil {
		return projectInfo{}, err
	}
	byShort, err := query.IndexByShortName(TodoWorkspaces(ps))
	if err != nil {
		return projectInfo{}, err
	}
	ws, ok := byShort[short]
	if !ok {
		return projectInfo{}, entity.NewStatusError(http.StatusNotFound, "not_found",
			fmt.Sprintf("no registered project has the short name %q", short))
	}
	p, _ := projectByName(ps, ws.Name)
	return listProjectInfos(ctx, []Project{p})[0], nil
}
