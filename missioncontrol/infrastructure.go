// Reads gavel's infrastructure from Mission Control's catalog: the config items, health
// checks and config changes its search selects, for the platform dashboard.
package missioncontrol

import (
	"context"

	"github.com/flanksource/incident-commander/clientapi"
)

// infrastructureLimit bounds how many resources of each kind one read returns.
const infrastructureLimit = 200

// Infrastructure is what Mission Control knows about gavel's infrastructure.
type Infrastructure struct {
	MissionControl string                       `json:"missionControl"`
	Configs        []clientapi.SelectedResource `json:"configs"`
	Checks         []clientapi.SelectedResource `json:"checks"`
	Changes        []clientapi.SelectedResource `json:"changes"`
}

// Infrastructure reads the config items, checks and config changes Search selects.
func (s *Source) Infrastructure(ctx context.Context) (Infrastructure, error) {
	selector := []clientapi.ResourceSelector{{Search: s.Search}}
	found, err := s.Client.SearchCatalog(ctx, clientapi.SearchResourcesRequest{
		Limit: infrastructureLimit, Configs: selector, Checks: selector, ConfigChanges: selector,
	})
	if err != nil {
		return Infrastructure{}, s.explain("gavel's infrastructure", err)
	}
	return Infrastructure{
		MissionControl: s.URL,
		Configs:        nonNil(found.Configs),
		Checks:         nonNil(found.Checks),
		Changes:        nonNil(found.ConfigChanges),
	}, nil
}

func nonNil(resources []clientapi.SelectedResource) []clientapi.SelectedResource {
	if resources == nil {
		return []clientapi.SelectedResource{}
	}
	return resources
}
