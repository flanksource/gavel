package ui

import (
	"context"
	todoruntime "github.com/flanksource/gavel/todos/runtime"
)

func (s *Server) catalogProjectCounts(ctx context.Context, hidden workspaceSet) ([]projectInfo, error) {
	projects, err := s.catalogProjects()
	if err != nil {
		return nil, err
	}
	providers := make([]*todoruntime.Provider, 0, len(projects))
	byDir := map[string]*todoruntime.Provider{}
	for i, project := range projects {
		projects[i].Hidden = hidden.has(project.Dir)
		if projects[i].Hidden || project.Error != "" {
			continue
		}
		provider, err := s.workspaceReadProvider(ctx, project.Dir)
		if err != nil {
			return nil, err
		}
		providers = append(providers, provider)
		byDir[project.Dir] = provider
	}
	counts, err := todoruntime.CountResolvedWorkspaces(ctx, providers)
	if err != nil {
		return nil, err
	}
	for index, project := range projects {
		if provider := byDir[project.Dir]; provider != nil {
			value := todoCounts{}
			for status, count := range counts[provider.Workspace().ID] {
				addTodoStatus(&value, status, count)
			}
			projects[index].TodoCounts = &value
		}
	}
	return projects, nil
}
